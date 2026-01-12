package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/linkasu/linka.type-backend/internal/dialoghelper"
	"github.com/linkasu/linka.type-backend/internal/id"
	"github.com/linkasu/linka.type-backend/internal/models"
	"github.com/linkasu/linka.type-backend/internal/store"
)

const (
	maxDialogChats        = 20
	maxDialogMessages     = 200
	maxDialogSuggestions  = 200
	maxDialogHistoryItems = 64
	maxBiographyChars     = 1800
	maxStatementsPerCat   = 12
	defaultMonthlyLimit   = 100
)

var ErrUsageLimitExceeded = errors.New("monthly inference limit exceeded")

type DialogChatInput struct {
	Title string
}

type DialogMessageInput struct {
	ID                 string
	Role               string
	Content            string
	Source             string
	Created            int64
	IncludeSuggestions bool
}

type DialogMessageResult struct {
	Message     models.DialogMessage `json:"message"`
	Suggestions []string             `json:"suggestions,omitempty"`
	Transcript  *string              `json:"transcript,omitempty"`
}

type DialogSuggestionApplyItem struct {
	ID           string  `json:"id"`
	CategoryID   *string `json:"categoryId,omitempty"`
	CategoryName *string `json:"categoryLabel,omitempty"`
}

type DialogSuggestionApplyResult struct {
	Created   []map[string]string `json:"created"`
	Applied   []string            `json:"applied"`
}

func (s *Service) ListDialogChats(ctx context.Context, userID string) ([]models.DialogChat, error) {
	return s.Store.ListDialogChats(ctx, userID)
}

func (s *Service) CreateDialogChat(ctx context.Context, userID string, input DialogChatInput) (models.DialogChat, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "Новый диалог"
	}
	now := time.Now().UnixMilli()
	chat := models.DialogChat{
		ID:        id.NewShort(),
		Title:     title,
		Created:   now,
		UpdatedAt: now,
	}

	chat, err := s.Store.UpsertDialogChat(ctx, userID, chat)
	if err != nil {
		return models.DialogChat{}, err
	}

	if err := s.trimDialogChats(ctx, userID); err != nil {
		return models.DialogChat{}, err
	}

	return chat, nil
}

func (s *Service) DeleteDialogChat(ctx context.Context, userID, chatID string) error {
	updatedAt := time.Now().UnixMilli()
	if err := s.Store.DeleteDialogChat(ctx, userID, chatID, updatedAt); err != nil {
		return err
	}
	if err := s.Store.DeleteDialogMessagesByChat(ctx, userID, chatID, updatedAt); err != nil {
		return err
	}
	_ = s.dismissDialogSuggestionsForChat(ctx, userID, chatID)
	return nil
}

func (s *Service) ListDialogMessages(ctx context.Context, userID, chatID string, limit int, before int64) ([]models.DialogMessage, error) {
	return s.Store.ListDialogMessages(ctx, userID, chatID, limit, before)
}

