package unit

import (
	"encoding/json"
	"testing"

	"linka.type-backend/websocket"

	"github.com/stretchr/testify/assert"
)

func TestWebSocketMessage(t *testing.T) {
	// Тест сериализации сообщения
	message := websocket.Message{
		Type:    "test",
		Payload: "test payload",
		UserID:  "user123",
	}

	data, err := json.Marshal(message)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Тест десериализации сообщения
	var decodedMessage websocket.Message
	err = json.Unmarshal(data, &decodedMessage)
	assert.NoError(t, err)
	assert.Equal(t, message.Type, decodedMessage.Type)
	assert.Equal(t, message.Payload, decodedMessage.Payload)
	assert.Equal(t, message.UserID, decodedMessage.UserID)
}

func TestWebSocketManager(t *testing.T) {
	manager := websocket.NewManager()
	assert.NotNil(t, manager)
}

func TestWebSocketManagerBroadcast(t *testing.T) {
	manager := websocket.NewManager()

	// Тест отправки сообщения несуществующему пользователю
	manager.BroadcastToUser("nonexistent", "test", "payload")
	// Не должно вызывать ошибок
}

func TestWebSocketMessageTypes(t *testing.T) {
	// Тест различных типов сообщений
	testCases := []struct {
		messageType string
		payload     interface{}
	}{
		{"category_update", map[string]interface{}{"action": "created", "category": "test"}},
		{"category_update", map[string]interface{}{"action": "updated", "category": "test"}},
		{"category_update", map[string]interface{}{"action": "deleted", "categoryId": "test"}},
		{"statement_update", map[string]interface{}{"action": "created", "statement": "test"}},
		{"statement_update", map[string]interface{}{"action": "updated", "statement": "test"}},
		{"statement_update", map[string]interface{}{"action": "deleted", "statementId": "test"}},
		{"ack", "Message received"},
	}

	for _, tc := range testCases {
		message := websocket.Message{
			Type:    tc.messageType,
			Payload: tc.payload,
			UserID:  "user123",
		}

		data, err := json.Marshal(message)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		var decodedMessage websocket.Message
		err = json.Unmarshal(data, &decodedMessage)
		assert.NoError(t, err)
		assert.Equal(t, tc.messageType, decodedMessage.Type)
	}
}
