package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message структура для WebSocket сообщений
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
	UserID  string      `json:"user_id,omitempty"`
}

// Client представляет WebSocket клиента
type Client struct {
	ID      string
	UserID  string
	Conn    *websocket.Conn
	Send    chan []byte
	Manager *Manager
	mu      sync.Mutex
}

// Manager управляет WebSocket подключениями
type Manager struct {
	clients     map[string]*Client  // clientID -> Client
	userClients map[string][]string // userID -> []clientID
	mu          sync.RWMutex
	upgrader    websocket.Upgrader
}

// NewManager создает новый WebSocket менеджер
func NewManager() *Manager {
	return &Manager{
		clients:     make(map[string]*Client),
		userClients: make(map[string][]string),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // В продакшене нужно настроить CORS
			},
		},
	}
}

// HandleWebSocket обрабатывает WebSocket подключения
func (m *Manager) HandleWebSocket(w http.ResponseWriter, r *http.Request, userID string) {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		ID:      generateClientID(),
		UserID:  userID,
		Conn:    conn,
		Send:    make(chan []byte, 256),
		Manager: m,
	}

	m.registerClient(client)
	go client.readPump()
	go client.writePump()
}

// registerClient регистрирует клиента
func (m *Manager) registerClient(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clients[client.ID] = client
	m.userClients[client.UserID] = append(m.userClients[client.UserID], client.ID)

	log.Printf("Client %s connected for user %s", client.ID, client.UserID)
}

// unregisterClient удаляет клиента
func (m *Manager) unregisterClient(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.clients[client.ID]; ok {
		delete(m.clients, client.ID)
		close(client.Send)

		// Удаляем клиента из списка пользователя
		if clients, ok := m.userClients[client.UserID]; ok {
			for i, clientID := range clients {
				if clientID == client.ID {
					m.userClients[client.UserID] = append(clients[:i], clients[i+1:]...)
					break
				}
			}
			// Если у пользователя больше нет клиентов, удаляем запись
			if len(m.userClients[client.UserID]) == 0 {
				delete(m.userClients, client.UserID)
			}
		}

		log.Printf("Client %s disconnected for user %s", client.ID, client.UserID)
	}
}

// BroadcastToUser отправляет сообщение всем клиентам пользователя
func (m *Manager) BroadcastToUser(userID string, messageType string, payload interface{}) {
	m.mu.RLock()
	clientIDs, exists := m.userClients[userID]
	m.mu.RUnlock()

	if !exists {
		return
	}

	message := Message{
		Type:    messageType,
		Payload: payload,
		UserID:  userID,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, clientID := range clientIDs {
		if client, ok := m.clients[clientID]; ok {
			select {
			case client.Send <- data:
			default:
				// Канал переполнен, удаляем клиента
				go m.unregisterClient(client)
			}
		}
	}
}

// readPump читает сообщения от клиента
func (c *Client) readPump() {
	defer func() {
		c.Manager.unregisterClient(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	if err := c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		log.Printf("Failed to set read deadline: %v", err)
		return
	}
	c.Conn.SetPongHandler(func(string) error {
		if err := c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			log.Printf("Failed to set read deadline in pong handler: %v", err)
		}
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Обрабатываем входящие сообщения
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		// Отправляем подтверждение
		response := Message{
			Type:    "ack",
			Payload: "Message received",
		}
		responseData, _ := json.Marshal(response)
		c.Send <- responseData
	}
}

// writePump отправляет сообщения клиенту
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				log.Printf("Failed to set write deadline: %v", err)
				return
			}
			if !ok {
				if err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					log.Printf("Failed to write close message: %v", err)
				}
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				log.Printf("Failed to write message: %v", err)
				return
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				log.Printf("Failed to set write deadline: %v", err)
				return
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// generateClientID генерирует уникальный ID для клиента
func generateClientID() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}
