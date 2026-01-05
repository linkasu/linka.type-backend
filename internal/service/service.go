package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/linkasu/linka.type-backend/internal/config"
	"github.com/linkasu/linka.type-backend/internal/defaults"
	"github.com/linkasu/linka.type-backend/internal/feature"
	"github.com/linkasu/linka.type-backend/internal/id"
	"github.com/linkasu/linka.type-backend/internal/models"
	"github.com/linkasu/linka.type-backend/internal/store"
)

// Service coordinates YDB and Firebase for dual-write and read-through.
type Service struct {
	Store        store.Store
	LegacyWriter store.LegacyWriter
	LegacyReader store.LegacyReader
	Feature      config.FeatureConfig
}

// CategoryInput captures category creation payload.
type CategoryInput struct {
	ID      string
	Label   string
	Created int64
	Default *bool
}

// CategoryPatch captures category updates.
type CategoryPatch struct {
	Label   *string
	Default *bool
}

// StatementInput captures statement creation payload.
type StatementInput struct {
	ID         string
	CategoryID string
	Text       string
	Created    int64
}

// StatementPatch captures statement updates.
type StatementPatch struct {
	Text *string
}

// UserStatePatch captures user state updates.
type UserStatePatch struct {
	Inited     *bool
	Quickes    []string
	QuickesSet bool
}

// QuestionInput captures onboarding question payloads.
type QuestionInput struct {
	UID        string   `json:"uid"`
	QuestionID string   `json:"question_id"`
	Value      string   `json:"value"`
	Label      string   `json:"label"`
	Phrases    []string `json:"phrases"`
	Category   string   `json:"category"`
	Type       string   `json:"type"`
}

// GlobalCategoryInput captures global category creation payload.
type GlobalCategoryInput struct {
	ID      string
	Label   string
	Created int64
	Default *bool
}

// GlobalCategoryPatch captures global category updates.
type GlobalCategoryPatch struct {
	Label   *string
	Default *bool
}

// FactoryQuestionInput captures factory question creation payload.
type FactoryQuestionInput struct {
	ID         string
	Label      string
	Phrases    []string
	Category   string
	Type       string
	OrderIndex int
}

// FactoryQuestionPatch captures factory question updates.
type FactoryQuestionPatch struct {
	Label      *string
	Phrases    []string
	Category   *string
	Type       *string
	OrderIndex *int
}

// ListCategories returns categories using the configured read source.
func (s *Service) ListCategories(ctx context.Context, userID string) ([]models.Category, error) {
	if s.useYDB(userID) {
		categories, err := s.Store.ListCategories(ctx, userID)
		if err != nil {
			return nil, err
		}
		if len(categories) == 0 && s.LegacyReader != nil {
			legacyCategories, legacyStatements, err := s.LegacyReader.FetchUserData(ctx, userID)
			if err != nil {
				return nil, err
			}
			if len(legacyCategories) > 0 || len(legacyStatements) > 0 {
				_ = s.seedUserData(ctx, userID, legacyCategories, legacyStatements)
				return legacyCategories, nil
			}
		}
		return categories, nil
	}

	if s.LegacyReader != nil {
		categories, _, err := s.LegacyReader.FetchUserData(ctx, userID)
		if err != nil {
			return nil, err
		}
		return categories, nil
	}

	return s.Store.ListCategories(ctx, userID)
}

// CreateCategory creates a category and mirrors it to Firebase.
func (s *Service) CreateCategory(ctx context.Context, userID string, input CategoryInput) (models.Category, error) {
	now := time.Now().UnixMilli()
	if input.ID == "" {
		input.ID = id.NewShort()
	}
	if input.Created == 0 {
		input.Created = now
	}
	category := models.Category{
		ID:        input.ID,
		Label:     input.Label,
		Created:   input.Created,
		Default:   input.Default,
		UpdatedAt: now,
	}

	category, err := s.Store.UpsertCategory(ctx, userID, category)
	if err != nil {
		return models.Category{}, err
	}

	if s.LegacyWriter != nil {
		if err := s.LegacyWriter.UpsertCategory(ctx, userID, category); err != nil {
			return models.Category{}, err
		}
	}

	_ = s.appendChange(ctx, userID, "category", category.ID, "upsert", category, now)

	return category, nil
}

