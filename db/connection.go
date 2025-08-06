package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitDB initializes the database connection
func InitDB() error {
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "postgres")
	password := getEnv("POSTGRES_PASSWORD", "postgres")
	dbname := getEnv("POSTGRES_DB", "linkatype")

	// Use Docker environment variables from docker-compose.yml
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}

	// Test the connection with retry logic
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		if err = DB.Ping(); err == nil {
			break
		}
		log.Printf("Attempt %d/%d: Failed to connect to database: %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(2 * time.Second)
		}
	}
	if err != nil {
		return fmt.Errorf("error connecting to database after %d attempts: %v", maxRetries, err)
	}

	log.Println("Successfully connected to PostgreSQL database")

	// Create tables if they don't exist
	if err := createTables(); err != nil {
		return fmt.Errorf("error creating tables: %v", err)
	}

	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// createTables creates the necessary tables if they don't exist
func createTables() error {
	// Create users table
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(255) PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		email_verified BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Create categories table
	createCategoriesTable := `
	CREATE TABLE IF NOT EXISTS categories (
		id VARCHAR(255) PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		user_id VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	// Create statements table
	createStatementsTable := `
	CREATE TABLE IF NOT EXISTS statements (
		id VARCHAR(255) PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		user_id VARCHAR(255) NOT NULL,
		category_id VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
	);`

	// Create migration logs table
	createMigrationLogsTable := `
	CREATE TABLE IF NOT EXISTS migration_logs (
		id SERIAL PRIMARY KEY,
		entity_type VARCHAR(50) NOT NULL,
		entity_id VARCHAR(255) NOT NULL,
		user_id VARCHAR(255) NOT NULL,
		action VARCHAR(50) NOT NULL,
		status VARCHAR(50) NOT NULL,
		error TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(entity_type, entity_id, user_id)
	);`

	// Create OTP codes table
	createOTPCodesTable := `
	CREATE TABLE IF NOT EXISTS otp_codes (
		id VARCHAR(255) PRIMARY KEY,
		email VARCHAR(255) NOT NULL,
		code VARCHAR(6) NOT NULL,
		type VARCHAR(20) NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		used BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Execute table creation queries
	queries := []string{createUsersTable, createCategoriesTable, createStatementsTable, createMigrationLogsTable, createOTPCodesTable}

	for _, query := range queries {
		if _, err := DB.Exec(query); err != nil {
			return fmt.Errorf("error creating table: %v", err)
		}
	}

	log.Println("Database tables created successfully")
	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
