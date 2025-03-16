// backend/websocket/client.go
package websocket

import (
    "encoding/json"
    "log"
    "time"

    "github.com/gorilla/websocket"
)

const (
    // Time allowed to write a message to the peer
    writeWait = 10 * time.Second

    // Time allowed to read the next pong message from the peer
    pongWait = 60 * time.Second

    // Send pings to peer with this period
    pingPeriod = (pongWait * 9) / 10

    // Maximum message size allowed
    maxMessageSize = 10000
)

// Client represents a connected WebSocket client
type Client struct {
    Hub     *Hub
    Conn    *websocket.Conn
    Send    chan []byte
    UserID  int
    IsTyping bool
}

// Message represents different types of messages exchanged over WebSocket
type Message struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"payload"`
}

// ChatMessage represents a private message between users
type ChatMessage struct {
    SenderID   int       `json:"senderId"`
    ReceiverID int       `json:"receiverId"`
    Content    string    `json:"content"`
    ImageURL   string    `json:"imageUrl"`
    CreatedAt  time.Time `json:"createdAt"`
    SenderName string    `json:"senderName"`
}

// PostMessage represents a new post notification
type PostMessage struct {
    PostID int `json:"postId"`
}

// CommentMessage represents a new comment notification
type CommentMessage struct {
    PostID    int `json:"postId"`
    CommentID int `json:"commentId"`
}

// TypingMessage indicates a user is typing
type TypingMessage struct {
    SenderID   int  `json:"senderId"`
    ReceiverID int  `json:"receiverId"`
    IsTyping   bool `json:"isTyping"`
}

// OnlineStatusMessage indicates a user's online status has changed
type OnlineStatusMessage struct {
    UserID int  `json:"userId"`
    Online bool `json:"online"`
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
    defer func() {
        c.Hub.Unregister <- c
        c.Conn.Close()
    }()
    
    c.Conn.SetReadLimit(maxMessageSize)
    c.Conn.SetReadDeadline(time.Now().Add(pongWait))
    c.Conn.SetPongHandler(func(string) error { 
        c.Conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil 
    })
    
    for {
        _, message, err := c.Conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("error: %v", err)
            }
            break
        }
        
        var msg Message
        if err := json.Unmarshal(message, &msg); err != nil {
            log.Printf("error unmarshaling message: %v", err)
            continue
        }
        
        c.Hub.Broadcast <- HubMessage{
            message: message,
            client:  c,
            msgType: msg.Type,
        }
    }
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.Conn.Close()
    }()
    
    for {
        select {
        case message, ok := <-c.Send:
            c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                // The hub closed the channel
                c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            w, err := c.Conn.NextWriter(websocket.TextMessage)
            if err != nil {
                return
            }
            w.Write(message)

            // Add queued chat messages to the current websocket message
            n := len(c.Send)
            for i := 0; i < n; i++ {
                w.Write([]byte{'\n'})
                w.Write(<-c.Send)
            }

            if err := w.Close(); err != nil {
                return
            }
        case <-ticker.C:
            c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}