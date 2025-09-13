package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"linka.type-backend/db"
	"linka.type-backend/models"
)

// CategoryRepository provides CRUD operations for Category entity
type CategoryRepository struct{}

// NewCategoryRepository creates a new CategoryRepository
func NewCategoryRepository() *CategoryRepository {
	return &CategoryRepository{}
}

// CreateCategory creates a new category
func (c *CategoryRepository) CreateCategory(category *models.Category) error {
	query := `
		INSERT INTO categories (id, title, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	now := time.Now()
	_, err := db.DB.Exec(query, category.ID, category.Title, category.UserID, now, now)
	if err != nil {
		return fmt.Errorf("error creating category: %v", err)
	}
	category.CreatedAt = now.Format(time.RFC3339)
	category.UpdatedAt = category.CreatedAt
	return nil
}

// GetCategoryByID retrieves a category by ID
func (c *CategoryRepository) GetCategoryByID(id string) (*models.Category, error) {
	query := `SELECT id, title, user_id, created_at, updated_at FROM categories WHERE id = $1`

	var category models.Category
	err := db.DB.QueryRow(query, id).Scan(&category.ID, &category.Title, &category.UserID, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("error getting category: %v", err)
	}

	return &category, nil
}

// GetCategoriesByUserID retrieves all categories for a specific user
func (c *CategoryRepository) GetCategoriesByUserID(userID string) ([]*models.Category, error) {
	query := `SELECT id, title, user_id, created_at, updated_at FROM categories WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := db.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting categories: %v", err)
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.Title, &category.UserID, &category.CreatedAt, &category.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error scanning category: %v", err)
		}
		categories = append(categories, &category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating categories: %v", err)
	}

	return categories, nil
}

// GetAllCategories retrieves all categories
func (c *CategoryRepository) GetAllCategories() ([]*models.Category, error) {
	query := `SELECT id, title, user_id, created_at, updated_at FROM categories ORDER BY created_at DESC`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting categories: %v", err)
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.Title, &category.UserID, &category.CreatedAt, &category.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error scanning category: %v", err)
		}
		categories = append(categories, &category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating categories: %v", err)
	}

	return categories, nil
}

// UpdateCategory updates an existing category
func (c *CategoryRepository) UpdateCategory(category *models.Category) error {
	query := `
		UPDATE categories 
		SET title = $2, user_id = $3, updated_at = $4
		WHERE id = $1
	`

	now := time.Now()
	result, err := db.DB.Exec(query, category.ID, category.Title, category.UserID, now)
	if err != nil {
		return fmt.Errorf("error updating category: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}
	category.UpdatedAt = now.Format(time.RFC3339)
	return nil
}

// DeleteCategory deletes a category by ID
func (c *CategoryRepository) DeleteCategory(id string) error {
	query := `DELETE FROM categories WHERE id = $1`

	result, err := db.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting category: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

// DeleteCategoriesByUserID deletes all categories for a specific user
func (c *CategoryRepository) DeleteCategoriesByUserID(userID string) error {
	query := `DELETE FROM categories WHERE user_id = $1`

	result, err := db.DB.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("error deleting categories: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no categories found for user")
	}

	return nil
}

// CategoryExists checks if a category exists by ID
func (c *CategoryRepository) CategoryExists(id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1)`

	var exists bool
	err := db.DB.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking category existence: %v", err)
	}

	return exists, nil
}
