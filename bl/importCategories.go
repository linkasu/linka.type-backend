package bl

import (
	"fmt"
	"log"
	"time"

	"linka.type-backend/db"
	"linka.type-backend/fb"
)

// ImportCategoriesResult содержит результат импорта категорий
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

// ImportError содержит информацию об ошибке импорта
type ImportError struct {
	CategoryID string `json:"categoryId"`
	UserID     string `json:"userId"`
	Error      string `json:"error"`
}

// ImportCategories импортирует категории пользователя из Firebase в PostgreSQL
// Поддерживает многократные запуски с инкрементальным обновлением
func ImportCategories(login, password string) (*ImportCategoriesResult, error) {
	startTime := time.Now()
	result := &ImportCategoriesResult{
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

	// Получаем категории из Firebase
	fbCategories, err := fb.GetCategories(user)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories from Firebase: %v", err)
	}

	log.Printf("Found %d categories in Firebase for user %s", len(fbCategories), user.UID)

	// Инициализируем трекер миграций
	migrationTracker := &db.MigrationTracker{}
	categoryCRUD := &db.CategoryCRUD{}

	// Обрабатываем каждую категорию
	for _, fbCategory := range fbCategories {
		result.TotalProcessed++

		// Проверяем статус последней миграции
		lastMigration, err := migrationTracker.GetLastMigrationStatus("category", fbCategory.ID, user.UID)
		if err != nil {
			errorMsg := fmt.Sprintf("failed to get migration status: %v", err)
			result.Errors = append(result.Errors, ImportError{
				CategoryID: fbCategory.ID,
				UserID:     user.UID,
				Error:      errorMsg,
			})
			result.Failed++
			log.Printf("Error processing category %s: %s", fbCategory.ID, errorMsg)
			continue
		}

		// Проверяем, существует ли категория в PostgreSQL
		existingCategory, err := categoryCRUD.GetCategoryByID(fbCategory.ID)
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

		// Определяем действие на основе существования и статуса миграции
		action := determineAction(lastMigration, existingCategory != nil)

		switch action {
		case "import":
			err = importNewCategory(fbCategory, categoryCRUD, migrationTracker)
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

		case "update":
			err = updateExistingCategory(fbCategory, categoryCRUD, migrationTracker)
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

		case "skip":
			result.Skipped++
			log.Printf("Skipped category %s (already up to date)", fbCategory.ID)
		}
	}

	// Получаем статистику миграций
	stats, err := migrationTracker.GetMigrationStats("category")
	if err != nil {
		log.Printf("Warning: failed to get migration stats: %v", err)
	} else {
		result.Stats = stats
	}

	result.Duration = time.Since(startTime)

	log.Printf("Import completed in %v. Processed: %d, Imported: %d, Updated: %d, Skipped: %d, Failed: %d",
		result.Duration, result.TotalProcessed, result.Imported, result.Updated, result.Skipped, result.Failed)

	return result, nil
}

// determineAction определяет действие на основе статуса миграции и существования категории
func determineAction(lastMigration *db.MigrationLog, existsInPostgres bool) string {
	if lastMigration == nil {
		// Нет записи о миграции - импортируем
		return "import"
	}

	if lastMigration.Status == "success" {
		if existsInPostgres {
			// Категория уже импортирована и существует - пропускаем
			return "skip"
		}
		// Категория была импортирована, но не существует - импортируем заново
		return "import"
	}

	if lastMigration.Status == "failed" {
		// Последняя попытка миграции не удалась - пробуем снова
		if existsInPostgres {
			return "update"
		}
		return "import"
	}

	// По умолчанию импортируем
	return "import"
}

// importNewCategory импортирует новую категорию
func importNewCategory(fbCategory *fb.Category, categoryCRUD *db.CategoryCRUD, migrationTracker *db.MigrationTracker) error {
	// Создаем категорию в PostgreSQL
	pgCategory := &db.Category{
		ID:     fbCategory.ID,
		Title:  fbCategory.Label,
		UserID: fbCategory.UserID,
	}

	err := categoryCRUD.CreateCategory(pgCategory)
	if err != nil {
		// Логируем неудачную попытку
		if logErr := migrationTracker.LogMigration("category", fbCategory.ID, fbCategory.UserID, "import", "failed", err.Error()); logErr != nil {
			// Log the error but don't fail the operation
			fmt.Printf("Failed to log migration: %v", logErr)
		}
		return fmt.Errorf("failed to create category: %v", err)
	}

	// Логируем успешную миграцию
	return migrationTracker.LogMigration("category", fbCategory.ID, fbCategory.UserID, "import", "success", "")
}

// updateExistingCategory обновляет существующую категорию
func updateExistingCategory(fbCategory *fb.Category, categoryCRUD *db.CategoryCRUD, migrationTracker *db.MigrationTracker) error {
	// Обновляем категорию в PostgreSQL
	pgCategory := &db.Category{
		ID:     fbCategory.ID,
		Title:  fbCategory.Label,
		UserID: fbCategory.UserID,
	}

	err := categoryCRUD.UpdateCategory(pgCategory)
	if err != nil {
		// Логируем неудачную попытку
		if logErr := migrationTracker.LogMigration("category", fbCategory.ID, fbCategory.UserID, "update", "failed", err.Error()); logErr != nil {
			// Log the error but don't fail the operation
			fmt.Printf("Failed to log migration: %v", logErr)
		}
		return fmt.Errorf("failed to update category: %v", err)
	}

	// Логируем успешную миграцию
	return migrationTracker.LogMigration("category", fbCategory.ID, fbCategory.UserID, "update", "success", "")
}

// ImportCategoriesForAllUsers импортирует категории для всех пользователей
// Полезно для массовой миграции
func ImportCategoriesForAllUsers() (*ImportCategoriesResult, error) {
	// TODO: Реализовать получение списка всех пользователей из Firebase
	// и импорт категорий для каждого пользователя
	return nil, fmt.Errorf("not implemented yet")
}

// GetImportStatus получает статус импорта для пользователя
func GetImportStatus(userID string) (map[string]interface{}, error) {
	migrationTracker := &db.MigrationTracker{}

	stats, err := migrationTracker.GetMigrationStats("category")
	if err != nil {
		return nil, fmt.Errorf("failed to get migration stats: %v", err)
	}

	// Получаем последние миграции для пользователя
	// TODO: Добавить метод для получения миграций по пользователю

	return map[string]interface{}{
		"stats":  stats,
		"userId": userID,
	}, nil
}
