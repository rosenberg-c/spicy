package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type AuthConfig struct {
	Local struct {
		Type string `json:"type"`
		Key  string `json:"key"`
	} `json:"local"`
	OpenAI struct {
		Type      string `json:"type"`
		Refresh   string `json:"refresh"`
		Access    string `json:"access"`
		Expires   int64  `json:"expires"`
		AccountID string `json:"accountId"`
	} `json:"openai"`
	Anthropic struct {
		Type    string `json:"type"`
		Refresh string `json:"refresh"`
		Access  string `json:"access"`
		Expires int64  `json:"expires"`
	} `json:"anthropic"`
}

// ValidateAuth checks if authentication is valid for the given model provider
func ValidateAuth(model string) error {
	if strings.TrimSpace(model) == "" {
		return fmt.Errorf("no model specified")
	}

	// Determine which provider based on model
	provider := getProvider(model)
	// Only openai and anthropic use OAuth tokens from opencode's auth.json.
	// Local servers (LM Studio, Ollama, ...) and API-key providers are
	// handled by opencode directly.
	if provider != "openai" && provider != "anthropic" {
		return nil
	}

	// Read auth config
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	authPath := filepath.Join(home, ".local", "share", "opencode", "auth.json")
	data, err := os.ReadFile(authPath)
	if err != nil {
		return fmt.Errorf(
			"failed to read auth config: %w\nPlease run: opencode auth login",
			err,
		)
	}

	var config AuthConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse auth config: %w", err)
	}

	// Check expiry for the provider
	var expires int64
	var providerName string
	var authType, refreshToken string

	switch provider {
	case "openai":
		expires = config.OpenAI.Expires
		providerName = "OpenAI"
		authType, refreshToken = config.OpenAI.Type, config.OpenAI.Refresh
	case "anthropic":
		expires = config.Anthropic.Expires
		providerName = "Anthropic"
		authType, refreshToken = config.Anthropic.Type, config.Anthropic.Refresh
	default:
		return nil
	}

	// OpenCode owns OAuth refresh; a stale access token can still be renewed.
	if authType == "oauth" && strings.TrimSpace(refreshToken) != "" {
		return nil
	}

	if expires == 0 {
		return fmt.Errorf(
			"%s auth token not found\nPlease run: opencode auth login",
			providerName,
		)
	}

	// Check if token is expired
	now := time.Now().UnixMilli()
	if expires < now {
		expiryTime := time.UnixMilli(expires)
		return fmt.Errorf(
			"%s token expired at %s\nPlease run: opencode auth login",
			providerName,
			expiryTime.Format("2006-01-02 15:04:05"),
		)
	}

	// Warn if expiring soon (within 5 minutes)
	buffer := int64(5 * 60 * 1000) // 5 minutes in milliseconds
	if expires-now < buffer {
		minutesLeft := (expires - now) / 1000 / 60
		fmt.Fprintf(
			os.Stderr,
			"Warning: %s token expires in %d minutes\n",
			providerName,
			minutesLeft,
		)
	}

	return nil
}

// getProvider extracts the provider from the model string
func getProvider(model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return ""
	}

	// Model IDs are "<provider>/<name>", so the provider is the first
	// path segment (the name itself may contain slashes).
	if idx := strings.Index(model, "/"); idx > 0 {
		return strings.ToLower(model[:idx])
	}

	// Bare model name without a provider prefix: infer from the name.
	switch {
	case contains(model, "openai"), contains(model, "gpt"):
		return "openai"
	case contains(model, "anthropic"), contains(model, "claude"):
		return "anthropic"
	default:
		return ""
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
