package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"forum/internal/auth"
	"forum/internal/database"
	"forum/internal/handlers"
	"forum/internal/middleware"
	"forum/internal/models"
)

const (
	PORT = ":8080"
)

// PageData represents the data passed to templates
type PageData struct {
	Title            string
	User             *models.User
	Posts            []models.Post
	Categories       []models.Category
	SelectedCategory string
}

func main() {
	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// Parse templates
	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	// Create handlers
	postHandler := handlers.NewPostHandler(db)
	commentHandler := handlers.NewCommentHandler(db)

	// Create router
	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Home page
	mux.Handle("/", middleware.OptionalAuth(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		user, _ := auth.UserFromContext(r.Context())
		category := r.URL.Query().Get("category")

		// Get categories
		rows, err := db.Query("SELECT id, name, description FROM categories")
		if err != nil {
			http.Error(w, "Error fetching categories", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var categories []models.Category
		for rows.Next() {
			var cat models.Category
			if err := rows.Scan(&cat.ID, &cat.Name, &cat.Description); err != nil {
				http.Error(w, "Error scanning categories", http.StatusInternalServerError)
				return
			}
			categories = append(categories, cat)
		}

		data := PageData{
			Title:            "Home",
			User:             user,
			Categories:       categories,
			SelectedCategory: category,
		}

		if err := templates.ExecuteTemplate(w, "base.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})))

	// Auth routes
	mux.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Handle registration
	})

	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Handle login
	})

	mux.HandleFunc("/api/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Handle logout
	})

	// Post routes
	mux.Handle("/api/posts", middleware.OptionalAuth(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			postHandler.GetPosts(w, r)
		case http.MethodPost:
			postHandler.CreatePost(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/posts/like", middleware.RequireAuth(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		postHandler.LikePost(w, r)
	})))

	// Comment routes
	mux.Handle("/api/comments", middleware.OptionalAuth(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			commentHandler.GetComments(w, r)
		case http.MethodPost:
			commentHandler.CreateComment(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/comments/like", middleware.RequireAuth(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		commentHandler.LikeComment(w, r)
	})))

	// Initialize server
	server := &http.Server{
		Addr:    PORT,
		Handler: mux,
	}

	// Channel for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server is starting on http://localhost%s\n", PORT)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v\n", err)
		}
	}()

	// Wait for interrupt signal
	<-stop
	log.Println("Shutting down server...")
	fmt.Println("Server stopped")
}
