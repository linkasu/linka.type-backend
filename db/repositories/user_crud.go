package repositories

import (
	"linka.type-backend/models"
)

// UserCRUD provides CRUD operations for User entity
// This is a wrapper around UserRepository for backward compatibility
type UserCRUD struct {
	repo *UserRepository
}

// NewUserCRUD creates a new UserCRUD
func NewUserCRUD() *UserCRUD {
	return &UserCRUD{
		repo: NewUserRepository(),
	}
}

// CreateUser creates a new user
func (u *UserCRUD) CreateUser(user *models.User) error {
	return u.repo.CreateUser(user)
}

// GetUserByID retrieves a user by ID
func (u *UserCRUD) GetUserByID(id string) (*models.User, error) {
	return u.repo.GetUserByID(id)
}

// GetUserByEmail retrieves a user by email
func (u *UserCRUD) GetUserByEmail(email string) (*models.User, error) {
	return u.repo.GetUserByEmail(email)
}

// GetAllUsers retrieves all users
func (u *UserCRUD) GetAllUsers() ([]*models.User, error) {
	return u.repo.GetAllUsers()
}

// UpdateUser updates an existing user
func (u *UserCRUD) UpdateUser(user *models.User) error {
	return u.repo.UpdateUser(user)
}

// DeleteUser deletes a user by ID
func (u *UserCRUD) DeleteUser(id string) error {
	return u.repo.DeleteUser(id)
}

// UserExists checks if a user exists by email
func (u *UserCRUD) UserExists(email string) (bool, error) {
	return u.repo.UserExists(email)
}

// VerifyUserEmail marks user email as verified
func (u *UserCRUD) VerifyUserEmail(userID string) error {
	return u.repo.VerifyUserEmail(userID)
}

// UpdateUserPassword updates user password
func (u *UserCRUD) UpdateUserPassword(userID, newPassword string) error {
	return u.repo.UpdateUserPassword(userID, newPassword)
}
