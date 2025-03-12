package oauth

import (
	"log"
	"os"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// Provider represents an OAuth provider
type Provider string

const (
	Google Provider = "google"
	GitHub Provider = "github"
)

var (
	config     map[Provider]*oauth2.Config
	configOnce sync.Once
)

// GetConfig returns the OAuth configuration after ensuring it's properly initialized
func GetConfig() map[Provider]*oauth2.Config {
	configOnce.Do(func() {
		log.Printf("Initializing OAuth configuration...")
		log.Printf("GOOGLE_CLIENT_ID: %s", os.Getenv("GOOGLE_CLIENT_ID"))
		log.Printf("GITHUB_CLIENT_ID: %s", os.Getenv("GITHUB_CLIENT_ID"))
		log.Printf("BASE_URL: %s", os.Getenv("BASE_URL"))

		config = map[Provider]*oauth2.Config{
			Google: {
				ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
				ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
				RedirectURL:  os.Getenv("BASE_URL") + "/api/public/auth/google/callback",
				Scopes: []string{
					"https://www.googleapis.com/auth/userinfo.email",
					"https://www.googleapis.com/auth/userinfo.profile",
				},
				Endpoint: google.Endpoint,
			},
			GitHub: {
				ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
				ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
				RedirectURL:  os.Getenv("BASE_URL") + "/api/public/auth/github/callback",
				Scopes: []string{
					"user:email",
					"read:user",
				},
				Endpoint: github.Endpoint,
			},
		}
	})
	return config
}
