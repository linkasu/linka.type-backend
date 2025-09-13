package auth

import (
	"log"
	"net/http"

	"linka.type-backend/auth"
	"linka.type-backend/db"

	"github.com/gin-gonic/gin"
)

// RefreshToken обрабатывает обновление токена
func RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Валидируем refresh token
	claims, err := auth.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		log.Printf("Invalid refresh token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Получаем пользователя из базы данных
	userCRUD := &db.UserCRUD{}
	user, err := userCRUD.GetUserByEmail(claims.Email)
	if err != nil {
		log.Printf("User not found for refresh token: %s", claims.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Генерируем новую пару токенов
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