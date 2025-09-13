package auth

import (
	"log"
	"net/http"

	"linka.type-backend/auth"
	"linka.type-backend/bl/services"

	"github.com/gin-gonic/gin"
)

// Login обрабатывает логин пользователя
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Login attempt for email: %s", req.Email)

	// Аутентифицируем пользователя через сервис
	authService := services.NewAuthService()
	user, err := authService.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		log.Printf("Authentication failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Импортируем данные из Firebase если необходимо
	importService := services.NewImportService()
	_, err = importService.ImportAllData(req.Email, req.Password)
	if err != nil {
		log.Printf("Failed to import data from Firebase: %v", err)
	}

	// Генерируем пару токенов (access + refresh)
	tokenPair, err := auth.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	response := LoginResponse{
		Token:        tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
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

