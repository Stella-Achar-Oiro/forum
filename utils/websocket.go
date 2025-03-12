package utils

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/Stella-Achar-Oiro/forum/models"
	"github.com/gorilla/websocket"
)

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	UserID   int
	Conn     *websocket.Conn
	mu       sync.Mutex
	IsActive bool
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// WebSocketHub maintains active connections and broadcasts messages
type WebSocketHub struct {
	clients    map[int]*WebSocketClient
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	mu         sync.Mutex
}

// NewWebSocketHub creates a new WebSocketHub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[int]*WebSocketClient),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
	}
}

// Run starts the WebSocketHub
func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			h.mu.Unlock()

			// Update user online status in the database
			_ = models.UpdateUserOnlineStatus(client.UserID, true)

			// Broadcast user online status change
			h.BroadcastUserStatus(client.UserID, true)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				client.IsActive = false
				delete(h.clients, client.UserID)
				client.Conn.Close()
			}
			h.mu.Unlock()

			// Update user online status in the database
			_ = models.UpdateUserOnlineStatus(client.UserID, false)

			// Broadcast user online status change
			h.BroadcastUserStatus(client.UserID, false)
		}
	}
}

// Register registers a new WebSocket client
func (h *WebSocketHub) Register(client *WebSocketClient) {
	client.IsActive = true
	h.register <- client
}

// Unregister unregisters a WebSocket client
func (h *WebSocketHub) Unregister(client *WebSocketClient) {
	h.unregister <- client
}

// SendToUser sends a message to a specific user
func (h *WebSocketHub) SendToUser(userID int, message WebSocketMessage) bool {
	h.mu.Lock()
	client, ok := h.clients[userID]
	h.mu.Unlock()

	if !ok || !client.IsActive {
		return false
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling WebSocket message: %v", err)
		return false
	}

	client.mu.Lock()
	defer client.mu.Unlock()

	if err := client.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("Error sending WebSocket message to user %d: %v", userID, err)
		client.IsActive = false
		return false
	}

	return true
}

// BroadcastUserStatus broadcasts a user's online status to all clients
func (h *WebSocketHub) BroadcastUserStatus(userID int, isOnline bool) {
	user, err := models.GetUserByID(userID)
	if err != nil {
		log.Printf("Error getting user for status broadcast: %v", err)
		return
	}

	userPublic := user.ToPublic()
	userPublic.IsOnline = isOnline

	payload, err := json.Marshal(userPublic)
	if err != nil {
		log.Printf("Error marshaling user status: %v", err)
		return
	}

	message := WebSocketMessage{
		Type:    "userStatus",
		Payload: payload,
	}

	h.mu.Lock()
	clients := make([]*WebSocketClient, 0, len(h.clients))
	for _, client := range h.clients {
		if client.IsActive {
			clients = append(clients, client)
		}
	}
	h.mu.Unlock()

	for _, client := range clients {
		h.SendToUser(client.UserID, message)
	}
}

// BroadcastNewMessage broadcasts a new private message
func (h *WebSocketHub) BroadcastNewMessage(message models.Message) {
	payload, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	wsMessage := WebSocketMessage{
		Type:    "newMessage",
		Payload: payload,
	}

	h.SendToUser(message.ReceiverID, wsMessage)
}

// IsUserOnline checks if a user is online
func (h *WebSocketHub) IsUserOnline(userID int) bool {
	h.mu.Lock()
	client, ok := h.clients[userID]
	h.mu.Unlock()
	return ok && client.IsActive
}
