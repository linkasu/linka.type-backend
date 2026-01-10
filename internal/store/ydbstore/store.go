package ydbstore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/linkasu/linka.type-backend/internal/models"
	"github.com/linkasu/linka.type-backend/internal/store"
	"github.com/linkasu/linka.type-backend/internal/ydb"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result/named"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
)

// Store implements store.Store backed by YDB.
type Store struct {
	client *ydb.Client
}

// New creates a YDB-backed store.
func New(client *ydb.Client) *Store {
	return &Store{client: client}
}

func (s *Store) ListCategories(ctx context.Context, userID string) ([]models.Category, error) {
	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
SELECT category_id, label, created_at, is_default, ai_use, updated_at
FROM categories
WHERE user_id = $user_id AND deleted_at IS NULL
ORDER BY created_at;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
	)

	var out []models.Category
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, params)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		for res.NextRow() {
			var (
				id        string
				label     string
				created   int64
				updated   *int64
				isDefault *bool
				aiUse     *bool
			)
			if err := res.ScanNamed(
				named.Required("category_id", &id),
				named.Required("label", &label),
				named.Required("created_at", &created),
				named.Optional("updated_at", &updated),
				named.Optional("is_default", &isDefault),
				named.Optional("ai_use", &aiUse),
			); err != nil {
				return err
			}
			cat := models.Category{
				ID:      id,
				Label:   label,
				Created: created,
				Default: isDefault,
			}
			if aiUse != nil {
				cat.AIUse = *aiUse
			}
			if updated != nil {
				cat.UpdatedAt = *updated
			} else {
				cat.UpdatedAt = created
			}
			out = append(out, cat)
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Store) UpsertCategory(ctx context.Context, userID string, category models.Category) (models.Category, error) {
	now := time.Now().UnixMilli()
	if category.Created == 0 {
		category.Created = now
	}
	if category.UpdatedAt == 0 {
		category.UpdatedAt = now
	}

	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $category_id AS Utf8;
DECLARE $label AS Utf8;
DECLARE $created_at AS Int64;
DECLARE $is_default AS Bool?;
DECLARE $updated_at AS Int64;
DECLARE $ai_use AS Bool?;
UPSERT INTO categories (user_id, category_id, label, created_at, is_default, ai_use, updated_at)
VALUES ($user_id, $category_id, $label, $created_at, $is_default, $ai_use, $updated_at);`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$category_id", types.UTF8Value(category.ID)),
		table.ValueParam("$label", types.UTF8Value(category.Label)),
		table.ValueParam("$created_at", types.Int64Value(category.Created)),
		table.ValueParam("$is_default", optionalBool(category.Default)),
		table.ValueParam("$ai_use", optionalBool(&category.AIUse)),
		table.ValueParam("$updated_at", types.Int64Value(category.UpdatedAt)),
	)

	err := s.execWrite(ctx, query, params)
	if err != nil {
		return models.Category{}, err
	}

	return category, nil
}

func (s *Store) DeleteCategory(ctx context.Context, userID, categoryID string, updatedAt int64) error {
	if updatedAt == 0 {
		updatedAt = time.Now().UnixMilli()
	}

	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $category_id AS Utf8;
DECLARE $updated_at AS Int64;
UPDATE categories
SET deleted_at = $updated_at, updated_at = $updated_at
WHERE user_id = $user_id AND category_id = $category_id;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$category_id", types.UTF8Value(categoryID)),
		table.ValueParam("$updated_at", types.Int64Value(updatedAt)),
	)

	return s.execWrite(ctx, query, params)
}

func (s *Store) ListStatements(ctx context.Context, userID, categoryID string) ([]models.Statement, error) {
	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $category_id AS Utf8;
SELECT statement_id, category_id, text, created_at, updated_at
FROM statements
WHERE user_id = $user_id AND category_id = $category_id AND deleted_at IS NULL
ORDER BY created_at;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$category_id", types.UTF8Value(categoryID)),
	)

	var out []models.Statement
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, params)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		for res.NextRow() {
			var (
				id      string
				catID   string
				text    string
				created int64
				updated *int64
			)
			if err := res.ScanNamed(
				named.Required("statement_id", &id),
				named.Required("category_id", &catID),
				named.Required("text", &text),
				named.Required("created_at", &created),
				named.Optional("updated_at", &updated),
			); err != nil {
				return err
			}
			stmt := models.Statement{
				ID:         id,
				CategoryID: catID,
				Text:       text,
				Created:    created,
			}
			if updated != nil {
				stmt.UpdatedAt = *updated
			} else {
				stmt.UpdatedAt = created
			}
			out = append(out, stmt)
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Store) ListAllStatements(ctx context.Context, userID string) ([]models.Statement, error) {
	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
SELECT statement_id, category_id, text, created_at, updated_at
FROM statements
WHERE user_id = $user_id AND deleted_at IS NULL
ORDER BY created_at;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
	)

	var out []models.Statement
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, params)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		for res.NextRow() {
			var (
				id      string
				catID   string
				text    string
				created int64
				updated *int64
			)
			if err := res.ScanNamed(
				named.Required("statement_id", &id),
				named.Required("category_id", &catID),
				named.Required("text", &text),
				named.Required("created_at", &created),
				named.Optional("updated_at", &updated),
			); err != nil {
				return err
			}
			stmt := models.Statement{
				ID:         id,
				CategoryID: catID,
				Text:       text,
				Created:    created,
			}
			if updated != nil {
				stmt.UpdatedAt = *updated
			} else {
				stmt.UpdatedAt = created
			}
			out = append(out, stmt)
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Store) UpsertStatement(ctx context.Context, userID string, statement models.Statement) (models.Statement, error) {
	now := time.Now().UnixMilli()
	if statement.Created == 0 {
		statement.Created = now
	}
	if statement.UpdatedAt == 0 {
		statement.UpdatedAt = now
	}

	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $category_id AS Utf8;
DECLARE $statement_id AS Utf8;
DECLARE $text AS Utf8;
DECLARE $created_at AS Int64;
DECLARE $updated_at AS Int64;
UPSERT INTO statements (user_id, category_id, statement_id, text, created_at, updated_at)
VALUES ($user_id, $category_id, $statement_id, $text, $created_at, $updated_at);`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$category_id", types.UTF8Value(statement.CategoryID)),
		table.ValueParam("$statement_id", types.UTF8Value(statement.ID)),
		table.ValueParam("$text", types.UTF8Value(statement.Text)),
		table.ValueParam("$created_at", types.Int64Value(statement.Created)),
		table.ValueParam("$updated_at", types.Int64Value(statement.UpdatedAt)),
	)

	err := s.execWrite(ctx, query, params)
	if err != nil {
		return models.Statement{}, err
	}

	return statement, nil
}

