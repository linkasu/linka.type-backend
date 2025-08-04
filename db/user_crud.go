package db

import (
	"database/sql"
	"fmt"
	"time"
)

// UserCRUD provides CRUD operations for User entity
type UserCRUD struct{}

// CreateUser creates a new user
func (u *UserCRUD) CreateUser(user *User) error {
	query := `
		INSERT INTO users (id, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	
	now := time.Now()
	_, err := DB.Exec(query, user.ID, user.Email, user.Password, now, now)
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}
	
	return nil
}

// GetUserByID retrieves a user by ID
func (u *UserCRUD) GetUserByID(id string) (*User, error) {
	query := `SELECT id, email, password FROM users WHERE id = $1`
	
	var user User
	err := DB.QueryRow(query, id).Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error getting user: %v", err)
	}
	
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (u *UserCRUD) GetUserByEmail(email string) (*User, error) {
	query := `SELECT id, email, password FROM users WHERE email = $1`
	
	var user User
	err := DB.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error getting user: %v", err)
	}
	
	return &user, nil
}

// GetAllUsers retrieves all users
func (u *UserCRUD) GetAllUsers() ([]*User, error) {
	query := `SELECT id, email, password FROM users ORDER BY created_at DESC`
	
	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting users: %v", err)
	}
	defer rows.Close()
	
	var users []*User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Email, &user.Password); err != nil {
			return nil, fmt.Errorf("error scanning user: %v", err)
		}
		users = append(users, &user)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %v", err)
	}
	
	return users, nil
}

// UpdateUser updates an existing user
func (u *UserCRUD) UpdateUser(user *User) error {
	query := `
		UPDATE users 
		SET email = $2, password = $3, updated_at = $4
		WHERE id = $1
	`
	
	now := time.Now()
	result, err := DB.Exec(query, user.ID, user.Email, user.Password, now)
	if err != nil {
		return fmt.Errorf("error updating user: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}

// DeleteUser deletes a user by ID
func (u *UserCRUD) DeleteUser(id string) error {
	query := `DELETE FROM users WHERE id = $1`
	
	result, err := DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting user: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}

// UserExists checks if a user exists by email
func (u *UserCRUD) UserExists(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	
	var exists bool
	err := DB.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking user existence: %v", err)
	}
	
	return exists, nil
} 