// UpdateCategory patches a category.
func (s *Service) UpdateCategory(ctx context.Context, userID, categoryID string, patch CategoryPatch) (models.Category, error) {
	category, err := s.findCategory(ctx, userID, categoryID)
	if err != nil {
		return models.Category{}, err
	}

	if patch.Label != nil {
		category.Label = *patch.Label
	}
	if patch.Default != nil {
		category.Default = patch.Default
	}
	category.UpdatedAt = time.Now().UnixMilli()

	category, err = s.Store.UpsertCategory(ctx, userID, category)
	if err != nil {
		return models.Category{}, err
	}
	if s.LegacyWriter != nil {
		if err := s.LegacyWriter.UpsertCategory(ctx, userID, category); err != nil {
			return models.Category{}, err
		}
	}

	_ = s.appendChange(ctx, userID, "category", category.ID, "upsert", category, category.UpdatedAt)

	return category, nil
}

// DeleteCategory deletes a category and its statements.
func (s *Service) DeleteCategory(ctx context.Context, userID, categoryID string) error {
	updatedAt := time.Now().UnixMilli()

	statements, _ := s.Store.ListStatements(ctx, userID, categoryID)
	if err := s.Store.DeleteCategory(ctx, userID, categoryID, updatedAt); err != nil {
		return err
	}
	for _, stmt := range statements {
		_ = s.Store.DeleteStatement(ctx, userID, categoryID, stmt.ID, updatedAt)
	}

	if s.LegacyWriter != nil {
		if err := s.LegacyWriter.DeleteCategory(ctx, userID, categoryID); err != nil {
			return err
		}
	}

	_ = s.appendChange(ctx, userID, "category", categoryID, "delete", map[string]string{"id": categoryID}, updatedAt)

	return nil
}

// ListStatements returns statements for a category.
func (s *Service) ListStatements(ctx context.Context, userID, categoryID string) ([]models.Statement, error) {
	if s.useYDB(userID) {
		statements, err := s.Store.ListStatements(ctx, userID, categoryID)
		if err != nil {
			return nil, err
		}
		if len(statements) == 0 && s.LegacyReader != nil {
			legacyCategories, legacyStatements, err := s.LegacyReader.FetchUserData(ctx, userID)
			if err != nil {
				return nil, err
			}
			if len(legacyCategories) > 0 || len(legacyStatements) > 0 {
				_ = s.seedUserData(ctx, userID, legacyCategories, legacyStatements)
				return filterStatements(legacyStatements, categoryID), nil
			}
		}
		return statements, nil
	}

	if s.LegacyReader != nil {
		_, statements, err := s.LegacyReader.FetchUserData(ctx, userID)
		if err != nil {
			return nil, err
		}
		return filterStatements(statements, categoryID), nil
	}

	return s.Store.ListStatements(ctx, userID, categoryID)
}

// CreateStatement creates a statement or runs onboarding generation.
func (s *Service) CreateStatement(ctx context.Context, userID string, input StatementInput) (models.Statement, error) {
	now := time.Now().UnixMilli()
	if input.ID == "" {
		input.ID = id.NewShort()
	}
	if input.Created == 0 {
		input.Created = now
	}

	statement := models.Statement{
		ID:         input.ID,
		CategoryID: input.CategoryID,
		Text:       input.Text,
		Created:    input.Created,
		UpdatedAt:  now,
	}

	statement, err := s.Store.UpsertStatement(ctx, userID, statement)
	if err != nil {
		return models.Statement{}, err
	}
	if s.LegacyWriter != nil {
		if err := s.LegacyWriter.UpsertStatement(ctx, userID, statement); err != nil {
			return models.Statement{}, err
		}
	}

	_ = s.appendChange(ctx, userID, "statement", statement.ID, "upsert", statement, now)

	return statement, nil
}