func (s *Store) DeleteStatement(ctx context.Context, userID, categoryID, statementID string, updatedAt int64) error {
	if updatedAt == 0 {
		updatedAt = time.Now().UnixMilli()
	}

	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $category_id AS Utf8;
DECLARE $statement_id AS Utf8;
DECLARE $updated_at AS Int64;
UPDATE statements
SET deleted_at = $updated_at, updated_at = $updated_at
WHERE user_id = $user_id AND category_id = $category_id AND statement_id = $statement_id;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$category_id", types.UTF8Value(categoryID)),
		table.ValueParam("$statement_id", types.UTF8Value(statementID)),
		table.ValueParam("$updated_at", types.Int64Value(updatedAt)),
	)

	return s.execWrite(ctx, query, params)
}

func (s *Store) GetUserState(ctx context.Context, userID string) (models.UserState, error) {
	state := models.UserState{}

	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
SELECT inited, preferences
FROM users
WHERE user_id = $user_id
LIMIT 1;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
	)

	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, params)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		if res.NextRow() {
			var (
				inited      bool
				preferences *string
			)
			if err := res.ScanNamed(
				named.Required("inited", &inited),
				named.Optional("preferences", &preferences),
			); err != nil {
				return err
			}
			state.Inited = inited
			if preferences != nil && *preferences != "" {
				if err := json.Unmarshal([]byte(*preferences), &state.Preferences); err != nil {
					return err
				}
			}
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return state, err
	}

	quickes, err := s.listQuickes(ctx, userID)
	if err != nil {
		return state, err
	}
	state.Quickes = quickes

	return state, nil
}

func (s *Store) SetUserState(ctx context.Context, userID string, state models.UserState, updatedAt int64) (models.UserState, error) {
	if updatedAt == 0 {
		updatedAt = time.Now().UnixMilli()
	}

	createdAt, err := s.getUserCreatedAt(ctx, userID)
	if err != nil && err != store.ErrNotFound {
		return state, err
	}
	if createdAt == 0 {
		createdAt = updatedAt
	}

	preferencesJSON := "{}"
	if state.Preferences != nil {
		serialized, err := json.Marshal(state.Preferences)
		if err != nil {
			return state, err
		}
		preferencesJSON = string(serialized)
	}

	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $created_at AS Int64;
DECLARE $inited AS Bool;
DECLARE $preferences AS JsonDocument;
DECLARE $deleted_at AS Int64?;
UPSERT INTO users (user_id, created_at, inited, preferences, deleted_at)
VALUES ($user_id, $created_at, $inited, $preferences, $deleted_at);`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$created_at", types.Int64Value(createdAt)),
		table.ValueParam("$inited", types.BoolValue(state.Inited)),
		table.ValueParam("$preferences", types.JSONDocumentValue(preferencesJSON)),
		table.ValueParam("$deleted_at", types.NullValue(types.TypeInt64)),
	)

	if err := s.execWrite(ctx, query, params); err != nil {
		return state, err
	}

	if len(state.Quickes) > 0 {
		quickes, err := s.SetQuickes(ctx, userID, state.Quickes, updatedAt)
		if err != nil {
			return state, err
		}
		state.Quickes = quickes
	}

	return state, nil
}

func (s *Store) SetQuickes(ctx context.Context, userID string, quickes []string, updatedAt int64) ([]string, error) {
	if updatedAt == 0 {
		updatedAt = time.Now().UnixMilli()
	}

	deleteQuery := s.withPrefix(`
DECLARE $user_id AS Utf8;
DELETE FROM quickes WHERE user_id = $user_id;`)
	deleteParams := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
	)
	if err := s.execWrite(ctx, deleteQuery, deleteParams); err != nil {
		return nil, err
	}

	for idx, text := range quickes {
		query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $slot AS Int64;
DECLARE $text AS Utf8;
DECLARE $updated_at AS Int64;
UPSERT INTO quickes (user_id, slot, text, updated_at)
VALUES ($user_id, $slot, $text, $updated_at);`)

		params := table.NewQueryParameters(
			table.ValueParam("$user_id", types.UTF8Value(userID)),
			table.ValueParam("$slot", types.Int64Value(int64(idx))),
			table.ValueParam("$text", types.UTF8Value(text)),
			table.ValueParam("$updated_at", types.Int64Value(updatedAt)),
		)

		if err := s.execWrite(ctx, query, params); err != nil {
			return nil, err
		}
	}

	return quickes, nil
}

func (s *Store) ListGlobalCategories(ctx context.Context, includeStatements bool) ([]models.GlobalCategory, error) {
	query := s.withPrefix(`
SELECT category_id, label, created_at, is_default, updated_at
FROM global_categories
WHERE deleted_at IS NULL
ORDER BY created_at;`)

	var out []models.GlobalCategory
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, nil)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		for res.NextRow() {
			var (
				id        string
				label     string
				created   int64
				updated   *int64
				isDefault *bool
			)
			if err := res.ScanNamed(
				named.Required("category_id", &id),
				named.Required("label", &label),
				named.Required("created_at", &created),
				named.Optional("updated_at", &updated),
				named.Optional("is_default", &isDefault),
			); err != nil {
				return err
			}
			cat := models.GlobalCategory{
				ID:      id,
				Label:   label,
				Created: created,
				Default: isDefault,
			}
			if updated != nil {
				cat.UpdatedAt = *updated
			} else {
				cat.UpdatedAt = created
			}
			out = append(out, cat)
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return nil, err
	}

	if includeStatements {
		for i := range out {
			statements, err := s.ListGlobalStatements(ctx, out[i].ID)
			if err != nil {
				return nil, err
			}
			out[i].Statements = statements
		}
	}

	return out, nil
}

