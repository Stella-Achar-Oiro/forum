package router

import (
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Stella-Achar-Oiro/forum/handlers"
	"github.com/Stella-Achar-Oiro/forum/utils"
	"github.com/gorilla/mux"
)

// New creates and configures a new router
func New() *mux.Router {
	r := mux.NewRouter()

	// Initialize SessionManager for secure session management
	sessionManager := utils.NewSessionManager()

	// Start a goroutine for session cleanup
	go func() {
		for {
			// Clean up expired and inactive sessions every hour
			time.Sleep(time.Hour)
			if err := sessionManager.CleanupSessions(); err != nil {
				log.Printf("Error cleaning up sessions: %v", err)
			}
		}
	}()

	// Initialize MIME types
	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".html", "text/html")

	// CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// API Routes
	api := r.PathPrefix("/api").Subrouter()

	// Add CSRF protection to the API routes
	api.Use(sessionManager.CSRFProtectionMiddleware)

	// Public routes (no authentication required)
	publicApi := api.PathPrefix("/public").Subrouter()

	// Auth routes
	publicApi.HandleFunc("/register", handlers.RegisterHandler).Methods("POST", "OPTIONS")
	publicApi.HandleFunc("/login", handlers.LoginHandler).Methods("POST", "OPTIONS")

	// OAuth routes - Add placeholders for Google and GitHub OAuth
	publicApi.HandleFunc("/auth/google", handlers.GoogleAuthHandler).Methods("GET", "OPTIONS")
	publicApi.HandleFunc("/auth/google/callback", handlers.GoogleAuthCallbackHandler).Methods("GET", "OPTIONS")
	publicApi.HandleFunc("/auth/github", handlers.GitHubAuthHandler).Methods("GET", "OPTIONS")
	publicApi.HandleFunc("/auth/github/callback", handlers.GitHubAuthCallbackHandler).Methods("GET", "OPTIONS")

	// Protected routes (authentication required)
	protectedApi := api.PathPrefix("").Subrouter()
	protectedApi.Use(sessionManager.RequireAuthentication)

	// User routes
	protectedApi.HandleFunc("/logout", handlers.LogoutHandler).Methods("POST", "OPTIONS")
	protectedApi.HandleFunc("/user", handlers.GetCurrentUserHandler).Methods("GET", "OPTIONS")

	// Post routes
	protectedApi.HandleFunc("/posts", handlers.CreatePostHandler).Methods("POST", "OPTIONS")
	protectedApi.HandleFunc("/posts", handlers.GetPostsHandler).Methods("GET", "OPTIONS")
	protectedApi.HandleFunc("/posts/{id:[0-9]+}", handlers.GetPostHandler).Methods("GET", "OPTIONS")
	protectedApi.HandleFunc("/posts/{id:[0-9]+}/like", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandlePostReaction(w, r, "like")
	}).Methods("POST", "OPTIONS")
	protectedApi.HandleFunc("/posts/{id:[0-9]+}/dislike", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandlePostReaction(w, r, "dislike")
	}).Methods("POST", "OPTIONS")
	protectedApi.HandleFunc("/posts/{id:[0-9]+}/react", handlers.AddPostReactionHandler).Methods("POST", "OPTIONS")
	protectedApi.HandleFunc("/categories", handlers.GetCategoriesHandler).Methods("GET", "OPTIONS")
	protectedApi.HandleFunc("/posts/category/{categoryId:[0-9]+}", handlers.GetPostsByCategoryHandler).Methods("GET", "OPTIONS")
	protectedApi.HandleFunc("/posts/user/created", handlers.GetUserCreatedPostsHandler).Methods("GET", "OPTIONS")
	protectedApi.HandleFunc("/posts/user/liked", handlers.GetUserLikedPostsHandler).Methods("GET", "OPTIONS")

	// Comment routes
	protectedApi.HandleFunc("/posts/{postId:[0-9]+}/comments", handlers.CreateCommentHandler).Methods("POST", "OPTIONS")
	protectedApi.HandleFunc("/posts/{postId:[0-9]+}/comments", handlers.GetCommentsHandler).Methods("GET", "OPTIONS")
	protectedApi.HandleFunc("/comments/{commentId:[0-9]+}/react", handlers.AddCommentReactionHandler).Methods("POST", "OPTIONS")

	// Message routes
	protectedApi.HandleFunc("/messages/users", handlers.GetMessageUsersHandler).Methods("GET", "OPTIONS")
	protectedApi.HandleFunc("/messages/users/{userId:[0-9]+}", handlers.GetMessagesHandler).Methods("GET", "OPTIONS")
	protectedApi.HandleFunc("/messages/users/{userId:[0-9]+}", handlers.CreateMessageHandler).Methods("POST", "OPTIONS")

	// Notification routes
	protectedApi.HandleFunc("/notifications", handlers.GetNotificationsHandler).Methods("GET", "OPTIONS")
	protectedApi.HandleFunc("/notifications/unread/count", handlers.GetUnreadNotificationCountHandler).Methods("GET", "OPTIONS")
	protectedApi.HandleFunc("/notifications/read", handlers.MarkNotificationReadHandler).Methods("POST", "OPTIONS")

	// Activity route
	protectedApi.HandleFunc("/user/activity", handlers.GetUserActivityHandler).Methods("GET", "OPTIONS")

	// WebSocket endpoint
	protectedApi.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub := utils.NewWebSocketHub()
		go hub.Run()
		// WebSocket connection handling will be implemented here
	}).Methods("GET")

	// Serve static files with proper MIME types
	fs := http.FileServer(http.Dir("static"))
	staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		ext := filepath.Ext(path)

		// Set content type based on file extension
		if mimeType := mime.TypeByExtension(ext); mimeType != "" {
			w.Header().Set("Content-Type", mimeType)
		}

		fs.ServeHTTP(w, r)
	})

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticHandler))

	// Serve index.html for all other routes
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Serving index.html for path: %s", r.URL.Path)
		w.Header().Set("Content-Type", "text/html")
		http.ServeFile(w, r, "static/index.html")
	})

	return r
}
