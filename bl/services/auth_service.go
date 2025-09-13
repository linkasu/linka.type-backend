package services

import (
	"crypto/md5"
	"fmt"
	"log"

	"linka.type-backend/db/repositories"
	"linka.type-backend/fb"
	"linka.type-backend/models"
	"linka.type-backend/utils"
)

// AuthService provides authentication business logic
type AuthService struct {
	userRepo *repositories.UserRepository
}

// NewAuthService creates a new AuthService
func NewAuthService() *AuthService {
	return &AuthService{
		userRepo: repositories.NewUserRepository(),
	}
}

// AuthenticateUser authenticates a user with email and password
func (s *AuthService) AuthenticateUser(email, password string) (*models.User, error) {
	// Try PostgreSQL first
	user, err := s.userRepo.GetUserByEmail(email)
	if err == nil {
		// User exists in PostgreSQL, check password
		if s.checkPostgreSQLPassword(user, password) {
			return user, nil
		}
	}

	// Try Firebase authentication
	firebaseAuth, err := fb.CheckPassword(email, password)
	if err != nil {
		return nil, fmt.Errorf("authentication failed")
	}

	// Handle Firebase authentication
	if firebaseAuth != nil {
		return s.handleFirebaseAuth(email, password, user)
	}

	return nil, fmt.Errorf("authentication failed")
}

// checkPostgreSQLPassword checks password in PostgreSQL
func (s *AuthService) checkPostgreSQLPassword(user *models.User, password string) bool {
	if user == nil {
		log.Printf("User not found in PostgreSQL, skipping password check")
		return false
	}

	hasher := utils.NewPasswordHasher()
	passwordOK := hasher.CheckPassword(user.Password, password) == nil
	log.Printf("Bcrypt password check result: %v", passwordOK)

	if !passwordOK {
		legacy := fmt.Sprintf("%x", md5.Sum([]byte(password)))
		log.Printf("Trying legacy MD5 check for user: %s", user.ID)
		if user.Password == legacy {
			passwordOK = true
			log.Printf("Legacy MD5 password match, updating to bcrypt")
			if newHash, err := hasher.HashPassword(password); err == nil {
				_ = s.userRepo.UpdateUserPassword(user.ID, newHash)
			}
		}
	}

	log.Printf("PostgreSQL password check final result: %v", passwordOK)
	return passwordOK
}

// handleFirebaseAuth handles successful Firebase authentication
func (s *AuthService) handleFirebaseAuth(email, password string, existingUser *models.User) (*models.User, error) {
	if existingUser == nil {
		log.Printf("User not found in PostgreSQL but Firebase auth successful, importing user")
		// Import user from Firebase
		return s.importUserFromFirebase(email, password)
	} else {
		log.Printf("User found in PostgreSQL but password incorrect, updating password and importing data")
		hasher := utils.NewPasswordHasher()
		if newHash, err := hasher.HashPassword(password); err == nil {
			_ = s.userRepo.UpdateUserPassword(existingUser.ID, newHash)
			log.Printf("Updated password for user: %s", existingUser.ID)
		}
		return existingUser, nil
	}
}

// importUserFromFirebase imports user from Firebase
func (s *AuthService) importUserFromFirebase(email, password string) (*models.User, error) {
	fbUser, err := fb.GetUser(email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from Firebase: %v", err)
	}

	// Create user in PostgreSQL
	hasher := utils.NewPasswordHasher()
	hashedPassword, err := hasher.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	user := &models.User{
		ID:            fbUser.UID,
		Email:         fbUser.Email,
		Password:      hashedPassword,
		EmailVerified: fbUser.EmailVerified,
	}

	err = s.userRepo.CreateUser(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return user, nil
}

// CreateUser creates a new user
func (s *AuthService) CreateUser(email, password string, emailVerified bool) (*models.User, error) {
	// Check if user already exists
	exists, err := s.userRepo.UserExists(email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %v", err)
	}

	if exists {
		return nil, fmt.Errorf("user already exists")
	}

	// Validate password strength
	hasher := utils.NewPasswordHasher()
	if ok, _ := hasher.ValidatePasswordStrength(password); !ok {
		return nil, fmt.Errorf("weak password")
	}

	// Hash password
	hashedPassword, err := hasher.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Create user
	user := &models.User{
		ID:            utils.GenerateID(),
		Email:         email,
		Password:      hashedPassword,
		EmailVerified: emailVerified,
	}

	err = s.userRepo.CreateUser(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return user, nil
}

// VerifyUserEmail verifies user email
func (s *AuthService) VerifyUserEmail(userID string) error {
	return s.userRepo.VerifyUserEmail(userID)
}

// UpdateUserPassword updates user password
func (s *AuthService) UpdateUserPassword(userID, newPassword string) error {
	// Validate password strength
	hasher := utils.NewPasswordHasher()
	if ok, _ := hasher.ValidatePasswordStrength(newPassword); !ok {
		return fmt.Errorf("weak password")
	}

	// Hash password
	hashedPassword, err := hasher.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	return s.userRepo.UpdateUserPassword(userID, hashedPassword)
}