func (s *Store) ListGlobalStatements(ctx context.Context, categoryID string) ([]models.Statement, error) {
	query := s.withPrefix(`
DECLARE $category_id AS Utf8;
SELECT statement_id, category_id, text, created_at, updated_at
FROM global_statements
WHERE category_id = $category_id AND deleted_at IS NULL
ORDER BY created_at;`)

	params := table.NewQueryParameters(
		table.ValueParam("$category_id", types.UTF8Value(categoryID)),
	)

	var out []models.Statement
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, params)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		for res.NextRow() {
			var (
				id      string
				catID   string
				text    string
				created int64
				updated *int64
			)
			if err := res.ScanNamed(
				named.Required("statement_id", &id),
				named.Required("category_id", &catID),
				named.Required("text", &text),
				named.Required("created_at", &created),
				named.Optional("updated_at", &updated),
			); err != nil {
				return err
			}
			stmt := models.Statement{
				ID:         id,
				CategoryID: catID,
				Text:       text,
				Created:    created,
			}
			if updated != nil {
				stmt.UpdatedAt = *updated
			} else {
				stmt.UpdatedAt = created
			}
			out = append(out, stmt)
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Store) ImportGlobalCategory(ctx context.Context, userID, categoryID string, force bool) (string, error) {
	exists, err := s.categoryExists(ctx, userID, categoryID)
	if err != nil {
		return "", err
	}
	if exists && !force {
		return "exists", nil
	}

	globals, err := s.ListGlobalCategories(ctx, false)
	if err != nil {
		return "", err
	}

	var globalCat *models.GlobalCategory
	for i := range globals {
		if globals[i].ID == categoryID {
			globalCat = &globals[i]
			break
		}
	}
	if globalCat == nil {
		return "", store.ErrNotFound
	}

	cat := models.Category{
		ID:        globalCat.ID,
		Label:     globalCat.Label,
		Created:   globalCat.Created,
		Default:   globalCat.Default,
		UpdatedAt: time.Now().UnixMilli(),
	}
	if _, err := s.UpsertCategory(ctx, userID, cat); err != nil {
		return "", err
	}

	statements, err := s.ListGlobalStatements(ctx, categoryID)
	if err != nil {
		return "", err
	}
	for _, stmt := range statements {
		stmt.CategoryID = categoryID
		stmt.UpdatedAt = time.Now().UnixMilli()
		if _, err := s.UpsertStatement(ctx, userID, stmt); err != nil {
			return "", err
		}
	}

	return "ok", nil
}

func (s *Store) ListFactoryQuestions(ctx context.Context) ([]models.FactoryQuestion, error) {
	query := s.withPrefix(`
SELECT question_id, label, phrases, category, type, order_index
FROM factory_questions
ORDER BY order_index;`)

	var out []models.FactoryQuestion
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, nil)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		for res.NextRow() {
			var (
				id       string
				label    string
				phrases  string
				category string
				qtype    string
				orderIdx int64
			)
			if err := res.ScanNamed(
				named.Required("question_id", &id),
				named.Required("label", &label),
				named.Required("phrases", &phrases),
				named.Required("category", &category),
				named.Required("type", &qtype),
				named.Required("order_index", &orderIdx),
			); err != nil {
				return err
			}
			var list []string
			if phrases != "" {
				if err := json.Unmarshal([]byte(phrases), &list); err != nil {
					return err
				}
			}
			out = append(out, models.FactoryQuestion{
				ID:         id,
				Label:      label,
				Phrases:    list,
				Category:   category,
				Type:       qtype,
				OrderIndex: int(orderIdx),
			})
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Store) IsAdmin(ctx context.Context, userID string) (bool, error) {
	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
SELECT user_id
FROM admins
WHERE user_id = $user_id
LIMIT 1;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
	)

	var exists bool
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, params)
		if err != nil {
			return err
		}
		defer res.Close()
		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		if res.NextRow() {
			exists = true
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *Store) DeleteUser(ctx context.Context, userID string, updatedAt int64) error {
	if updatedAt == 0 {
		updatedAt = time.Now().UnixMilli()
	}

	queries := []struct {
		query  string
		params *table.QueryParameters
	}{
		{
			query: s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $deleted_at AS Int64;
UPDATE users SET deleted_at = $deleted_at WHERE user_id = $user_id;`),
			params: table.NewQueryParameters(
				table.ValueParam("$user_id", types.UTF8Value(userID)),
				table.ValueParam("$deleted_at", types.Int64Value(updatedAt)),
			),
		},
		{
			query: s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $deleted_at AS Int64;
UPDATE categories SET deleted_at = $deleted_at, updated_at = $deleted_at WHERE user_id = $user_id;`),
			params: table.NewQueryParameters(
				table.ValueParam("$user_id", types.UTF8Value(userID)),
				table.ValueParam("$deleted_at", types.Int64Value(updatedAt)),
			),
		},
		{
			query: s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $deleted_at AS Int64;
UPDATE statements SET deleted_at = $deleted_at, updated_at = $deleted_at WHERE user_id = $user_id;`),
			params: table.NewQueryParameters(
				table.ValueParam("$user_id", types.UTF8Value(userID)),
				table.ValueParam("$deleted_at", types.Int64Value(updatedAt)),
			),
		},
		{
			query: s.withPrefix(`
DECLARE $user_id AS Utf8;
DELETE FROM quickes WHERE user_id = $user_id;`),
			params: table.NewQueryParameters(
				table.ValueParam("$user_id", types.UTF8Value(userID)),
			),
		},
		{
			query: s.withPrefix(`
DECLARE $user_id AS Utf8;
DELETE FROM changes WHERE user_id = $user_id;`),
			params: table.NewQueryParameters(
				table.ValueParam("$user_id", types.UTF8Value(userID)),
			),
		},
	}

	for _, item := range queries {
		if err := s.execWrite(ctx, item.query, item.params); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) AppendChange(ctx context.Context, userID string, change models.ChangeEvent) error {
	if change.Cursor == "" {
		change.Cursor = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if change.UpdatedAt == 0 {
		change.UpdatedAt = time.Now().UnixMilli()
	}

	payload := string(change.Payload)
	if payload == "" {
		payload = "{}"
	}

	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $cursor AS Utf8;
DECLARE $entity_type AS Utf8;
DECLARE $entity_id AS Utf8;
DECLARE $op AS Utf8;
DECLARE $payload AS JsonDocument;
DECLARE $updated_at AS Int64;
UPSERT INTO changes (user_id, cursor, entity_type, entity_id, op, payload, updated_at)
VALUES ($user_id, $cursor, $entity_type, $entity_id, $op, $payload, $updated_at);`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$cursor", types.UTF8Value(change.Cursor)),
		table.ValueParam("$entity_type", types.UTF8Value(change.EntityType)),
		table.ValueParam("$entity_id", types.UTF8Value(change.EntityID)),
		table.ValueParam("$op", types.UTF8Value(change.Op)),
		table.ValueParam("$payload", types.JSONDocumentValue(payload)),
		table.ValueParam("$updated_at", types.Int64Value(change.UpdatedAt)),
	)

	return s.execWrite(ctx, query, params)
}

