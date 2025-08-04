package auth

import (
	"testing"
)

func TestGenerateAndValidateToken(t *testing.T) {
	userID := "test_user_123"
	email := "test@example.com"

	// Генерируем токен
	token, err := GenerateToken(userID, email)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token is empty")
	}

	// Валидируем токен
	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	// Проверяем claims
	if claims.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("Expected Email %s, got %s", email, claims.Email)
	}
}

func TestValidateInvalidToken(t *testing.T) {
	// Тестируем неверный токен
	_, err := ValidateToken("invalid.token.here")
	if err == nil {
		t.Fatal("Expected error for invalid token, got nil")
	}
}

func TestTokenExpiration(t *testing.T) {
	// Устанавливаем короткое время жизни токена для теста
	originalSecret := jwtSecret
	defer func() { jwtSecret = originalSecret }()

	userID := "test_user_123"
	email := "test@example.com"

	// Генерируем токен
	token, err := GenerateToken(userID, email)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Валидируем токен сразу
	_, err = ValidateToken(token)
	if err != nil {
		t.Fatalf("Token should be valid immediately: %v", err)
	}
}

func TestSetJWTSecret(t *testing.T) {
	originalSecret := jwtSecret
	defer func() { jwtSecret = originalSecret }()

	newSecret := "new-secret-key"
	SetJWTSecret(newSecret)

	if string(jwtSecret) != newSecret {
		t.Errorf("Expected secret %s, got %s", newSecret, string(jwtSecret))
	}
}
