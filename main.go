package main

import (
	"log"
	"net/http"

	"github.com/Stella-Achar-Oiro/forum/database"
	"github.com/Stella-Achar-Oiro/forum/router"
	"github.com/Stella-Achar-Oiro/forum/utils"
)

func main() {
	// Load environment variables
	if err := utils.LoadEnv(); err != nil {
		log.Printf("Warning: Failed to load .env file: %v", err)
	}

	// Initialize database
	if err := database.Initialize(); err != nil {
		log.Fatal(err)
	}

	// Setup routes using Gorilla Mux
	r := router.New()

	// Start server
	port := utils.GetEnvWithDefault("PORT", "8080")
	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
