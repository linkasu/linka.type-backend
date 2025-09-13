package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"linka.type-backend/db"
	"linka.type-backend/models"
)

// UserRepository provides CRUD operations for User entity
type UserRepository struct{}

// NewUserRepository creates a new UserRepository
func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// CreateUser creates a new user
func (u *UserRepository) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (id, email, password, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	now := time.Now()
	_, err := db.DB.Exec(query, user.ID, user.Email, user.Password, user.EmailVerified, now, now)
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}

	return nil
}

// GetUserByID retrieves a user by ID
func (u *UserRepository) GetUserByID(id string) (*models.User, error) {
	query := `SELECT id, email, password, email_verified, created_at, updated_at FROM users WHERE id = $1`

	var user models.User
	err := db.DB.QueryRow(query, id).Scan(&user.ID, &user.Email, &user.Password, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error getting user: %v", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (u *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, email, password, email_verified, created_at, updated_at FROM users WHERE email = $1`

	var user models.User
	err := db.DB.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Password, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error getting user: %v", err)
	}

	return &user, nil
}

// GetAllUsers retrieves all users
func (u *UserRepository) GetAllUsers() ([]*models.User, error) {
	query := `SELECT id, email, password, email_verified, created_at, updated_at FROM users ORDER BY created_at DESC`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting users: %v", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.Password, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt); err != nil {
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
func (u *UserRepository) UpdateUser(user *models.User) error {
	query := `
		UPDATE users 
		SET email = $2, password = $3, email_verified = $4, updated_at = $5
		WHERE id = $1
	`

	now := time.Now()
	result, err := db.DB.Exec(query, user.ID, user.Email, user.Password, user.EmailVerified, now)
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
func (u *UserRepository) DeleteUser(id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := db.DB.Exec(query, id)
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
func (u *UserRepository) UserExists(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := db.DB.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking user existence: %v", err)
	}

	return exists, nil
}

// VerifyUserEmail marks user email as verified
func (u *UserRepository) VerifyUserEmail(userID string) error {
	query := `UPDATE users SET email_verified = true, updated_at = $2 WHERE id = $1`

	now := time.Now()
	result, err := db.DB.Exec(query, userID, now)
	if err != nil {
		return fmt.Errorf("error verifying user email: %v", err)
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

// UpdateUserPassword updates user password
func (u *UserRepository) UpdateUserPassword(userID, newPassword string) error {
	query := `UPDATE users SET password = $2, updated_at = $3 WHERE id = $1`

	now := time.Now()
	result, err := db.DB.Exec(query, userID, newPassword, now)
	if err != nil {
		return fmt.Errorf("error updating user password: %v", err)
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
