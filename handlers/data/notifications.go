package data

import (
	"log"

	"linka.type-backend/websocket"
)

// WebSocketManager глобальный менеджер WebSocket
var WebSocketManager *websocket.Manager

// SetWebSocketManager устанавливает WebSocket менеджер
func SetWebSocketManager(manager *websocket.Manager) {
	WebSocketManager = manager
}

// NotifyCategoryUpdate отправляет уведомление об обновлении категории
func NotifyCategoryUpdate(userID string, category interface{}, action string) {
	if WebSocketManager != nil {
		payload := map[string]interface{}{
			"action":   action,
			"category": category,
		}
		log.Printf("Sending category notification: userID=%s, action=%s", userID, action)
		WebSocketManager.BroadcastToUser(userID, "category_update", payload)
	} else {
		log.Printf("WebSocketManager is nil, cannot send notification")
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
		log.Printf("Sending statement notification: userID=%s, action=%s", userID, action)
		WebSocketManager.BroadcastToUser(userID, "statement_update", payload)
	} else {
		log.Printf("WebSocketManager is nil, cannot send notification")
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