// UpdateStatement patches a statement.
func (s *Service) UpdateStatement(ctx context.Context, userID, statementID string, patch StatementPatch) (models.Statement, error) {
	statement, err := s.findStatement(ctx, userID, statementID)
	if err != nil {
		return models.Statement{}, err
	}
	if patch.Text != nil {
		statement.Text = *patch.Text
	}
	statement.UpdatedAt = time.Now().UnixMilli()

	statement, err = s.Store.UpsertStatement(ctx, userID, statement)
	if err != nil {
		return models.Statement{}, err
	}
	if s.LegacyWriter != nil {
		if err := s.LegacyWriter.UpsertStatement(ctx, userID, statement); err != nil {
			return models.Statement{}, err
		}
	}

	_ = s.appendChange(ctx, userID, "statement", statement.ID, "upsert", statement, statement.UpdatedAt)

	return statement, nil
}

// DeleteStatement deletes a statement by ID.
func (s *Service) DeleteStatement(ctx context.Context, userID, statementID string) error {
	statement, err := s.findStatement(ctx, userID, statementID)
	if err != nil {
		return err
	}
	updatedAt := time.Now().UnixMilli()

	if err := s.Store.DeleteStatement(ctx, userID, statement.CategoryID, statementID, updatedAt); err != nil {
		return err
	}
	if s.LegacyWriter != nil {
		if err := s.LegacyWriter.DeleteStatement(ctx, userID, statement.CategoryID, statementID); err != nil {
			return err
		}
	}

	_ = s.appendChange(ctx, userID, "statement", statementID, "delete", map[string]string{"id": statementID}, updatedAt)

	return nil
}

// GetUserState returns inited and quickes with default fallbacks.
func (s *Service) GetUserState(ctx context.Context, userID string) (models.UserState, error) {
	state := models.UserState{}
	var err error

	if s.useYDB(userID) {
		state, err = s.Store.GetUserState(ctx, userID)
		if err != nil {
			return state, err
		}
		if s.LegacyReader != nil && len(state.Quickes) == 0 {
			legacyState, err := s.LegacyReader.GetUserState(ctx, userID)
			if err != nil {
				return state, err
			}
			if legacyState.Inited || len(legacyState.Quickes) > 0 {
				_ = s.seedUserState(ctx, userID, legacyState)
				state = legacyState
			}
		}
	} else if s.LegacyReader != nil {
		state, err = s.LegacyReader.GetUserState(ctx, userID)
		if err != nil {
			return state, err
		}
	} else {
		state, err = s.Store.GetUserState(ctx, userID)
		if err != nil {
			return state, err
		}
	}

	state.Quickes = normalizeQuickes(state.Quickes)

	return state, nil
}

// UpdateUserState updates inited/quickes and mirrors the result.
func (s *Service) UpdateUserState(ctx context.Context, userID string, patch UserStatePatch) (models.UserState, error) {
	current, err := s.GetUserState(ctx, userID)
	if err != nil {
		return current, err
	}

	if patch.Inited != nil {
		current.Inited = *patch.Inited
	}
	if patch.QuickesSet {
		current.Quickes = normalizeQuickes(patch.Quickes)
	}

	updatedAt := time.Now().UnixMilli()
	updated, err := s.Store.SetUserState(ctx, userID, current, updatedAt)
	if err != nil {
		return current, err
	}

	if s.LegacyWriter != nil {
		if err := s.LegacyWriter.SetUserState(ctx, userID, updated); err != nil {
			return updated, err
		}
	}

	_ = s.appendChange(ctx, userID, "user_state", userID, "upsert", updated, updatedAt)

	updated.Quickes = normalizeQuickes(updated.Quickes)
	return updated, nil
}

// SetQuickes updates quick phrases.
func (s *Service) SetQuickes(ctx context.Context, userID string, quickes []string) ([]string, error) {
	updatedAt := time.Now().UnixMilli()
	quickes = normalizeQuickes(quickes)

	updated, err := s.Store.SetQuickes(ctx, userID, quickes, updatedAt)
	if err != nil {
		return nil, err
	}
	if s.LegacyWriter != nil {
		if err := s.LegacyWriter.SetQuickes(ctx, userID, updated); err != nil {
			return nil, err
		}
	}

	_ = s.appendChange(ctx, userID, "quickes", userID, "upsert", updated, updatedAt)

	return updated, nil
}

