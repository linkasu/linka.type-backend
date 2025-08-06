package handlers

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"time"

	"linka.type-backend/auth"
	"linka.type-backend/db"
	"linka.type-backend/mail"
	"linka.type-backend/otp"
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

// VerifyEmailRequest структура для запроса верификации email
type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

// ResetPasswordRequest структура для запроса сброса пароля
type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordVerifyRequest структура для верификации OTP при сбросе пароля
type ResetPasswordVerifyRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

// ResetPasswordConfirmRequest структура для подтверждения нового пароля
type ResetPasswordConfirmRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Code     string `json:"code" binding:"required"`
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

	// Создаем пользователя (email не верифицирован)
	user := &db.User{
		ID:            utils.GenerateID(),
		Email:         req.Email,
		Password:      hashedPassword,
		EmailVerified: false,
	}

	if err := userCRUD.CreateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Генерируем OTP код
	otpCode, err := otp.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
		return
	}

	// Сохраняем OTP в базу данных
	otpCRUD := &db.OTPCRUD{}
	otpRecord := &db.OTPCode{
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

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful. Please check your email for verification code.",
		"user_id": user.ID,
	})
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

	// Проверяем, верифицирован ли email
	if !user.EmailVerified {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not verified. Please check your email for verification code."})
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

	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully",
		"token":   response.Token,
		"user":    response.User,
	})
}

// RequestPasswordReset обрабатывает запрос на сброс пароля
func RequestPasswordReset(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем, существует ли пользователь
	userCRUD := &db.UserCRUD{}
	_, err := userCRUD.GetUserByEmail(req.Email)
	if err != nil {
		// Не раскрываем информацию о существовании пользователя
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset code has been sent"})
		return
	}

	// Генерируем OTP код
	otpCode, err := otp.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
		return
	}

	// Удаляем старые OTP коды для этого email
	otpCRUD := &db.OTPCRUD{}
	if err := otpCRUD.DeleteOTPByEmail(req.Email); err != nil {
		log.Printf("Failed to delete old OTP codes: %v", err)
	}

	// Сохраняем новый OTP в базу данных
	otpRecord := &db.OTPCode{
		ID:        utils.GenerateID(),
		Email:     req.Email,
		Code:      otpCode,
		Type:      "reset_password",
		ExpiresAt: otp.GetOTPExpirationTime().Format(time.RFC3339),
		Used:      false,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	if err := otpCRUD.CreateOTP(otpRecord); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save OTP"})
		return
	}

	// Отправляем OTP на email
	if err := mail.SendOTPEmail(req.Email, otpCode, "reset_password"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If the email exists, a reset code has been sent",
	})
}

// VerifyPasswordResetOTP верифицирует OTP для сброса пароля
func VerifyPasswordResetOTP(c *gin.Context) {
	var req ResetPasswordVerifyRequest
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
	otpRecord, err := otpCRUD.GetOTPByCode(req.Code, req.Email, "reset_password")
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

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP verified successfully. You can now set a new password.",
		"otp_id":  otpRecord.ID,
	})
}

// ConfirmPasswordReset подтверждает сброс пароля
func ConfirmPasswordReset(c *gin.Context) {
	var req ResetPasswordConfirmRequest
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
	otpRecord, err := otpCRUD.GetOTPByCode(req.Code, req.Email, "reset_password")
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

	// Хешируем новый пароль
	hashedPassword := fmt.Sprintf("%x", md5.Sum([]byte(req.Password)))

	// Обновляем пароль пользователя
	if err := userCRUD.UpdateUserPassword(user.ID, hashedPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// Помечаем OTP как использованный
	if err := otpCRUD.MarkOTPAsUsed(otpRecord.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark OTP as used"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}
