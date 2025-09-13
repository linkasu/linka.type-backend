package services

import (
	"fmt"
	"strings"

	"linka.type-backend/db/repositories"
	"linka.type-backend/models"
	"linka.type-backend/utils"
)

// StatementService provides statement business logic
type StatementService struct {
	statementRepo *repositories.StatementRepository
	categoryRepo  *repositories.CategoryRepository
}

// NewStatementService creates a new StatementService
func NewStatementService() *StatementService {
	return &StatementService{
		statementRepo: repositories.NewStatementRepository(),
		categoryRepo:  repositories.NewCategoryRepository(),
	}
}

// CreateStatement creates a new statement
func (s *StatementService) CreateStatement(title, categoryID, userID string) (*models.Statement, error) {
	// Validate input
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}

	if strings.TrimSpace(categoryID) == "" {
		return nil, fmt.Errorf("category ID cannot be empty")
	}

	// Check if category exists and belongs to user
	category, err := s.categoryRepo.GetCategoryByID(categoryID)
	if err != nil {
		return nil, fmt.Errorf("category not found: %v", err)
	}

	if category.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}

	statement := &models.Statement{
		ID:         utils.GenerateID(),
		Title:      strings.TrimSpace(title),
		UserID:     userID,
		CategoryID: categoryID,
	}

	err = s.statementRepo.CreateStatement(statement)
	if err != nil {
		return nil, fmt.Errorf("failed to create statement: %v", err)
	}

	return statement, nil
}

// GetStatementByID gets a statement by ID
func (s *StatementService) GetStatementByID(id string) (*models.Statement, error) {
	return s.statementRepo.GetStatementByID(id)
}

// GetStatementsByUserID gets all statements for a user
func (s *StatementService) GetStatementsByUserID(userID string) ([]*models.Statement, error) {
	return s.statementRepo.GetStatementsByUserID(userID)
}

// GetStatementsByCategoryID gets all statements for a category
func (s *StatementService) GetStatementsByCategoryID(categoryID string) ([]*models.Statement, error) {
	return s.statementRepo.GetStatementsByCategoryID(categoryID)
}

// UpdateStatement updates a statement
func (s *StatementService) UpdateStatement(id, title, categoryID, userID string) (*models.Statement, error) {
	// Validate input
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}

	if strings.TrimSpace(categoryID) == "" {
		return nil, fmt.Errorf("category ID cannot be empty")
	}

	// Get existing statement
	existingStatement, err := s.statementRepo.GetStatementByID(id)
	if err != nil {
		return nil, fmt.Errorf("statement not found: %v", err)
	}

	// Check ownership
	if existingStatement.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}

	// Check if category exists and belongs to user
	category, err := s.categoryRepo.GetCategoryByID(categoryID)
	if err != nil {
		return nil, fmt.Errorf("category not found: %v", err)
	}

	if category.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}

	// Update statement
	statement := &models.Statement{
		ID:         id,
		Title:      strings.TrimSpace(title),
		UserID:     userID,
		CategoryID: categoryID,
		CreatedAt:  existingStatement.CreatedAt,
	}

	err = s.statementRepo.UpdateStatement(statement)
	if err != nil {
		return nil, fmt.Errorf("failed to update statement: %v", err)
	}

	return statement, nil
}

// DeleteStatement deletes a statement
func (s *StatementService) DeleteStatement(id, userID string) error {
	// Get existing statement
	existingStatement, err := s.statementRepo.GetStatementByID(id)
	if err != nil {
		return fmt.Errorf("statement not found: %v", err)
	}

	// Check ownership
	if existingStatement.UserID != userID {
		return fmt.Errorf("access denied")
	}

	return s.statementRepo.DeleteStatement(id)
}

// StatementExists checks if a statement exists
func (s *StatementService) StatementExists(id string) (bool, error) {
	return s.statementRepo.StatementExists(id)
}