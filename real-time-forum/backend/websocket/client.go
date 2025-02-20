package websocket

import (
    "bytes"
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

    // Maximum message size allowed from peer
    maxMessageSize = 512 * 1024 // 512KB
)

var (
    newline = []byte{'\n'}
    space   = []byte{' '}
)

// Client is a middleman between the websocket connection and the hub
type Client struct {
    Hub *Hub

    // The websocket connection
    conn *websocket.Conn

    // Buffered channel of outbound messages
    send chan *WebSocketMessage

    // User ID associated with this connection
    UserID int64
}

// NewClient creates a new client instance
func NewClient(hub *Hub, conn *websocket.Conn, userID int64) *Client {
    return &Client{
        Hub:    hub,
        conn:   conn,
        send:   make(chan *WebSocketMessage, 256),
        UserID: userID,
    }
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
    defer func() {
        c.Hub.unregister <- c
        c.conn.Close()
    }()

    c.conn.SetReadLimit(maxMessageSize)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("error: %v", err)
            }
            break
        }
        message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
        
        // Parse the message
        var wsMessage WebSocketMessage
        if err := json.Unmarshal(message, &wsMessage); err != nil {
            log.Printf("error parsing message: %v", err)
            continue
        }

        // Handle different message types
        switch wsMessage.Type {
        case "ping":
            // Handle ping messages
            c.conn.WriteMessage(websocket.PongMessage, nil)
        default:
            // For now, broadcast all other messages
            c.Hub.broadcast <- &wsMessage
        }
    }
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()

    for {
        select {
        case message, ok := <-c.send:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                // The hub closed the channel
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            // Write the message as JSON
            err := c.conn.WriteJSON(message)
            if err != nil {
                log.Printf("error writing message: %v", err)
                return
            }

        case <-ticker.C:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}

// ServeWs handles websocket requests from the peer
func ServeWs(hub *Hub, conn *websocket.Conn, userID int64) {
    client := NewClient(hub, conn, userID)

    // Register the client
    client.Hub.register <- client

    // Start the read and write pumps
    go client.WritePump()
    go client.ReadPump()
}