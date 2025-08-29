package main

import (
	"log"
	"net/http"
	"os"

	"linka.type-backend/auth"
	"linka.type-backend/db"
	"linka.type-backend/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// Инициализируем базу данных
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.CloseDB(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	// Инициализируем WebSocket менеджер
	handlers.InitWebSocketManager()

	// Настраиваем Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	})

	// Health check endpoint
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes
	api := router.Group("/api")
	{
		// Auth routes
		api.POST("/register", handlers.RegisterDirect)
		api.POST("/login", handlers.Login)
		api.GET("/profile", auth.AuthMiddleware(), func(c *gin.Context) {
			userID := auth.GetUserIDFromContext(c)
			email := auth.GetEmailFromContext(c)
			c.JSON(http.StatusOK, gin.H{
				"user_id": userID,
				"email":   email,
			})
		})

		// OTP routes
		api.POST("/auth/register", handlers.Register)
		api.POST("/auth/verify-email", handlers.VerifyEmail)
		api.POST("/auth/reset-password", handlers.RequestPasswordReset)
		api.POST("/auth/reset-password/verify", handlers.VerifyPasswordResetOTP)
		api.POST("/auth/reset-password/confirm", handlers.ConfirmPasswordReset)

		// Categories routes
		categories := api.Group("/categories")
		categories.Use(auth.AuthMiddleware())
		{
			categories.GET("", handlers.GetCategories)
			categories.GET("/:id", handlers.GetCategory)
			categories.POST("", handlers.CreateCategory)
			categories.PUT("/:id", handlers.UpdateCategory)
			categories.DELETE("/:id", handlers.DeleteCategory)
		}

		// Statements routes
		statements := api.Group("/statements")
		statements.Use(auth.AuthMiddleware())
		{
			statements.GET("", handlers.GetStatements)
			statements.GET("/:id", handlers.GetStatement)
			statements.POST("", handlers.CreateStatement)
			statements.PUT("/:id", handlers.UpdateStatement)
			statements.DELETE("/:id", handlers.DeleteStatement)
		}

		// WebSocket route
		api.GET("/ws", auth.AuthMiddleware(), handlers.HandleWebSocket)
	}

	// Запускаем сервер
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