func (s *Service) CreateDialogMessage(ctx context.Context, userID, chatID string, input DialogMessageInput, audio *dialoghelper.AudioPayload) (DialogMessageResult, error) {
	_, err := s.Store.GetDialogChat(ctx, userID, chatID)
	if err != nil {
		return DialogMessageResult{}, err
	}

	role := strings.ToLower(strings.TrimSpace(input.Role))
	if role != "speaker" && role != "disabled_person" {
		return DialogMessageResult{}, errors.New("unsupported message role")
	}

	content := strings.TrimSpace(input.Content)
	if audio == nil && content == "" {
		return DialogMessageResult{}, errors.New("message content is required")
	}

	now := time.Now().UnixMilli()
	if input.Created == 0 {
		input.Created = now
	}

	messageID := input.ID
	if messageID == "" {
		messageID = id.New()
	}

	message := models.DialogMessage{
		ID:        messageID,
		ChatID:    chatID,
		Role:      role,
		Content:   content,
		Source:    input.Source,
		Created:   input.Created,
		UpdatedAt: now,
	}
	if message.Source == "" && audio == nil {
		message.Source = "typed"
	}

	includeSuggestions := input.IncludeSuggestions || audio != nil
	var suggestions []string
	var transcript *string

	if role == "speaker" && includeSuggestions {
		dialogEnabled := s.DialogHelper != nil && s.DialogHelper.Available()
		if audio != nil && !dialogEnabled {
			return DialogMessageResult{}, errors.New("dialog helper not configured for audio")
		}
		if dialogEnabled {
			// Check monthly usage limit
			month := time.Now().Format("2006-01")
			usage, err := s.Store.GetUsageLimit(ctx, userID, month)
			if err != nil {
				return DialogMessageResult{}, err
			}
			limit := usage.Limit
			if limit == 0 {
				limit = defaultMonthlyLimit
			}
			if usage.InferenceCount >= limit {
				return DialogMessageResult{}, ErrUsageLimitExceeded
			}

			history, err := s.buildDialogHistory(ctx, userID, chatID, role, content, audio != nil)
			if err != nil {
				return DialogMessageResult{}, err
			}
			bio, err := s.buildDialogBiography(ctx, userID)
			if err != nil {
				return DialogMessageResult{}, err
			}
			resp, err := s.DialogHelper.Infer(ctx, dialoghelper.InferPayload{
				DisabledPersonBiography: bio,
				Messages:                history,
				Language:                "ru-RU",
				UserID:                  userID,
				DialogID:                chatID,
				StepID:                  messageID,
			}, audio)
			if err != nil {
				return DialogMessageResult{}, err
			}
			suggestions = resp.Response
			if resp.Transcript != nil {
				transcript = resp.Transcript
			}

			// Increment usage counter
			_, _ = s.Store.IncrementUsage(ctx, userID, month, defaultMonthlyLimit)
		}
	}

	if audio != nil {
		if transcript == nil || strings.TrimSpace(*transcript) == "" {
			return DialogMessageResult{}, errors.New("empty transcript")
		}
		message.Content = strings.TrimSpace(*transcript)
		message.Source = "audio"
	}

	if message.Content == "" {
		return DialogMessageResult{}, errors.New("message content is required")
	}

	stored, err := s.Store.UpsertDialogMessage(ctx, userID, message)
	if err != nil {
		return DialogMessageResult{}, err
	}

	if err := s.updateDialogChatMeta(ctx, userID, chatID, stored.Created); err != nil {
		return DialogMessageResult{}, err
	}

	if err := s.trimDialogMessages(ctx, userID, chatID); err != nil {
		return DialogMessageResult{}, err
	}

	// Save quick suggestions for persistence across page reloads
	if len(suggestions) > 0 {
		_ = s.saveInlineSuggestions(ctx, userID, chatID, stored.ID, suggestions)
	}

	// TODO: Bio analysis disabled - needs prompt tuning
	// Queue background job for bio analysis (extracts facts from dialog)
	// if role == "speaker" {
	// 	_ = s.enqueueDialogSuggestionJob(ctx, userID, chatID, stored.ID)
	// }

	return DialogMessageResult{
		Message:     stored,
		Suggestions: suggestions,
		Transcript:  transcript,
	}, nil
}

func (s *Service) ListDialogSuggestions(ctx context.Context, userID string, status string, limit int) ([]models.DialogSuggestion, error) {
	return s.Store.ListDialogSuggestions(ctx, userID, status, limit)
}

