package constants

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultModelFallback = "openai/gpt-5.3-codex"
	defaultModelEnvKey   = "SPICY_MODEL"
)

// DefaultModel is the default AI model used across all tools.
var DefaultModel = resolveDefaultModel()

func resolveDefaultModel() string {
	if model := strings.TrimSpace(os.Getenv(defaultModelEnvKey)); model != "" {
		return model
	}

	if model, ok := lookupDotEnv(defaultModelEnvKey); ok {
		return model
	}

	if model, ok := lookupHomeConfigDotEnv(defaultModelEnvKey); ok {
		return model
	}

	return defaultModelFallback
}

func lookupDotEnv(key string) (string, bool) {
	return lookupDotEnvFile(".env", key)
}

func lookupHomeConfigDotEnv(key string) (string, bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", false
	}

	path := filepath.Join(home, ".config", "spicy", ".env")
	return lookupDotEnvFile(path, key)
}

func lookupDotEnvFile(path, key string) (string, bool) {
	file, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimPrefix(line, "export ")
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		candidateKey := strings.TrimSpace(parts[0])
		if candidateKey != key {
			continue
		}

		value := strings.TrimSpace(parts[1])
		value = trimInlineComment(value)
		value = strings.Trim(value, `"'`)
		if value == "" {
			return "", false
		}

		return value, true
	}

	return "", false
}

func trimInlineComment(value string) string {
	if value == "" {
		return ""
	}

	if strings.HasPrefix(value, "\"") || strings.HasPrefix(value, "'") {
		return value
	}

	if idx := strings.Index(value, " #"); idx >= 0 {
		return strings.TrimSpace(value[:idx])
	}

	return value
}
