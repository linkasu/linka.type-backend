package otp

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// OTPCode представляет OTP код
type Code struct {
	Code      string    `json:"code"`
	Email     string    `json:"email"`
	Type      string    `json:"type"` // "registration" или "reset_password"
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
}

// GenerateOTP генерирует 6-значный OTP код
func GenerateOTP() (string, error) {
	// Генерируем случайное число от 100000 до 999999
	maxRange := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, maxRange)
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %v", err)
	}

	// Добавляем 100000 чтобы получить 6-значное число
	code := n.Add(n, big.NewInt(100000))
	return code.String(), nil
}

// IsOTPExpired проверяет, истек ли срок действия OTP
func IsOTPExpired(expiresAt time.Time) bool {
	return time.Now().After(expiresAt)
}

// GetOTPExpirationTime возвращает время истечения OTP (15 минут)
func GetOTPExpirationTime() time.Time {
	return time.Now().Add(15 * time.Minute)
}

// ValidateOTPCode проверяет корректность OTP кода
func ValidateOTPCode(code string) bool {
	if len(code) != 6 {
		return false
	}

	for _, char := range code {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}
