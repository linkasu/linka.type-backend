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
SELECT category_id, label, created_at, is_default, updated_at
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
			cat := models.Category{
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
UPSERT INTO categories (user_id, category_id, label, created_at, is_default, updated_at)
VALUES ($user_id, $category_id, $label, $created_at, $is_default, $updated_at);`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$category_id", types.UTF8Value(category.ID)),
		table.ValueParam("$label", types.UTF8Value(category.Label)),
		table.ValueParam("$created_at", types.Int64Value(category.Created)),
		table.ValueParam("$is_default", optionalBool(category.Default)),
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
SELECT inited
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
			var inited bool
			if err := res.ScanNamed(named.Required("inited", &inited)); err != nil {
				return err
			}
			state.Inited = inited
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

	query := s.withPrefix(`
DECLARE $user_id AS Utf8;
DECLARE $created_at AS Int64;
DECLARE $inited AS Bool;
DECLARE $deleted_at AS Int64?;
UPSERT INTO users (user_id, created_at, inited, deleted_at)
VALUES ($user_id, $created_at, $inited, $deleted_at);`)

	params := table.NewQueryParameters(
		table.ValueParam("$user_id", types.UTF8Value(userID)),
		table.ValueParam("$created_at", types.Int64Value(createdAt)),
		table.ValueParam("$inited", types.BoolValue(state.Inited)),
		table.ValueParam("$deleted_at", types.OptionalValue(types.NullValue(types.TypeInt64))),
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
		return types.OptionalValue(types.NullValue(types.TypeBool))
	}
	return types.OptionalValue(types.BoolValue(*val))
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
