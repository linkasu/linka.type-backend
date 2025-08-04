package main

import (
	"log"

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
	result, err := bl.ImportAllData("ivan@aacidov.ru", "nhjkkm1998")
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