// ListGlobalCategories returns global categories (optionally seeded from Firebase).
func (s *Service) ListGlobalCategories(ctx context.Context, includeStatements bool) ([]models.GlobalCategory, error) {
	categories, err := s.Store.ListGlobalCategories(ctx, includeStatements)
	if err != nil {
		return nil, err
	}
	if len(categories) == 0 && s.LegacyReader != nil {
		legacyCats, err := s.LegacyReader.ListGlobalCategories(ctx)
		if err != nil {
			return nil, err
		}
		if !includeStatements {
			for i := range legacyCats {
				legacyCats[i].Statements = nil
			}
		}
		return legacyCats, nil
	}
	return categories, nil
}

// ListGlobalStatements returns global statements for a category.
func (s *Service) ListGlobalStatements(ctx context.Context, categoryID string) ([]models.Statement, error) {
	return s.Store.ListGlobalStatements(ctx, categoryID)
}

// ImportGlobalCategory mirrors global category into user data.
func (s *Service) ImportGlobalCategory(ctx context.Context, userID, categoryID string, force bool) (string, error) {
	status, err := s.Store.ImportGlobalCategory(ctx, userID, categoryID, force)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) && s.LegacyReader != nil {
			return s.importGlobalFromLegacy(ctx, userID, categoryID, force)
		}
		return "", err
	}
	if status == "exists" {
		return status, nil
	}
	if s.LegacyWriter != nil {
		if err := s.LegacyWriter.ImportGlobalCategory(ctx, userID, categoryID); err != nil {
			return "", err
		}
	}

	return status, nil
}

func (s *Service) importGlobalFromLegacy(ctx context.Context, userID, categoryID string, force bool) (string, error) {
	if !force {
		if _, err := s.findCategory(ctx, userID, categoryID); err == nil {
			return "exists", nil
		} else if !errors.Is(err, store.ErrNotFound) {
			return "", err
		}
	}

	globals, err := s.LegacyReader.ListGlobalCategories(ctx)
	if err != nil {
		return "", err
	}
	var global *models.GlobalCategory
	for i := range globals {
		if globals[i].ID == categoryID {
			global = &globals[i]
			break
		}
	}
	if global == nil {
		return "", store.ErrNotFound
	}

	updatedAt := time.Now().UnixMilli()
	cat := models.Category{
		ID:        global.ID,
		Label:     global.Label,
		Created:   global.Created,
		Default:   global.Default,
		UpdatedAt: updatedAt,
	}
	if _, err := s.Store.UpsertCategory(ctx, userID, cat); err != nil {
		return "", err
	}

	for _, stmt := range global.Statements {
		stmt.CategoryID = global.ID
		stmt.UpdatedAt = updatedAt
		if _, err := s.Store.UpsertStatement(ctx, userID, stmt); err != nil {
			return "", err
		}
	}

	if s.LegacyWriter != nil {
		if err := s.LegacyWriter.ImportGlobalCategory(ctx, userID, categoryID); err != nil {
			return "", err
		}
	}
	return "ok", nil
}

// ListFactoryQuestions returns onboarding question templates.
func (s *Service) ListFactoryQuestions(ctx context.Context) ([]models.FactoryQuestion, error) {
	questions, err := s.Store.ListFactoryQuestions(ctx)
	if err != nil {
		return nil, err
	}
	if len(questions) == 0 && s.LegacyReader != nil {
		legacy, err := s.LegacyReader.ListFactoryQuestions(ctx)
		if err != nil {
			return nil, err
		}
		return legacy, nil
	}
	return questions, nil
}

// IsAdmin checks admin list.
func (s *Service) IsAdmin(ctx context.Context, userID string) (bool, error) {
	isAdmin, err := s.Store.IsAdmin(ctx, userID)
	if err != nil {
		return false, err
	}
	if !isAdmin && s.LegacyReader != nil {
		return s.LegacyReader.IsAdmin(ctx, userID)
	}
	return isAdmin, nil
}

