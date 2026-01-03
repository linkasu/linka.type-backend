package syncworker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"firebase.google.com/go/v4/db"
	"github.com/linkasu/linka.type-backend/internal/defaults"
	"github.com/linkasu/linka.type-backend/internal/models"
	"github.com/linkasu/linka.type-backend/internal/store"
	"github.com/linkasu/linka.type-backend/internal/ydb"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
	"golang.org/x/oauth2"
)

// Worker syncs Firebase RTDB data into YDB.
type Worker struct {
	ydbClient       *ydb.Client
	store           store.Store
	firebase        *db.Client
	legacy          store.LegacyReader
	streamBaseURL   string
	streamPath      string
	streamReconnect time.Duration
	tokenSource     oauth2.TokenSource
}

// New creates a sync worker.
func New(ydbClient *ydb.Client, store store.Store, firebase *db.Client, legacyReader store.LegacyReader) *Worker {
	return &Worker{ydbClient: ydbClient, store: store, firebase: firebase, legacy: legacyReader}
}

// EnableStream configures RTDB streaming for incremental updates.
func (w *Worker) EnableStream(baseURL string, tokenSource oauth2.TokenSource, path string, reconnect time.Duration) {
	w.streamBaseURL = baseURL
	w.tokenSource = tokenSource
	w.streamPath = path
	if reconnect > 0 {
		w.streamReconnect = reconnect
	}
}

// Run starts the periodic sync loop.
func (w *Worker) Run(ctx context.Context, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	if err := w.SyncOnce(ctx); err != nil {
		return err
	}

	if w.streamBaseURL != "" && w.tokenSource != nil {
		go w.runStream(ctx)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.SyncOnce(ctx); err != nil {
				return err
			}
		}
	}
}

// SyncOnce performs a full sync pass.
func (w *Worker) SyncOnce(ctx context.Context) error {
	if w.firebase == nil {
		return fmt.Errorf("firebase db client is nil")
	}
	if err := w.syncAdmins(ctx); err != nil {
		return err
	}
	if err := w.syncFactoryQuestions(ctx); err != nil {
		return err
	}
	if err := w.syncGlobalCategories(ctx); err != nil {
		return err
	}
	if err := w.syncUsers(ctx); err != nil {
		return err
	}
	return nil
}

func (w *Worker) syncAdmins(ctx context.Context) error {
	ref := w.firebase.NewRef("admins")
	var raw map[string]any
	if err := ref.Get(ctx, &raw); err != nil {
		return err
	}
	if err := w.exec(ctx, w.withPrefix("DELETE FROM admins;"), nil); err != nil {
		return err
	}
	for userID := range raw {
		query := w.withPrefix(`
DECLARE $user_id AS Utf8;
UPSERT INTO admins (user_id) VALUES ($user_id);`)
		params := table.NewQueryParameters(
			table.ValueParam("$user_id", types.UTF8Value(userID)),
		)
		if err := w.exec(ctx, query, params); err != nil {
			return err
		}
	}
	return nil
}

func (w *Worker) syncFactoryQuestions(ctx context.Context) error {
	ref := w.firebase.NewRef("factory/questions")
	var raw any
	if err := ref.Get(ctx, &raw); err != nil {
		return err
	}
	questions := parseFactoryQuestions(raw)

	if err := w.exec(ctx, w.withPrefix("DELETE FROM factory_questions;"), nil); err != nil {
		return err
	}

	for _, q := range questions {
		phrases, err := json.Marshal(q.Phrases)
		if err != nil {
			return err
		}
		query := w.withPrefix(`
DECLARE $question_id AS Utf8;
DECLARE $label AS Utf8;
DECLARE $phrases AS JsonDocument;
DECLARE $category AS Utf8;
DECLARE $type AS Utf8;
DECLARE $order_index AS Int64;
UPSERT INTO factory_questions (question_id, label, phrases, category, type, order_index)
VALUES ($question_id, $label, $phrases, $category, $type, $order_index);`)

		params := table.NewQueryParameters(
			table.ValueParam("$question_id", types.UTF8Value(q.ID)),
			table.ValueParam("$label", types.UTF8Value(q.Label)),
			table.ValueParam("$phrases", types.JSONDocumentValue(string(phrases))),
			table.ValueParam("$category", types.UTF8Value(q.Category)),
			table.ValueParam("$type", types.UTF8Value(q.Type)),
			table.ValueParam("$order_index", types.Int64Value(int64(q.OrderIndex))),
		)
		if err := w.exec(ctx, query, params); err != nil {
			return err
		}
	}

	return nil
}