func (s *Store) ListChanges(ctx context.Context, userID, cursor string, limit int) (string, []models.ChangeEvent, error) {
	if limit <= 0 {
		limit = 100
	}

	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $cursor AS Utf8;
DECLARE $limit AS Uint64;
SELECT cursor, entity_type, entity_id, op, payload, updated_at
FROM changes
WHERE user_id = $user_id AND cursor > $cursor
ORDER BY cursor
LIMIT $limit;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$cursor", types.UTF8Value(cursor)),
		table.ValueParam("$limit", types.Uint64Value(uint64(limit))),
	)

	var (
		changes    []models.ChangeEvent
		lastCursor string
	)

	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, params)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		for res.NextRow() {
			var (
				curs     string
				entityTy string
				entityID string
				op       string
				payload  string
				updated  int64
			)
			if err := res.ScanNamed(
				named.Required("cursor", &curs),
				named.Required("entity_type", &entityTy),
				named.Required("entity_id", &entityID),
				named.Required("op", &op),
				named.Required("payload", &payload),
				named.Required("updated_at", &updated),
			); err != nil {
				return err
			}
			changes = append(changes, models.ChangeEvent{
				Cursor:     curs,
				EntityType: entityTy,
				EntityID:   entityID,
				Op:         op,
				Payload:    json.RawMessage(payload),
				UpdatedAt:  updated,
			})
			lastCursor = curs
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return cursor, nil, err
	}

	if lastCursor == "" {
		lastCursor = cursor
	}

	return lastCursor, changes, nil
}

func (s *Store) withPrefix(query string) string {
	return fmt.Sprintf("PRAGMA TablePathPrefix(\"%s\");\n%s", s.client.Database(), query)
}

func (s *Store) execWrite(ctx context.Context, query string, params *table.QueryParameters) error {
	return s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, _, err := sess.Execute(ctx, table.DefaultTxControl(), query, params)
		return err
	}, table.WithIdempotent())
}

func optionalBool(val *bool) types.Value {
	if val == nil {
		return types.NullValue(types.TypeBool)
	}
	return types.OptionalValue(types.BoolValue(*val))
}

func optionalInt64(val int64) types.Value {
	if val == 0 {
		return types.NullValue(types.TypeInt64)
	}
	return types.OptionalValue(types.Int64Value(val))
}

func optionalString(val string) types.Value {
	if val == "" {
		return types.NullValue(types.TypeUTF8)
	}
	return types.OptionalValue(types.UTF8Value(val))
}

func optionalStringPtr(val *string) types.Value {
	if val == nil || *val == "" {
		return types.NullValue(types.TypeUTF8)
	}
	return types.OptionalValue(types.UTF8Value(*val))
}

func (s *Store) getUserCreatedAt(ctx context.Context, userID string) (int64, error) {
	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
SELECT created_at
FROM users
WHERE user_id = $user_id
LIMIT 1;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
	)

	var createdAt int64
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, params)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		if !res.NextRow() {
			return store.ErrNotFound
		}
		if err := res.ScanNamed(named.Required("created_at", &createdAt)); err != nil {
			return err
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return 0, err
	}

	return createdAt, nil
}

func (s *Store) listQuickes(ctx context.Context, userID string) ([]string, error) {
	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
SELECT slot, text
FROM quickes
WHERE user_id = $user_id
ORDER BY slot;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
	)

	var quickes []string
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, params)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		for res.NextRow() {
			var (
				slot int64
				text string
			)
			if err := res.ScanNamed(
				named.Required("slot", &slot),
				named.Required("text", &text),
			); err != nil {
				return err
			}
			if slot < 0 {
				continue
			}
			if int(slot) >= len(quickes) {
				newQuickes := make([]string, int(slot)+1)
				copy(newQuickes, quickes)
				quickes = newQuickes
			}
			quickes[slot] = text
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return nil, err
	}

	return quickes, nil
}

func (s *Store) categoryExists(ctx context.Context, userID, categoryID string) (bool, error) {
	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $category_id AS Utf8;
SELECT category_id
FROM categories
WHERE user_id = $user_id AND category_id = $category_id AND deleted_at IS NULL
LIMIT 1;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$category_id", types.UTF8Value(categoryID)),
	)

	var exists bool
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, params)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		if res.NextRow() {
			exists = true
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *Store) CountUsers(ctx context.Context, since time.Time) (int64, error) {
	sinceMs := since.UnixMilli()
	query := s.withPrefix(`
DECLARE $since AS Int64;
SELECT COUNT(*) as cnt
FROM users
WHERE created_at >= $since AND deleted_at IS NULL;`)

	params := table.NewQueryParameters(
		table.ValueParam("$since", types.Int64Value(sinceMs)),
	)

	var count int64
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, params)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		if !res.NextRow() {
			return nil
		}
		if err := res.ScanNamed(named.Required("cnt", &count)); err != nil {
			return err
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Store) CountCategories(ctx context.Context, since time.Time) (int64, error) {
	sinceMs := since.UnixMilli()
	query := s.withPrefix(`
DECLARE $since AS Int64;
SELECT COUNT(*) as cnt
FROM categories
WHERE created_at >= $since AND deleted_at IS NULL;`)

	params := table.NewQueryParameters(
		table.ValueParam("$since", types.Int64Value(sinceMs)),
	)

	var count int64
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, params)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		if !res.NextRow() {
			return nil
		}
		if err := res.ScanNamed(named.Required("cnt", &count)); err != nil {
			return err
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Store) CountStatements(ctx context.Context, since time.Time) (int64, error) {
	sinceMs := since.UnixMilli()
	query := s.withPrefix(`
DECLARE $since AS Int64;
SELECT COUNT(*) as cnt
FROM statements
WHERE created_at >= $since AND deleted_at IS NULL;`)

	params := table.NewQueryParameters(
		table.ValueParam("$since", types.Int64Value(sinceMs)),
	)

	var count int64
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, params)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		if !res.NextRow() {
			return nil
		}
		if err := res.ScanNamed(named.Required("cnt", &count)); err != nil {
			return err
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Store) ListAdmins(ctx context.Context) ([]string, error) {
	query := s.withPrefix(`
SELECT user_id
FROM admins
ORDER BY user_id;`)

	var admins []string
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, nil)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		for res.NextRow() {
			var userID string
			if err := res.ScanNamed(named.Required("user_id", &userID)); err != nil {
				return err
			}
			admins = append(admins, userID)
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return nil, err
	}

	return admins, nil
}

