package utils

import (
	"strings"
	"testing"
)

func TestPasswordHasher_HashPassword(t *testing.T) {
	hasher := NewPasswordHasher()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid password",
			password: "MySecurePassword123!",
			wantErr:  false,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "Special characters",
			password: "P@ssw0rd!@#$%^&*()",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := hasher.HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && hash == "" {
				t.Error("HashPassword() returned empty hash")
			}

			// Проверяем, что хеш начинается с $2a$ (bcrypt)
			if !strings.HasPrefix(hash, "$2a$") {
				t.Errorf("HashPassword() returned invalid hash format: %s", hash)
			}
		})
	}
}

func TestPasswordHasher_CheckPassword(t *testing.T) {
	hasher := NewPasswordHasher()

	// Создаем хеш для тестирования
	password := "TestPassword123!"
	hash, err := hasher.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to create hash for testing: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "Correct password",
			password: password,
			hash:     hash,
			wantErr:  false,
		},
		{
			name:     "Wrong password",
			password: "WrongPassword123!",
			hash:     hash,
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash,
			wantErr:  true,
		},
		{
			name:     "Invalid hash",
			password: password,
			hash:     "invalid_hash",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hasher.CheckPassword(tt.hash, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPasswordHasher_GenerateRandomPassword(t *testing.T) {
	hasher := NewPasswordHasher()

	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "Default length",
			length: 12,
		},
		{
			name:   "Short length",
			length: 8,
		},
		{
			name:   "Long length",
			length: 20,
		},
		{
			name:   "Very short length (should default to 8)",
			length: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := hasher.GenerateRandomPassword(tt.length)
			if err != nil {
				t.Errorf("GenerateRandomPassword() error = %v", err)
				return
			}

			// Проверяем длину
			expectedLength := tt.length
			if tt.length < 8 {
				expectedLength = 8
			}

			if len(password) != expectedLength {
				t.Errorf("GenerateRandomPassword() length = %d, want %d", len(password), expectedLength)
			}

			// Проверяем, что пароль содержит разные типы символов
			hasLower, hasUpper, hasDigit, hasSpecial := false, false, false, false
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

			// Проверяем, что пароль содержит хотя бы один тип символов
			if !hasLower && !hasUpper && !hasDigit && !hasSpecial {
				t.Error("Generated password contains no valid characters")
			}

			// Генерируем несколько паролей и проверяем, что они разные
			passwords := make(map[string]bool)
			for i := 0; i < 10; i++ {
				pwd, err := hasher.GenerateRandomPassword(tt.length)
				if err != nil {
					t.Errorf("Failed to generate password %d: %v", i, err)
					continue
				}
				if passwords[pwd] {
					t.Errorf("Generated duplicate password: %s", pwd)
				}
				passwords[pwd] = true
			}
		})
	}
}

func TestPasswordHasher_ValidatePasswordStrength(t *testing.T) {
	hasher := NewPasswordHasher()

	tests := []struct {
		name       string
		password   string
		wantValid  bool
		wantErrors []string
	}{
		{
			name:       "Strong password",
			password:   "MySecurePass123!",
			wantValid:  true,
			wantErrors: []string{},
		},
		{
			name:       "Too short",
			password:   "Abc1!",
			wantValid:  false,
			wantErrors: []string{"Password must be at least 8 characters long"},
		},
		{
			name:       "No letters",
			password:   "12345678!",
			wantValid:  false,
			wantErrors: []string{"Password must contain at least one letter"},
		},
		{
			name:       "No digits",
			password:   "MySecurePass!",
			wantValid:  false,
			wantErrors: []string{"Password must contain at least one digit"},
		},
		{
			name:       "No special characters",
			password:   "MySecurePass123",
			wantValid:  false,
			wantErrors: []string{"Password must contain at least one special character"},
		},
		{
			name:      "Multiple issues",
			password:  "abc",
			wantValid: false,
			wantErrors: []string{
				"Password must be at least 8 characters long",
				"Password must contain at least one digit",
				"Password must contain at least one special character",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, errors := hasher.ValidatePasswordStrength(tt.password)

			if isValid != tt.wantValid {
				t.Errorf("ValidatePasswordStrength() isValid = %v, want %v", isValid, tt.wantValid)
			}

			if len(errors) != len(tt.wantErrors) {
				t.Errorf("ValidatePasswordStrength() errors count = %d, want %d", len(errors), len(tt.wantErrors))
			}

			// Проверяем, что все ожидаемые ошибки присутствуют
			for _, expectedError := range tt.wantErrors {
				found := false
				for _, actualError := range errors {
					if actualError == expectedError {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ValidatePasswordStrength() missing error: %s", expectedError)
				}
			}
		})
	}
}

func TestPasswordHasher_GetPasswordStrength(t *testing.T) {
	hasher := NewPasswordHasher()

	tests := []struct {
		name     string
		password string
		wantMin  int
		wantMax  int
	}{
		{
			name:     "Very weak password",
			password: "abc",
			wantMin:  0,
			wantMax:  20,
		},
		{
			name:     "Weak password",
			password: "abcdefgh",
			wantMin:  20,
			wantMax:  40,
		},
		{
			name:     "Medium password",
			password: "MyPass123",
			wantMin:  40,
			wantMax:  70,
		},
		{
			name:     "Strong password",
			password: "MySecurePass123!",
			wantMin:  60,
			wantMax:  100,
		},
		{
			name:     "Very strong password",
			password: "MyVerySecurePassword123!@#",
			wantMin:  80,
			wantMax:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strength := hasher.GetPasswordStrength(tt.password)

			if strength < tt.wantMin || strength > tt.wantMax {
				t.Errorf("GetPasswordStrength() = %d, want between %d and %d", strength, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestPasswordHasher_GetPasswordStrengthText(t *testing.T) {
	hasher := NewPasswordHasher()

	tests := []struct {
		name     string
		strength int
		want     string
	}{
		{
			name:     "Very Weak",
			strength: 10,
			want:     "Very Weak",
		},
		{
			name:     "Weak",
			strength: 30,
			want:     "Weak",
		},
		{
			name:     "Medium",
			strength: 50,
			want:     "Medium",
		},
		{
			name:     "Strong",
			strength: 70,
			want:     "Strong",
		},
		{
			name:     "Very Strong",
			strength: 90,
			want:     "Very Strong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasher.GetPasswordStrengthText(tt.strength)
			if got != tt.want {
				t.Errorf("GetPasswordStrengthText() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Benchmark тесты для производительности
func BenchmarkPasswordHasher_HashPassword(b *testing.B) {
	hasher := NewPasswordHasher()
	password := "MySecurePassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := hasher.HashPassword(password)
		if err != nil {
			b.Fatalf("HashPassword failed: %v", err)
		}
	}
}

func BenchmarkPasswordHasher_CheckPassword(b *testing.B) {
	hasher := NewPasswordHasher()
	password := "MySecurePassword123!"
	hash, err := hasher.HashPassword(password)
	if err != nil {
		b.Fatalf("Failed to create hash for benchmark: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := hasher.CheckPassword(hash, password)
		if err != nil {
			b.Fatalf("CheckPassword failed: %v", err)
		}
	}
}

func BenchmarkPasswordHasher_GenerateRandomPassword(b *testing.B) {
	hasher := NewPasswordHasher()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := hasher.GenerateRandomPassword(12)
		if err != nil {
			b.Fatalf("GenerateRandomPassword failed: %v", err)
		}
	}
}
