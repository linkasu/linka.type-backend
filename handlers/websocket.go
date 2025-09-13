package handlers

import (
	"net/http"

	"linka.type-backend/auth"
	"linka.type-backend/handlers/data"
	"linka.type-backend/websocket"

	"github.com/gin-gonic/gin"
)

// WebSocketManager глобальный менеджер WebSocket
var WebSocketManager *websocket.Manager

// InitWebSocketManager инициализирует WebSocket менеджер
func InitWebSocketManager() {
	WebSocketManager = websocket.NewManager()
	data.SetWebSocketManager(WebSocketManager)
}

// HandleWebSocket обрабатывает WebSocket подключения
func HandleWebSocket(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Обновляем HTTP соединение до WebSocket
	WebSocketManager.HandleWebSocket(c.Writer, c.Request, userID)
}