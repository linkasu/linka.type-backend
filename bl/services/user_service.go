package services

import (
	"linka.type-backend/db/repositories"
	"linka.type-backend/models"
)

// UserService provides user business logic
type UserService struct {
	userRepo *repositories.UserRepository
}

// NewUserService creates a new UserService
func NewUserService() *UserService {
	return &UserService{
		userRepo: repositories.NewUserRepository(),
	}
}

// GetUserByID gets a user by ID
func (s *UserService) GetUserByID(id string) (*models.User, error) {
	return s.userRepo.GetUserByID(id)
}

// GetUserByEmail gets a user by email
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	return s.userRepo.GetUserByEmail(email)
}

// GetAllUsers gets all users
func (s *UserService) GetAllUsers() ([]*models.User, error) {
	return s.userRepo.GetAllUsers()
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(user *models.User) error {
	return s.userRepo.UpdateUser(user)
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(id string) error {
	return s.userRepo.DeleteUser(id)
}

// UserExists checks if a user exists
func (s *UserService) UserExists(email string) (bool, error) {
	return s.userRepo.UserExists(email)
}