func (s *Store) AddAdmin(ctx context.Context, userID string) error {
	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
UPSERT INTO admins (user_id) VALUES ($user_id);`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
	)

	return s.execWrite(ctx, query, params)
}

func (s *Store) RemoveAdmin(ctx context.Context, userID string) error {
	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DELETE FROM admins WHERE user_id = $user_id;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
	)

	return s.execWrite(ctx, query, params)
}

func (s *Store) CreateClientKey(ctx context.Context, key store.ClientKey) error {
	query := s.withPrefix(`
DECLARE $key_hash AS Utf8;
DECLARE $client_id AS Utf8;
DECLARE $status AS Utf8;
DECLARE $created_at AS Int64;
UPSERT INTO client_keys (key_hash, client_id, status, created_at, revoked_at)
VALUES ($key_hash, $client_id, $status, $created_at, NULL);`)

	params := table.NewQueryParameters(
		table.ValueParam("$key_hash", types.UTF8Value(key.KeyHash)),
		table.ValueParam("$client_id", types.UTF8Value(key.ClientID)),
		table.ValueParam("$status", types.UTF8Value(key.Status)),
		table.ValueParam("$created_at", types.Int64Value(key.CreatedAt)),
	)

	return s.execWrite(ctx, query, params)
}

func (s *Store) ListClientKeys(ctx context.Context) ([]store.ClientKey, error) {
	query := s.withPrefix(`
SELECT key_hash, client_id, status, created_at, revoked_at
FROM client_keys
ORDER BY created_at DESC;`)

	var keys []store.ClientKey
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, nil)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		for res.NextRow() {
			var (
				keyHash  string
				clientID string
				status   string
				created  int64
				revoked  *int64
			)
			if err := res.ScanNamed(
				named.Required("key_hash", &keyHash),
				named.Required("client_id", &clientID),
				named.Required("status", &status),
				named.Required("created_at", &created),
				named.Optional("revoked_at", &revoked),
			); err != nil {
				return err
			}
			keys = append(keys, store.ClientKey{
				KeyHash:   keyHash,
				ClientID:  clientID,
				Status:    status,
				CreatedAt: created,
				RevokedAt: revoked,
			})
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func (s *Store) RevokeClientKey(ctx context.Context, keyHash string) error {
	now := time.Now().UnixMilli()
	query := s.withPrefix(`
DECLARE $key_hash AS Utf8;
DECLARE $revoked_at AS Int64;
UPDATE client_keys
SET status = 'revoked', revoked_at = $revoked_at
WHERE key_hash = $key_hash;`)

	params := table.NewQueryParameters(
		table.ValueParam("$key_hash", types.UTF8Value(keyHash)),
		table.ValueParam("$revoked_at", types.Int64Value(now)),
	)

	return s.execWrite(ctx, query, params)
}

func (s *Store) UpsertGlobalCategory(ctx context.Context, category models.GlobalCategory) (models.GlobalCategory, error) {
	now := time.Now().UnixMilli()
	if category.Created == 0 {
		category.Created = now
	}
	if category.UpdatedAt == 0 {
		category.UpdatedAt = now
	}

	query := s.withPrefix(`
DECLARE $category_id AS Utf8;
DECLARE $label AS Utf8;
DECLARE $created_at AS Int64;
DECLARE $is_default AS Bool?;
DECLARE $updated_at AS Int64;
UPSERT INTO global_categories (category_id, label, created_at, is_default, updated_at)
VALUES ($category_id, $label, $created_at, $is_default, $updated_at);`)

	params := table.NewQueryParameters(
		table.ValueParam("$category_id", types.UTF8Value(category.ID)),
		table.ValueParam("$label", types.UTF8Value(category.Label)),
		table.ValueParam("$created_at", types.Int64Value(category.Created)),
		table.ValueParam("$is_default", optionalBool(category.Default)),
		table.ValueParam("$updated_at", types.Int64Value(category.UpdatedAt)),
	)

	err := s.execWrite(ctx, query, params)
	if err != nil {
		return models.GlobalCategory{}, err
	}

	return category, nil
}

func (s *Store) DeleteGlobalCategory(ctx context.Context, categoryID string, updatedAt int64) error {
	if updatedAt == 0 {
		updatedAt = time.Now().UnixMilli()
	}

	queries := []struct {
		query  string
		params *table.QueryParameters
	}{
		{
			query: s.withPrefix(`
DECLARE $category_id AS Utf8;
DECLARE $deleted_at AS Int64;
UPDATE global_categories
SET deleted_at = $deleted_at, updated_at = $deleted_at
WHERE category_id = $category_id;`),
			params: table.NewQueryParameters(
				table.ValueParam("$category_id", types.UTF8Value(categoryID)),
				table.ValueParam("$deleted_at", types.Int64Value(updatedAt)),
			),
		},
		{
			query: s.withPrefix(`
DECLARE $category_id AS Utf8;
DECLARE $deleted_at AS Int64;
UPDATE global_statements
SET deleted_at = $deleted_at, updated_at = $deleted_at
WHERE category_id = $category_id;`),
			params: table.NewQueryParameters(
				table.ValueParam("$category_id", types.UTF8Value(categoryID)),
				table.ValueParam("$deleted_at", types.Int64Value(updatedAt)),
			),
		},
	}

	for _, item := range queries {
		if err := s.execWrite(ctx, item.query, item.params); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) UpsertFactoryQuestion(ctx context.Context, question models.FactoryQuestion) (models.FactoryQuestion, error) {
	phrasesJSON, err := json.Marshal(question.Phrases)
	if err != nil {
		return models.FactoryQuestion{}, err
	}

	query := s.withPrefix(`
DECLARE $question_id AS Utf8;
DECLARE $label AS Utf8;
DECLARE $phrases AS JsonDocument;
DECLARE $category AS Utf8;
DECLARE $type AS Utf8;
DECLARE $order_index AS Int64;
UPSERT INTO factory_questions (question_id, label, phrases, category, type, order_index)
VALUES ($question_id, $label, $phrases, $category, $type, $order_index);`)

	params := table.NewQueryParameters(
		table.ValueParam("$question_id", types.UTF8Value(question.ID)),
		table.ValueParam("$label", types.UTF8Value(question.Label)),
		table.ValueParam("$phrases", types.JSONDocumentValue(string(phrasesJSON))),
		table.ValueParam("$category", types.UTF8Value(question.Category)),
		table.ValueParam("$type", types.UTF8Value(question.Type)),
		table.ValueParam("$order_index", types.Int64Value(int64(question.OrderIndex))),
	)

	err = s.execWrite(ctx, query, params)
	if err != nil {
		return models.FactoryQuestion{}, err
	}

	return question, nil
}

func (s *Store) DeleteFactoryQuestion(ctx context.Context, questionID string) error {
	query := s.withPrefix(`
DECLARE $question_id AS Utf8;
DELETE FROM factory_questions
WHERE question_id = $question_id;`)

	params := table.NewQueryParameters(
		table.ValueParam("$question_id", types.UTF8Value(questionID)),
	)

	return s.execWrite(ctx, query, params)
}

func (s *Store) ListDialogChats(ctx context.Context, userID string) ([]models.DialogChat, error) {
	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
SELECT chat_id, title, created_at, updated_at, last_message_at, message_count
FROM dialog_chats
WHERE user_id = $user_id AND deleted_at IS NULL
ORDER BY updated_at DESC;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
	)

	var out []models.DialogChat
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, params)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		for res.NextRow() {
			var (
				id            string
				title         string
				created       int64
				updated       *int64
				lastMessageAt *int64
				messageCount  *int64
			)
			if err := res.ScanNamed(
				named.Required("chat_id", &id),
				named.Required("title", &title),
				named.Required("created_at", &created),
				named.Optional("updated_at", &updated),
				named.Optional("last_message_at", &lastMessageAt),
				named.Optional("message_count", &messageCount),
			); err != nil {
				return err
			}
			chat := models.DialogChat{
				ID:      id,
				Title:   title,
				Created: created,
			}
			if updated != nil {
				chat.UpdatedAt = *updated
			} else {
				chat.UpdatedAt = created
			}
			if lastMessageAt != nil {
				chat.LastMessageAt = *lastMessageAt
			}
			if messageCount != nil {
				chat.MessageCount = *messageCount
			}
			out = append(out, chat)
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Store) GetDialogChat(ctx context.Context, userID, chatID string) (models.DialogChat, error) {
	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $chat_id AS Utf8;
SELECT chat_id, title, created_at, updated_at, last_message_at, message_count
FROM dialog_chats
WHERE user_id = $user_id AND chat_id = $chat_id AND deleted_at IS NULL
LIMIT 1;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$chat_id", types.UTF8Value(chatID)),
	)

	var out models.DialogChat
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, params)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		if !res.NextRow() {
			return store.ErrNotFound
		}

		var (
			id            string
			title         string
			created       int64
			updated       *int64
			lastMessageAt *int64
			messageCount  *int64
		)
		if err := res.ScanNamed(
			named.Required("chat_id", &id),
			named.Required("title", &title),
			named.Required("created_at", &created),
			named.Optional("updated_at", &updated),
			named.Optional("last_message_at", &lastMessageAt),
			named.Optional("message_count", &messageCount),
		); err != nil {
			return err
		}
		out = models.DialogChat{
			ID:      id,
			Title:   title,
			Created: created,
		}
		if updated != nil {
			out.UpdatedAt = *updated
		} else {
			out.UpdatedAt = created
		}
		if lastMessageAt != nil {
			out.LastMessageAt = *lastMessageAt
		}
		if messageCount != nil {
			out.MessageCount = *messageCount
		}

		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return models.DialogChat{}, err
	}

	return out, nil
}

func (s *Store) UpsertDialogChat(ctx context.Context, userID string, chat models.DialogChat) (models.DialogChat, error) {
	now := time.Now().UnixMilli()
	if chat.Created == 0 {
		chat.Created = now
	}
	if chat.UpdatedAt == 0 {
		chat.UpdatedAt = now
	}

	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $chat_id AS Utf8;
DECLARE $title AS Utf8;
DECLARE $created_at AS Int64;
DECLARE $updated_at AS Int64;
DECLARE $last_message_at AS Int64?;
DECLARE $message_count AS Int64?;
UPSERT INTO dialog_chats (user_id, chat_id, title, created_at, updated_at, last_message_at, message_count)
VALUES ($user_id, $chat_id, $title, $created_at, $updated_at, $last_message_at, $message_count);`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$chat_id", types.UTF8Value(chat.ID)),
		table.ValueParam("$title", types.UTF8Value(chat.Title)),
		table.ValueParam("$created_at", types.Int64Value(chat.Created)),
		table.ValueParam("$updated_at", types.Int64Value(chat.UpdatedAt)),
		table.ValueParam("$last_message_at", optionalInt64(chat.LastMessageAt)),
		table.ValueParam("$message_count", optionalInt64(chat.MessageCount)),
	)

	if err := s.execWrite(ctx, query, params); err != nil {
		return models.DialogChat{}, err
	}

	return chat, nil
}

