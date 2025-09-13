package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// DB is the global database connection
var DB *sql.DB

// InitDB initializes the database connection
func InitDB() error {
	// Get database connection string from environment variable
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		// Fallback to individual environment variables
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		sslmode := os.Getenv("DB_SSLMODE")

		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "5432"
		}
		if user == "" {
			user = "postgres"
		}
		if password == "" {
			password = "postgres"
		}
		if dbname == "" {
			dbname = "linka_type"
		}
		if sslmode == "" {
			sslmode = "disable"
		}

		connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode)
	}

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}

	// Test the connection
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("error pinging database: %v", err)
	}

	log.Println("Database connection established successfully")
	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}