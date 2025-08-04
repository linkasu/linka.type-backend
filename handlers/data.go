package handlers

import (
	"net/http"

	"linka.type-backend/auth"
	"linka.type-backend/db"
	"linka.type-backend/utils"

	"github.com/gin-gonic/gin"
)

// CreateStatementRequest структура для создания statement
type CreateStatementRequest struct {
	Title      string `json:"title" binding:"required"`
	CategoryID string `json:"categoryId" binding:"required"`
}

// UpdateStatementRequest структура для обновления statement
type UpdateStatementRequest struct {
	Title      string `json:"title" binding:"required"`
	CategoryID string `json:"categoryId" binding:"required"`
}

// CreateCategoryRequest структура для создания категории
type CreateCategoryRequest struct {
	Title string `json:"title" binding:"required"`
}

// UpdateCategoryRequest структура для обновления категории
type UpdateCategoryRequest struct {
	Title string `json:"title" binding:"required"`
}

// GetStatements получает все statements пользователя
func GetStatements(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	statementCRUD := &db.StatementCRUD{}
	statements, err := statementCRUD.GetStatementsByUserID(userID)
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

	statementCRUD := &db.StatementCRUD{}
	statement, err := statementCRUD.GetStatementByID(statementID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Statement not found"})
		return
	}

	// Проверяем, что statement принадлежит пользователю
	if statement.UserId != userID {
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

	// Проверяем, что категория принадлежит пользователю
	categoryCRUD := &db.CategoryCRUD{}
	category, err := categoryCRUD.GetCategoryByID(req.CategoryID)
	if err != nil || category.UserId != userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		return
	}

	statement := &db.Statement{
		ID:         utils.GenerateID(),
		Title:      req.Title,
		UserId:     userID,
		CategoryId: req.CategoryID,
	}

	statementCRUD := &db.StatementCRUD{}
	if err := statementCRUD.CreateStatement(statement); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create statement"})
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

	// Проверяем, что statement принадлежит пользователю
	statementCRUD := &db.StatementCRUD{}
	existingStatement, err := statementCRUD.GetStatementByID(statementID)
	if err != nil || existingStatement.UserId != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Statement not found"})
		return
	}

	// Проверяем, что категория принадлежит пользователю
	categoryCRUD := &db.CategoryCRUD{}
	category, err := categoryCRUD.GetCategoryByID(req.CategoryID)
	if err != nil || category.UserId != userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		return
	}

	statement := &db.Statement{
		ID:         statementID,
		Title:      req.Title,
		UserId:     userID,
		CategoryId: req.CategoryID,
	}

	if err := statementCRUD.UpdateStatement(statement); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update statement"})
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

	// Проверяем, что statement принадлежит пользователю
	statementCRUD := &db.StatementCRUD{}
	statement, err := statementCRUD.GetStatementByID(statementID)
	if err != nil || statement.UserId != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Statement not found"})
		return
	}

	if err := statementCRUD.DeleteStatement(statementID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete statement"})
		return
	}

	// Отправляем WebSocket уведомление
	NotifyStatementDeleted(userID, statementID)

	c.JSON(http.StatusOK, gin.H{"message": "Statement deleted successfully"})
}

// GetCategories получает все категории пользователя
func GetCategories(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	categoryCRUD := &db.CategoryCRUD{}
	categories, err := categoryCRUD.GetCategoriesByUserID(userID)
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

	categoryCRUD := &db.CategoryCRUD{}
	category, err := categoryCRUD.GetCategoryByID(categoryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	// Проверяем, что категория принадлежит пользователю
	if category.UserId != userID {
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

	category := &db.Category{
		ID:     utils.GenerateID(),
		Title:  req.Title,
		UserId: userID,
	}

	categoryCRUD := &db.CategoryCRUD{}
	if err := categoryCRUD.CreateCategory(category); err != nil {
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

	// Проверяем, что категория принадлежит пользователю
	categoryCRUD := &db.CategoryCRUD{}
	existingCategory, err := categoryCRUD.GetCategoryByID(categoryID)
	if err != nil || existingCategory.UserId != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	category := &db.Category{
		ID:     categoryID,
		Title:  req.Title,
		UserId: userID,
	}

	if err := categoryCRUD.UpdateCategory(category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
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

	// Проверяем, что категория принадлежит пользователю
	categoryCRUD := &db.CategoryCRUD{}
	category, err := categoryCRUD.GetCategoryByID(categoryID)
	if err != nil || category.UserId != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	if err := categoryCRUD.DeleteCategory(categoryID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}

	// Отправляем WebSocket уведомление
	NotifyCategoryDeleted(userID, categoryID)

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}