func (s *Store) DeleteDialogChat(ctx context.Context, userID, chatID string, updatedAt int64) error {
	if updatedAt == 0 {
		updatedAt = time.Now().UnixMilli()
	}
	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $chat_id AS Utf8;
DECLARE $updated_at AS Int64;
UPDATE dialog_chats
SET deleted_at = $updated_at, updated_at = $updated_at
WHERE user_id = $user_id AND chat_id = $chat_id;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$chat_id", types.UTF8Value(chatID)),
		table.ValueParam("$updated_at", types.Int64Value(updatedAt)),
	)

	return s.execWrite(ctx, query, params)
}

func (s *Store) ListDialogMessages(ctx context.Context, userID, chatID string, limit int, before int64) ([]models.DialogMessage, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}

	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $chat_id AS Utf8;
DECLARE $before AS Int64;
DECLARE $limit AS Int64;
SELECT message_id, role, content, source, created_at, updated_at
FROM dialog_messages
WHERE user_id = $user_id AND chat_id = $chat_id AND deleted_at IS NULL
  AND ($before = 0 OR created_at < $before)
ORDER BY created_at DESC
LIMIT $limit;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$chat_id", types.UTF8Value(chatID)),
		table.ValueParam("$before", types.Int64Value(before)),
		table.ValueParam("$limit", types.Int64Value(int64(limit))),
	)

	var out []models.DialogMessage
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, params)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		for res.NextRow() {
			var (
				id      string
				role    string
				content string
				source  *string
				created int64
				updated *int64
			)
			if err := res.ScanNamed(
				named.Required("message_id", &id),
				named.Required("role", &role),
				named.Required("content", &content),
				named.Optional("source", &source),
				named.Required("created_at", &created),
				named.Optional("updated_at", &updated),
			); err != nil {
				return err
			}
			msg := models.DialogMessage{
				ID:      id,
				ChatID:  chatID,
				Role:    role,
				Content: content,
				Created: created,
			}
			if source != nil {
				msg.Source = *source
			}
			if updated != nil {
				msg.UpdatedAt = *updated
			} else {
				msg.UpdatedAt = created
			}
			out = append(out, msg)
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return nil, err
	}

	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}

	return out, nil
}

func (s *Store) ListOldestDialogMessages(ctx context.Context, userID, chatID string, limit int) ([]models.DialogMessage, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}

	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $chat_id AS Utf8;
