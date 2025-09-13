package data

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"linka.type-backend/auth"
	"linka.type-backend/bl/services"
	"linka.type-backend/models"

	"github.com/gin-gonic/gin"
)

// generateCategoryHash генерирует хеш для категории (ID + title)
func generateCategoryHash(category *models.Category) string {
	data := category.ID + category.Title
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// generateAllCategoriesHash генерирует хеш для всех категорий пользователя
func generateAllCategoriesHash(categories []*models.Category) string {
	// Сортируем категории по ID для консистентности
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].ID < categories[j].ID
	})

	var data strings.Builder
	for _, category := range categories {
		data.WriteString(category.ID)
		data.WriteString(category.Title)
	}

	hash := sha256.Sum256([]byte(data.String()))
	return fmt.Sprintf("%x", hash)
}

// GetCategoryHash возвращает хеш для конкретной категории или всех категорий
func GetCategoryHash(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	categoryID := c.Query("id")

	if categoryID != "" {
		// Хеш для конкретной категории
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

		hash := generateCategoryHash(category)
		c.JSON(http.StatusOK, gin.H{
			"categoryId": categoryID,
			"hash":       hash,
		})
	} else {
		// Хеш для всех категорий пользователя
		categoryService := services.NewCategoryService()
		categories, err := categoryService.GetCategoriesByUserID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get categories"})
			return
		}

		hash := generateAllCategoriesHash(categories)
		c.JSON(http.StatusOK, gin.H{
			"hash": hash,
		})
	}
}