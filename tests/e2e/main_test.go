package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"linka.type-backend/auth"
	"linka.type-backend/db"
	"linka.type-backend/handlers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestData структура для тестовых данных
type TestData struct {
	UserID      string
	Email       string
	Password    string
	Token       string
	CategoryID  string
	StatementID string
}

// setupTestServer создает тестовый сервер
func setupTestServer() *gin.Engine {
	// Инициализируем базу данных
	if err := db.InitDB(); err != nil {
		panic(fmt.Sprintf("Failed to initialize database: %v", err))
	}

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

	// Группа для публичных маршрутов
	public := r.Group("/api")
	{
		public.POST("/register", handlers.Register)
		public.POST("/login", handlers.Login)
		public.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}

	// Группа для защищенных маршрутов
	protected := r.Group("/api")
	protected.Use(auth.JWTAuthMiddleware())
	{
		protected.GET("/profile", func(c *gin.Context) {
			userID := auth.GetUserIDFromContext(c)
			email := auth.GetEmailFromContext(c)

			c.JSON(http.StatusOK, gin.H{
				"user_id": userID,
				"email":   email,
			})
		})

		protected.GET("/categories", handlers.GetCategories)
		protected.GET("/categories/:id", handlers.GetCategory)
		protected.POST("/categories", handlers.CreateCategory)
		protected.PUT("/categories/:id", handlers.UpdateCategory)
		protected.DELETE("/categories/:id", handlers.DeleteCategory)

		protected.GET("/statements", handlers.GetStatements)
		protected.GET("/statements/:id", handlers.GetStatement)
		protected.POST("/statements", handlers.CreateStatement)
		protected.PUT("/statements/:id", handlers.UpdateStatement)
		protected.DELETE("/statements/:id", handlers.DeleteStatement)
	}

	return r
}

// makeRequest вспомогательная функция для выполнения HTTP запросов
func makeRequest(router *gin.Engine, method, url string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			panic(err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// TestE2EFlow полный e2e тест
func TestE2EFlow(t *testing.T) {
	router := setupTestServer()
	defer db.CloseDB()

	testData := &TestData{
		Email:    "test@example.com",
		Password: "password123",
	}

	// 1. Тест регистрации
	t.Run("Register User", func(t *testing.T) {
		registerBody := map[string]interface{}{
			"email":    testData.Email,
			"password": testData.Password,
		}

		w := makeRequest(router, "POST", "/api/register", registerBody, "")

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Contains(t, response, "token")
		assert.Contains(t, response, "user")

		user := response["user"].(map[string]interface{})
		testData.UserID = user["id"].(string)
		testData.Token = response["token"].(string)

		assert.Equal(t, testData.Email, user["email"])
	})

	// 2. Тест логина
	t.Run("Login User", func(t *testing.T) {
		loginBody := map[string]interface{}{
			"email":    testData.Email,
			"password": testData.Password,
		}

		w := makeRequest(router, "POST", "/api/login", loginBody, "")

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Contains(t, response, "token")
		assert.Contains(t, response, "user")

		user := response["user"].(map[string]interface{})
		assert.Equal(t, testData.Email, user["email"])
	})

	// 3. Тест получения профиля
	t.Run("Get Profile", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/profile", nil, testData.Token)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, testData.UserID, response["user_id"])
		assert.Equal(t, testData.Email, response["email"])
	})

	// 4. Тест создания категории
	t.Run("Create Category", func(t *testing.T) {
		categoryBody := map[string]interface{}{
			"title": "Test Category",
		}

		w := makeRequest(router, "POST", "/api/categories", categoryBody, testData.Token)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Contains(t, response, "id")
		assert.Equal(t, "Test Category", response["title"])
		assert.Equal(t, testData.UserID, response["userId"])

		testData.CategoryID = response["id"].(string)
	})

	// 5. Тест получения категорий
	t.Run("Get Categories", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/categories", nil, testData.Token)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Contains(t, response, "categories")
		categories := response["categories"].([]interface{})
		assert.Len(t, categories, 1)

		category := categories[0].(map[string]interface{})
		assert.Equal(t, testData.CategoryID, category["id"])
		assert.Equal(t, "Test Category", category["title"])
	})

	// 6. Тест создания statement
	t.Run("Create Statement", func(t *testing.T) {
		statementBody := map[string]interface{}{
			"title":      "Test Statement",
			"categoryId": testData.CategoryID,
		}

		w := makeRequest(router, "POST", "/api/statements", statementBody, testData.Token)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Contains(t, response, "id")
		assert.Equal(t, "Test Statement", response["text"])
		assert.Equal(t, testData.UserID, response["userId"])
		assert.Equal(t, testData.CategoryID, response["categoryId"])

		testData.StatementID = response["id"].(string)
	})

	// 7. Тест получения statements
	t.Run("Get Statements", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/statements", nil, testData.Token)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Contains(t, response, "statements")
		statements := response["statements"].([]interface{})
		assert.Len(t, statements, 1)

		statement := statements[0].(map[string]interface{})
		assert.Equal(t, testData.StatementID, statement["id"])
		assert.Equal(t, "Test Statement", statement["text"])
	})

	// 8. Тест обновления statement
	t.Run("Update Statement", func(t *testing.T) {
		updateBody := map[string]interface{}{
			"title":      "Updated Statement",
			"categoryId": testData.CategoryID,
		}

		w := makeRequest(router, "PUT", "/api/statements/"+testData.StatementID, updateBody, testData.Token)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "Updated Statement", response["text"])
	})

	// 9. Тест удаления statement
	t.Run("Delete Statement", func(t *testing.T) {
		w := makeRequest(router, "DELETE", "/api/statements/"+testData.StatementID, nil, testData.Token)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "Statement deleted successfully", response["message"])
	})

	// 10. Тест удаления категории
	t.Run("Delete Category", func(t *testing.T) {
		w := makeRequest(router, "DELETE", "/api/categories/"+testData.CategoryID, nil, testData.Token)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "Category deleted successfully", response["message"])
	})
}

