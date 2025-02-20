package websocket

import (
	"sync"
)

// Event types for WebSocket messages
const (
	EventNewPost    = "new_post"
	EventUpdatePost = "update_post"
	EventDeletePost = "delete_post"
	EventNewComment = "new_comment"
	EventNewMessage = "new_message"
	EventUserStatus = "user_status"
)

// WebSocketMessage represents the structure of messages sent through WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// User to client mapping
	userClients map[int64][]*Client

	// Channel for broadcasting messages
	broadcast chan *WebSocketMessage

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe operations
	mutex sync.RWMutex
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		userClients: make(map[int64][]*Client),
		broadcast:   make(chan *WebSocketMessage),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			if client.UserID != 0 {
				h.userClients[client.UserID] = append(h.userClients[client.UserID], client)
				// Broadcast user online status
				h.broadcastUserStatus(client.UserID, true)
			}
			h.mutex.Unlock()

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				if client.UserID != 0 {
					h.removeUserClient(client)
					// Broadcast user offline status
					h.broadcastUserStatus(client.UserID, false)
				}
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
					if client.UserID != 0 {
						h.removeUserClient(client)
					}
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// BroadcastNewPost broadcasts a new post to all connected clients
func (h *Hub) BroadcastNewPost(post *models.Post) {
	message := &WebSocketMessage{
		Type:    EventNewPost,
		Payload: post,
	}
	h.broadcast <- message
}

// BroadcastNewComment broadcasts a new comment to all connected clients
func (h *Hub) BroadcastNewComment(comment *models.Comment) {
	message := &WebSocketMessage{
		Type:    EventNewComment,
		Payload: comment,
	}
	h.broadcast <- message
}

// SendPrivateMessage sends a message to a specific user
func (h *Hub) SendPrivateMessage(message *models.Message) {
	wsMessage := &WebSocketMessage{
		Type:    EventNewMessage,
		Payload: message,
	}

	h.mutex.RLock()
	if clients, ok := h.userClients[message.ReceiverID]; ok {
		for _, client := range clients {
			select {
			case client.send <- wsMessage:
			default:
				// If send buffer is full, close the connection
				close(client.send)
				delete(h.clients, client)
				h.removeUserClient(client)
			}
		}
	}
	h.mutex.RUnlock()
}

// GetOnlineUsers returns a list of online user IDs
func (h *Hub) GetOnlineUsers() []int64 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	onlineUsers := make([]int64, 0, len(h.userClients))
	for userID := range h.userClients {
		onlineUsers = append(onlineUsers, userID)
	}
	return onlineUsers
}

// Helper functions

func (h *Hub) removeUserClient(client *Client) {
	if clients, ok := h.userClients[client.UserID]; ok {
		newClients := make([]*Client, 0)
		for _, c := range clients {
			if c != client {
				newClients = append(newClients, c)
			}
		}
		if len(newClients) > 0 {
			h.userClients[client.UserID] = newClients
		} else {
			delete(h.userClients, client.UserID)
		}
	}
}

func (h *Hub) broadcastUserStatus(userID int64, online bool) {
	message := &WebSocketMessage{
		Type: EventUserStatus,
		Payload: struct {
			UserID int64 `json:"user_id"`
			Online bool  `json:"online"`
		}{
			UserID: userID,
			Online: online,
		},
	}
	h.broadcast <- message
}
