package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{
			name:   "exact match",
			s:      "openai",
			substr: "openai",
			want:   true,
		},
		{
			name:   "prefix match",
			s:      "openai/gpt-4",
			substr: "openai",
			want:   true,
		},
		{
			name:   "suffix match",
			s:      "provider/openai",
			substr: "openai",
			want:   true,
		},
		{
			name:   "middle match",
			s:      "vendor/openai/model",
			substr: "openai",
			want:   true,
		},
		{
			name:   "substring in model name",
			s:      "gpt-4",
			substr: "gpt",
			want:   true,
		},
		{
			name:   "substring in hyphenated name",
			s:      "claude-3-opus",
			substr: "claude",
			want:   true,
		},
		{
			name:   "no match",
			s:      "anthropic",
			substr: "openai",
			want:   false,
		},
		{
			name:   "empty substring",
			s:      "openai",
			substr: "",
			want:   true,
		},
		{
			name:   "empty string",
			s:      "",
			substr: "openai",
			want:   false,
		},
		{
			name:   "both empty",
			s:      "",
			substr: "",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf(
					"contains(%q, %q) = %v, want %v",
					tt.s,
					tt.substr,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestGetProvider(t *testing.T) {
	tests := []struct {
		name  string
		model string
		want  string
	}{
		{
			name:  "openai exact",
			model: "openai",
			want:  "openai",
		},
		{
			name:  "openai with path",
			model: "openai/gpt-4",
			want:  "openai",
		},
		{
			name:  "gpt model",
			model: "gpt-4",
			want:  "openai",
		},
		{
			name:  "gpt with prefix",
			model: "provider/gpt",
			want:  "openai",
		},
		{
			name:  "anthropic exact",
			model: "anthropic",
			want:  "anthropic",
		},
		{
			name:  "anthropic with path",
			model: "anthropic/claude-3",
			want:  "anthropic",
		},
		{
			name:  "claude model",
			model: "claude-3-opus",
			want:  "anthropic",
		},
		{
			name:  "claude with prefix",
			model: "provider/claude",
			want:  "anthropic",
		},
		{
			name:  "unknown model",
			model: "gemini",
			want:  "",
		},
		{
			name:  "empty string",
			model: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getProvider(tt.model)
			if got != tt.want {
				t.Errorf(
					"getProvider(%q) = %q, want %q",
					tt.model,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestValidateAuth(t *testing.T) {
	// Create temp directory for test auth configs
	tempDir := t.TempDir()

	// Helper to create auth config file
	createAuthConfig := func(t *testing.T, config AuthConfig) string {
		t.Helper()
		authDir := filepath.Join(
			tempDir,
			".local",
			"share",
			"opencode",
		)
		if err := os.MkdirAll(authDir, 0755); err != nil {
			t.Fatalf("create auth dir: %v", err)
		}

		authPath := filepath.Join(authDir, "auth.json")
		data, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("marshal config: %v", err)
		}

		if err := os.WriteFile(authPath, data, 0644); err != nil {
			t.Fatalf("write auth file: %v", err)
		}

		return tempDir
	}

	// Future time (1 hour from now)
	futureTime := time.Now().Add(1 * time.Hour).UnixMilli()

	// Past time (1 hour ago)
	pastTime := time.Now().Add(-1 * time.Hour).UnixMilli()

	// Soon time (3 minutes from now)
	soonTime := time.Now().Add(3 * time.Minute).UnixMilli()

	tests := []struct {
		name        string
		model       string
		config      AuthConfig
		wantErr     bool
		errContains string
		setupHome   bool
	}{
		{
			name:  "valid openai auth",
			model: "openai/gpt-4",
			config: AuthConfig{
				OpenAI: struct {
					Type      string `json:"type"`
					Refresh   string `json:"refresh"`
					Access    string `json:"access"`
					Expires   int64  `json:"expires"`
					AccountID string `json:"accountId"`
				}{
					Type:    "oauth",
					Expires: futureTime,
				},
			},
			wantErr:   false,
			setupHome: true,
		},
		{
			name:  "valid anthropic auth",
			model: "anthropic/claude-3",
			config: AuthConfig{
				Anthropic: struct {
					Type    string `json:"type"`
					Refresh string `json:"refresh"`
					Access  string `json:"access"`
					Expires int64  `json:"expires"`
				}{
					Type:    "oauth",
					Expires: futureTime,
				},
			},
			wantErr:   false,
			setupHome: true,
		},
		{
			name:  "expired openai token",
			model: "gpt-4",
			config: AuthConfig{
				OpenAI: struct {
					Type      string `json:"type"`
					Refresh   string `json:"refresh"`
					Access    string `json:"access"`
					Expires   int64  `json:"expires"`
					AccountID string `json:"accountId"`
				}{
					Type:    "oauth",
					Expires: pastTime,
				},
			},
			wantErr:     true,
			errContains: "token expired",
			setupHome:   true,
		},
		{
			name:  "missing openai token",
			model: "gpt-4",
			config: AuthConfig{
				OpenAI: struct {
					Type      string `json:"type"`
					Refresh   string `json:"refresh"`
					Access    string `json:"access"`
					Expires   int64  `json:"expires"`
					AccountID string `json:"accountId"`
				}{
					Type:    "oauth",
					Expires: 0,
				},
			},
			wantErr:     true,
			errContains: "token not found",
			setupHome:   true,
		},
		{
			name:  "unknown provider returns error",
			model: "gemini/pro",
			config: AuthConfig{
				OpenAI: struct {
					Type      string `json:"type"`
					Refresh   string `json:"refresh"`
					Access    string `json:"access"`
					Expires   int64  `json:"expires"`
					AccountID string `json:"accountId"`
				}{
					Type:    "oauth",
					Expires: pastTime,
				},
			},
			wantErr:     true,
			errContains: "unknown model provider",
			setupHome:   false,
		},
		{
			name:  "token expiring soon",
			model: "claude-3",
			config: AuthConfig{
				Anthropic: struct {
					Type    string `json:"type"`
					Refresh string `json:"refresh"`
					Access  string `json:"access"`
					Expires int64  `json:"expires"`
				}{
					Type:    "oauth",
					Expires: soonTime,
				},
			},
			wantErr:   false,
			setupHome: true,
		},
		{
			name:        "empty model string",
			model:       "",
			config:      AuthConfig{},
			wantErr:     true,
			errContains: "unknown model provider",
			setupHome:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupHome {
				// Set HOME to temp directory for this test
				home := createAuthConfig(t, tt.config)
				oldHome := os.Getenv("HOME")
				os.Setenv("HOME", home)
				defer os.Setenv("HOME", oldHome)
			}

			err := ValidateAuth(tt.model)

			if (err != nil) != tt.wantErr {
				t.Errorf(
					"ValidateAuth(%q) error = %v, wantErr %v",
					tt.model,
					err,
					tt.wantErr,
				)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf(
						"error = %v, want error containing %q",
						err,
						tt.errContains,
					)
				}
			}
		})
	}
}

func TestValidateAuth_MissingFile(t *testing.T) {
	// Create temp directory without auth file
	tempDir := t.TempDir()

	// Set HOME to temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	err := ValidateAuth("openai/gpt-4")
	if err == nil {
		t.Error("ValidateAuth() expected error for missing auth file")
	}

	if !strings.Contains(err.Error(), "failed to read auth config") {
		t.Errorf(
			"error = %v, want error containing %q",
			err,
			"failed to read auth config",
		)
	}
}

func TestValidateAuth_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()

	// Create auth directory
	authDir := filepath.Join(tempDir, ".local", "share", "opencode")
	if err := os.MkdirAll(authDir, 0755); err != nil {
		t.Fatalf("create auth dir: %v", err)
	}

	// Write invalid JSON
	authPath := filepath.Join(authDir, "auth.json")
	if err := os.WriteFile(authPath, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("write auth file: %v", err)
	}

	// Set HOME to temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	err := ValidateAuth("openai/gpt-4")
	if err == nil {
		t.Error("ValidateAuth() expected error for invalid JSON")
	}

	if !strings.Contains(err.Error(), "failed to parse auth config") {
		t.Errorf(
			"error = %v, want error containing %q",
			err,
			"failed to parse auth config",
		)
	}
}