// DeleteUser deletes YDB data and optionally Firebase RTDB data.
func (s *Service) DeleteUser(ctx context.Context, userID string, deleteFirebase bool) error {
	updatedAt := time.Now().UnixMilli()
	if err := s.Store.DeleteUser(ctx, userID, updatedAt); err != nil {
		return err
	}
	if deleteFirebase && s.LegacyWriter != nil {
		return s.LegacyWriter.DeleteUserData(ctx, userID)
	}
	return nil
}

// OnboardingPhrases generates initial phrases based on questions.
func (s *Service) OnboardingPhrases(ctx context.Context, userID string, questions []QuestionInput) (string, error) {
	state, err := s.GetUserState(ctx, userID)
	if err != nil {
		return "", err
	}
	if state.Inited {
		return "ok", nil
	}

	factory, err := s.ListFactoryQuestions(ctx)
	if err != nil {
		return "", err
	}
	factoryByID := make(map[string]models.FactoryQuestion, len(factory))
	for _, q := range factory {
		factoryByID[q.ID] = q
	}

	categories, err := s.ListCategories(ctx, userID)
	if err != nil {
		return "", err
	}
	categoryByLabel := make(map[string]models.Category, len(categories))
	for _, cat := range categories {
		categoryByLabel[cat.Label] = cat
	}

	for _, q := range questions {
		value := strings.TrimSpace(q.Value)
		if value == "" {
			continue
		}

		phrases := q.Phrases
		categoryName := q.Category
		if phrases == nil || categoryName == "" {
			if q.QuestionID != "" {
				if tmpl, ok := factoryByID[q.QuestionID]; ok {
					if phrases == nil {
						phrases = tmpl.Phrases
					}
					if categoryName == "" {
						categoryName = tmpl.Category
					}
				}
			}
			if q.UID != "" {
				if tmpl, ok := factoryByID[q.UID]; ok {
					if phrases == nil {
						phrases = tmpl.Phrases
					}
					if categoryName == "" {
						categoryName = tmpl.Category
					}
				}
			}
		}

		if len(phrases) == 0 || categoryName == "" {
			continue
		}

		catID, err := s.getOrCreateCategory(ctx, userID, categoryName, categoryByLabel)
		if err != nil {
			return "", err
		}

		for _, phrase := range phrases {
			text := strings.ReplaceAll(phrase, "%%", value)
			if strings.TrimSpace(text) == "" {
				continue
			}
			_, err := s.CreateStatement(ctx, userID, StatementInput{
				CategoryID: catID,
				Text:       text,
			})
			if err != nil {
				return "", err
			}
		}
	}

	patch := UserStatePatch{Inited: boolPtr(true)}
	_, err = s.UpdateUserState(ctx, userID, patch)
	if err != nil {
		return "", err
	}

	return "ok", nil
}

func (s *Service) useYDB(userID string) bool {
	return feature.UseYDB(userID, s.Feature)
}

