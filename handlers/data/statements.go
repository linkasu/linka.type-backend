package data

import (
	"net/http"
	"strings"

	"linka.type-backend/auth"
	"linka.type-backend/bl/services"

	"github.com/gin-gonic/gin"
)

// GetStatements получает все statements пользователя
func GetStatements(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	statementService := services.NewStatementService()
	statements, err := statementService.GetStatementsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get statements"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"statements": statements})
}

// GetStatement получает конкретный statement
func GetStatement(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	statementID := c.Param("id")

	statementService := services.NewStatementService()
	statement, err := statementService.GetStatementByID(statementID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Statement not found"})
		return
	}

	// Проверяем, что statement принадлежит пользователю
	if statement.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, statement)
}

// CreateStatement создает новый statement
func CreateStatement(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req CreateStatementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Дополнительная валидация
	if strings.TrimSpace(req.Title) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title cannot be empty"})
		return
	}

	if strings.TrimSpace(req.CategoryID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category ID cannot be empty"})
		return
	}

	statementService := services.NewStatementService()
	statement, err := statementService.CreateStatement(req.Title, req.CategoryID, userID)
	if err != nil {
		if err.Error() == "access denied" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create statement"})
		}
		return
	}

	// Отправляем WebSocket уведомление
	NotifyStatementCreated(userID, statement)

	c.JSON(http.StatusCreated, statement)
}

// UpdateStatement обновляет statement
func UpdateStatement(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	statementID := c.Param("id")

	var req UpdateStatementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Дополнительная валидация
	if strings.TrimSpace(req.Title) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title cannot be empty"})
		return
	}

	if strings.TrimSpace(req.CategoryID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category ID cannot be empty"})
		return
	}

	statementService := services.NewStatementService()
	statement, err := statementService.UpdateStatement(statementID, req.Title, req.CategoryID, userID)
	if err != nil {
		if err.Error() == "statement not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Statement not found"})
		} else if err.Error() == "access denied" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update statement"})
		}
		return
	}

	// Отправляем WebSocket уведомление
	NotifyStatementUpdated(userID, statement)

	c.JSON(http.StatusOK, statement)
}

// DeleteStatement удаляет statement
func DeleteStatement(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	statementID := c.Param("id")

	statementService := services.NewStatementService()
	err := statementService.DeleteStatement(statementID, userID)
	if err != nil {
		if err.Error() == "statement not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Statement not found"})
		} else if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete statement"})
		}
		return
	}

	// Отправляем WebSocket уведомление
	NotifyStatementDeleted(userID, statementID)

	c.JSON(http.StatusOK, gin.H{"message": "Statement deleted successfully"})
}