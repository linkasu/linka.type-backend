package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"linka.type-backend/auth"
	"linka.type-backend/handlers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupTestRouter создает тестовый роутер без базы данных
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
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

// TestHealthCheck тест health check endpoint
func TestHealthCheck(t *testing.T) {
	router := setupTestRouter()

	w := makeRequest(router, "GET", "/api/health", nil, "")

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
}

// TestAuthenticationFlow тест аутентификации (без базы данных)
func TestAuthenticationFlow(t *testing.T) {
	router := setupTestRouter()

	// Тест регистрации с неверными данными
	t.Run("Register with invalid email", func(t *testing.T) {
		registerBody := map[string]interface{}{
			"email":    "invalid-email",
			"password": "password123",
		}

		w := makeRequest(router, "POST", "/api/register", registerBody, "")

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Тест регистрации с коротким паролем
	t.Run("Register with short password", func(t *testing.T) {
		registerBody := map[string]interface{}{
			"email":    "test@example.com",
			"password": "123",
		}

		w := makeRequest(router, "POST", "/api/register", registerBody, "")

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Тест доступа к защищенному endpoint без токена
	t.Run("Access protected endpoint without token", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/profile", nil, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Тест доступа к защищенному endpoint с неверным токеном
	t.Run("Access protected endpoint with invalid token", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/profile", nil, "invalid-token")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Тест доступа к защищенному endpoint с неверным форматом токена
	t.Run("Access protected endpoint with invalid token format", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/profile", nil, "Bearer invalid-token")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestJWTTokenGeneration тест генерации JWT токенов
func TestJWTTokenGeneration(t *testing.T) {
	userID := "test_user_123"
	email := "test@example.com"

	// Генерируем токен
	token, err := auth.GenerateToken(userID, email)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Валидируем токен
	claims, err := auth.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)

	// Тестируем доступ к защищенному endpoint с валидным токеном
	router := setupTestRouter()
	w := makeRequest(router, "GET", "/api/profile", nil, token)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, userID, response["user_id"])
	assert.Equal(t, email, response["email"])
}

// TestCORSHeaders тест CORS заголовков
func TestCORSHeaders(t *testing.T) {
	router := setupTestRouter()

	// Тест OPTIONS запроса
	req, err := http.NewRequest("OPTIONS", "/api/health", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
}
