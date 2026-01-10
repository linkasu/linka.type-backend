package dialoghelper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

type DialogMessage struct {
	Role    string `json:"role"`
	Content string `json:"content,omitempty"`
	Audio   bool   `json:"audio,omitempty"`
}

type InferOptions struct {
	MaxCandidates int     `json:"max_candidates,omitempty"`
	Temperature   float64 `json:"temperature,omitempty"`
}

type InferPayload struct {
	DisabledPersonBiography string          `json:"disabled_person_biography,omitempty"`
	Messages                []DialogMessage `json:"messages"`
	Language                string          `json:"language,omitempty"`
	UserID                  string          `json:"user_id,omitempty"`
	DialogID                string          `json:"dialog_id,omitempty"`
	StepID                  string          `json:"step_id,omitempty"`
	Options                 *InferOptions   `json:"options,omitempty"`
}

type InferResponse struct {
	Transcript *string  `json:"transcript,omitempty"`
	Response   []string `json:"response"`
}

type AudioPayload struct {
	Data        []byte
	Filename    string
	ContentType string
}

func New(baseURL, apiKey string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		client:  &http.Client{Timeout: timeout},
	}
}

func (c *Client) Available() bool {
	return c != nil && c.baseURL != "" && c.apiKey != ""
}

func (c *Client) Infer(ctx context.Context, payload InferPayload, audio *AudioPayload) (InferResponse, error) {
	if !c.Available() {
		return InferResponse{}, fmt.Errorf("dialog helper is not configured")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return InferResponse{}, err
	}
	if err := writer.WriteField("payload", string(payloadBytes)); err != nil {
		return InferResponse{}, err
	}

	if audio != nil && len(audio.Data) > 0 {
		contentType := strings.TrimSpace(audio.ContentType)
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		filename := audio.Filename
		if filename == "" {
			filename = "audio"
		}
		partHeader := make(textproto.MIMEHeader)
		partHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="audio"; filename="%s"`, filename))
		partHeader.Set("Content-Type", contentType)
		part, err := writer.CreatePart(partHeader)
		if err != nil {
			return InferResponse{}, err
		}
		if _, err := part.Write(audio.Data); err != nil {
			return InferResponse{}, err
		}
	}

	if err := writer.Close(); err != nil {
		return InferResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/infer", body)
	if err != nil {
		return InferResponse{}, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-API-KEY", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return InferResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return InferResponse{}, fmt.Errorf("dialog helper error: %s", strings.TrimSpace(string(data)))
	}

	var out InferResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return InferResponse{}, err
	}

	return out, nil
}
