package main

import (
	"log"
	"net/http"
	"time"

	"real-time-forum/backend/database"
	"real-time-forum/backend/handlers"
	"real-time-forum/backend/middleware"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In development, allow all origins
		return true
	},
}

func setupRoutes() http.Handler {
	// Create new mux
	mux := http.NewServeMux()

	// Serve static files
	fileServer := http.FileServer(http.Dir("frontend"))
	mux.Handle("/", fileServer)

	// Auth routes - no authentication required
	mux.HandleFunc("/api/auth/register", handlers.RegisterHandler)
	mux.HandleFunc("/api/auth/login", handlers.LoginHandler)

	// Protected routes - require authentication
	protectedMux := http.NewServeMux()

	// Auth related
	protectedMux.HandleFunc("/api/auth/logout", handlers.LogoutHandler)

	// Posts & Comments
	protectedMux.HandleFunc("/api/posts", handlers.PostsHandler)
	protectedMux.HandleFunc("/api/posts/", handlers.SpecificPostHandler) // For /api/posts/{id}
	protectedMux.HandleFunc("/api/comments", handlers.CommentsHandler)
	protectedMux.HandleFunc("/api/comments/", handlers.SpecificCommentHandler) // For /api/comments/{id}

	// Websocket endpoint for chat
	protectedMux.HandleFunc("/ws", handleWebSocket)

	// Apply authentication middleware to protected routes
	mux.Handle("/api/", middleware.AuthMiddleware(protectedMux))

	// Apply general middleware to all routes
	handler := applyCORS(mux)
	handler = applyLogging(handler)

	return handler
}

// Middleware functions
func applyCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func applyLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf(
			"%s %s %s",
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	})
}

// Placeholder handlers (to be implemented)
func handlePosts(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement posts handling
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func handleSpecificPost(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement specific post handling
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func handleComments(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement comments handling
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// Global hub instance
var wsHub *websocket.Hub

func init() {
	wsHub = websocket.NewHub()
	go wsHub.Run()
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}

	// Start serving the WebSocket connection
	websocket.ServeWs(wsHub, conn, user.ID)
}

func main() {
	// Initialize database
	err := database.InitDB("data/forum.db")
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer database.CloseDB()

	// Setup routes
	handler := setupRoutes()

	// Create server
	server := &http.Server{
		Addr:         ":8000",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	log.Printf("Server starting on http://localhost%s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
