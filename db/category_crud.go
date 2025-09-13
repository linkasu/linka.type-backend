package db

import (
	"linka.type-backend/db/repositories"
)

// CategoryCRUD provides CRUD operations for Category entity
// This is a wrapper around CategoryRepository for backward compatibility
type CategoryCRUD struct {
	repo *repositories.CategoryRepository
}

// NewCategoryCRUD creates a new CategoryCRUD
func NewCategoryCRUD() *CategoryCRUD {
	return &CategoryCRUD{
		repo: repositories.NewCategoryRepository(),
	}
}

// CreateCategory creates a new category
func (c *CategoryCRUD) CreateCategory(category *Category) error {
	return c.repo.CreateCategory(category)
}

// GetCategoryByID retrieves a category by ID
func (c *CategoryCRUD) GetCategoryByID(id string) (*Category, error) {
	return c.repo.GetCategoryByID(id)
}

// GetCategoriesByUserID retrieves all categories for a specific user
func (c *CategoryCRUD) GetCategoriesByUserID(userID string) ([]*Category, error) {
	return c.repo.GetCategoriesByUserID(userID)
}

// GetAllCategories retrieves all categories
func (c *CategoryCRUD) GetAllCategories() ([]*Category, error) {
	return c.repo.GetAllCategories()
}

// UpdateCategory updates an existing category
func (c *CategoryCRUD) UpdateCategory(category *Category) error {
	return c.repo.UpdateCategory(category)
}

// DeleteCategory deletes a category by ID
func (c *CategoryCRUD) DeleteCategory(id string) error {
	return c.repo.DeleteCategory(id)
}

// DeleteCategoriesByUserID deletes all categories for a specific user
func (c *CategoryCRUD) DeleteCategoriesByUserID(userID string) error {
	return c.repo.DeleteCategoriesByUserID(userID)
}

// CategoryExists checks if a category exists by ID
func (c *CategoryCRUD) CategoryExists(id string) (bool, error) {
	return c.repo.CategoryExists(id)
}