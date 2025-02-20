package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Define the base content for essential files
var fileContents = map[string]string{
	"go.mod": `module real-time-forum

go 1.21

require (
	github.com/gorilla/websocket v1.5.1
	github.com/mattn/go-sqlite3 v1.14.22
	golang.org/x/crypto v0.19.0
	github.com/google/uuid v1.6.0
)`,

	"main.go": `package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("Starting Real-Time Forum server...")
	// TODO: Initialize server
}`,

	"Makefile": `run:
	go run main.go

build:
	go build -o forum

test:
	go test ./...`,

	".gitignore": `# Binary
forum
real-time-forum

# Database
*.db

# IDE
.vscode/
.idea/

# OS
.DS_Store`,

	"README.md": `# Real-Time Forum

A real-time forum application with private messaging capabilities.

## Features
- User authentication
- Post creation and comments
- Real-time private messaging
- WebSocket communication

## Setup
1. Clone the repository
2. Run make build
3. Run ./forum

## Technologies
- Go
- SQLite
- WebSocket
- JavaScript (Vanilla)
- HTML/CSS`,
}

// Project structure with directories and files
var structure = []string{
	// Backend structure
	"backend/config/config.go",
	"backend/database/sqlite.go",
	"backend/database/schema.sql",
	"backend/models/user.go",
	"backend/models/post.go",
	"backend/models/comment.go",
	"backend/models/message.go",
	"backend/handlers/auth.go",
	"backend/handlers/post.go",
	"backend/handlers/comment.go",
	"backend/handlers/message.go",
	"backend/middleware/auth.go",
	"backend/middleware/logging.go",
	"backend/websocket/hub.go",
	"backend/websocket/client.go",
	"backend/websocket/message.go",
	"backend/utils/validator.go",
	"backend/utils/security.go",
	"backend/utils/helpers.go",
	// Frontend structure
	"frontend/static/css/main.css",
	"frontend/static/css/auth.css",
	"frontend/static/css/posts.css",
	"frontend/static/css/chat.css",
	"frontend/static/js/main.js",
	"frontend/static/js/router.js",
	"frontend/static/js/websocket.js",
	"frontend/static/js/auth/login.js",
	"frontend/static/js/auth/register.js",
	"frontend/static/js/posts/create.js",
	"frontend/static/js/posts/feed.js",
	"frontend/static/js/posts/comment.js",
	"frontend/static/js/chat/messages.js",
	"frontend/static/js/chat/users.js",
	"frontend/static/img/.gitkeep",
	"frontend/index.html",
	// Data directory
	"data/.gitkeep",
}

func main() {
	projectName := "real-time-forum"
	
	// Create project root directory
	err := os.MkdirAll(projectName, 0755)
	if err != nil {
		fmt.Printf("Error creating project directory: %v\n", err)
		return
	}

	// Create directory structure
	for _, path := range structure {
		fullPath := filepath.Join(projectName, path)
		dir := filepath.Dir(fullPath)
		
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			continue
		}

		// Create empty file
		file, err := os.Create(fullPath)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", fullPath, err)
			continue
		}
		file.Close()
	}

	// Create files with content
	for filename, content := range fileContents {
		fullPath := filepath.Join(projectName, filename)
		err := os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			fmt.Printf("Error writing file %s: %v\n", fullPath, err)
			continue
		}
	}

	fmt.Println("Project structure created successfully!")
	fmt.Println("Next steps:")
	fmt.Println("1. cd", projectName)
	fmt.Println("2. go mod tidy")
	fmt.Println("3. make run")
}