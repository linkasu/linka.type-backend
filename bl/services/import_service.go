package services

import (
	"fmt"
	"log"
	"time"

	"linka.type-backend/db/repositories"
	"linka.type-backend/fb"
	"linka.type-backend/models"
	"linka.type-backend/utils"
)

// ImportService provides import business logic
type ImportService struct {
	userRepo      *repositories.UserRepository
	categoryRepo  *repositories.CategoryRepository
	statementRepo *repositories.StatementRepository
}

// NewImportService creates a new ImportService
func NewImportService() *ImportService {
	return &ImportService{
		userRepo:      repositories.NewUserRepository(),
		categoryRepo:  repositories.NewCategoryRepository(),
		statementRepo: repositories.NewStatementRepository(),
	}
}

// ImportAllData imports all data from Firebase
func (s *ImportService) ImportAllData(email, password string) (*ImportAllDataResult, error) {
	startTime := time.Now()
	result := &ImportAllDataResult{
		StartTime: startTime,
	}

	// Get user from Firebase
	user, err := fb.GetUser(email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from Firebase: %v", err)
	}

	// Check password
	_, err = fb.CheckPassword(email, password)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %v", err)
	}

	// Import user and categories
	err = s.ImportUser(email, password)
	if err != nil {
		return nil, fmt.Errorf("failed to import user and categories: %v", err)
	}

	// Make sure user exists in PostgreSQL before importing statements
	existingUser, err := s.userRepo.GetUserByID(user.UID)
	if err != nil || existingUser == nil {
		return nil, fmt.Errorf("user not found in PostgreSQL after import: %v", err)
	}

	// Import statements
	statementsResult, err := s.ImportStatements(email, password)
	if err != nil {
		return nil, fmt.Errorf("failed to import statements: %v", err)
	}

	result.StatementsResult = statementsResult
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	log.Printf("Complete data import finished in %v", result.Duration)

	return result, nil
}

// ImportUser imports user and categories from Firebase
func (s *ImportService) ImportUser(email, password string) error {
	// Get user from Firebase
	user, err := fb.GetUser(email)
	if err != nil {
		return fmt.Errorf("failed to get user from Firebase: %v", err)
	}

	// Check password
	_, err = fb.CheckPassword(email, password)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %v", err)
	}

	// Import user in PostgreSQL
	hasher := utils.NewPasswordHasher()
	hashedPassword, err := hasher.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	pgUser := &models.User{
		ID:       user.UID,
		Email:    user.Email,
		Password: hashedPassword,
	}

	// Check if user exists
	existingUser, err := s.userRepo.GetUserByID(user.UID)
	if err != nil && err.Error() != "user not found" {
		return fmt.Errorf("failed to check existing user: %v", err)
	}

	if existingUser == nil {
		// Create new user
		err = s.userRepo.CreateUser(pgUser)
		if err != nil {
			return fmt.Errorf("failed to create user: %v", err)
		}
		log.Printf("Successfully created user %s", user.UID)
	} else {
		// Update existing user
		err = s.userRepo.UpdateUser(pgUser)
		if err != nil {
			return fmt.Errorf("failed to update user: %v", err)
		}
		log.Printf("Successfully updated user %s", user.UID)
	}

	// Import categories
	_, err = s.ImportCategories(email, password)
	if err != nil {
		return fmt.Errorf("failed to import categories: %v", err)
	}

	return nil
}

// ImportCategories imports categories from Firebase
func (s *ImportService) ImportCategories(email, password string) (*ImportCategoriesResult, error) {
	startTime := time.Now()
	result := &ImportCategoriesResult{
		Errors: []ImportError{},
		Stats:  make(map[string]int),
	}

	// Get user from Firebase
	user, err := fb.GetUser(email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from Firebase: %v", err)
	}

	// Check password
	_, err = fb.CheckPassword(email, password)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %v", err)
	}

	// Get categories from Firebase
	fbCategories, err := fb.GetCategories(user)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories from Firebase: %v", err)
	}

	log.Printf("Found %d categories in Firebase for user %s", len(fbCategories), user.UID)

	// Process each category
	for _, fbCategory := range fbCategories {
		result.TotalProcessed++

		// Check if category exists in PostgreSQL
		existingCategory, err := s.categoryRepo.GetCategoryByID(fbCategory.ID)
		if err != nil && err.Error() != "category not found" {
			errorMsg := fmt.Sprintf("failed to check existing category: %v", err)
			result.Errors = append(result.Errors, ImportError{
				CategoryID: fbCategory.ID,
				UserID:     user.UID,
				Error:      errorMsg,
			})
			result.Failed++
			log.Printf("Error checking category %s: %s", fbCategory.ID, errorMsg)
			continue
		}

		if existingCategory == nil {
			// Import new category
			pgCategory := &models.Category{
				ID:     fbCategory.ID,
				Title:  fbCategory.Label,
				UserID: fbCategory.UserID,
			}

			err = s.categoryRepo.CreateCategory(pgCategory)
			if err != nil {
				result.Failed++
				result.Errors = append(result.Errors, ImportError{
					CategoryID: fbCategory.ID,
					UserID:     user.UID,
					Error:      err.Error(),
				})
				log.Printf("Failed to import category %s: %v", fbCategory.ID, err)
			} else {
				result.Imported++
				log.Printf("Successfully imported category %s", fbCategory.ID)
			}
		} else {
			// Update existing category
			pgCategory := &models.Category{
				ID:     fbCategory.ID,
				Title:  fbCategory.Label,
				UserID: fbCategory.UserID,
			}

			err = s.categoryRepo.UpdateCategory(pgCategory)
			if err != nil {
				result.Failed++
				result.Errors = append(result.Errors, ImportError{
					CategoryID: fbCategory.ID,
					UserID:     user.UID,
					Error:      err.Error(),
				})
				log.Printf("Failed to update category %s: %v", fbCategory.ID, err)
			} else {
				result.Updated++
				log.Printf("Successfully updated category %s", fbCategory.ID)
			}
		}
	}

	result.Duration = time.Since(startTime)

	log.Printf("Categories import completed in %v. Processed: %d, Imported: %d, Updated: %d, Failed: %d",
		result.Duration, result.TotalProcessed, result.Imported, result.Updated, result.Failed)

	return result, nil
}

