// backend/routes/websocket.go
package routes

import (
	"net/http"
	"strconv"

	websocketPkg "forum/backend/websocket"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// HandleWebSocket upgrades HTTP connection to WebSocket and registers client
func HandleWebSocket(hub *websocketPkg.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from query parameter
		userIDStr := r.URL.Query().Get("userId")
		if userIDStr == "" {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}

		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Upgrade connection
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "Could not upgrade connection", http.StatusInternalServerError)
			return
		}

		// Create client
		client := &websocketPkg.Client{
			Hub:    hub,
			Conn:   conn,
			Send:   make(chan []byte, 256),
			UserID: userID,
		}

		// Register client
		client.Hub.Register <- client

		// Start goroutines for reading and writing
		go client.ReadPump()
		go client.WritePump()
	}
}