DECLARE $limit AS Int64;
SELECT message_id, role, content, source, created_at, updated_at
FROM dialog_messages
WHERE user_id = $user_id AND chat_id = $chat_id AND deleted_at IS NULL
ORDER BY created_at ASC
LIMIT $limit;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$chat_id", types.UTF8Value(chatID)),
		table.ValueParam("$limit", types.Int64Value(int64(limit))),
	)

	var out []models.DialogMessage
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, params)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		for res.NextRow() {
			var (
				id      string
				role    string
				content string
				source  *string
				created int64
				updated *int64
			)
			if err := res.ScanNamed(
				named.Required("message_id", &id),
				named.Required("role", &role),
				named.Required("content", &content),
				named.Optional("source", &source),
				named.Required("created_at", &created),
				named.Optional("updated_at", &updated),
			); err != nil {
				return err
			}
			msg := models.DialogMessage{
				ID:      id,
				ChatID:  chatID,
				Role:    role,
				Content: content,
				Created: created,
			}
			if source != nil {
				msg.Source = *source
			}
			if updated != nil {
				msg.UpdatedAt = *updated
			} else {
				msg.UpdatedAt = created
			}
			out = append(out, msg)
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Store) CountDialogMessages(ctx context.Context, userID, chatID string) (int64, error) {
	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $chat_id AS Utf8;
SELECT COUNT(*) AS total
FROM dialog_messages
WHERE user_id = $user_id AND chat_id = $chat_id AND deleted_at IS NULL;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$chat_id", types.UTF8Value(chatID)),
	)

	var total int64
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, params)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		if !res.NextRow() {
			total = 0
			return nil
		}
		return res.ScanNamed(named.Required("total", &total))
	}, table.WithIdempotent())
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (s *Store) UpsertDialogMessage(ctx context.Context, userID string, message models.DialogMessage) (models.DialogMessage, error) {
	now := time.Now().UnixMilli()
	if message.Created == 0 {
		message.Created = now
	}
	if message.UpdatedAt == 0 {
		message.UpdatedAt = now
	}

	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $chat_id AS Utf8;
DECLARE $message_id AS Utf8;
DECLARE $role AS Utf8;
DECLARE $content AS Utf8;
DECLARE $source AS Utf8?;
DECLARE $created_at AS Int64;
DECLARE $updated_at AS Int64;
UPSERT INTO dialog_messages (user_id, chat_id, message_id, role, content, source, created_at, updated_at)
VALUES ($user_id, $chat_id, $message_id, $role, $content, $source, $created_at, $updated_at);`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$chat_id", types.UTF8Value(message.ChatID)),
		table.ValueParam("$message_id", types.UTF8Value(message.ID)),
		table.ValueParam("$role", types.UTF8Value(message.Role)),
		table.ValueParam("$content", types.UTF8Value(message.Content)),
		table.ValueParam("$source", optionalString(message.Source)),
		table.ValueParam("$created_at", types.Int64Value(message.Created)),
		table.ValueParam("$updated_at", types.Int64Value(message.UpdatedAt)),
	)

	if err := s.execWrite(ctx, query, params); err != nil {
		return models.DialogMessage{}, err
	}

	return message, nil
}

func (s *Store) DeleteDialogMessage(ctx context.Context, userID, chatID, messageID string, updatedAt int64) error {
	if updatedAt == 0 {
		updatedAt = time.Now().UnixMilli()
	}

	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $chat_id AS Utf8;
DECLARE $message_id AS Utf8;
DECLARE $updated_at AS Int64;
UPDATE dialog_messages
SET deleted_at = $updated_at, updated_at = $updated_at
WHERE user_id = $user_id AND chat_id = $chat_id AND message_id = $message_id;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$chat_id", types.UTF8Value(chatID)),
		table.ValueParam("$message_id", types.UTF8Value(messageID)),
		table.ValueParam("$updated_at", types.Int64Value(updatedAt)),
	)

	return s.execWrite(ctx, query, params)
}

func (s *Store) DeleteDialogMessagesByChat(ctx context.Context, userID, chatID string, updatedAt int64) error {
	if updatedAt == 0 {
		updatedAt = time.Now().UnixMilli()
	}
	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $chat_id AS Utf8;
DECLARE $updated_at AS Int64;
UPDATE dialog_messages
SET deleted_at = $updated_at, updated_at = $updated_at
WHERE user_id = $user_id AND chat_id = $chat_id;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$chat_id", types.UTF8Value(chatID)),
		table.ValueParam("$updated_at", types.Int64Value(updatedAt)),
	)

	return s.execWrite(ctx, query, params)
}

func (s *Store) ListDialogSuggestions(ctx context.Context, userID string, status string, limit int) ([]models.DialogSuggestion, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $status AS Utf8;
DECLARE $limit AS Int64;
SELECT suggestion_id, chat_id, message_id, text, status, category_id, created_at, updated_at
FROM dialog_suggestions
WHERE user_id = $user_id AND status = $status
ORDER BY created_at DESC
LIMIT $limit;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$status", types.UTF8Value(status)),
		table.ValueParam("$limit", types.Int64Value(int64(limit))),
	)

	var out []models.DialogSuggestion
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, params)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		for res.NextRow() {
			var (
				id        string
				chatID    *string
				messageID *string
				text      string
				statusVal string
				category  *string
				created   int64
				updated   *int64
			)
			if err := res.ScanNamed(
				named.Required("suggestion_id", &id),
				named.Optional("chat_id", &chatID),
				named.Optional("message_id", &messageID),
				named.Required("text", &text),
				named.Required("status", &statusVal),
				named.Optional("category_id", &category),
				named.Required("created_at", &created),
				named.Optional("updated_at", &updated),
			); err != nil {
				return err
			}
			suggestion := models.DialogSuggestion{
				ID:      id,
				Text:    text,
				Status:  statusVal,
				Created: created,
			}
			if chatID != nil {
				suggestion.ChatID = *chatID
			}
			if messageID != nil {
				suggestion.MessageID = *messageID
			}
			if category != nil {
				suggestion.CategoryID = category
			}
			if updated != nil {
				suggestion.UpdatedAt = *updated
			} else {
				suggestion.UpdatedAt = created
			}
			out = append(out, suggestion)
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Store) CountDialogSuggestions(ctx context.Context, userID string, status string) (int64, error) {
	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $status AS Utf8;
SELECT COUNT(*) AS total
FROM dialog_suggestions
WHERE user_id = $user_id AND status = $status;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$status", types.UTF8Value(status)),
	)

	var total int64
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, params)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		if !res.NextRow() {
			total = 0
			return nil
		}
		return res.ScanNamed(named.Required("total", &total))
	}, table.WithIdempotent())
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (s *Store) UpsertDialogSuggestion(ctx context.Context, userID string, suggestion models.DialogSuggestion) (models.DialogSuggestion, error) {
	now := time.Now().UnixMilli()
	if suggestion.Created == 0 {
		suggestion.Created = now
	}
	if suggestion.UpdatedAt == 0 {
		suggestion.UpdatedAt = now
	}

	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $suggestion_id AS Utf8;
DECLARE $chat_id AS Utf8?;
DECLARE $message_id AS Utf8?;
DECLARE $text AS Utf8;
DECLARE $status AS Utf8;
DECLARE $category_id AS Utf8?;
DECLARE $created_at AS Int64;
DECLARE $updated_at AS Int64;
UPSERT INTO dialog_suggestions (user_id, suggestion_id, chat_id, message_id, text, status, category_id, created_at, updated_at)
VALUES ($user_id, $suggestion_id, $chat_id, $message_id, $text, $status, $category_id, $created_at, $updated_at);`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$suggestion_id", types.UTF8Value(suggestion.ID)),
		table.ValueParam("$chat_id", optionalString(suggestion.ChatID)),
		table.ValueParam("$message_id", optionalString(suggestion.MessageID)),
		table.ValueParam("$text", types.UTF8Value(suggestion.Text)),
		table.ValueParam("$status", types.UTF8Value(suggestion.Status)),
		table.ValueParam("$category_id", optionalStringPtr(suggestion.CategoryID)),
		table.ValueParam("$created_at", types.Int64Value(suggestion.Created)),
		table.ValueParam("$updated_at", types.Int64Value(suggestion.UpdatedAt)),
	)

	if err := s.execWrite(ctx, query, params); err != nil {
		return models.DialogSuggestion{}, err
	}

	return suggestion, nil
}