// TestAuthenticationErrors тесты ошибок аутентификации
func TestAuthenticationErrors(t *testing.T) {
	router := setupTestServer()
	defer db.CloseDB()

	t.Run("Register with invalid email", func(t *testing.T) {
		registerBody := map[string]interface{}{
			"email":    "invalid-email",
			"password": "password123",
		}

		w := makeRequest(router, "POST", "/api/register", registerBody, "")

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Login with wrong password", func(t *testing.T) {
		// Сначала регистрируем пользователя
		registerBody := map[string]interface{}{
			"email":    "test2@example.com",
			"password": "password123",
		}

		makeRequest(router, "POST", "/api/register", registerBody, "")

		// Пытаемся войти с неправильным паролем
		loginBody := map[string]interface{}{
			"email":    "test2@example.com",
			"password": "wrongpassword",
		}

		w := makeRequest(router, "POST", "/api/login", loginBody, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Access protected endpoint without token", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/profile", nil, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Access protected endpoint with invalid token", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/profile", nil, "invalid-token")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestDataValidation тесты валидации данных
func TestDataValidation(t *testing.T) {
	router := setupTestServer()
	defer db.CloseDB()

	// Регистрируем пользователя для тестов
	registerBody := map[string]interface{}{
		"email":    "test3@example.com",
		"password": "password123",
	}

	w := makeRequest(router, "POST", "/api/register", registerBody, "")
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	token := response["token"].(string)

	t.Run("Create category without title", func(t *testing.T) {
		categoryBody := map[string]interface{}{
			"title": "",
		}

		w := makeRequest(router, "POST", "/api/categories", categoryBody, token)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Create statement without title", func(t *testing.T) {
		statementBody := map[string]interface{}{
			"title":      "",
			"categoryId": "some-category-id",
		}

		w := makeRequest(router, "POST", "/api/statements", statementBody, token)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Create statement without categoryId", func(t *testing.T) {
		statementBody := map[string]interface{}{
			"title": "Test Statement",
		}

		w := makeRequest(router, "POST", "/api/statements", statementBody, token)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestHealthCheck тест health check endpoint
func TestHealthCheck(t *testing.T) {
	router := setupTestServer()
	defer db.CloseDB()

	w := makeRequest(router, "GET", "/api/health", nil, "")

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
}
