package config

import (
	"fmt"
	"strings"
)

// ValidateAPIKey performs basic validation on the API key format
func ValidateAPIKey(apiKey string) error {
	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// PostHog Personal API keys are prefixed with phx_ (phc_ is for Project API keys which won't work here)
	if !strings.HasPrefix(apiKey, "phx_") {
		if strings.HasPrefix(apiKey, "phc_") {
			return fmt.Errorf("Project API keys (phc_) cannot be used. You need a Personal API key (phx_) from Settings → Personal API Keys")
		}
		return fmt.Errorf("Personal API key should start with 'phx_'")
	}

	// Basic length check (PostHog keys are typically 40+ characters)
	if len(apiKey) < 20 {
		return fmt.Errorf("API key appears to be too short")
	}

	return nil
}

// ValidateInstanceURL performs basic validation on the instance URL
func ValidateInstanceURL(url string) error {
	url = strings.TrimSpace(url)

	if url == "" {
		return fmt.Errorf("instance URL cannot be empty")
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("instance URL must start with http:// or https://")
	}

	// Remove trailing slash for consistency
	url = strings.TrimSuffix(url, "/")

	return nil
}

// NormalizeInstanceURL normalizes the instance URL by removing trailing slash
func NormalizeInstanceURL(url string) string {
	return strings.TrimSuffix(strings.TrimSpace(url), "/")
}