func (w *Worker) syncGlobalCategories(ctx context.Context) error {
	ref := w.firebase.NewRef("global/Category")
	var raw map[string]any
	if err := ref.Get(ctx, &raw); err != nil {
		return err
	}

	if err := w.exec(ctx, w.withPrefix("DELETE FROM global_categories;"), nil); err != nil {
		return err
	}
	if err := w.exec(ctx, w.withPrefix("DELETE FROM global_statements;"), nil); err != nil {
		return err
	}

	now := time.Now().UnixMilli()
	for key, rawCat := range raw {
		cat, ok := rawCat.(map[string]any)
		if !ok {
			continue
		}
		categoryID := str(cat["id"], key)
		label := str(cat["label"], "")
		created := int64From(cat["created"], now)
		updatedAt := created
		isDefault := boolPtrFrom(cat["default"])

		query := w.withPrefix(`
DECLARE $category_id AS Utf8;
DECLARE $label AS Utf8;
DECLARE $created_at AS Int64;
DECLARE $is_default AS Bool?;
DECLARE $updated_at AS Int64;
UPSERT INTO global_categories (category_id, label, created_at, is_default, updated_at)
VALUES ($category_id, $label, $created_at, $is_default, $updated_at);`)
		params := table.NewQueryParameters(
			table.ValueParam("$category_id", types.UTF8Value(categoryID)),
			table.ValueParam("$label", types.UTF8Value(label)),
			table.ValueParam("$created_at", types.Int64Value(created)),
			table.ValueParam("$is_default", optionalBool(isDefault)),
			table.ValueParam("$updated_at", types.Int64Value(updatedAt)),
		)
		if err := w.exec(ctx, query, params); err != nil {
			return err
		}

		statements, _ := cat["statements"].(map[string]any)
		for stmtKey, rawStmt := range statements {
			stmtMap, ok := rawStmt.(map[string]any)
			if !ok {
				continue
			}
			statementID := str(stmtMap["id"], stmtKey)
			text := str(stmtMap["text"], "")
			stmtCreated := int64From(stmtMap["created"], now)
			stmtUpdated := stmtCreated

			stmtQuery := w.withPrefix(`
DECLARE $category_id AS Utf8;
DECLARE $statement_id AS Utf8;
DECLARE $text AS Utf8;
DECLARE $created_at AS Int64;
DECLARE $updated_at AS Int64;
UPSERT INTO global_statements (category_id, statement_id, text, created_at, updated_at)
VALUES ($category_id, $statement_id, $text, $created_at, $updated_at);`)
			stmtParams := table.NewQueryParameters(
				table.ValueParam("$category_id", types.UTF8Value(categoryID)),
				table.ValueParam("$statement_id", types.UTF8Value(statementID)),
				table.ValueParam("$text", types.UTF8Value(text)),
				table.ValueParam("$created_at", types.Int64Value(stmtCreated)),
				table.ValueParam("$updated_at", types.Int64Value(stmtUpdated)),
			)
			if err := w.exec(ctx, stmtQuery, stmtParams); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *Worker) syncUsers(ctx context.Context) error {
	ref := w.firebase.NewRef("users")
	var raw map[string]any
	if err := ref.Get(ctx, &raw); err != nil {
		return err
	}
	if raw == nil {
		return nil
	}

	for userID := range raw {
		categories, statements, err := w.fetchUserData(ctx, userID)
		if err != nil {
			return err
		}
		state, err := w.fetchUserState(ctx, userID)
		if err != nil {
			return err
		}

		for _, cat := range categories {
			if cat.Created == 0 {
				cat.Created = time.Now().UnixMilli()
			}
			if cat.UpdatedAt == 0 {
				cat.UpdatedAt = cat.Created
			}
			if _, err := w.store.UpsertCategory(ctx, userID, cat); err != nil {
				return err
			}
		}
		for _, stmt := range statements {
			if stmt.Created == 0 {
				stmt.Created = time.Now().UnixMilli()
			}
			if stmt.UpdatedAt == 0 {
				stmt.UpdatedAt = stmt.Created
			}
			if _, err := w.store.UpsertStatement(ctx, userID, stmt); err != nil {
				return err
			}
		}

		if len(state.Quickes) == 0 {
			state.Quickes = defaults.DefaultQuickes
		}
		if _, err := w.store.SetUserState(ctx, userID, state, time.Now().UnixMilli()); err != nil {
			return err
		}
	}

	return nil
}

func (w *Worker) fetchUserData(ctx context.Context, userID string) ([]models.Category, []models.Statement, error) {
	if w.legacy == nil {
		return nil, nil, fmt.Errorf("legacy reader is nil")
	}
	return w.legacy.FetchUserData(ctx, userID)
}

func (w *Worker) fetchUserState(ctx context.Context, userID string) (models.UserState, error) {
	if w.legacy == nil {
		return models.UserState{}, fmt.Errorf("legacy reader is nil")
	}
	return w.legacy.GetUserState(ctx, userID)
}

func (w *Worker) withPrefix(query string) string {
	return fmt.Sprintf("PRAGMA TablePathPrefix(\"%s\");\n%s", w.ydbClient.Database(), query)
}

func (w *Worker) exec(ctx context.Context, query string, params *table.QueryParameters) error {
	return w.ydbClient.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, _, err := sess.Execute(ctx, table.DefaultTxControl(), query, params)
		return err
	}, table.WithIdempotent())
}

