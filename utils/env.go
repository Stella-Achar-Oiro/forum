package utils

import (
	"bufio"
	"log"
	"os"
	"strings"
)

// LoadEnv loads environment variables from .env file
func LoadEnv() error {
	log.Printf("Loading environment variables from .env file...")

	file, err := os.Open(".env")
	if err != nil {
		log.Printf("Error opening .env file: %v", err)
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first = sign
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Set environment variable
		err := os.Setenv(key, value)
		if err != nil {
			log.Printf("Failed to set environment variable %s: %v", key, err)
		} else {
			log.Printf("Set environment variable: %s", key)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading .env file: %v", err)
		return err
	}

	return nil
}

// GetEnvWithDefault gets an environment variable or returns a default value if not set
func GetEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
