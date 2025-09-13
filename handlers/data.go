package handlers

import (
	"github.com/gin-gonic/gin"
	"linka.type-backend/handlers/data"
)

// GetStatements получает все statements пользователя
func GetStatements(c *gin.Context) {
	data.GetStatements(c)
}

// GetStatement получает конкретный statement
func GetStatement(c *gin.Context) {
	data.GetStatement(c)
}

// CreateStatement создает новый statement
func CreateStatement(c *gin.Context) {
	data.CreateStatement(c)
}

// UpdateStatement обновляет statement
func UpdateStatement(c *gin.Context) {
	data.UpdateStatement(c)
}

// DeleteStatement удаляет statement
func DeleteStatement(c *gin.Context) {
	data.DeleteStatement(c)
}

// GetCategories получает все категории пользователя
func GetCategories(c *gin.Context) {
	data.GetCategories(c)
}

// GetCategory получает конкретную категорию
func GetCategory(c *gin.Context) {
	data.GetCategory(c)
}

// CreateCategory создает новую категорию
func CreateCategory(c *gin.Context) {
	data.CreateCategory(c)
}

// UpdateCategory обновляет категорию
func UpdateCategory(c *gin.Context) {
	data.UpdateCategory(c)
}

// DeleteCategory удаляет категорию
func DeleteCategory(c *gin.Context) {
	data.DeleteCategory(c)
}

// GetCategoryHash возвращает хеш для конкретной категории или всех категорий
func GetCategoryHash(c *gin.Context) {
	data.GetCategoryHash(c)
}

// CreateEvent создает новое событие
func CreateEvent(c *gin.Context) {
	data.CreateEvent(c)
}
