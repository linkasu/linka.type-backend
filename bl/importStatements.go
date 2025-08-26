package bl

import (
	"fmt"
	"log"
	"time"

	"linka.type-backend/db"
	"linka.type-backend/fb"
)

// ImportStatementsResult содержит результат импорта statements
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

// ImportStatements импортирует statements пользователя из Firebase в PostgreSQL
// Поддерживает многократные запуски с инкрементальным обновлением
func ImportStatements(login string, password string) (*ImportStatementsResult, error) {
	startTime := time.Now()
	result := &ImportStatementsResult{
		Errors: []ImportError{},
		Stats:  make(map[string]int),
	}

	// Получаем пользователя из Firebase
	user, err := fb.GetUser(login)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from Firebase: %v", err)
	}

	// Проверяем пароль
	_, err = fb.CheckPassword(login, password)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %v", err)
	}

	// Получаем категории пользователя из Firebase
	fbCategories, err := fb.GetCategories(user)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories from Firebase: %v", err)
	}

	log.Printf("Found %d categories in Firebase for user %s", len(fbCategories), user.UID)

	// Инициализируем трекер миграций
	migrationTracker := &db.MigrationTracker{}
	statementCRUD := &db.StatementCRUD{}

	// Обрабатываем statements для каждой категории
	for _, fbCategory := range fbCategories {
		// Получаем statements для категории
		fbStatements, err := fbCategory.GetStatements()
		if err != nil {
			log.Printf("Failed to get statements for category %s: %v", fbCategory.ID, err)
			continue
		}

		log.Printf("Found %d statements in category %s", len(fbStatements), fbCategory.ID)

		// Обрабатываем каждое statement
		for _, fbStatement := range fbStatements {
			// Убеждаемся, что используем правильный UserID (из категории)
			fbStatement.UserID = fbCategory.UserID
			result.TotalProcessed++

			// Проверяем статус последней миграции
			lastMigration, err := migrationTracker.GetLastMigrationStatus("statement", fbStatement.ID, user.UID)
			if err != nil {
				errorMsg := fmt.Sprintf("failed to get migration status: %v", err)
				result.Errors = append(result.Errors, ImportError{
					CategoryID: fbStatement.CategoryID,
					UserID:     user.UID,
					Error:      errorMsg,
				})
				result.Failed++
				log.Printf("Error processing statement %s: %s", fbStatement.ID, errorMsg)
				continue
			}

			// Проверяем, существует ли statement в PostgreSQL
			existingStatement, err := statementCRUD.GetStatementByID(fbStatement.ID)
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

			// Определяем действие на основе существования и статуса миграции
			action := determineAction(lastMigration, existingStatement != nil)

			switch action {
			case "import":
				err = importNewStatement(fbStatement, statementCRUD, migrationTracker)
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

			case "update":
				err = updateExistingStatement(fbStatement, statementCRUD, migrationTracker)
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

			case "skip":
				result.Skipped++
				log.Printf("Skipped statement %s (already up to date)", fbStatement.ID)
			}
		}
	}

	// Получаем статистику миграций
	stats, err := migrationTracker.GetMigrationStats("statement")
	if err != nil {
		log.Printf("Warning: failed to get migration stats: %v", err)
	} else {
		result.Stats = stats
	}

	result.Duration = time.Since(startTime)

	log.Printf("Statements import completed in %v. Processed: %d, Imported: %d, Updated: %d, Skipped: %d, Failed: %d",
		result.Duration, result.TotalProcessed, result.Imported, result.Updated, result.Skipped, result.Failed)

	return result, nil
}

// importNewStatement импортирует новое statement
func importNewStatement(fbStatement *fb.Statement, statementCRUD *db.StatementCRUD, migrationTracker *db.MigrationTracker) error {
	// Создаем statement в PostgreSQL
	pgStatement := &db.Statement{
		ID:         fbStatement.ID,
		Title:      fbStatement.Text,
		UserID:     fbStatement.UserID,
		CategoryID: fbStatement.CategoryID,
	}

	err := statementCRUD.CreateStatement(pgStatement)
	if err != nil {
		// Логируем неудачную попытку
		if logErr := migrationTracker.LogMigration("statement", fbStatement.ID, fbStatement.UserID, "import", "failed", err.Error()); logErr != nil {
			// Log the error but don't fail the operation
			fmt.Printf("Failed to log migration: %v", logErr)
		}
		return fmt.Errorf("failed to create statement: %v", err)
	}

	// Логируем успешную миграцию
	return migrationTracker.LogMigration("statement", fbStatement.ID, fbStatement.UserID, "import", "success", "")
}

// updateExistingStatement обновляет существующее statement
func updateExistingStatement(fbStatement *fb.Statement, statementCRUD *db.StatementCRUD, migrationTracker *db.MigrationTracker) error {
	// Обновляем statement в PostgreSQL
	pgStatement := &db.Statement{
		ID:         fbStatement.ID,
		Title:      fbStatement.Text,
		UserID:     fbStatement.UserID,
		CategoryID: fbStatement.CategoryID,
	}

	err := statementCRUD.UpdateStatement(pgStatement)
	if err != nil {
		// Логируем неудачную попытку
		if logErr := migrationTracker.LogMigration("statement", fbStatement.ID, fbStatement.UserID, "update", "failed", err.Error()); logErr != nil {
			// Log the error but don't fail the operation
			fmt.Printf("Failed to log migration: %v", logErr)
		}
		return fmt.Errorf("failed to update statement: %v", err)
	}

	// Логируем успешную миграцию
	return migrationTracker.LogMigration("statement", fbStatement.ID, fbStatement.UserID, "update", "success", "")
}

// ImportAllData импортирует пользователя, категории и statements
func ImportAllData(login, password string) (*ImportAllDataResult, error) {
	startTime := time.Now()
	result := &ImportAllDataResult{
		StartTime: startTime,
	}

	// Получаем пользователя из Firebase для проверки
	user, err := fb.GetUser(login)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from Firebase: %v", err)
	}

	// Проверяем пароль
	_, err = fb.CheckPassword(login, password)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %v", err)
	}

	// Импортируем пользователя и категории
	err = ImportUser(login, password)
	if err != nil {
		return nil, fmt.Errorf("failed to import user and categories: %v", err)
	}

	// Убеждаемся, что пользователь существует в PostgreSQL перед импортом statements
	userCRUD := &db.UserCRUD{}
	existingUser, err := userCRUD.GetUserByID(user.UID)
	if err != nil || existingUser == nil {
		return nil, fmt.Errorf("user not found in PostgreSQL after import: %v", err)
	}

	// Импортируем statements
	statementsResult, err := ImportStatements(login, password)
	if err != nil {
		return nil, fmt.Errorf("failed to import statements: %v", err)
	}

	result.StatementsResult = statementsResult
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	log.Printf("Complete data import finished in %v", result.Duration)

	return result, nil
}

// ImportAllDataResult содержит результат полного импорта данных
type ImportAllDataResult struct {
	StatementsResult *ImportStatementsResult `json:"statementsResult"`
	Duration         time.Duration           `json:"duration"`
	StartTime        time.Time               `json:"startTime"`
	EndTime          time.Time               `json:"endTime"`
}
