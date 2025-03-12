package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	GiphyAPIEndpoint = "https://api.giphy.com/v1/gifs"
)

// GiphyConfig holds the configuration for Giphy API
type GiphyConfig struct {
	APIKey string
}

// GiphyGIF represents a GIF from Giphy
type GiphyGIF struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Width  string `json:"width"`
	Height string `json:"height"`
}

// GiphySearchResponse represents the response from Giphy search API
type GiphySearchResponse struct {
	Data []struct {
		ID     string `json:"id"`
		Images struct {
			Original struct {
				URL    string `json:"url"`
				Width  string `json:"width"`
				Height string `json:"height"`
			} `json:"original"`
		} `json:"images"`
	} `json:"data"`
}

// SearchGiphy searches for GIFs on Giphy
func SearchGiphy(query string, apiKey string) ([]GiphyGIF, error) {
	// Build the URL
	baseURL := fmt.Sprintf("%s/search", GiphyAPIEndpoint)
	params := url.Values{}
	params.Add("api_key", apiKey)
	params.Add("q", query)
	params.Add("limit", "10")
	params.Add("rating", "g")

	// Make the request
	resp, err := http.Get(baseURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("failed to search Giphy: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var searchResp GiphySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse Giphy response: %v", err)
	}

	// Convert to GiphyGIF slice
	var gifs []GiphyGIF
	for _, item := range searchResp.Data {
		gifs = append(gifs, GiphyGIF{
			ID:     item.ID,
			URL:    item.Images.Original.URL,
			Width:  item.Images.Original.Width,
			Height: item.Images.Original.Height,
		})
	}

	return gifs, nil
}

// GetGiphyGIF gets a specific GIF by ID
func GetGiphyGIF(id string, apiKey string) (*GiphyGIF, error) {
	// Build the URL
	baseURL := fmt.Sprintf("%s/%s", GiphyAPIEndpoint, id)
	params := url.Values{}
	params.Add("api_key", apiKey)

	// Make the request
	resp, err := http.Get(baseURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("failed to get Giphy GIF: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var gifResp struct {
		Data struct {
			ID     string `json:"id"`
			Images struct {
				Original struct {
					URL    string `json:"url"`
					Width  string `json:"width"`
					Height string `json:"height"`
				} `json:"original"`
			} `json:"images"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&gifResp); err != nil {
		return nil, fmt.Errorf("failed to parse Giphy response: %v", err)
	}

	return &GiphyGIF{
		ID:     gifResp.Data.ID,
		URL:    gifResp.Data.Images.Original.URL,
		Width:  gifResp.Data.Images.Original.Width,
		Height: gifResp.Data.Images.Original.Height,
	}, nil
}
