package gpt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	yandexLLMEndpoint   = "https://llm.api.cloud.yandex.net/foundationModels/v1/completion"
	metadataURL         = "http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token"
	metadataHeaderKey   = "Metadata-Flavor"
	metadataHeaderVal   = "Google"
)

// AnalyzePrompt is the system prompt for extracting user facts from dialog
const AnalyzePrompt = `Ты анализируешь диалог между человеком с нарушением речи и его собеседником.
Из диалога извлеки НОВЫЕ факты о человеке с нарушением речи, которые можно добавить в его биографию.

Факты должны быть:
- Конкретными (имена, места, предпочтения, привычки)
- Полезными для будущих диалогов
- О самом человеке, не о собеседнике

Верни строго JSON: {"facts": ["факт1", "факт2", "факт3"]}
Если новых фактов нет, верни: {"facts": []}
Максимум 5 фактов. Используй русский язык.`

type Client struct {
	folderID string
	modelURI string
	client   *http.Client
	mu       sync.Mutex
	token    string
	tokenExp time.Time
}

func NewClient(folderID, modelURI string) *Client {
	if modelURI == "" {
		modelURI = fmt.Sprintf("gpt://%s/yandexgpt-lite/latest", folderID)
	}
	return &Client{
		folderID: folderID,
		modelURI: modelURI,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Available() bool {
	return c.folderID != ""
}

type AnalyzeResult struct {
	Facts []string `json:"facts"`
}

func (c *Client) Analyze(ctx context.Context, biography string, messages []DialogMessage) (AnalyzeResult, error) {
	if !c.Available() {
		return AnalyzeResult{}, errors.New("gpt client not configured")
	}

	userPrompt := buildUserPrompt(biography, messages)
	if userPrompt == "" {
		return AnalyzeResult{}, errors.New("empty dialog")
	}

	token, err := c.getToken(ctx)
	if err != nil {
		return AnalyzeResult{}, fmt.Errorf("get token: %w", err)
	}

	body := map[string]interface{}{
		"modelUri": c.modelURI,
		"completionOptions": map[string]interface{}{
			"stream":      false,
			"temperature": 0.3,
			"maxTokens":   "500",
		},
		"messages": []map[string]string{
			{"role": "system", "text": AnalyzePrompt},
			{"role": "user", "text": userPrompt},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return AnalyzeResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, yandexLLMEndpoint, bytes.NewReader(payload))
	if err != nil {
		return AnalyzeResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("x-folder-id", c.folderID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return AnalyzeResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return AnalyzeResult{}, fmt.Errorf("llm request failed: %d", resp.StatusCode)
	}

	var decoded struct {
		Result struct {
			Alternatives []struct {
				Message struct {
					Text string `json:"text"`
				} `json:"message"`
			} `json:"alternatives"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return AnalyzeResult{}, err
	}

	message := ""
	if len(decoded.Result.Alternatives) > 0 {
		message = decoded.Result.Alternatives[0].Message.Text
	}

	return parseAnalyzeResult(message), nil
}

func (c *Client) getToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UTC()
	if c.token != "" && now.Add(time.Minute).Before(c.tokenExp) {
		return c.token, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metadataURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set(metadataHeaderKey, metadataHeaderVal)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New("metadata token request failed")
	}

	var payload struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.AccessToken == "" {
		return "", errors.New("empty token")
	}

	c.token = payload.AccessToken
	c.tokenExp = now.Add(time.Duration(payload.ExpiresIn) * time.Second)

	return c.token, nil
}

type DialogMessage struct {
	Role    string
	Content string
}

func buildUserPrompt(biography string, messages []DialogMessage) string {
	var sb strings.Builder

	if biography != "" {
		sb.WriteString("Текущая биография:\n")
		sb.WriteString(biography)
		sb.WriteString("\n\n")
	}

	sb.WriteString("Диалог:\n")
	for _, msg := range messages {
		role := msg.Role
		if role == "disabled_person" {
			role = "Пользователь"
		} else if role == "speaker" {
			role = "Собеседник"
		}
		sb.WriteString(role)
		sb.WriteString(": ")
		sb.WriteString(msg.Content)
		sb.WriteString("\n")
	}

	return sb.String()
}

func parseAnalyzeResult(message string) AnalyzeResult {
	message = strings.TrimSpace(message)

	// Try to find JSON in message
	start := strings.Index(message, "{")
	end := strings.LastIndex(message, "}")
	if start != -1 && end > start {
		message = message[start : end+1]
	}

	var result AnalyzeResult
	if err := json.Unmarshal([]byte(message), &result); err != nil {
		return AnalyzeResult{}
	}

	// Filter empty facts
	filtered := make([]string, 0, len(result.Facts))
	for _, fact := range result.Facts {
		fact = strings.TrimSpace(fact)
		if fact != "" {
			filtered = append(filtered, fact)
		}
	}
	result.Facts = filtered

	return result
}
