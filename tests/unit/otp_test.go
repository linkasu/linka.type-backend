package unit

import (
	"testing"
	"time"

	"linka.type-backend/otp"
)

func TestGenerateOTP(t *testing.T) {
	code, err := otp.GenerateOTP()
	if err != nil {
		t.Fatalf("Failed to generate OTP: %v", err)
	}

	if len(code) != 6 {
		t.Errorf("Expected OTP length 6, got %d", len(code))
	}

	// Проверяем, что код состоит только из цифр
	for _, char := range code {
		if char < '0' || char > '9' {
			t.Errorf("OTP contains non-digit character: %c", char)
		}
	}
}

func TestValidateOTPCode(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"valid 6-digit code", "123456", true},
		{"valid 6-digit code with zeros", "000000", true},
		{"too short", "12345", false},
		{"too long", "1234567", false},
		{"contains letters", "12345a", false},
		{"contains special chars", "12345!", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := otp.ValidateOTPCode(tt.code)
			if result != tt.expected {
				t.Errorf("ValidateOTPCode(%s) = %v, expected %v", tt.code, result, tt.expected)
			}
		})
	}
}

func TestIsOTPExpired(t *testing.T) {
	// Тест для истекшего OTP
	pastTime := time.Now().Add(-1 * time.Hour)
	if !otp.IsOTPExpired(pastTime) {
		t.Error("Expected expired OTP to be marked as expired")
	}

	// Тест для действующего OTP
	futureTime := time.Now().Add(1 * time.Hour)
	if otp.IsOTPExpired(futureTime) {
		t.Error("Expected valid OTP to not be marked as expired")
	}
}

func TestGetOTPExpirationTime(t *testing.T) {
	expirationTime := otp.GetOTPExpirationTime()
	now := time.Now()

	// Проверяем, что время истечения в будущем
	if expirationTime.Before(now) {
		t.Error("Expiration time should be in the future")
	}

	// Проверяем, что время истечения примерно через 15 минут
	expectedMin := now.Add(14 * time.Minute)
	expectedMax := now.Add(16 * time.Minute)

	if expirationTime.Before(expectedMin) || expirationTime.After(expectedMax) {
		t.Errorf("Expiration time %v should be around 15 minutes from now", expirationTime)
	}
}
