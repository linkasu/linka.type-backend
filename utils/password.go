package utils

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher предоставляет методы для работы с паролями
type PasswordHasher struct{}

// NewPasswordHasher создает новый экземпляр PasswordHasher
func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{}
}

// HashPassword хеширует пароль с использованием bcrypt
func (ph *PasswordHasher) HashPassword(password string) (string, error) {
	// Генерируем соль и хешируем пароль
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %v", err)
	}

	return string(hashedBytes), nil
}

// CheckPassword проверяет пароль против хеша
func (ph *PasswordHasher) CheckPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("password verification failed: %v", err)
	}

	return nil
}

// GenerateRandomPassword генерирует случайный пароль заданной длины
func (ph *PasswordHasher) GenerateRandomPassword(length int) (string, error) {
	if length < 8 {
		length = 8 // Минимальная длина пароля
	}

	// Символы для генерации пароля
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"

	// Генерируем случайные байты
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random password: %v", err)
	}

	// Преобразуем байты в символы
	password := make([]byte, length)
	for i := range password {
		password[i] = charset[randomBytes[i]%byte(len(charset))]
	}

	return string(password), nil
}

// ValidatePasswordStrength проверяет сложность пароля
func (ph *PasswordHasher) ValidatePasswordStrength(password string) (bool, []string) {
	var errors []string

	// Проверяем минимальную длину
	if len(password) < 8 {
		errors = append(errors, "Password must be at least 8 characters long")
	}

	// Проверяем наличие букв
	hasLetter := false
	for _, char := range password {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			hasLetter = true
			break
		}
	}
	if !hasLetter {
		errors = append(errors, "Password must contain at least one letter")
	}

	// Проверяем наличие цифр
	hasDigit := false
	for _, char := range password {
		if char >= '0' && char <= '9' {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		errors = append(errors, "Password must contain at least one digit")
	}

	// Проверяем наличие специальных символов
	hasSpecial := false
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	for _, char := range password {
		for _, special := range specialChars {
			if char == special {
				hasSpecial = true
				break
			}
		}
		if hasSpecial {
			break
		}
	}
	if !hasSpecial {
		errors = append(errors, "Password must contain at least one special character")
	}

	return len(errors) == 0, errors
}

// GetPasswordStrength возвращает оценку сложности пароля (0-100)
func (ph *PasswordHasher) GetPasswordStrength(password string) int {
	score := 0

	// Базовая оценка за длину
	if len(password) >= 8 {
		score += 20
	}
	if len(password) >= 12 {
		score += 10
	}
	if len(password) >= 16 {
		score += 10
	}

	// Оценка за разнообразие символов
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		if char >= 'a' && char <= 'z' {
			hasLower = true
		} else if char >= 'A' && char <= 'Z' {
			hasUpper = true
		} else if char >= '0' && char <= '9' {
			hasDigit = true
		} else {
			hasSpecial = true
		}
	}

	if hasLower {
		score += 15
	}
	if hasUpper {
		score += 15
	}
	if hasDigit {
		score += 15
	}
	if hasSpecial {
		score += 15
	}

	// Ограничиваем максимальную оценку
	if score > 100 {
		score = 100
	}

	return score
}

// GetPasswordStrengthText возвращает текстовое описание сложности пароля
func (ph *PasswordHasher) GetPasswordStrengthText(strength int) string {
	switch {
	case strength >= 80:
		return "Very Strong"
	case strength >= 60:
		return "Strong"
	case strength >= 40:
		return "Medium"
	case strength >= 20:
		return "Weak"
	default:
		return "Very Weak"
	}
}
