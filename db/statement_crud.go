package db

import (
	"database/sql"
	"fmt"
	"time"
)

// StatementCRUD provides CRUD operations for Statement entity
type StatementCRUD struct{}

// CreateStatement creates a new statement
func (s *StatementCRUD) CreateStatement(statement *Statement) error {
	query := `
		INSERT INTO statements (id, title, user_id, category_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	
	now := time.Now()
	_, err := DB.Exec(query, statement.ID, statement.Title, statement.UserId, statement.CategoryId, now, now)
	if err != nil {
		return fmt.Errorf("error creating statement: %v", err)
	}
	
	return nil
}

// GetStatementByID retrieves a statement by ID
func (s *StatementCRUD) GetStatementByID(id string) (*Statement, error) {
	query := `SELECT id, title, user_id, category_id FROM statements WHERE id = $1`
	
	var statement Statement
	err := DB.QueryRow(query, id).Scan(&statement.ID, &statement.Title, &statement.UserId, &statement.CategoryId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("statement not found")
		}
		return nil, fmt.Errorf("error getting statement: %v", err)
	}
	
	return &statement, nil
}

// GetStatementsByUserID retrieves all statements for a specific user
func (s *StatementCRUD) GetStatementsByUserID(userID string) ([]*Statement, error) {
	query := `SELECT id, title, user_id, category_id FROM statements WHERE user_id = $1 ORDER BY created_at DESC`
	
	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting statements: %v", err)
	}
	defer rows.Close()
	
	var statements []*Statement
	for rows.Next() {
		var statement Statement
		if err := rows.Scan(&statement.ID, &statement.Title, &statement.UserId, &statement.CategoryId); err != nil {
			return nil, fmt.Errorf("error scanning statement: %v", err)
		}
		statements = append(statements, &statement)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating statements: %v", err)
	}
	
	return statements, nil
}

// GetStatementsByCategoryID retrieves all statements for a specific category
func (s *StatementCRUD) GetStatementsByCategoryID(categoryID string) ([]*Statement, error) {
	query := `SELECT id, title, user_id, category_id FROM statements WHERE category_id = $1 ORDER BY created_at DESC`
	
	rows, err := DB.Query(query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("error getting statements: %v", err)
	}
	defer rows.Close()
	
	var statements []*Statement
	for rows.Next() {
		var statement Statement
		if err := rows.Scan(&statement.ID, &statement.Title, &statement.UserId, &statement.CategoryId); err != nil {
			return nil, fmt.Errorf("error scanning statement: %v", err)
		}
		statements = append(statements, &statement)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating statements: %v", err)
	}
	
	return statements, nil
}

// GetAllStatements retrieves all statements
func (s *StatementCRUD) GetAllStatements() ([]*Statement, error) {
	query := `SELECT id, title, user_id, category_id FROM statements ORDER BY created_at DESC`
	
	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting statements: %v", err)
	}
	defer rows.Close()
	
	var statements []*Statement
	for rows.Next() {
		var statement Statement
		if err := rows.Scan(&statement.ID, &statement.Title, &statement.UserId, &statement.CategoryId); err != nil {
			return nil, fmt.Errorf("error scanning statement: %v", err)
		}
		statements = append(statements, &statement)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating statements: %v", err)
	}
	
	return statements, nil
}

// UpdateStatement updates an existing statement
func (s *StatementCRUD) UpdateStatement(statement *Statement) error {
	query := `
		UPDATE statements 
		SET title = $2, user_id = $3, category_id = $4, updated_at = $5
		WHERE id = $1
	`
	
	now := time.Now()
	result, err := DB.Exec(query, statement.ID, statement.Title, statement.UserId, statement.CategoryId, now)
	if err != nil {
		return fmt.Errorf("error updating statement: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("statement not found")
	}
	
	return nil
}

// DeleteStatement deletes a statement by ID
func (s *StatementCRUD) DeleteStatement(id string) error {
	query := `DELETE FROM statements WHERE id = $1`
	
	result, err := DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting statement: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("statement not found")
	}
	
	return nil
}

// DeleteStatementsByUserID deletes all statements for a specific user
func (s *StatementCRUD) DeleteStatementsByUserID(userID string) error {
	query := `DELETE FROM statements WHERE user_id = $1`
	
	result, err := DB.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("error deleting statements: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("no statements found for user")
	}
	
	return nil
}

// DeleteStatementsByCategoryID deletes all statements for a specific category
func (s *StatementCRUD) DeleteStatementsByCategoryID(categoryID string) error {
	query := `DELETE FROM statements WHERE category_id = $1`
	
	result, err := DB.Exec(query, categoryID)
	if err != nil {
		return fmt.Errorf("error deleting statements: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("no statements found for category")
	}
	
	return nil
}

// StatementExists checks if a statement exists by ID
func (s *StatementCRUD) StatementExists(id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM statements WHERE id = $1)`
	
	var exists bool
	err := DB.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking statement existence: %v", err)
	}
	
	return exists, nil
} 