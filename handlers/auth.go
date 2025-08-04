package handlers

import (
	"crypto/md5"
	"fmt"
	"net/http"

	"linka.type-backend/auth"
	"linka.type-backend/db"
	"linka.type-backend/utils"

	"github.com/gin-gonic/gin"
)

// LoginRequest структура для запроса логина
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest структура для запроса регистрации
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginResponse структура для ответа при логине
type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

// Register обрабатывает регистрацию пользователя
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем, существует ли пользователь
	userCRUD := &db.UserCRUD{}
	exists, err := userCRUD.UserExists(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// Хешируем пароль (MD5 как в существующей системе)
	hashedPassword := fmt.Sprintf("%x", md5.Sum([]byte(req.Password)))

	// Создаем пользователя
	user := &db.User{
		ID:       utils.GenerateID(), // Генерируем уникальный ID
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := userCRUD.CreateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Генерируем JWT токен
	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	response := LoginResponse{
		Token: token,
		User: struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		}{
			ID:    user.ID,
			Email: user.Email,
		},
	}

	c.JSON(http.StatusCreated, response)
}

// Login обрабатывает логин пользователя
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем пользователя по email
	userCRUD := &db.UserCRUD{}
	user, err := userCRUD.GetUserByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Проверяем пароль
	hashedPassword := fmt.Sprintf("%x", md5.Sum([]byte(req.Password)))
	if user.Password != hashedPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Генерируем JWT токен
	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	response := LoginResponse{
		Token: token,
		User: struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		}{
			ID:    user.ID,
			Email: user.Email,
		},
	}

	c.JSON(http.StatusOK, response)
}
