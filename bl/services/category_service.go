package services

import (
	"fmt"
	"strings"

	"linka.type-backend/db/repositories"
	"linka.type-backend/models"
	"linka.type-backend/utils"
)

// CategoryService provides category business logic
type CategoryService struct {
	categoryRepo *repositories.CategoryRepository
}

// NewCategoryService creates a new CategoryService
func NewCategoryService() *CategoryService {
	return &CategoryService{
		categoryRepo: repositories.NewCategoryRepository(),
	}
}

// CreateCategory creates a new category
func (s *CategoryService) CreateCategory(title, userID string) (*models.Category, error) {
	// Validate input
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}

	category := &models.Category{
		ID:     utils.GenerateID(),
		Title:  strings.TrimSpace(title),
		UserID: userID,
	}

	err := s.categoryRepo.CreateCategory(category)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %v", err)
	}

	return category, nil
}

// GetCategoryByID gets a category by ID
func (s *CategoryService) GetCategoryByID(id string) (*models.Category, error) {
	return s.categoryRepo.GetCategoryByID(id)
}

// GetCategoriesByUserID gets all categories for a user
func (s *CategoryService) GetCategoriesByUserID(userID string) ([]*models.Category, error) {
	return s.categoryRepo.GetCategoriesByUserID(userID)
}

// UpdateCategory updates a category
func (s *CategoryService) UpdateCategory(id, title, userID string) (*models.Category, error) {
	// Validate input
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}

	// Get existing category
	existingCategory, err := s.categoryRepo.GetCategoryByID(id)
	if err != nil {
		return nil, fmt.Errorf("category not found: %v", err)
	}

	// Check ownership
	if existingCategory.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}

	// Update category
	category := &models.Category{
		ID:        id,
		Title:     strings.TrimSpace(title),
		UserID:    userID,
		CreatedAt: existingCategory.CreatedAt,
	}

	err = s.categoryRepo.UpdateCategory(category)
	if err != nil {
		return nil, fmt.Errorf("failed to update category: %v", err)
	}

	return category, nil
}

// DeleteCategory deletes a category
func (s *CategoryService) DeleteCategory(id, userID string) error {
	// Get existing category
	existingCategory, err := s.categoryRepo.GetCategoryByID(id)
	if err != nil {
		return fmt.Errorf("category not found: %v", err)
	}

	// Check ownership
	if existingCategory.UserID != userID {
		return fmt.Errorf("access denied")
	}

	return s.categoryRepo.DeleteCategory(id)
}

// CategoryExists checks if a category exists
func (s *CategoryService) CategoryExists(id string) (bool, error) {
	return s.categoryRepo.CategoryExists(id)
}