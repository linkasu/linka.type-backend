package bl

import (
	"linka.type-backend/bl/services"
)

// ImportUser импортирует пользователя и его категории из Firebase в PostgreSQL
func ImportUser(login, password string) error {
	importService := services.NewImportService()
	return importService.ImportUser(login, password)
}