func optionalBool(val *bool) types.Value {
	if val == nil {
		return types.OptionalValue(types.NullValue(types.TypeBool))
	}
	return types.OptionalValue(types.BoolValue(*val))
}

func str(raw any, fallback string) string {
	if value, ok := raw.(string); ok && value != "" {
		return value
	}
	return fallback
}

func int64From(raw any, fallback int64) int64 {
	switch value := raw.(type) {
	case int64:
		return value
	case int:
		return int64(value)
	case float64:
		return int64(value)
	default:
		return fallback
	}
}

func boolPtrFrom(raw any) *bool {
	if value, ok := raw.(bool); ok {
		return &value
	}
	return nil
}

func parseFactoryQuestions(raw any) []models.FactoryQuestion {
	questions := make([]models.FactoryQuestion, 0)
	switch value := raw.(type) {
	case []any:
		for idx, item := range value {
			q := questionFromAny(item)
			if q.ID == "" {
				q.ID = fmt.Sprintf("%d", idx)
			}
			if q.OrderIndex == 0 {
				q.OrderIndex = idx
			}
			questions = append(questions, q)
		}
	case map[string]any:
		idx := 0
		for key, item := range value {
			q := questionFromAny(item)
			if q.ID == "" {
				q.ID = key
			}
			if q.OrderIndex == 0 {
				q.OrderIndex = idx
			}
			questions = append(questions, q)
			idx++
		}
	}
	return questions
}

func questionFromAny(raw any) models.FactoryQuestion {
	data, ok := raw.(map[string]any)
	if !ok {
		return models.FactoryQuestion{}
	}

	q := models.FactoryQuestion{
		Label:    str(data["label"], ""),
		Category: str(data["category"], ""),
		Type:     str(data["type"], ""),
	}
	if uid := str(data["uid"], ""); uid != "" {
		q.ID = uid
	}
	if id := str(data["id"], ""); id != "" && q.ID == "" {
		q.ID = id
	}

	if rawOrder, ok := data["order_index"]; ok {
		q.OrderIndex = int(int64From(rawOrder, 0))
	}
	if rawOrder, ok := data["orderIndex"]; ok && q.OrderIndex == 0 {
		q.OrderIndex = int(int64From(rawOrder, 0))
	}

	q.Phrases = toStringSlice(data["phrases"])
	return q
}

func toStringSlice(raw any) []string {
	switch value := raw.(type) {
	case []string:
		return value
	case []any:
		out := make([]string, 0, len(value))
		for _, item := range value {
			if strVal, ok := item.(string); ok {
				out = append(out, strVal)
			}
		}
		return out
	default:
		return nil
	}
}
