package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"linka.type-backend/auth"
	"linka.type-backend/handlers"
	"linka.type-backend/websocket"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupWebSocketTestRouter создает тестовый роутер с WebSocket
func setupWebSocketTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Инициализируем WebSocket менеджер
	handlers.InitWebSocketManager()

	// Группа для защищенных маршрутов
	protected := r.Group("/api")
	protected.Use(auth.JWTAuthMiddleware())
	{
		protected.GET("/ws", handlers.HandleWebSocket)
	}

	return r
}

// TestWebSocketEndpoint тест WebSocket endpoint
func TestWebSocketEndpoint(t *testing.T) {
	router := setupWebSocketTestRouter()

	// Генерируем токен для тестирования
	userID := "test_user_123"
	email := "test@example.com"
	token, err := auth.GenerateToken(userID, email)
	assert.NoError(t, err)

	// Тест доступа к WebSocket endpoint без токена
	t.Run("Access WebSocket without token", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/ws", nil, "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Тест доступа к WebSocket endpoint с неверным токеном
	t.Run("Access WebSocket with invalid token", func(t *testing.T) {
		w := makeRequest(router, "GET", "/api/ws", nil, "invalid-token")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Тест доступа к WebSocket endpoint с валидным токеном
	t.Run("Access WebSocket with valid token", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/ws", nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Проверяем, что endpoint доступен (не 401/403)
		assert.NotEqual(t, http.StatusUnauthorized, w.Code)
		assert.NotEqual(t, http.StatusForbidden, w.Code)
	})
}

// TestWebSocketNotifications тест уведомлений WebSocket
func TestWebSocketNotifications(t *testing.T) {
	// Инициализируем WebSocket менеджер
	handlers.InitWebSocketManager()

	userID := "test_user_123"

	// Тест уведомления о создании категории
	t.Run("Notify category created", func(t *testing.T) {
		category := map[string]interface{}{
			"id":     "category_123",
			"title":  "Test Category",
			"userId": userID,
		}

		handlers.NotifyCategoryCreated(userID, category)
		// Не должно вызывать ошибок
	})

	// Тест уведомления об обновлении категории
	t.Run("Notify category updated", func(t *testing.T) {
		category := map[string]interface{}{
			"id":     "category_123",
			"title":  "Updated Category",
			"userId": userID,
		}

		handlers.NotifyCategoryUpdated(userID, category)
		// Не должно вызывать ошибок
	})

	// Тест уведомления об удалении категории
	t.Run("Notify category deleted", func(t *testing.T) {
		categoryID := "category_123"
		handlers.NotifyCategoryDeleted(userID, categoryID)
		// Не должно вызывать ошибок
	})

	// Тест уведомления о создании statement
	t.Run("Notify statement created", func(t *testing.T) {
		statement := map[string]interface{}{
			"id":         "statement_123",
			"text":       "Test Statement",
			"userId":     userID,
			"categoryId": "category_123",
		}

		handlers.NotifyStatementCreated(userID, statement)
		// Не должно вызывать ошибок
	})

	// Тест уведомления об обновлении statement
	t.Run("Notify statement updated", func(t *testing.T) {
		statement := map[string]interface{}{
			"id":         "statement_123",
			"text":       "Updated Statement",
			"userId":     userID,
			"categoryId": "category_123",
		}

		handlers.NotifyStatementUpdated(userID, statement)
		// Не должно вызывать ошибок
	})

	// Тест уведомления об удалении statement
	t.Run("Notify statement deleted", func(t *testing.T) {
		statementID := "statement_123"
		handlers.NotifyStatementDeleted(userID, statementID)
		// Не должно вызывать ошибок
	})
}

// TestWebSocketMessageFormat тест формата WebSocket сообщений
func TestWebSocketMessageFormat(t *testing.T) {
	// Тест формата сообщения о создании категории
	category := map[string]interface{}{
		"id":     "category_123",
		"title":  "Test Category",
		"userId": "user_123",
	}

	expectedMessage := websocket.Message{
		Type: "category_update",
		Payload: map[string]interface{}{
			"action":   "created",
			"category": category,
		},
		UserID: "user_123",
	}

	data, err := json.Marshal(expectedMessage)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Проверяем, что сообщение можно десериализовать
	var decodedMessage websocket.Message
	err = json.Unmarshal(data, &decodedMessage)
	assert.NoError(t, err)
	assert.Equal(t, "category_update", decodedMessage.Type)
	assert.Equal(t, "user_123", decodedMessage.UserID)

	// Проверяем payload
	payload, ok := decodedMessage.Payload.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "created", payload["action"])
}
