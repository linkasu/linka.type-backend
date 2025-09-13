package auth

import (
	"log"
	"net/http"
	"time"

	"linka.type-backend/db"
	"linka.type-backend/db/repositories"
	"linka.type-backend/mail"
	"linka.type-backend/otp"
	"linka.type-backend/utils"

	"github.com/gin-gonic/gin"
)

// RequestPasswordReset обрабатывает запрос на сброс пароля
func RequestPasswordReset(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем, существует ли пользователь
	userCRUD := &repositories.UserCRUD{}
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
	otpCRUD := &repositories.OTPCRUD{}
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
	otpCRUD := &repositories.OTPCRUD{}
	otpRecord, err := otpCRUD.GetOTPByCodeAnyStatus(req.Code, req.Email, "reset_password")
	if err != nil {
		// Если OTP не найден, это ошибка клиента (400), иначе внутренняя ошибка (500)
		if err.Error() == "OTP not found" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired OTP code"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
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
	otpCRUD := &repositories.OTPCRUD{}
	otpRecord, err := otpCRUD.GetOTPByCodeAnyStatus(req.Code, req.Email, "reset_password")
	if err != nil {
		// Если OTP не найден, это ошибка клиента (400), иначе внутренняя ошибка (500)
		if err.Error() == "OTP not found" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired OTP code"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
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
	userCRUD := &repositories.UserCRUD{}
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