func (s *Service) ApplyDialogSuggestions(ctx context.Context, userID string, items []DialogSuggestionApplyItem) (DialogSuggestionApplyResult, error) {
	if len(items) == 0 {
		return DialogSuggestionApplyResult{}, errors.New("items are required")
	}

	pending, err := s.Store.ListDialogSuggestions(ctx, userID, "pending", maxDialogSuggestions)
	if err != nil {
		return DialogSuggestionApplyResult{}, err
	}
	pendingByID := make(map[string]models.DialogSuggestion, len(pending))
	for _, suggestion := range pending {
		pendingByID[suggestion.ID] = suggestion
	}

	categories, err := s.ListCategories(ctx, userID)
	if err != nil {
		return DialogSuggestionApplyResult{}, err
	}
	categoryByLabel := make(map[string]models.Category, len(categories))
	for _, cat := range categories {
		categoryByLabel[strings.ToLower(strings.TrimSpace(cat.Label))] = cat
	}

	result := DialogSuggestionApplyResult{}
	for _, item := range items {
		suggestion, ok := pendingByID[item.ID]
		if !ok {
			continue
		}

		var categoryID string
		if item.CategoryID != nil && strings.TrimSpace(*item.CategoryID) != "" {
			categoryID = strings.TrimSpace(*item.CategoryID)
		} else if item.CategoryName != nil && strings.TrimSpace(*item.CategoryName) != "" {
			label := strings.TrimSpace(*item.CategoryName)
			if existing, ok := categoryByLabel[strings.ToLower(label)]; ok {
				categoryID = existing.ID
			} else {
				category, err := s.CreateCategory(ctx, userID, CategoryInput{
					Label: label,
				})
				if err != nil {
					return DialogSuggestionApplyResult{}, err
				}
				categoryByLabel[strings.ToLower(label)] = category
				categoryID = category.ID
			}
		} else {
			return DialogSuggestionApplyResult{}, errors.New("categoryId or categoryLabel is required")
		}

		statement, err := s.CreateStatement(ctx, userID, StatementInput{
			CategoryID: categoryID,
			Text:       suggestion.Text,
			Created:    time.Now().UnixMilli(),
		})
		if err != nil {
			return DialogSuggestionApplyResult{}, err
		}

		suggestion.Status = "accepted"
		suggestion.CategoryID = &categoryID
		suggestion.UpdatedAt = time.Now().UnixMilli()
		if _, err := s.Store.UpsertDialogSuggestion(ctx, userID, suggestion); err != nil {
			return DialogSuggestionApplyResult{}, err
		}

		result.Created = append(result.Created, map[string]string{
			"categoryId":  categoryID,
			"statementId": statement.ID,
		})
		result.Applied = append(result.Applied, suggestion.ID)
	}

	return result, nil
}

