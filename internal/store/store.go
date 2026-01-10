package store

import (
	"context"
	"errors"
	"time"

	"github.com/linkasu/linka.type-backend/internal/models"
)

var ErrNotFound = errors.New("not found")

// ClientKey represents a client API key.
type ClientKey struct {
	KeyHash  string
	ClientID string
	Status   string
	CreatedAt int64
	RevokedAt *int64
}

// Store defines the core data operations backed by YDB.
type Store interface {
	ListCategories(ctx context.Context, userID string) ([]models.Category, error)
	UpsertCategory(ctx context.Context, userID string, category models.Category) (models.Category, error)
	DeleteCategory(ctx context.Context, userID, categoryID string, updatedAt int64) error

	ListStatements(ctx context.Context, userID, categoryID string) ([]models.Statement, error)
	ListAllStatements(ctx context.Context, userID string) ([]models.Statement, error)
	UpsertStatement(ctx context.Context, userID string, statement models.Statement) (models.Statement, error)
	DeleteStatement(ctx context.Context, userID, categoryID, statementID string, updatedAt int64) error

	GetUserState(ctx context.Context, userID string) (models.UserState, error)
	SetUserState(ctx context.Context, userID string, state models.UserState, updatedAt int64) (models.UserState, error)
	SetQuickes(ctx context.Context, userID string, quickes []string, updatedAt int64) ([]string, error)

	ListGlobalCategories(ctx context.Context, includeStatements bool) ([]models.GlobalCategory, error)
	ListGlobalStatements(ctx context.Context, categoryID string) ([]models.Statement, error)
	ImportGlobalCategory(ctx context.Context, userID, categoryID string, force bool) (string, error)
	UpsertGlobalCategory(ctx context.Context, category models.GlobalCategory) (models.GlobalCategory, error)
	DeleteGlobalCategory(ctx context.Context, categoryID string, updatedAt int64) error

	ListFactoryQuestions(ctx context.Context) ([]models.FactoryQuestion, error)
	UpsertFactoryQuestion(ctx context.Context, question models.FactoryQuestion) (models.FactoryQuestion, error)
	DeleteFactoryQuestion(ctx context.Context, questionID string) error

	IsAdmin(ctx context.Context, userID string) (bool, error)
	DeleteUser(ctx context.Context, userID string, updatedAt int64) error

	AppendChange(ctx context.Context, userID string, change models.ChangeEvent) error
	ListChanges(ctx context.Context, userID, cursor string, limit int) (nextCursor string, changes []models.ChangeEvent, err error)

	// Admin methods
	CountUsers(ctx context.Context, since time.Time) (int64, error)
	CountCategories(ctx context.Context, since time.Time) (int64, error)
	CountStatements(ctx context.Context, since time.Time) (int64, error)
	ListAdmins(ctx context.Context) ([]string, error)
	AddAdmin(ctx context.Context, userID string) error
	RemoveAdmin(ctx context.Context, userID string) error
	CreateClientKey(ctx context.Context, key ClientKey) error
	ListClientKeys(ctx context.Context) ([]ClientKey, error)
	RevokeClientKey(ctx context.Context, keyHash string) error

	// Dialog helper data
	ListDialogChats(ctx context.Context, userID string) ([]models.DialogChat, error)
	GetDialogChat(ctx context.Context, userID, chatID string) (models.DialogChat, error)
	UpsertDialogChat(ctx context.Context, userID string, chat models.DialogChat) (models.DialogChat, error)
	DeleteDialogChat(ctx context.Context, userID, chatID string, updatedAt int64) error

	ListDialogMessages(ctx context.Context, userID, chatID string, limit int, before int64) ([]models.DialogMessage, error)
	ListOldestDialogMessages(ctx context.Context, userID, chatID string, limit int) ([]models.DialogMessage, error)
	CountDialogMessages(ctx context.Context, userID, chatID string) (int64, error)
	UpsertDialogMessage(ctx context.Context, userID string, message models.DialogMessage) (models.DialogMessage, error)
	DeleteDialogMessage(ctx context.Context, userID, chatID, messageID string, updatedAt int64) error
	DeleteDialogMessagesByChat(ctx context.Context, userID, chatID string, updatedAt int64) error

	ListDialogSuggestions(ctx context.Context, userID string, status string, limit int) ([]models.DialogSuggestion, error)
	CountDialogSuggestions(ctx context.Context, userID string, status string) (int64, error)
	UpsertDialogSuggestion(ctx context.Context, userID string, suggestion models.DialogSuggestion) (models.DialogSuggestion, error)
	DeleteDialogSuggestion(ctx context.Context, userID, suggestionID string) error

	ListDialogSuggestionJobs(ctx context.Context, status string, limit int) ([]models.DialogSuggestionJob, error)
	UpsertDialogSuggestionJob(ctx context.Context, job models.DialogSuggestionJob) error
	UpdateDialogSuggestionJob(ctx context.Context, job models.DialogSuggestionJob) error
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
	DeleteUserData(ctx context.Context, userID string) error
}

// LegacyReader reads data from Firebase RTDB for seeding and sync.
type LegacyReader interface {
	FetchUserData(ctx context.Context, userID string) ([]models.Category, []models.Statement, error)
	GetUserState(ctx context.Context, userID string) (models.UserState, error)
	ListGlobalCategories(ctx context.Context) ([]models.GlobalCategory, error)
	ListFactoryQuestions(ctx context.Context) ([]models.FactoryQuestion, error)
	IsAdmin(ctx context.Context, userID string) (bool, error)
}
