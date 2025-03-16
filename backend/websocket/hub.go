// backend/websocket/hub.go
package websocket

import (
    "database/sql"
    "encoding/json"
    "log"
)

// HubMessage combines a message with its sender and type
type HubMessage struct {
    message []byte
    client  *Client
    msgType string
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
    // Registered clients
    Clients map[*Client]bool

    // User ID to client mapping for direct messaging
    UserClients map[int]*Client

    // Inbound messages from clients
    Broadcast chan HubMessage

    // Register requests from clients
    Register chan *Client

    // Unregister requests from clients
    Unregister chan *Client
    
    // Database connection
    DB *sql.DB

	// Work queue for handling messages
    messageQueue chan HubMessage
    
    // Number of worker goroutines
    workerCount int
}

// NewHub creates a new hub
func NewHub(db *sql.DB) *Hub {
    return &Hub{
        Broadcast:    make(chan HubMessage),
        Register:     make(chan *Client),
        Unregister:   make(chan *Client),
        Clients:      make(map[*Client]bool),
        UserClients:  make(map[int]*Client),
        DB:           db,
        messageQueue: make(chan HubMessage, 100), // Buffered channel for message queue
        workerCount:  4, // Number of worker goroutines
    }
}

// Run starts the hub
func (h *Hub) Run() {
    // Start worker goroutines
    for i := 0; i < h.workerCount; i++ {
        go h.worker()
    }
    
    for {
        select {
        case client := <-h.Register:
            h.Clients[client] = true
            h.UserClients[client.UserID] = client
            
            // Broadcast online status
            onlineMsg := OnlineStatusMessage{
                UserID: client.UserID,
                Online: true,
            }
            payload, _ := json.Marshal(onlineMsg)
            msg := Message{
                Type:    "online_status",
                Payload: payload,
            }
            msgBytes, _ := json.Marshal(msg)
            h.broadcastToAll(msgBytes, nil)
            
        case client := <-h.Unregister:
            if _, ok := h.Clients[client]; ok {
                delete(h.Clients, client)
                delete(h.UserClients, client.UserID)
                close(client.Send)
                
                // Broadcast offline status
                offlineMsg := OnlineStatusMessage{
                    UserID: client.UserID,
                    Online: false,
                }
                payload, _ := json.Marshal(offlineMsg)
                msg := Message{
                    Type:    "online_status",
                    Payload: payload,
                }
                msgBytes, _ := json.Marshal(msg)
                h.broadcastToAll(msgBytes, nil)
            }
            
        case hubMsg := <-h.Broadcast:
            // Enqueue message for processing by worker goroutines
            h.messageQueue <- hubMsg
        }
    }
}

// worker is a goroutine that processes messages from the queue
func (h *Hub) worker() {
    for hubMsg := range h.messageQueue {
        switch hubMsg.msgType {
        case "chat_message":
            h.handleChatMessage(hubMsg)
        case "typing":
            h.handleTypingMessage(hubMsg)
        case "new_post":
            h.handleNewPostMessage(hubMsg)
        case "new_comment":
            h.handleNewCommentMessage(hubMsg)
        }
    }
}

// broadcastToAll sends a message to all connected clients except the sender
func (h *Hub) broadcastToAll(message []byte, sender *Client) {
    for client := range h.Clients {
        if client != sender {
            select {
            case client.Send <- message:
            default:
                close(client.Send)
                delete(h.Clients, client)
                delete(h.UserClients, client.UserID)
            }
        }
    }
}

// handleChatMessage processes a chat message
func (h *Hub) handleChatMessage(hubMsg HubMessage) {
    var msg Message
    if err := json.Unmarshal(hubMsg.message, &msg); err != nil {
        log.Printf("error unmarshaling message: %v", err)
        return
    }
    
    var chatMsg ChatMessage
    if err := json.Unmarshal(msg.Payload, &chatMsg); err != nil {
        log.Printf("error unmarshaling chat message: %v", err)
        return
    }
    
    // Send message to the target user if online
    if receiver, ok := h.UserClients[chatMsg.ReceiverID]; ok {
        select {
        case receiver.Send <- hubMsg.message:
        default:
            close(receiver.Send)
            delete(h.Clients, receiver)
            delete(h.UserClients, chatMsg.ReceiverID)
        }
    }
}

// handleTypingMessage processes a typing indicator message
func (h *Hub) handleTypingMessage(hubMsg HubMessage) {
    var msg Message
    if err := json.Unmarshal(hubMsg.message, &msg); err != nil {
        log.Printf("error unmarshaling message: %v", err)
        return
    }
    
    var typingMsg TypingMessage
    if err := json.Unmarshal(msg.Payload, &typingMsg); err != nil {
        log.Printf("error unmarshaling typing message: %v", err)
        return
    }
    
    // Update client typing state
    hubMsg.client.IsTyping = typingMsg.IsTyping
    
    // Send typing status to the target user if online
    if receiver, ok := h.UserClients[typingMsg.ReceiverID]; ok {
        select {
        case receiver.Send <- hubMsg.message:
        default:
            close(receiver.Send)
            delete(h.Clients, receiver)
            delete(h.UserClients, typingMsg.ReceiverID)
        }
    }
}

// handleNewPostMessage broadcasts a new post notification to all users
func (h *Hub) handleNewPostMessage(hubMsg HubMessage) {
    h.broadcastToAll(hubMsg.message, hubMsg.client)
}

// handleNewCommentMessage broadcasts a new comment notification to all users
func (h *Hub) handleNewCommentMessage(hubMsg HubMessage) {
    h.broadcastToAll(hubMsg.message, hubMsg.client)
}