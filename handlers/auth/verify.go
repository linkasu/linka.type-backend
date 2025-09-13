package auth

import (
	"log"
	"net/http"
	"time"

	"linka.type-backend/auth"
	"linka.type-backend/db"
	"linka.type-backend/mail"
	"linka.type-backend/otp"

	"github.com/gin-gonic/gin"
)

// VerifyEmail обрабатывает верификацию email
func VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем корректность OTP кода
	if !otp.ValidateOTPCode(req.Code) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTP code format"})
		return
	}

	// Получаем OTP из базы данных
	otpCRUD := &db.OTPCRUD{}
	otpRecord, err := otpCRUD.GetOTPByCode(req.Code, req.Email, "registration")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if otpRecord == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired OTP code"})
		return
	}

	// Проверяем, не истек ли срок действия
	expiresAt, err := time.Parse(time.RFC3339, otpRecord.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid date format"})
		return
	}

	if otp.IsOTPExpired(expiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP code has expired"})
		return
	}

	// Получаем пользователя
	userCRUD := &db.UserCRUD{}
	user, err := userCRUD.GetUserByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Верифицируем email пользователя
	if err := userCRUD.VerifyUserEmail(user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email"})
		return
	}

	// Помечаем OTP как использованный
	if err := otpCRUD.MarkOTPAsUsed(otpRecord.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark OTP as used"})
		return
	}

	// Отправляем приветственное письмо
	go func() {
		if err := mail.SendWelcomeEmail(req.Email); err != nil {
			log.Printf("Failed to send welcome email: %v", err)
		}
	}()

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

	c.JSON(http.StatusOK, gin.H{
		"message":       "Email verified successfully",
		"token":         response.Token,
		"refresh_token": response.RefreshToken,
		"user":          response.User,
	})
}