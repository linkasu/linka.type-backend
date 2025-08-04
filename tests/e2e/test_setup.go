package main

import (
	"fmt"
	"os"
	"testing"
)

// setupTestEnvironment настраивает тестовую среду
func setupTestEnvironment(t *testing.T) {
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
		os.Setenv("POSTGRES_PORT", "5433") // Используем порт тестовой БД для локальной разработки
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
	
	// Debug: print environment variables
	fmt.Printf("Database connection settings:\n")
	fmt.Printf("  POSTGRES_HOST: %s\n", os.Getenv("POSTGRES_HOST"))
	fmt.Printf("  POSTGRES_PORT: %s\n", os.Getenv("POSTGRES_PORT"))
	fmt.Printf("  POSTGRES_USER: %s\n", os.Getenv("POSTGRES_USER"))
	fmt.Printf("  POSTGRES_DB: %s\n", os.Getenv("POSTGRES_DB"))
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
