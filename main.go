// main.go
package main

import (
    "log"
    "net/http"
    
    "forum/backend/controllers"
    "forum/backend/database"
    "forum/backend/middleware"
    "forum/backend/routes"
    "forum/backend/websocket"
)

func main() {
    // Initialize database
    db := database.InitDB()
    defer db.Close()
    
    // Initialize controllers
    authController := &controllers.AuthController{DB: db}
    postController := &controllers.PostController{DB: db}
    messageController := &controllers.MessageController{DB: db}
    profileController := &controllers.ProfileController{DB: db}

	// Initialize upload controller
	uploadController := &controllers.UploadController{DB: db}
	uploadController.Init()

    // Initialize WebSocket hub
    hub := websocket.NewHub(db)
    go hub.Run()
    
    // Static files
    http.Handle("/", http.FileServer(http.Dir("./frontend")))
	
	// Serve uploaded files
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./frontend/uploads"))))
    
    // Auth routes
    http.HandleFunc("/api/register", authController.Register)
    http.HandleFunc("/api/login", authController.Login)
    http.HandleFunc("/api/logout", authController.Logout)
    http.HandleFunc("/api/me", authController.GetCurrentUser)
    
    // Post routes (with authentication)
    http.HandleFunc("/api/posts", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodGet {
            postController.GetAllPosts(w, r)
        } else {
            middleware.AuthMiddleware(db, func(w http.ResponseWriter, r *http.Request) {
                userID, ok := middleware.GetUserID(r)
                if !ok {
                    http.Error(w, "Unauthorized", http.StatusUnauthorized)
                    return
                }
                postController.CreatePost(w, r, userID)
            })(w, r)
        }
    })
    
    http.HandleFunc("/api/post", postController.GetPost)
    
    http.HandleFunc("/api/comments", middleware.AuthMiddleware(db, func(w http.ResponseWriter, r *http.Request) {
        userID, ok := middleware.GetUserID(r)
        if !ok {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        postController.CreateComment(w, r, userID)
    }))
    
    // Message routes (with authentication)
    http.HandleFunc("/api/messages", middleware.AuthMiddleware(db, func(w http.ResponseWriter, r *http.Request) {
        userID, ok := middleware.GetUserID(r)
        if !ok {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        messageController.GetMessages(w, r, userID)
    }))
    
    http.HandleFunc("/api/send-message", middleware.AuthMiddleware(db, func(w http.ResponseWriter, r *http.Request) {
        userID, ok := middleware.GetUserID(r)
        if !ok {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        messageController.SendMessage(w, r, userID)
    }))
    
    http.HandleFunc("/api/chats", middleware.AuthMiddleware(db, func(w http.ResponseWriter, r *http.Request) {
        userID, ok := middleware.GetUserID(r)
        if !ok {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        messageController.GetChats(w, r, userID)
    }))
    
    // WebSocket route
    http.HandleFunc("/ws", routes.HandleWebSocket(hub))

	// Profile routes
	http.HandleFunc("/api/profile", profileController.GetProfile)
	http.HandleFunc("/api/update-profile", middleware.AuthMiddleware(db, func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		profileController.UpdateProfile(w, r, userID)
	}))

	// Upload routes
	http.HandleFunc("/api/upload-image", middleware.AuthMiddleware(db, func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		uploadController.UploadImage(w, r, userID)
	}))

	http.HandleFunc("/api/upload-avatar", middleware.AuthMiddleware(db, func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		uploadController.UploadAvatar(w, r, userID)
	}))
    
    // Start server
    log.Println("Server started on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}