package unit

import (
	"testing"

	"linka.type-backend/auth"

	"github.com/stretchr/testify/assert"
)

func TestJWTTokenGeneration(t *testing.T) {
	userID := "test_user_123"
	email := "test@example.com"

	// Генерируем токен
	token, err := auth.GenerateToken(userID, email)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Валидируем токен
	claims, err := auth.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
}

func TestJWTTokenValidation(t *testing.T) {
	// Тест с неверным токеном
	_, err := auth.ValidateToken("invalid.token.here")
	assert.Error(t, err)

	// Тест с пустым токеном
	_, err = auth.ValidateToken("")
	assert.Error(t, err)
}

func TestJWTSecretChange(t *testing.T) {
	userID := "test_user_123"
	email := "test@example.com"

	// Генерируем токен с дефолтным секретом
	token1, err := auth.GenerateToken(userID, email)
	assert.NoError(t, err)

	// Меняем секрет
	auth.SetJWTSecret("new-secret-key")

	// Генерируем токен с новым секретом
	token2, err := auth.GenerateToken(userID, email)
	assert.NoError(t, err)

	// Токены должны быть разными
	assert.NotEqual(t, token1, token2)

	// Валидируем токен с новым секретом
	claims, err := auth.ValidateToken(token2)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
}
