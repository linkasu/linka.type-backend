package store

import (
	"context"
	"errors"

	"github.com/linkasu/linka.type-backend/internal/models"
)

var ErrNotFound = errors.New("not found")

// Store defines the core data operations backed by YDB.
type Store interface {
	ListCategories(ctx context.Context, userID string) ([]models.Category, error)
	UpsertCategory(ctx context.Context, userID string, category models.Category) (models.Category, error)
	DeleteCategory(ctx context.Context, userID, categoryID string, updatedAt int64) error

	ListStatements(ctx context.Context, userID, categoryID string) ([]models.Statement, error)
	UpsertStatement(ctx context.Context, userID string, statement models.Statement) (models.Statement, error)
	DeleteStatement(ctx context.Context, userID, categoryID, statementID string, updatedAt int64) error

	GetUserState(ctx context.Context, userID string) (models.UserState, error)
	SetUserState(ctx context.Context, userID string, state models.UserState, updatedAt int64) (models.UserState, error)
	SetQuickes(ctx context.Context, userID string, quickes []string, updatedAt int64) ([]string, error)

	ListGlobalCategories(ctx context.Context, includeStatements bool) ([]models.GlobalCategory, error)
	ListGlobalStatements(ctx context.Context, categoryID string) ([]models.Statement, error)
	ImportGlobalCategory(ctx context.Context, userID, categoryID string, force bool) (string, error)

	ListFactoryQuestions(ctx context.Context) ([]models.FactoryQuestion, error)

	IsAdmin(ctx context.Context, userID string) (bool, error)
	DeleteUser(ctx context.Context, userID string, updatedAt int64) error

	AppendChange(ctx context.Context, userID string, change models.ChangeEvent) error
	ListChanges(ctx context.Context, userID, cursor string, limit int) (nextCursor string, changes []models.ChangeEvent, err error)
}

// LegacyWriter mirrors writes to Firebase RTDB.
type LegacyWriter interface {
	UpsertCategory(ctx context.Context, userID string, category models.Category) error
	DeleteCategory(ctx context.Context, userID, categoryID string) error
	UpsertStatement(ctx context.Context, userID string, statement models.Statement) error
	DeleteStatement(ctx context.Context, userID, categoryID, statementID string) error
	SetUserState(ctx context.Context, userID string, state models.UserState) error
	SetQuickes(ctx context.Context, userID string, quickes []string) error
	ImportGlobalCategory(ctx context.Context, userID, categoryID string) error
}

// LegacyReader reads data from Firebase RTDB for seeding and sync.
type LegacyReader interface {
	FetchUserData(ctx context.Context, userID string) ([]models.Category, []models.Statement, error)
	GetUserState(ctx context.Context, userID string) (models.UserState, error)
	ListGlobalCategories(ctx context.Context) ([]models.GlobalCategory, error)
	ListFactoryQuestions(ctx context.Context) ([]models.FactoryQuestion, error)
	IsAdmin(ctx context.Context, userID string) (bool, error)
}