func (s *Service) DismissDialogSuggestions(ctx context.Context, userID string, ids []string) error {
	if len(ids) == 0 {
		return errors.New("ids are required")
	}

	pending, err := s.Store.ListDialogSuggestions(ctx, userID, "pending", maxDialogSuggestions)
	if err != nil {
		return err
	}
	pendingByID := make(map[string]models.DialogSuggestion, len(pending))
	for _, suggestion := range pending {
		pendingByID[suggestion.ID] = suggestion
	}

	for _, id := range ids {
		suggestion, ok := pendingByID[id]
		if !ok {
			continue
		}
		suggestion.Status = "dismissed"
		suggestion.UpdatedAt = time.Now().UnixMilli()
		if _, err := s.Store.UpsertDialogSuggestion(ctx, userID, suggestion); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) trimDialogChats(ctx context.Context, userID string) error {
	chats, err := s.Store.ListDialogChats(ctx, userID)
	if err != nil {
		return err
	}
	if len(chats) <= maxDialogChats {
		return nil
	}

	sort.Slice(chats, func(i, j int) bool {
		return chats[i].UpdatedAt < chats[j].UpdatedAt
	})

	over := len(chats) - maxDialogChats
	for i := 0; i < over; i++ {
		if err := s.DeleteDialogChat(ctx, userID, chats[i].ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) updateDialogChatMeta(ctx context.Context, userID, chatID string, lastMessageAt int64) error {
	chat, err := s.Store.GetDialogChat(ctx, userID, chatID)
	if err != nil {
		return err
	}
	count, err := s.Store.CountDialogMessages(ctx, userID, chatID)
	if err != nil {
		return err
	}
	chat.LastMessageAt = lastMessageAt
	chat.MessageCount = count
	chat.UpdatedAt = time.Now().UnixMilli()
	_, err = s.Store.UpsertDialogChat(ctx, userID, chat)
	return err
}

func (s *Service) trimDialogMessages(ctx context.Context, userID, chatID string) error {
	count, err := s.Store.CountDialogMessages(ctx, userID, chatID)
	if err != nil {
		return err
	}
	if count <= maxDialogMessages {
		return nil
	}

	toRemove := int(count - maxDialogMessages)
	oldest, err := s.Store.ListOldestDialogMessages(ctx, userID, chatID, toRemove)
	if err != nil {
		return err
	}
	updatedAt := time.Now().UnixMilli()
	for _, msg := range oldest {
		if err := s.Store.DeleteDialogMessage(ctx, userID, chatID, msg.ID, updatedAt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) saveInlineSuggestions(ctx context.Context, userID, chatID, messageID string, suggestions []string) error {
	now := time.Now().UnixMilli()

	// Get existing texts to avoid duplicates
	existingTexts := make(map[string]struct{})

	// Check statements
	statements, _ := s.Store.ListAllStatements(ctx, userID)
	for _, stmt := range statements {
		existingTexts[strings.ToLower(strings.TrimSpace(stmt.Text))] = struct{}{}
	}

	// Check existing suggestions
	for _, status := range []string{"pending", "accepted", "dismissed"} {
		existing, _ := s.Store.ListDialogSuggestions(ctx, userID, status, maxDialogSuggestions)
		for _, sug := range existing {
			existingTexts[strings.ToLower(strings.TrimSpace(sug.Text))] = struct{}{}
		}
	}

	// Save new unique suggestions
	for _, text := range suggestions {
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		normalized := strings.ToLower(text)
		if _, exists := existingTexts[normalized]; exists {
			continue
		}
		existingTexts[normalized] = struct{}{}

		suggestion := models.DialogSuggestion{
			ID:        id.New(),
			ChatID:    chatID,
			MessageID: messageID,
			Text:      text,
			Status:    "pending",
			Created:   now,
			UpdatedAt: now,
		}
		if _, err := s.Store.UpsertDialogSuggestion(ctx, userID, suggestion); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) enqueueDialogSuggestionJob(ctx context.Context, userID, chatID, messageID string) error {
	job := models.DialogSuggestionJob{
		ID:        id.New(),
		UserID:    userID,
		ChatID:    chatID,
		MessageID: messageID,
		Status:    "pending",
		Attempts:  0,
		Created:   time.Now().UnixMilli(),
	}
	job.UpdatedAt = job.Created
	return s.Store.UpsertDialogSuggestionJob(ctx, job)
}

func (s *Service) buildDialogHistory(ctx context.Context, userID, chatID, role, content string, hasAudio bool) ([]dialoghelper.DialogMessage, error) {
	history, err := s.Store.ListDialogMessages(ctx, userID, chatID, maxDialogHistoryItems-1, 0)
	if err != nil {
		return nil, err
	}
	out := make([]dialoghelper.DialogMessage, 0, len(history)+1)
	for _, msg := range history {
		out = append(out, dialoghelper.DialogMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	if hasAudio {
		out = append(out, dialoghelper.DialogMessage{
			Role:  role,
			Audio: true,
		})
		return out, nil
	}

	out = append(out, dialoghelper.DialogMessage{
		Role:    role,
		Content: content,
	})
	return out, nil
}

func (s *Service) buildDialogBiography(ctx context.Context, userID string) (string, error) {
	categories, err := s.ListCategories(ctx, userID)
	if err != nil {
		return "", err
	}

	lines := make([]string, 0)
	currentLen := 0

	for _, cat := range categories {
		if !cat.AIUse {
			continue
		}
		statements, err := s.ListStatements(ctx, userID, cat.ID)
		if err != nil && !errors.Is(err, store.ErrNotFound) {
			return "", err
		}
		if len(statements) == 0 {
			continue
		}
		texts := make([]string, 0, len(statements))
		for _, stmt := range statements {
			if len(texts) >= maxStatementsPerCat {
				break
			}
			text := strings.TrimSpace(stmt.Text)
			if text != "" {
				texts = append(texts, text)
			}
		}
		if len(texts) == 0 {
			continue
		}
		line := fmt.Sprintf("%s: %s", strings.TrimSpace(cat.Label), strings.Join(texts, "; "))
		if line == "" {
			continue
		}
		lineLen := utf8.RuneCountInString(line)
		if currentLen+lineLen > maxBiographyChars {
			break
		}
		lines = append(lines, line)
		currentLen += lineLen
	}

	return strings.Join(lines, "\n"), nil
}

func (s *Service) dismissDialogSuggestionsForChat(ctx context.Context, userID, chatID string) error {
	pending, err := s.Store.ListDialogSuggestions(ctx, userID, "pending", maxDialogSuggestions)
	if err != nil {
		return err
	}
	for _, suggestion := range pending {
		if suggestion.ChatID != chatID {
			continue
		}
		suggestion.Status = "dismissed"
		suggestion.UpdatedAt = time.Now().UnixMilli()
		if _, err := s.Store.UpsertDialogSuggestion(ctx, userID, suggestion); err != nil {
			return err
		}
	}
	return nil
}
