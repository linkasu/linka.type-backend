package bl

import (
	"linka.type-backend/bl/services"
)

// ImportStatementsResult содержит результат импорта statements
type ImportStatementsResult = services.ImportStatementsResult

// ImportStatements импортирует statements пользователя из Firebase в PostgreSQL
func ImportStatements(login, password string) (*ImportStatementsResult, error) {
	importService := services.NewImportService()
	return importService.ImportStatements(login, password)
}

// ImportAllData импортирует пользователя, категории и statements
func ImportAllData(login, password string) (*ImportAllDataResult, error) {
	importService := services.NewImportService()
	return importService.ImportAllData(login, password)
}

// ImportAllDataResult содержит результат полного импорта данных
type ImportAllDataResult = services.ImportAllDataResult