func (s *Service) appendChange(ctx context.Context, userID, entityType, entityID, op string, payload any, updatedAt int64) error {
	if s.Store == nil {
		return nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.Store.AppendChange(ctx, userID, models.ChangeEvent{
		Cursor:     id.New(),
		EntityType: entityType,
		EntityID:   entityID,
		Op:         op,
		Payload:    data,
		UpdatedAt:  updatedAt,
	})
}

func (s *Service) seedUserData(ctx context.Context, userID string, categories []models.Category, statements []models.Statement) error {
	for _, cat := range categories {
		if cat.Created == 0 {
			cat.Created = time.Now().UnixMilli()
		}
		if cat.UpdatedAt == 0 {
			cat.UpdatedAt = cat.Created
		}
		if _, err := s.Store.UpsertCategory(ctx, userID, cat); err != nil {
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
		if _, err := s.Store.UpsertStatement(ctx, userID, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) seedUserState(ctx context.Context, userID string, state models.UserState) error {
	state.Quickes = normalizeQuickes(state.Quickes)
	_, err := s.Store.SetUserState(ctx, userID, state, time.Now().UnixMilli())
	return err
}

func (s *Service) findCategory(ctx context.Context, userID, categoryID string) (models.Category, error) {
	categories, err := s.ListCategories(ctx, userID)
	if err != nil {
		return models.Category{}, err
	}
	for _, cat := range categories {
		if cat.ID == categoryID {
			return cat, nil
		}
	}
	return models.Category{}, store.ErrNotFound
}

func (s *Service) findStatement(ctx context.Context, userID, statementID string) (models.Statement, error) {
	if s.LegacyReader != nil && !s.useYDB(userID) {
		_, statements, err := s.LegacyReader.FetchUserData(ctx, userID)
		if err != nil {
			return models.Statement{}, err
		}
		for _, stmt := range statements {
			if stmt.ID == statementID {
				return stmt, nil
			}
		}
		return models.Statement{}, store.ErrNotFound
	}

	categories, err := s.Store.ListCategories(ctx, userID)
	if err != nil {
		return models.Statement{}, err
	}
	if len(categories) == 0 && s.LegacyReader != nil {
		legacyCategories, statements, err := s.LegacyReader.FetchUserData(ctx, userID)
		if err != nil {
			return models.Statement{}, err
		}
		for _, stmt := range statements {
			if stmt.ID == statementID {
				_ = s.seedUserData(ctx, userID, legacyCategories, statements)
				return stmt, nil
			}
		}
		return models.Statement{}, store.ErrNotFound
	}

	for _, cat := range categories {
		statements, err := s.Store.ListStatements(ctx, userID, cat.ID)
		if err != nil {
			return models.Statement{}, err
		}
		for _, stmt := range statements {
			if stmt.ID == statementID {
				return stmt, nil
			}
		}
	}

	return models.Statement{}, store.ErrNotFound
}

func (s *Service) getOrCreateCategory(ctx context.Context, userID, label string, cache map[string]models.Category) (string, error) {
	if cat, ok := cache[label]; ok {
		return cat.ID, nil
	}

	isDefault := len(cache) == 0
	category, err := s.CreateCategory(ctx, userID, CategoryInput{
		Label:   label,
		Default: boolPtr(isDefault),
	})
	if err != nil {
		return "", err
	}
	cache[label] = category
	return category.ID, nil
}

func normalizeQuickes(quickes []string) []string {
	defaultsList := defaults.DefaultQuickes
	out := make([]string, len(defaultsList))
	for i := range defaultsList {
		if i < len(quickes) && strings.TrimSpace(quickes[i]) != "" {
			out[i] = quickes[i]
		} else {
			out[i] = defaultsList[i]
		}
	}
	return out
}

func filterStatements(statements []models.Statement, categoryID string) []models.Statement {
	filtered := make([]models.Statement, 0)
	for _, stmt := range statements {
		if stmt.CategoryID == categoryID {
			filtered = append(filtered, stmt)
		}
	}
	return filtered
}

func boolPtr(val bool) *bool {
	return &val
}

// AdminStats returns statistics for admin panel.
func (s *Service) AdminStats(ctx context.Context, since time.Time) (map[string]interface{}, error) {
	users, err := s.Store.CountUsers(ctx, since)
	if err != nil {
		return nil, err
	}
	categories, err := s.Store.CountCategories(ctx, since)
	if err != nil {
		return nil, err
	}
	statements, err := s.Store.CountStatements(ctx, since)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"total_users":     users,
		"total_categories": categories,
		"total_statements": statements,
		"since":           since.Format(time.RFC3339),
		"window_seconds":  int(time.Since(since).Seconds()),
	}, nil
}

// ListAdmins returns all admin user IDs.
func (s *Service) ListAdmins(ctx context.Context) ([]string, error) {
	return s.Store.ListAdmins(ctx)
}

// AddAdmin adds a user to the admin list.
func (s *Service) AddAdmin(ctx context.Context, userID string) error {
	return s.Store.AddAdmin(ctx, userID)
}

// RemoveAdmin removes a user from the admin list.
func (s *Service) RemoveAdmin(ctx context.Context, userID string) error {
	return s.Store.RemoveAdmin(ctx, userID)
}

// CreateClientKey creates a new client API key.
func (s *Service) CreateClientKey(ctx context.Context, key store.ClientKey) error {
	return s.Store.CreateClientKey(ctx, key)
}

// ListClientKeys returns all client keys.
func (s *Service) ListClientKeys(ctx context.Context) ([]store.ClientKey, error) {
	return s.Store.ListClientKeys(ctx)
}

// RevokeClientKey revokes a client key.
func (s *Service) RevokeClientKey(ctx context.Context, keyHash string) error {
	return s.Store.RevokeClientKey(ctx, keyHash)
}

// CreateGlobalCategory creates a global category.
func (s *Service) CreateGlobalCategory(ctx context.Context, input GlobalCategoryInput) (models.GlobalCategory, error) {
	now := time.Now().UnixMilli()
	if input.ID == "" {
		input.ID = id.NewShort()
	}
	if input.Created == 0 {
		input.Created = now
	}
	category := models.GlobalCategory{
		ID:        input.ID,
		Label:     input.Label,
		Created:   input.Created,
		Default:   input.Default,
		UpdatedAt: now,
	}
	return s.Store.UpsertGlobalCategory(ctx, category)
}

// UpdateGlobalCategory updates a global category.
func (s *Service) UpdateGlobalCategory(ctx context.Context, categoryID string, patch GlobalCategoryPatch) (models.GlobalCategory, error) {
	categories, err := s.Store.ListGlobalCategories(ctx, false)
	if err != nil {
		return models.GlobalCategory{}, err
	}
	var category *models.GlobalCategory
	for i := range categories {
		if categories[i].ID == categoryID {
			category = &categories[i]
			break
		}
	}
	if category == nil {
		return models.GlobalCategory{}, store.ErrNotFound
	}

	if patch.Label != nil {
		category.Label = *patch.Label
	}
	if patch.Default != nil {
		category.Default = patch.Default
	}
	category.UpdatedAt = time.Now().UnixMilli()

	return s.Store.UpsertGlobalCategory(ctx, *category)
}

// DeleteGlobalCategory deletes a global category.
func (s *Service) DeleteGlobalCategory(ctx context.Context, categoryID string) error {
	updatedAt := time.Now().UnixMilli()
	return s.Store.DeleteGlobalCategory(ctx, categoryID, updatedAt)
}

// CreateFactoryQuestion creates a factory question.
func (s *Service) CreateFactoryQuestion(ctx context.Context, input FactoryQuestionInput) (models.FactoryQuestion, error) {
	if input.ID == "" {
		input.ID = id.NewShort()
	}
	question := models.FactoryQuestion{
		ID:         input.ID,
		Label:      input.Label,
		Phrases:    input.Phrases,
		Category:   input.Category,
		Type:       input.Type,
		OrderIndex: input.OrderIndex,
	}
	return s.Store.UpsertFactoryQuestion(ctx, question)
}

// UpdateFactoryQuestion updates a factory question.
func (s *Service) UpdateFactoryQuestion(ctx context.Context, questionID string, patch FactoryQuestionPatch) (models.FactoryQuestion, error) {
	questions, err := s.Store.ListFactoryQuestions(ctx)
	if err != nil {
		return models.FactoryQuestion{}, err
	}
	var question *models.FactoryQuestion
	for i := range questions {
		if questions[i].ID == questionID {
			question = &questions[i]
			break
		}
	}
	if question == nil {
		return models.FactoryQuestion{}, store.ErrNotFound
	}

	if patch.Label != nil {
		question.Label = *patch.Label
	}
	if patch.Phrases != nil {
		question.Phrases = patch.Phrases
	}
	if patch.Category != nil {
		question.Category = *patch.Category
	}
	if patch.Type != nil {
		question.Type = *patch.Type
	}
	if patch.OrderIndex != nil {
		question.OrderIndex = *patch.OrderIndex
	}

	return s.Store.UpsertFactoryQuestion(ctx, *question)
}

// DeleteFactoryQuestion deletes a factory question.
func (s *Service) DeleteFactoryQuestion(ctx context.Context, questionID string) error {
	return s.Store.DeleteFactoryQuestion(ctx, questionID)
}
