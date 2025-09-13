package auth

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