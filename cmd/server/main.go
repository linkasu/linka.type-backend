package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"linka.type-backend/auth"
	"linka.type-backend/db"
	"linka.type-backend/handlers"

	"github.com/gin-gonic/gin"
)

// getEnv получает переменную окружения или возвращает дефолт
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getCORSOrigins получает список разрешенных origins из переменной окружения
func getCORSOrigins() []string {
	origins := getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:8080")
	return strings.Split(origins, ",")
}

// getCORSMethods получает разрешенные HTTP методы
func getCORSMethods() string {
	return getEnv("CORS_METHODS", "GET, POST, PUT, DELETE, OPTIONS")
}

// getCORSHeaders получает разрешенные заголовки
func getCORSHeaders() string {
	return getEnv("CORS_HEADERS", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

// corsMiddleware создает CORS middleware с настройками из переменных окружения
func corsMiddleware() gin.HandlerFunc {
	allowedOrigins := getCORSOrigins()
	allowedMethods := getCORSMethods()
	allowedHeaders := getCORSHeaders()

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Проверяем, разрешен ли origin
		originAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				originAllowed = true
				break
			}
		}

		if originAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", allowedMethods)
		c.Header("Access-Control-Allow-Headers", allowedHeaders)
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

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

	// CORS middleware с настройками из переменных окружения
	router.Use(corsMiddleware())

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
		api.POST("/refresh-token", handlers.RefreshToken)
		api.GET("/profile", auth.AuthMiddleware(), func(c *gin.Context) {
			userID := auth.GetUserIDFromContext(c)
			email := auth.GetEmailFromContext(c)
			c.JSON(http.StatusOK, gin.H{
				"user_id": userID,
				"email":   email,
			})
		})

		// OTP routes
		api.POST("/auth/login", handlers.Login)
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

		// Hash route
		api.GET("/hash", auth.AuthMiddleware(), handlers.GetCategoryHash)

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

		// Events routes
		events := api.Group("/events")
		events.Use(auth.AuthMiddleware())
		{
			events.POST("", handlers.CreateEvent)
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
