package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Stella-Achar-Oiro/forum/models"
	"github.com/Stella-Achar-Oiro/forum/utils"
	"github.com/Stella-Achar-Oiro/forum/utils/oauth"
)

// GoogleUserInfo represents a Google user profile
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// GitHubUserInfo represents a GitHub user profile
type GitHubUserInfo struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// GitHubEmailInfo represents a GitHub user email
type GitHubEmailInfo struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// OAuthLoginHandler initiates OAuth login flow
func OAuthLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get provider from URL path
	provider := strings.TrimPrefix(r.URL.Path, "/api/auth/")
	provider = strings.TrimSuffix(provider, "/login")

	// Get OAuth config
	config, ok := oauth.GetConfig()[oauth.Provider(provider)]
	if !ok {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid OAuth provider: %s", provider))
		return
	}

	// Validate OAuth configuration
	if config.ClientID == "" || config.ClientSecret == "" {
		log.Printf("OAuth configuration missing for provider %s", provider)
		utils.RespondWithError(w, http.StatusInternalServerError, "OAuth configuration error")
		return
	}

	// Generate state token to prevent CSRF
	state := utils.GenerateRandomString(32)
	utils.SetStateToken(w, state)

	// Redirect to OAuth provider
	url := config.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// OAuthCallbackHandler handles OAuth callback
func OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get provider from URL path
	provider := strings.TrimPrefix(r.URL.Path, "/api/auth/")
	provider = strings.TrimSuffix(provider, "/callback")

	// Get OAuth config
	config, ok := oauth.GetConfig()[oauth.Provider(provider)]
	if !ok {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid OAuth provider: %s", provider))
		return
	}

	// Check for error parameter
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		errDescription := r.URL.Query().Get("error_description")
		log.Printf("OAuth error from provider %s: %s - %s", provider, errMsg, errDescription)
		utils.RespondWithError(w, http.StatusBadRequest, "Authentication failed")
		return
	}

	// Verify state token
	state := r.URL.Query().Get("state")
	if state == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing state parameter")
		return
	}
	if !utils.VerifyStateToken(r, state) {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid state token")
		return
	}

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing authorization code")
		return
	}

	// Exchange code for token
	token, err := config.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("Failed to exchange code for token: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to authenticate")
		return
	}

	if !token.Valid() {
		log.Printf("Invalid token received from provider %s", provider)
		utils.RespondWithError(w, http.StatusInternalServerError, "Invalid authentication token")
		return
	}

	// Get user info based on provider
	var email, name string

	switch oauth.Provider(provider) {
	case oauth.Google:
		googleUserInfo, err := getGoogleUserInfo(token.AccessToken)
		if err != nil {
			log.Printf("Failed to get user info from Google: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user info")
			return
		}
		email = googleUserInfo.Email
		name = googleUserInfo.Name
		if name == "" {
			name = strings.Split(email, "@")[0]
		}

	case oauth.GitHub:
		githubUserInfo, err := getGitHubUserInfo(token.AccessToken)
		if err != nil {
			log.Printf("Failed to get user info from GitHub: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user info")
			return
		}
		email = githubUserInfo.Email

		// If email is not available in user info, get emails
		if email == "" {
			emails, err := getGitHubEmails(token.AccessToken)
			if err != nil {
				log.Printf("Failed to get GitHub emails: %v", err)
				utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user email")
				return
			}

			// Find primary email
			for _, emailInfo := range emails {
				if emailInfo.Primary && emailInfo.Verified {
					email = emailInfo.Email
					break
				}
			}

			// If no primary email found, use the first verified email
			if email == "" {
				for _, emailInfo := range emails {
					if emailInfo.Verified {
						email = emailInfo.Email
						break
					}
				}
			}
		}

		name = githubUserInfo.Name
		if name == "" {
			name = githubUserInfo.Login
		}
		if name == "" {
			name = strings.Split(email, "@")[0]
		}

	default:
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid OAuth provider")
		return
	}

	if email == "" {
		log.Printf("No email found for user from %s", provider)
		utils.RespondWithError(w, http.StatusInternalServerError, "No email address provided")
		return
	}

	// Check if user exists
	user, err := models.GetUserByEmail(email)
	if err != nil {
		// Create new user
		nameParts := strings.Split(name, " ")
		firstName := nameParts[0]
		lastName := ""
		if len(nameParts) > 1 {
			lastName = strings.Join(nameParts[1:], " ")
		}

		user = models.User{
			Email:     email,
			Nickname:  strings.Split(email, "@")[0],
			FirstName: firstName,
			LastName:  lastName,
			Age:       0, // Default age
			Gender:    "unspecified",
		}

		userID, err := models.CreateUser(user, utils.GenerateRandomString(32))
		if err != nil {
			log.Printf("Failed to create user: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
			return
		}
		user.ID = userID
		log.Printf("Created new user with ID %d from %s login", userID, provider)
	} else {
		log.Printf("Existing user %d logged in via %s", user.ID, provider)
	}

	// Create session
	cookie, err := utils.CreateSessionCookie(user.ID)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	http.SetCookie(w, &cookie)

	// Redirect to home page or dashboard
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// GoogleAuthHandler initiates Google OAuth flow
func GoogleAuthHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting Google OAuth flow")

	// Get OAuth configuration
	configs := oauth.GetConfig()
	googleConfig := configs[oauth.Google]

	// Generate a random state token
	state := utils.GenerateRandomString(32)
	utils.SetStateToken(w, state)

	// Redirect to Google's OAuth 2.0 server
	url := googleConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleAuthCallbackHandler handles the callback from Google
func GoogleAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received Google OAuth callback")

	// Get the state and code
	state := r.FormValue("state")
	code := r.FormValue("code")

	// Verify state token
	if !utils.VerifyStateToken(r, state) {
		log.Println("Invalid OAuth state")
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid OAuth state")
		return
	}

	// Get OAuth configuration
	configs := oauth.GetConfig()
	googleConfig := configs[oauth.Google]

	// Exchange authorization code for token
	token, err := googleConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Code exchange failed: %s", err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Code exchange failed")
		return
	}

	// Get user info
	userInfo, err := getGoogleUserInfo(token.AccessToken)
	if err != nil {
		log.Printf("Failed to get user info: %s", err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user info")
		return
	}

	// Check if user exists
	user, err := models.GetUserByEmail(userInfo.Email)
	if err != nil {
		// User doesn't exist, create a new one
		log.Printf("Creating new user for Google account: %s", userInfo.Email)

		// Generate a random age between 18-65 for demonstration (this would be collected properly in a real app)
		age := 18 + (time.Now().Nanosecond() % 47)

		// Extract nickname from email (part before @)
		emailParts := strings.Split(userInfo.Email, "@")
		nickname := emailParts[0]

		newUser := models.User{
			Nickname:  nickname, // Use part of email as nickname
			Email:     userInfo.Email,
			FirstName: userInfo.GivenName,
			LastName:  userInfo.FamilyName,
			Age:       age,
			Gender:    "prefer_not_to_say", // Default, user can update later
		}

		// Create user with a random secure password
		randomPassword := utils.GenerateRandomString(16)
		userID, err := models.CreateUser(newUser, randomPassword)
		if err != nil {
			log.Printf("Failed to create user: %s", err.Error())
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
			return
		}

		// Get the newly created user
		user, err = models.GetUserByID(userID)
		if err != nil {
			log.Printf("Failed to get created user: %s", err.Error())
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user")
			return
		}
	}

	// Create a session
	cookie, err := utils.CreateSessionCookie(user.ID)
	if err != nil {
		log.Printf("Failed to create session: %s", err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	// Set the cookie
	http.SetCookie(w, &cookie)

	// Update user's online status
	_ = models.UpdateUserOnlineStatus(user.ID, true)

	// Redirect to the main page
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// GitHubAuthHandler initiates GitHub OAuth flow
func GitHubAuthHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting GitHub OAuth flow")

	// Get OAuth configuration
	configs := oauth.GetConfig()
	githubConfig := configs[oauth.GitHub]

	// Generate a random state token
	state := utils.GenerateRandomString(32)
	utils.SetStateToken(w, state)

	// Redirect to GitHub's OAuth server
	url := githubConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GitHubAuthCallbackHandler handles the callback from GitHub
func GitHubAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received GitHub OAuth callback")

	// Get the state and code
	state := r.FormValue("state")
	code := r.FormValue("code")

	// Verify state token
	if !utils.VerifyStateToken(r, state) {
		log.Println("Invalid OAuth state")
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid OAuth state")
		return
	}

	// Get OAuth configuration
	configs := oauth.GetConfig()
	githubConfig := configs[oauth.GitHub]

	// Exchange authorization code for token
	token, err := githubConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Code exchange failed: %s", err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Code exchange failed")
		return
	}

	// Get user info
	userInfo, err := getGitHubUserInfo(token.AccessToken)
	if err != nil {
		log.Printf("Failed to get user info: %s", err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user info")
		return
	}

	// If email is not available in user info, get emails
	if userInfo.Email == "" {
		emails, err := getGitHubEmails(token.AccessToken)
		if err != nil {
			log.Printf("Failed to get emails: %s", err.Error())
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user email")
			return
		}

		// Find primary email
		for _, email := range emails {
			if email.Primary && email.Verified {
				userInfo.Email = email.Email
				break
			}
		}

		// If no primary email found, use the first verified email
		if userInfo.Email == "" {
			for _, email := range emails {
				if email.Verified {
					userInfo.Email = email.Email
					break
				}
			}
		}
	}

	// Check if user exists
	user, err := models.GetUserByEmail(userInfo.Email)
	if err != nil {
		// User doesn't exist, create a new one
		log.Printf("Creating new user for GitHub account: %s", userInfo.Email)

		// Split name into first and last name
		firstName, lastName := userInfo.Name, ""
		if userInfo.Name != "" {
			names := splitName(userInfo.Name)
			if len(names) > 0 {
				firstName = names[0]
				if len(names) > 1 {
					lastName = names[len(names)-1]
				}
			}
		}

		// Generate a random age between 18-65 for demonstration (this would be collected properly in a real app)
		age := 18 + (time.Now().Nanosecond() % 47)

		newUser := models.User{
			Nickname:  userInfo.Login,
			Email:     userInfo.Email,
			FirstName: firstName,
			LastName:  lastName,
			Age:       age,
			Gender:    "prefer_not_to_say", // Default, user can update later
		}

		// Create user with a random secure password
		randomPassword := utils.GenerateRandomString(16)
		userID, err := models.CreateUser(newUser, randomPassword)
		if err != nil {
			log.Printf("Failed to create user: %s", err.Error())
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
			return
		}

		// Get the newly created user
		user, err = models.GetUserByID(userID)
		if err != nil {
			log.Printf("Failed to get created user: %s", err.Error())
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user")
			return
		}
	}

	// Create a session
	cookie, err := utils.CreateSessionCookie(user.ID)
	if err != nil {
		log.Printf("Failed to create session: %s", err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	// Set the cookie
	http.SetCookie(w, &cookie)

	// Update user's online status
	_ = models.UpdateUserOnlineStatus(user.ID, true)

	// Redirect to the main page
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// Helper functions

// getGoogleUserInfo fetches user information from Google
func getGoogleUserInfo(accessToken string) (*GoogleUserInfo, error) {
	// Fetch user data
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %s", err.Error())
	}
	defer resp.Body.Close()

	// Parse response
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %s", err.Error())
	}

	// Unmarshal JSON
	var userInfo GoogleUserInfo
	if err := json.Unmarshal(data, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %s", err.Error())
	}

	return &userInfo, nil
}

// getGitHubUserInfo fetches user information from GitHub
func getGitHubUserInfo(accessToken string) (*GitHubUserInfo, error) {
	// Create request
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err.Error())
	}

	// Set headers
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %s", err.Error())
	}
	defer resp.Body.Close()

	// Parse response
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %s", err.Error())
	}

	// Unmarshal JSON
	var userInfo GitHubUserInfo
	if err := json.Unmarshal(data, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %s", err.Error())
	}

	return &userInfo, nil
}

// getGitHubEmails fetches user emails from GitHub
func getGitHubEmails(accessToken string) ([]GitHubEmailInfo, error) {
	// Create request
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err.Error())
	}

	// Set headers
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get emails: %s", err.Error())
	}
	defer resp.Body.Close()

	// Parse response
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %s", err.Error())
	}

	// Unmarshal JSON
	var emails []GitHubEmailInfo
	if err := json.Unmarshal(data, &emails); err != nil {
		return nil, fmt.Errorf("failed to parse emails: %s", err.Error())
	}

	return emails, nil
}

// splitName splits a full name into parts
func splitName(fullName string) []string {
	// Simple implementation for demonstration purposes
	// In a real app, you might want to use a more sophisticated name parser
	var names []string
	var current string

	for _, c := range fullName {
		if c == ' ' {
			if current != "" {
				names = append(names, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}

	if current != "" {
		names = append(names, current)
	}

	return names
}