func (s *Store) DeleteDialogSuggestion(ctx context.Context, userID, suggestionID string) error {
	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $suggestion_id AS Utf8;
DELETE FROM dialog_suggestions
WHERE user_id = $user_id AND suggestion_id = $suggestion_id;`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$suggestion_id", types.UTF8Value(suggestionID)),
	)

	return s.execWrite(ctx, query, params)
}

func (s *Store) ListDialogSuggestionJobs(ctx context.Context, status string, limit int) ([]models.DialogSuggestionJob, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query := s.withPrefix(`
DECLARE $status AS Utf8;
DECLARE $limit AS Int64;
SELECT job_id, user_id, chat_id, message_id, status, attempts, last_error, created_at, updated_at
FROM dialog_suggestion_jobs
WHERE status = $status
ORDER BY created_at
LIMIT $limit;`)

	params := table.NewQueryParameters(
		table.ValueParam("$status", types.UTF8Value(status)),
		table.ValueParam("$limit", types.Int64Value(int64(limit))),
	)

	var out []models.DialogSuggestionJob
	err := s.client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
		_, res, err := sess.Execute(ctx, table.OnlineReadOnlyTxControl(), query, params)
		if err != nil {
			return err
		}
		defer res.Close()

		if err := res.NextResultSetErr(ctx); err != nil {
			return err
		}
		for res.NextRow() {
			var (
				id        string
				userID    string
				chatID    string
				messageID string
				statusVal string
				attempts  int64
				lastError *string
				created   int64
				updated   *int64
			)
			if err := res.ScanNamed(
				named.Required("job_id", &id),
				named.Required("user_id", &userID),
				named.Required("chat_id", &chatID),
				named.Required("message_id", &messageID),
				named.Required("status", &statusVal),
				named.Required("attempts", &attempts),
				named.Optional("last_error", &lastError),
				named.Required("created_at", &created),
				named.Optional("updated_at", &updated),
			); err != nil {
				return err
			}
			job := models.DialogSuggestionJob{
				ID:        id,
				UserID:    userID,
				ChatID:    chatID,
				MessageID: messageID,
				Status:    statusVal,
				Attempts:  int(attempts),
				Created:   created,
			}
			if lastError != nil {
				job.LastError = lastError
			}
			if updated != nil {
				job.UpdatedAt = *updated
			} else {
				job.UpdatedAt = created
			}
			out = append(out, job)
		}
		return res.Err()
	}, table.WithIdempotent())
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Store) UpsertDialogSuggestionJob(ctx context.Context, job models.DialogSuggestionJob) error {
	now := time.Now().UnixMilli()
	if job.Created == 0 {
		job.Created = now
	}
	if job.UpdatedAt == 0 {
		job.UpdatedAt = now
	}

	query := s.withPrefix(`
DECLARE $job_id AS Utf8;
DECLARE $user_id AS Utf8;
DECLARE $chat_id AS Utf8;
DECLARE $message_id AS Utf8;
DECLARE $status AS Utf8;
DECLARE $attempts AS Int64;
DECLARE $last_error AS Utf8?;
DECLARE $created_at AS Int64;
DECLARE $updated_at AS Int64;
UPSERT INTO dialog_suggestion_jobs (job_id, user_id, chat_id, message_id, status, attempts, last_error, created_at, updated_at)
VALUES ($job_id, $user_id, $chat_id, $message_id, $status, $attempts, $last_error, $created_at, $updated_at);`)

	params := table.NewQueryParameters(
		table.ValueParam("$job_id", types.UTF8Value(job.ID)),
		table.ValueParam("$user_id", types.UTF8Value(job.UserID)),
		table.ValueParam("$chat_id", types.UTF8Value(job.ChatID)),
		table.ValueParam("$message_id", types.UTF8Value(job.MessageID)),
		table.ValueParam("$status", types.UTF8Value(job.Status)),
		table.ValueParam("$attempts", types.Int64Value(int64(job.Attempts))),
		table.ValueParam("$last_error", optionalStringPtr(job.LastError)),
		table.ValueParam("$created_at", types.Int64Value(job.Created)),
		table.ValueParam("$updated_at", types.Int64Value(job.UpdatedAt)),
	)

	return s.execWrite(ctx, query, params)
}

func (s *Store) UpdateDialogSuggestionJob(ctx context.Context, job models.DialogSuggestionJob) error {
	job.UpdatedAt = time.Now().UnixMilli()

	query := s.withPrefix(`
DECLARE $job_id AS Utf8;
DECLARE $status AS Utf8;
DECLARE $attempts AS Int64;
DECLARE $last_error AS Utf8?;
DECLARE $updated_at AS Int64;
UPDATE dialog_suggestion_jobs
SET status = $status, attempts = $attempts, last_error = $last_error, updated_at = $updated_at
WHERE job_id = $job_id;`)

	params := table.NewQueryParameters(
		table.ValueParam("$job_id", types.UTF8Value(job.ID)),
		table.ValueParam("$status", types.UTF8Value(job.Status)),
		table.ValueParam("$attempts", types.Int64Value(int64(job.Attempts))),
		table.ValueParam("$last_error", optionalStringPtr(job.LastError)),
		table.ValueParam("$updated_at", types.Int64Value(job.UpdatedAt)),
	)

	return s.execWrite(ctx, query, params)
}
