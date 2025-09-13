package db

import (
	"linka.type-backend/db/repositories"
)

// StatementCRUD provides CRUD operations for Statement entity
// This is a wrapper around StatementRepository for backward compatibility
type StatementCRUD struct {
	repo *repositories.StatementRepository
}

// NewStatementCRUD creates a new StatementCRUD
func NewStatementCRUD() *StatementCRUD {
	return &StatementCRUD{
		repo: repositories.NewStatementRepository(),
	}
}

// CreateStatement creates a new statement
func (s *StatementCRUD) CreateStatement(statement *Statement) error {
	return s.repo.CreateStatement(statement)
}

// GetStatementByID retrieves a statement by ID
func (s *StatementCRUD) GetStatementByID(id string) (*Statement, error) {
	return s.repo.GetStatementByID(id)
}

// GetStatementsByUserID retrieves all statements for a specific user
func (s *StatementCRUD) GetStatementsByUserID(userID string) ([]*Statement, error) {
	return s.repo.GetStatementsByUserID(userID)
}

// GetStatementsByCategoryID retrieves all statements for a specific category
func (s *StatementCRUD) GetStatementsByCategoryID(categoryID string) ([]*Statement, error) {
	return s.repo.GetStatementsByCategoryID(categoryID)
}

// GetAllStatements retrieves all statements
func (s *StatementCRUD) GetAllStatements() ([]*Statement, error) {
	return s.repo.GetAllStatements()
}

// UpdateStatement updates an existing statement
func (s *StatementCRUD) UpdateStatement(statement *Statement) error {
	return s.repo.UpdateStatement(statement)
}

// DeleteStatement deletes a statement by ID
func (s *StatementCRUD) DeleteStatement(id string) error {
	return s.repo.DeleteStatement(id)
}

// DeleteStatementsByUserID deletes all statements for a specific user
func (s *StatementCRUD) DeleteStatementsByUserID(userID string) error {
	return s.repo.DeleteStatementsByUserID(userID)
}

// DeleteStatementsByCategoryID deletes all statements for a specific category
func (s *StatementCRUD) DeleteStatementsByCategoryID(categoryID string) error {
	return s.repo.DeleteStatementsByCategoryID(categoryID)
}

// StatementExists checks if a statement exists by ID
func (s *StatementCRUD) StatementExists(id string) (bool, error) {
	return s.repo.StatementExists(id)
}