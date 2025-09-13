package handlers

import (
	"linka.type-backend/handlers/auth"
	"github.com/gin-gonic/gin"
)

// RegisterDirect обрабатывает прямую регистрацию пользователя без OTP
func RegisterDirect(c *gin.Context) {
	auth.RegisterDirect(c)
}

// Register обрабатывает регистрацию пользователя
func Register(c *gin.Context) {
	auth.Register(c)
}

// Login обрабатывает логин пользователя
func Login(c *gin.Context) {
	auth.Login(c)
}

// VerifyEmail обрабатывает верификацию email
func VerifyEmail(c *gin.Context) {
	auth.VerifyEmail(c)
}

// RequestPasswordReset обрабатывает запрос на сброс пароля
func RequestPasswordReset(c *gin.Context) {
	auth.RequestPasswordReset(c)
}

// VerifyPasswordResetOTP верифицирует OTP для сброса пароля
func VerifyPasswordResetOTP(c *gin.Context) {
	auth.VerifyPasswordResetOTP(c)
}

// ConfirmPasswordReset подтверждает сброс пароля
func ConfirmPasswordReset(c *gin.Context) {
	auth.ConfirmPasswordReset(c)
}

// RefreshToken обрабатывает обновление токена
func RefreshToken(c *gin.Context) {
	auth.RefreshToken(c)
}