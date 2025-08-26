package e2e

import (
	"os"
	"testing"
)

// setupTestEnvironment настраивает тестовую среду
func setupTestEnvironment(_ *testing.T) {
	// Устанавливаем переменные окружения для тестов
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_PORT", "5433") // Используем порт тестовой БД
	os.Setenv("POSTGRES_USER", "postgres")
	os.Setenv("POSTGRES_PASSWORD", "postgres")
	os.Setenv("POSTGRES_DB", "linkatype_test")
}

// setupTestEnvironmentNoT настраивает тестовую среду без параметра testing.T
func setupTestEnvironmentNoT() {
	// Устанавливаем переменные окружения для тестов
	// Используем значения из окружения или дефолтные для локальной разработки
	if os.Getenv("POSTGRES_HOST") == "" {
		os.Setenv("POSTGRES_HOST", "localhost")
	}
	if os.Getenv("POSTGRES_PORT") == "" {
		os.Setenv("POSTGRES_PORT", "5433") // Используем порт тестовой БД для CI/CD
	}
	if os.Getenv("POSTGRES_USER") == "" {
		os.Setenv("POSTGRES_USER", "postgres")
	}
	if os.Getenv("POSTGRES_PASSWORD") == "" {
		os.Setenv("POSTGRES_PASSWORD", "postgres")
	}
	if os.Getenv("POSTGRES_DB") == "" {
		os.Setenv("POSTGRES_DB", "linkatype_test")
	}

	// Настройка mail для тестов (используем mock значения)
	if os.Getenv("MAIL_SERVER") == "" {
		os.Setenv("MAIL_SERVER", "smtp.test.com")
	}
	if os.Getenv("MAIL_PORT") == "" {
		os.Setenv("MAIL_PORT", "587")
	}
	if os.Getenv("MAIL_ADDRESS") == "" {
		os.Setenv("MAIL_ADDRESS", "test@example.com")
	}
	if os.Getenv("MAIL_PASSWORD") == "" {
		os.Setenv("MAIL_PASSWORD", "test_password")
	}

	// Включаем режим тестирования для мокирования email
	os.Setenv("TEST_MODE", "true")
}

// TestMain настраивает тестовую среду перед запуском тестов
func TestMain(m *testing.M) {
	// Настройка тестовой среды
	setupTestEnvironment(&testing.T{})

	// Запуск тестов
	code := m.Run()

	// Очистка после тестов
	os.Exit(code)
}
