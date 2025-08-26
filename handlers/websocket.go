package handlers

import (
	"net/http"

	"linka.type-backend/auth"
	"linka.type-backend/websocket"

	"github.com/gin-gonic/gin"
)

// WebSocketManager глобальный менеджер WebSocket
var WebSocketManager *websocket.Manager

// InitWebSocketManager инициализирует WebSocket менеджер
func InitWebSocketManager() {
	WebSocketManager = websocket.NewManager()
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

// NotifyCategoryUpdate отправляет уведомление об обновлении категории
func NotifyCategoryUpdate(userID string, category interface{}, action string) {
	if WebSocketManager != nil {
		payload := map[string]interface{}{
			"action":   action,
			"category": category,
		}
		WebSocketManager.BroadcastToUser(userID, "category_update", payload)
	}
}

// NotifyCategoryCreated отправляет уведомление о создании категории
func NotifyCategoryCreated(userID string, category interface{}) {
	NotifyCategoryUpdate(userID, category, "created")
}

// NotifyCategoryUpdated отправляет уведомление об обновлении категории
func NotifyCategoryUpdated(userID string, category interface{}) {
	NotifyCategoryUpdate(userID, category, "updated")
}

// NotifyCategoryDeleted отправляет уведомление об удалении категории
func NotifyCategoryDeleted(userID, categoryID string) {
	payload := map[string]interface{}{
		"action":     "deleted",
		"categoryId": categoryID,
	}
	if WebSocketManager != nil {
		WebSocketManager.BroadcastToUser(userID, "category_update", payload)
	}
}

// NotifyStatementUpdate отправляет уведомление об обновлении statement
func NotifyStatementUpdate(userID string, statement interface{}, action string) {
	if WebSocketManager != nil {
		payload := map[string]interface{}{
			"action":    action,
			"statement": statement,
		}
		WebSocketManager.BroadcastToUser(userID, "statement_update", payload)
	}
}

// NotifyStatementCreated отправляет уведомление о создании statement
func NotifyStatementCreated(userID string, statement interface{}) {
	NotifyStatementUpdate(userID, statement, "created")
}

// NotifyStatementUpdated отправляет уведомление об обновлении statement
func NotifyStatementUpdated(userID string, statement interface{}) {
	NotifyStatementUpdate(userID, statement, "updated")
}

// NotifyStatementDeleted отправляет уведомление об удалении statement
func NotifyStatementDeleted(userID, statementID string) {
	payload := map[string]interface{}{
		"action":      "deleted",
		"statementId": statementID,
	}
	if WebSocketManager != nil {
		WebSocketManager.BroadcastToUser(userID, "statement_update", payload)
	}
}
