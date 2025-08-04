package main

import (
	"log"
	"net/http"

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
	defer db.CloseDB()

	// Создаем Gin роутер
	r := gin.Default()

	// Добавляем CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Группа для публичных маршрутов (без авторизации)
	public := r.Group("/api")
	{
		// Регистрация и логин
		public.POST("/register", handlers.Register)
		public.POST("/login", handlers.Login)

		// Health check
		public.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}

	// Группа для защищенных маршрутов (с JWT авторизацией)
	protected := r.Group("/api")
	protected.Use(auth.JWTAuthMiddleware())
	{
		// Statements
		protected.GET("/statements", handlers.GetStatements)
		protected.GET("/statements/:id", handlers.GetStatement)
		protected.POST("/statements", handlers.CreateStatement)
		protected.PUT("/statements/:id", handlers.UpdateStatement)
		protected.DELETE("/statements/:id", handlers.DeleteStatement)

		// Categories
		protected.GET("/categories", handlers.GetCategories)
		protected.GET("/categories/:id", handlers.GetCategory)
		protected.POST("/categories", handlers.CreateCategory)
		protected.PUT("/categories/:id", handlers.UpdateCategory)
		protected.DELETE("/categories/:id", handlers.DeleteCategory)

		// Профиль пользователя
		protected.GET("/profile", func(c *gin.Context) {
			userID := auth.GetUserIDFromContext(c)
			email := auth.GetEmailFromContext(c)

			c.JSON(http.StatusOK, gin.H{
				"user_id": userID,
				"email":   email,
			})
		})
	}

	// Запускаем сервер
	port := ":8080"
	log.Printf("Server starting on port %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
