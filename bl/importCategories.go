package bl

import (
	"fmt"

	"linka.type-backend/bl/services"
)

// ImportCategoriesResult содержит результат импорта категорий
type ImportCategoriesResult = services.ImportCategoriesResult

// ImportError содержит информацию об ошибке импорта
type ImportError = services.ImportError

// ImportCategories импортирует категории пользователя из Firebase в PostgreSQL
func ImportCategories(login, password string) (*ImportCategoriesResult, error) {
	importService := services.NewImportService()
	return importService.ImportCategories(login, password)
}

// ImportCategoriesForAllUsers импортирует категории для всех пользователей
func ImportCategoriesForAllUsers() (*ImportCategoriesResult, error) {
	// TODO: Реализовать получение списка всех пользователей из Firebase
	// и импорт категорий для каждого пользователя
	return nil, fmt.Errorf("not implemented yet")
}

// GetImportStatus получает статус импорта для пользователя
func GetImportStatus(userID string) (map[string]interface{}, error) {
	// TODO: Добавить метод для получения статуса импорта
	return map[string]interface{}{
		"userId": userID,
	}, nil
}