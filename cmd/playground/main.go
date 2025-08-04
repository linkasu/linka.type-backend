package main

import (
	"log"
	"os"

	"linka.type-backend/bl"
	"linka.type-backend/db"
)

func main() {
	// Инициализируем базу данных
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	log.Println("Starting complete data import...")

	// Импортируем все данные (пользователь, категории, statements)
	email := getEnv("IMPORT_EMAIL", "test@example.com")
	password := getEnv("IMPORT_PASSWORD", "password123")
	result, err := bl.ImportAllData(email, password)
	if err != nil {
		log.Fatalf("Failed to import data: %v", err)
	}

	log.Printf("Complete import finished in %v", result.Duration)
	if result.StatementsResult != nil {
		log.Printf("Statements: %d imported, %d updated, %d skipped, %d failed",
			result.StatementsResult.Imported,
			result.StatementsResult.Updated,
			result.StatementsResult.Skipped,
			result.StatementsResult.Failed)
	}
	log.Println("Complete data import finished successfully!")
}

// getEnv получает переменную окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
