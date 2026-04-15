package constants

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveDefaultModel(t *testing.T) {
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	t.Setenv(defaultModelEnvKey, "")

	t.Run("uses environment variable", func(t *testing.T) {
		t.Setenv(defaultModelEnvKey, "openai/gpt-4.1")
		if got := resolveDefaultModel(); got != "openai/gpt-4.1" {
			t.Fatalf("resolveDefaultModel() = %q, want %q", got, "openai/gpt-4.1")
		}
	})

	t.Run("uses dot env variable", func(t *testing.T) {
		t.Setenv(defaultModelEnvKey, "")

		tempDir := t.TempDir()
		envPath := filepath.Join(tempDir, ".env")
		if err := os.WriteFile(envPath, []byte("SPICY_MODEL=openai/gpt-4o\n"), 0o644); err != nil {
			t.Fatalf("write .env: %v", err)
		}

		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("chdir: %v", err)
		}
		defer func() {
			_ = os.Chdir(originalWD)
		}()

		if got := resolveDefaultModel(); got != "openai/gpt-4o" {
			t.Fatalf("resolveDefaultModel() = %q, want %q", got, "openai/gpt-4o")
		}
	})

	t.Run("falls back to default", func(t *testing.T) {
		t.Setenv(defaultModelEnvKey, "")

		tempDir := t.TempDir()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("chdir: %v", err)
		}
		defer func() {
			_ = os.Chdir(originalWD)
		}()

		if got := resolveDefaultModel(); got != defaultModelFallback {
			t.Fatalf("resolveDefaultModel() = %q, want %q", got, defaultModelFallback)
		}
	})

	t.Run("uses home config dot env variable", func(t *testing.T) {
		t.Setenv(defaultModelEnvKey, "")

		tempHome := t.TempDir()
		t.Setenv("HOME", tempHome)

		homeConfigDir := filepath.Join(tempHome, ".config", "spicy")
		if err := os.MkdirAll(homeConfigDir, 0o755); err != nil {
			t.Fatalf("mkdir home config dir: %v", err)
		}

		homeEnvPath := filepath.Join(homeConfigDir, ".env")
		if err := os.WriteFile(homeEnvPath, []byte("SPICY_MODEL=openai/gpt-4.5\n"), 0o644); err != nil {
			t.Fatalf("write home .env: %v", err)
		}

		tempDir := t.TempDir()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("chdir: %v", err)
		}
		defer func() {
			_ = os.Chdir(originalWD)
		}()

		if got := resolveDefaultModel(); got != "openai/gpt-4.5" {
			t.Fatalf("resolveDefaultModel() = %q, want %q", got, "openai/gpt-4.5")
		}
	})

	t.Run("prefers local dot env over home config", func(t *testing.T) {
		t.Setenv(defaultModelEnvKey, "")

		tempHome := t.TempDir()
		t.Setenv("HOME", tempHome)

		homeConfigDir := filepath.Join(tempHome, ".config", "spicy")
		if err := os.MkdirAll(homeConfigDir, 0o755); err != nil {
			t.Fatalf("mkdir home config dir: %v", err)
		}

		homeEnvPath := filepath.Join(homeConfigDir, ".env")
		if err := os.WriteFile(homeEnvPath, []byte("SPICY_MODEL=openai/gpt-home\n"), 0o644); err != nil {
			t.Fatalf("write home .env: %v", err)
		}

		localDir := t.TempDir()
		localEnvPath := filepath.Join(localDir, ".env")
		if err := os.WriteFile(localEnvPath, []byte("SPICY_MODEL=openai/gpt-local\n"), 0o644); err != nil {
			t.Fatalf("write local .env: %v", err)
		}

		if err := os.Chdir(localDir); err != nil {
			t.Fatalf("chdir: %v", err)
		}
		defer func() {
			_ = os.Chdir(originalWD)
		}()

		if got := resolveDefaultModel(); got != "openai/gpt-local" {
			t.Fatalf("resolveDefaultModel() = %q, want %q", got, "openai/gpt-local")
		}
	})
}

func TestLookupDotEnv(t *testing.T) {
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")
	content := []byte("\n# comment\nexport SPICY_MODEL=\"openai/gpt-4.1\"\nOTHER=value\n")
	if err := os.WriteFile(envPath, content, 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	got, ok := lookupDotEnv(defaultModelEnvKey)
	if !ok {
		t.Fatalf("lookupDotEnv() returned ok=false")
	}
	if got != "openai/gpt-4.1" {
		t.Fatalf("lookupDotEnv() = %q, want %q", got, "openai/gpt-4.1")
	}
}
