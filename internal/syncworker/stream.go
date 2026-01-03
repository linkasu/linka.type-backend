package syncworker

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/linkasu/linka.type-backend/internal/defaults"
	"github.com/linkasu/linka.type-backend/internal/id"
	"github.com/linkasu/linka.type-backend/internal/models"
	"golang.org/x/oauth2"
)

type streamMessage struct {
	Path string          `json:"path"`
	Data json.RawMessage `json:"data"`
}

func (w *Worker) runStream(ctx context.Context) {
	reconnect := w.streamReconnect
	if reconnect <= 0 {
		reconnect = 5 * time.Second
	}

	for {
		err := w.streamOnce(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("rtdb stream error", "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(reconnect):
		}
	}
}

func (w *Worker) streamOnce(ctx context.Context) error {
	if w.streamBaseURL == "" || w.tokenSource == nil {
		return nil
	}

	url := buildStreamURL(w.streamBaseURL, w.streamPath)
	client := oauth2.NewClient(ctx, w.tokenSource)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/event-stream")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("stream status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var eventName string
	var data strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			payload := strings.TrimSpace(data.String())
			if eventName != "" && payload != "" {
				if err := w.handleStreamEvent(ctx, eventName, payload); err != nil {
					return err
				}
			}
			eventName = ""
			data.Reset()
			continue
		}
		if strings.HasPrefix(line, "event:") {
			eventName = strings.TrimSpace(line[len("event:"):])
			continue
		}
		if strings.HasPrefix(line, "data:") {
			if data.Len() > 0 {
				data.WriteByte('\n')
			}
			data.WriteString(strings.TrimSpace(line[len("data:"):]))
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return fmt.Errorf("stream ended")
}

func (w *Worker) handleStreamEvent(ctx context.Context, event, payload string) error {
	switch event {
	case "put", "patch":
		// continue
	case "keep-alive":
		return nil
	default:
		return nil
	}

	var msg streamMessage
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		return err
	}
	return w.applyStreamEvent(ctx, msg.Path, msg.Data)
}

func (w *Worker) applyStreamEvent(ctx context.Context, path string, data json.RawMessage) error {
	root := strings.Trim(w.streamPath, "/")
	relative := strings.Trim(path, "/")
	if root == "" {
		if !strings.HasPrefix(relative, "users") {
			return nil
		}
		relative = strings.TrimPrefix(relative, "users")
		relative = strings.Trim(relative, "/")
		return w.applyUserPath(ctx, relative, data)
	}

	if root == "users" {
		return w.applyUserPath(ctx, relative, data)
	}

	// Unsupported stream root.
	return nil
}

func (w *Worker) applyUserPath(ctx context.Context, path string, data json.RawMessage) error {
	if path == "" {
		return nil
	}

	parts := strings.Split(path, "/")
	userID := parts[0]
	if userID == "" {
		return nil
	}

	if len(parts) == 1 {
		if isNullData(data) {
			return w.deleteUser(ctx, userID)
		}
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		return w.applyUserSnapshot(ctx, userID, raw)
	}

	switch parts[1] {
	case "Category":
		return w.applyCategoryPath(ctx, userID, parts[2:], data)
	case "quickes":
		return w.applyQuickes(ctx, userID, data)
	case "inited":
		return w.applyInited(ctx, userID, data)
	default:
		return nil
	}
}

func (w *Worker) applyCategoryPath(ctx context.Context, userID string, parts []string, data json.RawMessage) error {
	if len(parts) == 0 {
		return nil
	}
	categoryID := parts[0]
	if categoryID == "" {
		return nil
	}

	if len(parts) == 1 {
		if isNullData(data) {
			return w.deleteCategory(ctx, userID, categoryID)
		}
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		return w.upsertCategoryFromMap(ctx, userID, categoryID, raw)
	}

	if len(parts) >= 2 && parts[1] == "statements" {
		if len(parts) == 2 {
			if isNullData(data) {
				return nil
			}
			var raw map[string]any
			if err := json.Unmarshal(data, &raw); err != nil {
				return err
			}
			for stmtKey, rawStmt := range raw {
				stmtMap, ok := rawStmt.(map[string]any)
				if !ok {
					continue
				}
				if err := w.upsertStatementFromMap(ctx, userID, categoryID, stmtKey, stmtMap); err != nil {
					return err
				}
			}
			return nil
		}

		statementID := parts[2]
		if statementID == "" {
			return nil
		}
		if isNullData(data) {
			return w.deleteStatement(ctx, userID, categoryID, statementID)
		}
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		return w.upsertStatementFromMap(ctx, userID, categoryID, statementID, raw)
	}

	return nil
}

func (w *Worker) applyUserSnapshot(ctx context.Context, userID string, raw map[string]any) error {
	if raw == nil {
		return nil
	}

	if cats, ok := raw["Category"].(map[string]any); ok {
		for catKey, rawCat := range cats {
			catMap, ok := rawCat.(map[string]any)
			if !ok {
				continue
			}
			if err := w.upsertCategoryFromMap(ctx, userID, catKey, catMap); err != nil {
				return err
			}
		}
	}

	if rawQuickes, ok := raw["quickes"]; ok {
		if err := w.applyQuickes(ctx, userID, mustJSON(rawQuickes)); err != nil {
			return err
		}
	}

	if rawInited, ok := raw["inited"]; ok {
		if err := w.applyInited(ctx, userID, mustJSON(rawInited)); err != nil {
			return err
		}
	}

	return nil
}

