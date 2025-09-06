package handlers

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"time"

	"linka.type-backend/auth"
	"linka.type-backend/bl"
	"linka.type-backend/db"
	"linka.type-backend/fb"
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
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	User         struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

// RefreshTokenRequest структура для запроса обновления токена
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RegisterDirect обрабатывает прямую регистрацию пользователя без OTP
func RegisterDirect(c *gin.Context) {
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

	// Проверяем силу пароля и хешируем (bcrypt)
	hasher := utils.NewPasswordHasher()
	if ok, _ := hasher.ValidatePasswordStrength(req.Password); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Weak password"})
		return
	}
	hashedPassword, err := hasher.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Создаем пользователя (email верифицирован автоматически)
	user := &db.User{
		ID:            utils.GenerateID(),
		Email:         req.Email,
		Password:      hashedPassword,
		EmailVerified: true, // Автоматически верифицируем для прямой регистрации
	}

	if err := userCRUD.CreateUser(user); err != nil {
		log.Printf("Error creating user: %v", err)
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

	// Проверяем силу пароля и хешируем (bcrypt)
	hasher := utils.NewPasswordHasher()
	if ok, _ := hasher.ValidatePasswordStrength(req.Password); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Weak password"})
		return
	}
	hashedPassword, err := hasher.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Создаем пользователя (email не верифицирован)
	user := &db.User{
		ID:            utils.GenerateID(),
		Email:         req.Email,
		Password:      hashedPassword,
		EmailVerified: false,
	}

	if err := userCRUD.CreateUser(user); err != nil {
		log.Printf("Error creating user: %v", err)
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

	log.Printf("Login attempt for email: %s", req.Email)

	// Получаем пользователя по email
	userCRUD := &db.UserCRUD{}
	user, err := userCRUD.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("User not found in PostgreSQL: %s, error: %v", req.Email, err)
		// Пользователь не найден в PostgreSQL, но может существовать в Firebase
		// Продолжаем проверку Firebase
	} else {
		log.Printf("User found in PostgreSQL: %s", user.ID)
	}

	// Проверяем пароль в PostgreSQL
	passwordOK := checkPostgreSQLPassword(user, req.Password, userCRUD)

	// Если пароль в PostgreSQL неверный, пробуем Firebase
	var firebasePasswordOK *fb.FirebaseAuthResponse
	if !passwordOK {
		firebasePasswordOK = checkFirebasePassword(req.Email, req.Password)
	}

	// Если аутентификация не прошла нигде, возвращаем ошибку
	if !passwordOK && firebasePasswordOK == nil {
		log.Printf("Authentication failed in both PostgreSQL and Firebase")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Обрабатываем успешную аутентификацию
	if passwordOK {
		log.Printf("Authentication successful via PostgreSQL")
	} else if firebasePasswordOK != nil {
		log.Printf("Authentication successful via Firebase")
		handleFirebaseAuth(user, req, userCRUD)
	}

	// Импортируем данные из Firebase если необходимо
	if firebasePasswordOK != nil {
		importFirebaseData(req.Email, req.Password, user, userCRUD, c)
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

	// Получаем OTP из базы данных (включая уже использованный, чтобы корректно отработать повтор)
	otpCRUD := &db.OTPCRUD{}
	otpRecord, err := otpCRUD.GetOTPByCodeAnyStatus(req.Code, req.Email, "reset_password")
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

	// Если код ещё не использован — помечаем использованным, чтобы сделать одноразовым
	if !otpRecord.Used {
		if err := otpCRUD.MarkOTPAsUsed(otpRecord.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark OTP as used"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP code has already been used"})
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

	// Получаем OTP из базы данных независимо от статуса used,
	// так как мы допускаем, что код мог быть отмечен использованным на этапе verify
	otpCRUD := &db.OTPCRUD{}
	otpRecord, err := otpCRUD.GetOTPByCodeAnyStatus(req.Code, req.Email, "reset_password")
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

	// На шаге confirm повторная проверка used не блокирует, т.к. verify уже мог отметить код

	// Получаем пользователя
	userCRUD := &db.UserCRUD{}
	user, err := userCRUD.GetUserByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Проверяем силу нового пароля и хешируем (bcrypt)
	hasher := utils.NewPasswordHasher()
	if ok, _ := hasher.ValidatePasswordStrength(req.Password); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Weak password"})
		return
	}
	hashedPassword, err := hasher.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Обновляем пароль пользователя
	if err := userCRUD.UpdateUserPassword(user.ID, hashedPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// Повторно отмечать не нужно; если verify не вызывали, код уже отмечен выше. Если вызывали — уже used.

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}

// checkPostgreSQLPassword проверяет пароль в PostgreSQL
func checkPostgreSQLPassword(user *db.User, password string, userCRUD *db.UserCRUD) bool {
	if user == nil {
		log.Printf("User not found in PostgreSQL, skipping password check")
		return false
	}

	hasher := utils.NewPasswordHasher()
	passwordOK := hasher.CheckPassword(user.Password, password) == nil
	log.Printf("Bcrypt password check result: %v", passwordOK)

	if !passwordOK {
		legacy := fmt.Sprintf("%x", md5.Sum([]byte(password)))
		log.Printf("Trying legacy MD5 check for user: %s", user.ID)
		if user.Password == legacy {
			passwordOK = true
			log.Printf("Legacy MD5 password match, updating to bcrypt")
			if newHash, err := hasher.HashPassword(password); err == nil {
				_ = userCRUD.UpdateUserPassword(user.ID, newHash)
			}
		}
	}

	log.Printf("PostgreSQL password check final result: %v", passwordOK)
	return passwordOK
}

// checkFirebasePassword проверяет пароль в Firebase
func checkFirebasePassword(email, password string) *fb.FirebaseAuthResponse {
	log.Printf("PostgreSQL password check failed, trying Firebase authentication")
	firebasePasswordOK, err := fb.CheckPassword(email, password)
	if err != nil {
		log.Printf("Firebase password check error: %v", err)
		return nil
	}
	log.Printf("Firebase authentication successful")
	return firebasePasswordOK
}

// handleFirebaseAuth обрабатывает успешную аутентификацию через Firebase
func handleFirebaseAuth(user *db.User, req LoginRequest, userCRUD *db.UserCRUD) {
	if user == nil {
		log.Printf("User not found in PostgreSQL but Firebase auth successful, importing user")
	} else {
		log.Printf("User found in PostgreSQL but password incorrect, updating password and importing data")
		hasher := utils.NewPasswordHasher()
		if newHash, err := hasher.HashPassword(req.Password); err == nil {
			_ = userCRUD.UpdateUserPassword(user.ID, newHash)
			log.Printf("Updated password for user: %s", user.ID)
		}
	}
}

// importFirebaseData импортирует данные из Firebase
func importFirebaseData(email, password string, user *db.User, userCRUD *db.UserCRUD, c *gin.Context) {
	fbUser, err := fb.GetUser(email)
	if err == nil && fbUser != nil {
		importResult, err := bl.ImportAllData(email, password)
		if err != nil {
			log.Printf("Failed to import data from Firebase: %v", err)
		} else {
			log.Printf("Successfully imported data from Firebase: %+v", importResult)
		}

		if user == nil {
			user, err = userCRUD.GetUserByEmail(email)
			if err != nil {
				log.Printf("Failed to get imported user from PostgreSQL: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user after import"})
				return
			}
			log.Printf("Successfully retrieved imported user: %s", user.ID)
		}
	}
}

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
