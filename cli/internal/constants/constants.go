package constants

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultModelFallback = "openai/gpt-5.3-codex"
	localConfigFileName  = ".spicy.config.json"
)

type spicyConfig struct {
	DefaultModel string            `json:"default_model"`
	Models       map[string]string `json:"models"`
}

func resolveDefaultModel() string {
	config, ok := loadConfig()
	if ok {
		if model := strings.TrimSpace(config.DefaultModel); model != "" {
			return model
		}
	}

	return defaultModelFallback
}

// ModelForCommand returns the configured model for a command.
// Command-specific config overrides default_model.
func ModelForCommand(command string) string {
	config, ok := loadConfig()
	if ok && config.Models != nil {
		if model := strings.TrimSpace(config.Models[command]); model != "" {
			return model
		}
	}

	if ok {
		if model := strings.TrimSpace(config.DefaultModel); model != "" {
			return model
		}
	}

	return defaultModelFallback
}

// loadConfig returns the merged model configuration for the current
// invocation: a per-repo local config file (discovered from the current
// working directory) overlays the global config file under $HOME.
// Callers apply defaultModelFallback when no value remains.
func loadConfig() (spicyConfig, bool) {
	global, globalOK := loadGlobalConfig()

	local, localOK := loadLocalConfig()
	if !localOK {
		return global, globalOK
	}

	return mergeConfig(global, local), true
}

func loadGlobalConfig() (spicyConfig, bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return spicyConfig{}, false
	}

	return loadConfigFile(filepath.Join(home, ".config", "spicy", "config.json"))
}

func loadLocalConfig() (spicyConfig, bool) {
	startDir, err := os.Getwd()
	if err != nil {
		return spicyConfig{}, false
	}

	path, ok := findLocalConfigFile(startDir)
	if !ok {
		return spicyConfig{}, false
	}

	return loadConfigFile(path)
}

// findLocalConfigFile walks from startDir up to the filesystem root and
// returns the first existing regular file named localConfigFileName.
func findLocalConfigFile(startDir string) (string, bool) {
	dir := startDir
	for {
		path := filepath.Join(dir, localConfigFileName)
		if info, err := os.Stat(path); err == nil && info.Mode().IsRegular() {
			return path, true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// loadConfigFile reads a single config file. Missing, unreadable, or invalid
// files return the zero config so callers can fall back silently.
func loadConfigFile(path string) (spicyConfig, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return spicyConfig{}, false
	}

	var config spicyConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return spicyConfig{}, false
	}

	return config, true
}

// mergeConfig overlays a local config onto a global config per key. Non-empty
// local values win; global values are preserved where the local value is
// absent or empty.
func mergeConfig(global, local spicyConfig) spicyConfig {
	merged := spicyConfig{DefaultModel: global.DefaultModel}
	if strings.TrimSpace(local.DefaultModel) != "" {
		merged.DefaultModel = local.DefaultModel
	}

	if len(global.Models) == 0 && len(local.Models) == 0 {
		return merged
	}

	merged.Models = make(map[string]string, len(global.Models)+len(local.Models))
	for command, model := range global.Models {
		merged.Models[command] = model
	}
	for command, model := range local.Models {
		if strings.TrimSpace(model) != "" {
			merged.Models[command] = model
		}
	}

	return merged
}
