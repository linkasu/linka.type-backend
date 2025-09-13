package auth

import (
	"log"
	"net/http"
	"time"

	"linka.type-backend/auth"
	"linka.type-backend/bl/services"
	"linka.type-backend/db/repositories"
	"linka.type-backend/mail"
	"linka.type-backend/models"
	"linka.type-backend/otp"
	"linka.type-backend/utils"

	"github.com/gin-gonic/gin"
)

// RegisterDirect обрабатывает прямую регистрацию пользователя без OTP
func RegisterDirect(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Создаем пользователя через сервис
	authService := services.NewAuthService()
	user, err := authService.CreateUser(req.Email, req.Password, true)
	if err != nil {
		if err.Error() == "user already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		} else if err.Error() == "weak password" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Weak password"})
		} else {
			log.Printf("Error creating user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		}
		return
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

	c.JSON(http.StatusCreated, response)
}

// Register обрабатывает регистрацию пользователя
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Создаем пользователя через сервис
	authService := services.NewAuthService()
	user, err := authService.CreateUser(req.Email, req.Password, false)
	if err != nil {
		if err.Error() == "user already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		} else if err.Error() == "weak password" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Weak password"})
		} else {
			log.Printf("Error creating user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		}
		return
	}

	// Генерируем OTP код
	otpCode, err := otp.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
		return
	}

	// Сохраняем OTP в базу данных
	otpCRUD := &repositories.OTPCRUD{}
	otpRecord := &models.OTPCode{
		ID:        utils.GenerateID(),
		Email:     req.Email,
		Code:      otpCode,
		Type:      "registration",
		ExpiresAt: otp.GetOTPExpirationTime().Format(time.RFC3339),
		Used:      false,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	if err := otpCRUD.CreateOTP(otpRecord); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save OTP"})
		return
	}

	// Отправляем OTP на email
	if err := mail.SendOTPEmail(req.Email, otpCode, "registration"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP email"})
		return
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

	c.JSON(http.StatusCreated, response)
}