func (w *Worker) upsertCategoryFromMap(ctx context.Context, userID, categoryID string, raw map[string]any) error {
	now := time.Now().UnixMilli()
	categoryID = str(raw["id"], categoryID)
	category := models.Category{
		ID:        categoryID,
		Label:     str(raw["label"], ""),
		Created:   int64From(raw["created"], now),
		Default:   boolPtrFrom(raw["default"]),
		UpdatedAt: now,
	}

	if _, err := w.store.UpsertCategory(ctx, userID, category); err != nil {
		return err
	}
	if err := w.appendChange(ctx, userID, "category", categoryID, "upsert", category, now); err != nil {
		return err
	}

	if rawStatements, ok := raw["statements"].(map[string]any); ok {
		for stmtKey, rawStmt := range rawStatements {
			stmtMap, ok := rawStmt.(map[string]any)
			if !ok {
				continue
			}
			if err := w.upsertStatementFromMap(ctx, userID, categoryID, stmtKey, stmtMap); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *Worker) upsertStatementFromMap(ctx context.Context, userID, categoryID, statementID string, raw map[string]any) error {
	now := time.Now().UnixMilli()
	statementID = str(raw["id"], statementID)
	statement := models.Statement{
		ID:         statementID,
		CategoryID: str(raw["categoryId"], categoryID),
		Text:       str(raw["text"], ""),
		Created:    int64From(raw["created"], now),
		UpdatedAt:  now,
	}

	if _, err := w.store.UpsertStatement(ctx, userID, statement); err != nil {
		return err
	}
	return w.appendChange(ctx, userID, "statement", statementID, "upsert", statement, now)
}

func (w *Worker) deleteCategory(ctx context.Context, userID, categoryID string) error {
	updatedAt := time.Now().UnixMilli()
	statements, err := w.store.ListStatements(ctx, userID, categoryID)
	if err != nil {
		return err
	}
	if err := w.store.DeleteCategory(ctx, userID, categoryID, updatedAt); err != nil {
		return err
	}
	if err := w.appendChange(ctx, userID, "category", categoryID, "delete", map[string]string{"id": categoryID}, updatedAt); err != nil {
		return err
	}

	for _, stmt := range statements {
		if err := w.store.DeleteStatement(ctx, userID, categoryID, stmt.ID, updatedAt); err != nil {
			return err
		}
		if err := w.appendChange(ctx, userID, "statement", stmt.ID, "delete", map[string]string{"id": stmt.ID}, updatedAt); err != nil {
			return err
		}
	}

	return nil
}

func (w *Worker) deleteStatement(ctx context.Context, userID, categoryID, statementID string) error {
	updatedAt := time.Now().UnixMilli()
	if err := w.store.DeleteStatement(ctx, userID, categoryID, statementID, updatedAt); err != nil {
		return err
	}
	return w.appendChange(ctx, userID, "statement", statementID, "delete", map[string]string{"id": statementID}, updatedAt)
}

func (w *Worker) deleteUser(ctx context.Context, userID string) error {
	updatedAt := time.Now().UnixMilli()
	if err := w.store.DeleteUser(ctx, userID, updatedAt); err != nil {
		return err
	}
	return w.appendChange(ctx, userID, "user", userID, "delete", map[string]string{"id": userID}, updatedAt)
}

func (w *Worker) applyQuickes(ctx context.Context, userID string, data json.RawMessage) error {
	if isNullData(data) {
		return nil
	}
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	quickes := parseQuickes(raw)
	if len(quickes) == 0 {
		quickes = defaults.DefaultQuickes
	}
	updatedAt := time.Now().UnixMilli()
	updated, err := w.store.SetQuickes(ctx, userID, quickes, updatedAt)
	if err != nil {
		return err
	}
	return w.appendChange(ctx, userID, "quickes", userID, "upsert", updated, updatedAt)
}

func (w *Worker) applyInited(ctx context.Context, userID string, data json.RawMessage) error {
	if isNullData(data) {
		return nil
	}
	var inited bool
	if err := json.Unmarshal(data, &inited); err != nil {
		return err
	}

	state, err := w.store.GetUserState(ctx, userID)
	if err != nil {
		return err
	}
	state.Inited = inited
	updatedAt := time.Now().UnixMilli()
	updated, err := w.store.SetUserState(ctx, userID, state, updatedAt)
	if err != nil {
		return err
	}
	return w.appendChange(ctx, userID, "user_state", userID, "upsert", updated, updatedAt)
}

func (w *Worker) appendChange(ctx context.Context, userID, entityType, entityID, op string, payload any, updatedAt int64) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return w.store.AppendChange(ctx, userID, models.ChangeEvent{
		Cursor:     id.New(),
		EntityType: entityType,
		EntityID:   entityID,
		Op:         op,
		Payload:    data,
		UpdatedAt:  updatedAt,
	})
}

func isNullData(data json.RawMessage) bool {
	trimmed := strings.TrimSpace(string(data))
	return trimmed == "" || trimmed == "null"
}

func parseQuickes(raw any) []string {
	switch val := raw.(type) {
	case nil:
		return nil
	case []string:
		return val
	case []any:
		out := make([]string, 0, len(val))
		for _, item := range val {
			if strVal, ok := item.(string); ok {
				out = append(out, strVal)
			}
		}
		return out
	case map[string]any:
		out := make([]string, 0, len(val))
		for i := 0; i < len(val); i++ {
			if strVal, ok := val[fmt.Sprintf("%d", i)].(string); ok {
				out = append(out, strVal)
			}
		}
		return out
	default:
		return nil
	}
}

func buildStreamURL(baseURL, path string) string {
	base := strings.TrimRight(baseURL, "/")
	path = strings.Trim(path, "/")
	if path != "" {
		base = base + "/" + path
	}
	return base + ".json"
}

func mustJSON(raw any) json.RawMessage {
	data, _ := json.Marshal(raw)
	return data
}
