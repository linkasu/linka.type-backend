package auth

import (
	"net/http"
	"strings"

	"linka.type-backend/db/repositories"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware проверяет JWT токен
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Проверяем формат "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Добавляем информацию о пользователе в контекст
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Next()
	}
}

// EmailVerifiedMiddleware проверяет, что email пользователя верифицирован
func EmailVerifiedMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Получаем пользователя из базы данных
		userCRUD := &repositories.UserCRUD{}
		user, err := userCRUD.GetUserByID(userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			c.Abort()
			return
		}

		// Проверяем верификацию email
		if !user.EmailVerified {
			c.JSON(http.StatusForbidden, gin.H{"error": "Email not verified"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserIDFromContext получает user_id из контекста Gin
func GetUserIDFromContext(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return userID.(string)
	}
	return ""
}

// GetEmailFromContext получает email из контекста Gin
func GetEmailFromContext(c *gin.Context) string {
	if email, exists := c.Get("user_email"); exists {
		return email.(string)
	}
	return ""
}

// JWTAuthMiddleware является алиасом для AuthMiddleware
func JWTAuthMiddleware() gin.HandlerFunc {
	return AuthMiddleware()
}
