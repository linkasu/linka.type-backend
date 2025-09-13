package data

import (
	"net/http"
	"strings"

	"linka.type-backend/auth"
	"linka.type-backend/bl/services"

	"github.com/gin-gonic/gin"
)

// GetCategories получает все категории пользователя
func GetCategories(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	categoryService := services.NewCategoryService()
	categories, err := categoryService.GetCategoriesByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

// GetCategory получает конкретную категорию
func GetCategory(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	categoryID := c.Param("id")

	categoryService := services.NewCategoryService()
	category, err := categoryService.GetCategoryByID(categoryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	// Проверяем, что категория принадлежит пользователю
	if category.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// CreateCategory создает новую категорию
func CreateCategory(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Дополнительная валидация
	if strings.TrimSpace(req.Title) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title cannot be empty"})
		return
	}

	categoryService := services.NewCategoryService()
	category, err := categoryService.CreateCategory(req.Title, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
		return
	}

	// Отправляем WebSocket уведомление
	NotifyCategoryCreated(userID, category)

	c.JSON(http.StatusCreated, category)
}

// UpdateCategory обновляет категорию
func UpdateCategory(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	categoryID := c.Param("id")

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Дополнительная валидация
	if strings.TrimSpace(req.Title) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title cannot be empty"})
		return
	}

	categoryService := services.NewCategoryService()
	category, err := categoryService.UpdateCategory(categoryID, req.Title, userID)
	if err != nil {
		if err.Error() == "category not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		} else if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		}
		return
	}

	// Отправляем WebSocket уведомление
	NotifyCategoryUpdated(userID, category)

	c.JSON(http.StatusOK, category)
}

// DeleteCategory удаляет категорию
func DeleteCategory(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	categoryID := c.Param("id")

	categoryService := services.NewCategoryService()
	err := categoryService.DeleteCategory(categoryID, userID)
	if err != nil {
		if err.Error() == "category not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		} else if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		}
		return
	}

	// Отправляем WebSocket уведомление
	NotifyCategoryDeleted(userID, categoryID)

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}