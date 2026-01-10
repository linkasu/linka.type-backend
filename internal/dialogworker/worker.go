package dialogworker

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/linkasu/linka.type-backend/internal/dialoghelper"
	"github.com/linkasu/linka.type-backend/internal/id"
	"github.com/linkasu/linka.type-backend/internal/models"
	"github.com/linkasu/linka.type-backend/internal/store"
)

const (
	maxHistoryMessages   = 64
	maxBiographyChars    = 1800
	maxStatementsPerCat  = 12
	maxSuggestionBatch   = 10
	maxSuggestionStorage = 200
)

type Worker struct {
	store  store.Store
	client *dialoghelper.Client
	logger *slog.Logger
}

func New(store store.Store, client *dialoghelper.Client, logger *slog.Logger) *Worker {
	if logger == nil {
		logger = slog.Default()
	}
	return &Worker{
		store:  store,
		client: client,
		logger: logger,
	}
}

func (w *Worker) Run(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		interval = 10 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if err := w.process(ctx); err != nil {
			w.logger.Error("dialog worker cycle failed", "error", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (w *Worker) process(ctx context.Context) error {
	if w.client == nil || !w.client.Available() {
		return nil
	}

	jobs, err := w.store.ListDialogSuggestionJobs(ctx, "pending", 50)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		job.Attempts++
		job.Status = "processing"
		if err := w.store.UpdateDialogSuggestionJob(ctx, job); err != nil {
			w.logger.Warn("failed to mark job processing", "job_id", job.ID, "error", err)
			continue
		}

		if err := w.processJob(ctx, job); err != nil {
			errMsg := err.Error()
			job.Status = "failed"
			job.LastError = &errMsg
			_ = w.store.UpdateDialogSuggestionJob(ctx, job)
			w.logger.Warn("failed to process job", "job_id", job.ID, "error", err)
			continue
		}

		job.Status = "done"
		job.LastError = nil
		_ = w.store.UpdateDialogSuggestionJob(ctx, job)
	}

	return nil
}

func (w *Worker) processJob(ctx context.Context, job models.DialogSuggestionJob) error {
	messages, err := w.store.ListDialogMessages(ctx, job.UserID, job.ChatID, maxHistoryMessages, 0)
	if err != nil {
		return err
	}
	if len(messages) == 0 {
		return nil
	}

	history := make([]dialoghelper.DialogMessage, 0, len(messages))
	for _, msg := range messages {
		history = append(history, dialoghelper.DialogMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	bio, err := w.buildBiography(ctx, job.UserID)
	if err != nil {
		return err
	}

	resp, err := w.client.Infer(ctx, dialoghelper.InferPayload{
		DisabledPersonBiography: bio,
		Messages:                history,
		Language:                "ru-RU",
		UserID:                  job.UserID,
		DialogID:                job.ChatID,
		StepID:                  job.ID,
	}, nil)
	if err != nil {
		return err
	}

	candidates := normalizeSuggestions(resp.Response)
	if len(candidates) == 0 {
		return nil
	}

	existing, err := w.collectExistingTexts(ctx, job.UserID)
	if err != nil {
		return err
	}

	available := maxSuggestionStorage
	if count, err := w.store.CountDialogSuggestions(ctx, job.UserID, "pending"); err == nil {
		available = maxSuggestionStorage - int(count)
	}
	if available <= 0 {
		return nil
	}

	added := 0
	for _, text := range candidates {
		if added >= available || added >= maxSuggestionBatch {
			break
		}
		if _, ok := existing[normalizeText(text)]; ok {
			continue
		}
		suggestion := models.DialogSuggestion{
			ID:        id.New(),
			ChatID:    job.ChatID,
			MessageID: job.MessageID,
			Text:      text,
			Status:    "pending",
			Created:   time.Now().UnixMilli(),
		}
		suggestion.UpdatedAt = suggestion.Created
		if _, err := w.store.UpsertDialogSuggestion(ctx, job.UserID, suggestion); err != nil {
			return err
		}
		existing[normalizeText(text)] = struct{}{}
		added++
	}

	return nil
}

func (w *Worker) collectExistingTexts(ctx context.Context, userID string) (map[string]struct{}, error) {
	out := make(map[string]struct{})

	statements, err := w.store.ListAllStatements(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, stmt := range statements {
		if text := normalizeText(stmt.Text); text != "" {
			out[text] = struct{}{}
		}
	}

	for _, status := range []string{"pending", "accepted", "dismissed"} {
		suggestions, err := w.store.ListDialogSuggestions(ctx, userID, status, maxSuggestionStorage)
		if err != nil {
			return nil, err
		}
		for _, suggestion := range suggestions {
			if text := normalizeText(suggestion.Text); text != "" {
				out[text] = struct{}{}
			}
		}
	}

	return out, nil
}

func (w *Worker) buildBiography(ctx context.Context, userID string) (string, error) {
	categories, err := w.store.ListCategories(ctx, userID)
	if err != nil {
		return "", err
	}

	lines := make([]string, 0)
	currentLen := 0
	for _, cat := range categories {
		if !cat.AIUse {
			continue
		}
		statements, err := w.store.ListStatements(ctx, userID, cat.ID)
		if err != nil && !errors.Is(err, store.ErrNotFound) {
			return "", err
		}
		if len(statements) == 0 {
			continue
		}

		parts := make([]string, 0, len(statements))
		for _, stmt := range statements {
			if len(parts) >= maxStatementsPerCat {
				break
			}
			if text := strings.TrimSpace(stmt.Text); text != "" {
				parts = append(parts, text)
			}
		}
		if len(parts) == 0 {
			continue
		}
		line := strings.TrimSpace(cat.Label)
		if line == "" {
			continue
		}
		line = line + ": " + strings.Join(parts, "; ")
		lineLen := utf8.RuneCountInString(line)
		if currentLen+lineLen > maxBiographyChars {
			break
		}
		lines = append(lines, line)
		currentLen += lineLen
	}

	return strings.Join(lines, "\n"), nil
}

func normalizeSuggestions(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if text := strings.TrimSpace(item); text != "" {
			out = append(out, text)
		}
	}
	return out
}

func normalizeText(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