// ImportStatements imports statements from Firebase
func (s *ImportService) ImportStatements(email, password string) (*ImportStatementsResult, error) {
	startTime := time.Now()
	result := &ImportStatementsResult{
		Errors: []ImportError{},
		Stats:  make(map[string]int),
	}

	// Get user from Firebase
	user, err := fb.GetUser(email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from Firebase: %v", err)
	}

	// Check password
	_, err = fb.CheckPassword(email, password)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %v", err)
	}

	// Get categories from Firebase
	fbCategories, err := fb.GetCategories(user)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories from Firebase: %v", err)
	}

	log.Printf("Found %d categories in Firebase for user %s", len(fbCategories), user.UID)

	// Process statements for each category
	for _, fbCategory := range fbCategories {
		// Get statements for category
		fbStatements, err := fbCategory.GetStatements()
		if err != nil {
			log.Printf("Failed to get statements for category %s: %v", fbCategory.ID, err)
			continue
		}

		log.Printf("Found %d statements in category %s", len(fbStatements), fbCategory.ID)

		// Process each statement
		for _, fbStatement := range fbStatements {
			// Ensure we use the correct UserID (from category)
			fbStatement.UserID = fbCategory.UserID
			result.TotalProcessed++

			// Check if statement exists in PostgreSQL
			existingStatement, err := s.statementRepo.GetStatementByID(fbStatement.ID)
			if err != nil && err.Error() != "statement not found" {
				errorMsg := fmt.Sprintf("failed to check existing statement: %v", err)
				result.Errors = append(result.Errors, ImportError{
					CategoryID: fbStatement.CategoryID,
					UserID:     user.UID,
					Error:      errorMsg,
				})
				result.Failed++
				log.Printf("Error checking statement %s: %s", fbStatement.ID, errorMsg)
				continue
			}

			if existingStatement == nil {
				// Import new statement
				pgStatement := &models.Statement{
					ID:         fbStatement.ID,
					Title:      fbStatement.Text,
					UserID:     fbStatement.UserID,
					CategoryID: fbStatement.CategoryID,
				}

				err = s.statementRepo.CreateStatement(pgStatement)
				if err != nil {
					result.Failed++
					result.Errors = append(result.Errors, ImportError{
						CategoryID: fbStatement.CategoryID,
						UserID:     user.UID,
						Error:      err.Error(),
					})
					log.Printf("Failed to import statement %s: %v", fbStatement.ID, err)
				} else {
					result.Imported++
					log.Printf("Successfully imported statement %s", fbStatement.ID)
				}
			} else {
				// Update existing statement
				pgStatement := &models.Statement{
					ID:         fbStatement.ID,
					Title:      fbStatement.Text,
					UserID:     fbStatement.UserID,
					CategoryID: fbStatement.CategoryID,
				}

				err = s.statementRepo.UpdateStatement(pgStatement)
				if err != nil {
					result.Failed++
					result.Errors = append(result.Errors, ImportError{
						CategoryID: fbStatement.CategoryID,
						UserID:     user.UID,
						Error:      err.Error(),
					})
					log.Printf("Failed to update statement %s: %v", fbStatement.ID, err)
				} else {
					result.Updated++
					log.Printf("Successfully updated statement %s", fbStatement.ID)
				}
			}
		}
	}

	result.Duration = time.Since(startTime)

	log.Printf("Statements import completed in %v. Processed: %d, Imported: %d, Updated: %d, Failed: %d",
		result.Duration, result.TotalProcessed, result.Imported, result.Updated, result.Failed)

	return result, nil
}

// ImportError represents an import error
type ImportError struct {
	CategoryID string `json:"categoryId"`
	UserID     string `json:"userId"`
	Error      string `json:"error"`
}

// ImportCategoriesResult represents the result of importing categories
type ImportCategoriesResult struct {
	TotalProcessed int            `json:"totalProcessed"`
	Imported       int            `json:"imported"`
	Updated        int            `json:"updated"`
	Skipped        int            `json:"skipped"`
	Failed         int            `json:"failed"`
	Errors         []ImportError  `json:"errors"`
	Stats          map[string]int `json:"stats"`
	Duration       time.Duration  `json:"duration"`
}

// ImportStatementsResult represents the result of importing statements
type ImportStatementsResult struct {
	TotalProcessed int            `json:"totalProcessed"`
	Imported       int            `json:"imported"`
	Updated        int            `json:"updated"`
	Skipped        int            `json:"skipped"`
	Failed         int            `json:"failed"`
	Errors         []ImportError  `json:"errors"`
	Stats          map[string]int `json:"stats"`
	Duration       time.Duration  `json:"duration"`
}

// ImportAllDataResult represents the result of importing all data
type ImportAllDataResult struct {
	StatementsResult *ImportStatementsResult `json:"statementsResult"`
	Duration         time.Duration           `json:"duration"`
	StartTime        time.Time               `json:"startTime"`
	EndTime          time.Time               `json:"endTime"`
}