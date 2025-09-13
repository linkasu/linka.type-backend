package repositories

import (
	"linka.type-backend/models"
)

// StatementCRUD provides CRUD operations for Statement entity
// This is a wrapper around StatementRepository for backward compatibility
type StatementCRUD struct {
	repo *models.StatementRepository
}

// NewStatementCRUD creates a new StatementCRUD
func NewStatementCRUD() *models.StatementCRUD {
	return &StatementCRUD{
		repo: NewStatementRepository(),
	}
}

// CreateStatement creates a new statement
func (s *models.StatementCRUD) CreateStatement(statement *models.Statement) error {
	return s.repo.CreateStatement(statement)
}

// GetStatementByID retrieves a statement by ID
func (s *models.StatementCRUD) GetStatementByID(id string) (*models.Statement, error) {
	return s.repo.GetStatementByID(id)
}

// GetStatementsByUserID retrieves all statements for a specific user
func (s *models.StatementCRUD) GetStatementsByUserID(userID string) ([]*models.Statement, error) {
	return s.repo.GetStatementsByUserID(userID)
}

// GetStatementsByCategoryID retrieves all statements for a specific category
func (s *models.StatementCRUD) GetStatementsByCategoryID(categoryID string) ([]*models.Statement, error) {
	return s.repo.GetStatementsByCategoryID(categoryID)
}

// GetAllStatements retrieves all statements
func (s *models.StatementCRUD) GetAllStatements() ([]*models.Statement, error) {
	return s.repo.GetAllStatements()
}

// UpdateStatement updates an existing statement
func (s *models.StatementCRUD) UpdateStatement(statement *models.Statement) error {
	return s.repo.UpdateStatement(statement)
}

// DeleteStatement deletes a statement by ID
func (s *models.StatementCRUD) DeleteStatement(id string) error {
	return s.repo.DeleteStatement(id)
}

// DeleteStatementsByUserID deletes all statements for a specific user
func (s *models.StatementCRUD) DeleteStatementsByUserID(userID string) error {
	return s.repo.DeleteStatementsByUserID(userID)
}

// DeleteStatementsByCategoryID deletes all statements for a specific category
func (s *models.StatementCRUD) DeleteStatementsByCategoryID(categoryID string) error {
	return s.repo.DeleteStatementsByCategoryID(categoryID)
}

// StatementExists checks if a statement exists by ID
func (s *models.StatementCRUD) StatementExists(id string) (bool, error) {
	return s.repo.StatementExists(id)
}
