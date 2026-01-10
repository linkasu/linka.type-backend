package legacy

import (
	"context"
	"fmt"

	"firebase.google.com/go/v4/db"
	"github.com/linkasu/linka.type-backend/internal/models"
	"github.com/linkasu/linka.type-backend/internal/store"
)

// Writer mirrors changes into Firebase RTDB.
type Writer struct {
	db *db.Client
}

// New creates a legacy writer.
func New(client *db.Client) (*Writer, error) {
	if client == nil {
		return nil, fmt.Errorf("firebase db client is nil")
	}
	return &Writer{db: client}, nil
}

func (w *Writer) UpsertCategory(ctx context.Context, userID string, category models.Category) error {
	ref := w.db.NewRef(fmt.Sprintf("users/%s/Category/%s", userID, category.ID))
	payload := map[string]any{
		"id":      category.ID,
		"label":   category.Label,
		"created": category.Created,
		"aiUse":   category.AIUse,
	}
	if category.Default != nil {
		payload["default"] = *category.Default
	}
	return ref.Set(ctx, payload)
}

func (w *Writer) DeleteCategory(ctx context.Context, userID, categoryID string) error {
	ref := w.db.NewRef(fmt.Sprintf("users/%s/Category/%s", userID, categoryID))
	return ref.Delete(ctx)
}

func (w *Writer) UpsertStatement(ctx context.Context, userID string, statement models.Statement) error {
	ref := w.db.NewRef(fmt.Sprintf("users/%s/Category/%s/statements/%s", userID, statement.CategoryID, statement.ID))
	payload := map[string]any{
		"id":         statement.ID,
		"categoryId": statement.CategoryID,
		"text":       statement.Text,
		"created":    statement.Created,
	}
	return ref.Set(ctx, payload)
}

func (w *Writer) DeleteStatement(ctx context.Context, userID, categoryID, statementID string) error {
	ref := w.db.NewRef(fmt.Sprintf("users/%s/Category/%s/statements/%s", userID, categoryID, statementID))
	return ref.Delete(ctx)
}

func (w *Writer) SetUserState(ctx context.Context, userID string, state models.UserState) error {
	ref := w.db.NewRef(fmt.Sprintf("users/%s", userID))
	updates := map[string]any{
		"inited": state.Inited,
	}
	if state.Quickes != nil {
		updates["quickes"] = state.Quickes
	}
	if state.Preferences != nil {
		updates["preferences"] = state.Preferences
	}
	return ref.Update(ctx, updates)
}

func (w *Writer) SetQuickes(ctx context.Context, userID string, quickes []string) error {
	ref := w.db.NewRef(fmt.Sprintf("users/%s/quickes", userID))
	return ref.Set(ctx, quickes)
}

func (w *Writer) ImportGlobalCategory(ctx context.Context, userID, categoryID string) error {
	globalRef := w.db.NewRef(fmt.Sprintf("global/Category/%s", categoryID))
	var payload map[string]any
	if err := globalRef.Get(ctx, &payload); err != nil {
		return err
	}
	if payload == nil {
		return store.ErrNotFound
	}
	userRef := w.db.NewRef(fmt.Sprintf("users/%s/Category/%s", userID, categoryID))
	return userRef.Set(ctx, payload)
}

func (w *Writer) DeleteUserData(ctx context.Context, userID string) error {
	ref := w.db.NewRef(fmt.Sprintf("users/%s", userID))
	return ref.Delete(ctx)
}
