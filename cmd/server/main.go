package main

import (
	"log"
	"os"

	"linka.type-backend/auth"
	"linka.type-backend/db"
	"linka.type-backend/handlers"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные окружения
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Инициализируем базу данных
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.CloseDB(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	// Настраиваем Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API маршруты
	api := r.Group("/api")
	{
		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		// Аутентификация (без верификации)
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", handlers.Register)
			authGroup.POST("/login", handlers.Login)
			authGroup.POST("/verify-email", handlers.VerifyEmail)
			authGroup.POST("/reset-password", handlers.RequestPasswordReset)
			authGroup.POST("/reset-password/verify", handlers.VerifyPasswordResetOTP)
			authGroup.POST("/reset-password/confirm", handlers.ConfirmPasswordReset)
		}

		// Защищенные маршруты (требуют аутентификации и верификации email)
		protected := api.Group("/")
		protected.Use(auth.AuthMiddleware(), auth.EmailVerifiedMiddleware())
		{
			// Statements
			protected.GET("/statements", handlers.GetStatements)
			protected.POST("/statements", handlers.CreateStatement)
			protected.PUT("/statements/:id", handlers.UpdateStatement)
			protected.DELETE("/statements/:id", handlers.DeleteStatement)

			// Categories
			protected.GET("/categories", handlers.GetCategories)
			protected.POST("/categories", handlers.CreateCategory)
			protected.PUT("/categories/:id", handlers.UpdateCategory)
			protected.DELETE("/categories/:id", handlers.DeleteCategory)
		}

		// WebSocket (требует только аутентификации)
		ws := api.Group("/ws")
		ws.Use(auth.AuthMiddleware())
		{
			ws.GET("", handlers.HandleWebSocket)
		}
	}

	// Получаем порт из переменных окружения
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
