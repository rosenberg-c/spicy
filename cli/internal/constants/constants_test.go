package constants

import (
	"os"
	"path/filepath"
	"testing"
)

// @req CORE-CLI-005
func TestResolveDefaultModel(t *testing.T) {
	t.Run("uses global config default model when no local config exists", func(t *testing.T) {
		writeGlobalConfig(t, `{"default_model":"openai/gpt-4.1"}`)
		chdirClean(t)

		if got := resolveDefaultModel(); got != "openai/gpt-4.1" {
			t.Fatalf("resolveDefaultModel() = %q, want %q", got, "openai/gpt-4.1")
		}
	})

	t.Run("local default_model overrides global", func(t *testing.T) {
		writeGlobalConfig(t, `{"default_model":"openai/gpt-global"}`)
		writeLocalConfig(t, `{"default_model":"openai/gpt-local"}`)

		if got := resolveDefaultModel(); got != "openai/gpt-local" {
			t.Fatalf("resolveDefaultModel() = %q, want %q", got, "openai/gpt-local")
		}
	})

	t.Run("empty local default_model keeps global", func(t *testing.T) {
		writeGlobalConfig(t, `{"default_model":"openai/gpt-global"}`)
		writeLocalConfig(t, `{"default_model":""}`)

		if got := resolveDefaultModel(); got != "openai/gpt-global" {
			t.Fatalf("resolveDefaultModel() = %q, want %q", got, "openai/gpt-global")
		}
	})

	t.Run("invalid local json falls back to global", func(t *testing.T) {
		writeGlobalConfig(t, `{"default_model":"openai/gpt-global"}`)
		writeLocalConfig(t, `{invalid json`)

		if got := resolveDefaultModel(); got != "openai/gpt-global" {
			t.Fatalf("resolveDefaultModel() = %q, want %q", got, "openai/gpt-global")
		}
	})

	t.Run("local config is discovered from a subdirectory of the repo", func(t *testing.T) {
		writeGlobalConfig(t, `{"default_model":"openai/gpt-global"}`)

		repoDir := t.TempDir()
		writeLocalConfigFile(t, filepath.Join(repoDir, localConfigFileName),
			`{"default_model":"openai/gpt-local"}`)

		subDir := filepath.Join(repoDir, "nested", "deep")
		if err := os.MkdirAll(subDir, 0o755); err != nil {
			t.Fatalf("mkdir sub dir: %v", err)
		}
		chdir(t, subDir)

		if got := resolveDefaultModel(); got != "openai/gpt-local" {
			t.Fatalf("resolveDefaultModel() = %q, want %q", got, "openai/gpt-local")
		}
	})

	t.Run("falls back to built-in default when no config exists", func(t *testing.T) {
		setTempHome(t)
		chdirClean(t)

		if got := resolveDefaultModel(); got != defaultModelFallback {
			t.Fatalf("resolveDefaultModel() = %q, want %q", got, defaultModelFallback)
		}
	})

	t.Run("ignores invalid global config json", func(t *testing.T) {
		writeGlobalConfig(t, `{invalid json`)
		chdirClean(t)

		if got := resolveDefaultModel(); got != defaultModelFallback {
			t.Fatalf("resolveDefaultModel() = %q, want %q", got, defaultModelFallback)
		}
	})
}

// @req CORE-CLI-005, CORE-CLI-009
func TestModelForCommand(t *testing.T) {
	t.Run("uses command specific model over default model", func(t *testing.T) {
		writeGlobalConfig(t, `{"default_model":"openai/gpt-default","models":{"ask":"openai/gpt-ask"}}`)
		chdirClean(t)

		if got := ModelForCommand("ask"); got != "openai/gpt-ask" {
			t.Fatalf("ModelForCommand(ask) = %q, want %q", got, "openai/gpt-ask")
		}
	})

	t.Run("uses default model when command model is absent", func(t *testing.T) {
		writeGlobalConfig(t, `{"default_model":"openai/gpt-default"}`)
		chdirClean(t)

		if got := ModelForCommand("explain"); got != "openai/gpt-default" {
			t.Fatalf(
				"ModelForCommand(explain) = %q, want %q",
				got,
				"openai/gpt-default",
			)
		}
	})

	t.Run("falls back to built in default when config is missing", func(t *testing.T) {
		setTempHome(t)
		chdirClean(t)

		if got := ModelForCommand("gitmessage"); got != defaultModelFallback {
			t.Fatalf("ModelForCommand(gitmessage) = %q, want %q", got, defaultModelFallback)
		}
	})

	t.Run("local models entry overrides global models entry", func(t *testing.T) {
		writeGlobalConfig(t,
			`{"default_model":"openai/gpt-default","models":{"ask":"openai/gpt-global-ask"}}`)
		writeLocalConfig(t, `{"models":{"ask":"openai/gpt-local-ask"}}`)

		if got := ModelForCommand("ask"); got != "openai/gpt-local-ask" {
			t.Fatalf("ModelForCommand(ask) = %q, want %q", got, "openai/gpt-local-ask")
		}
	})

	t.Run("global only models keys are preserved", func(t *testing.T) {
		writeGlobalConfig(t,
			`{"models":{"ask":"openai/gpt-global-ask","tutor":"openai/gpt-global-tutor"}}`)
		writeLocalConfig(t, `{"models":{"ask":"openai/gpt-local-ask"}}`)

		if got := ModelForCommand("tutor"); got != "openai/gpt-global-tutor" {
			t.Fatalf("ModelForCommand(tutor) = %q, want %q", got, "openai/gpt-global-tutor")
		}
	})

	t.Run("empty local models entry keeps global entry", func(t *testing.T) {
		writeGlobalConfig(t, `{"models":{"ask":"openai/gpt-global-ask"}}`)
		writeLocalConfig(t, `{"models":{"ask":""}}`)

		if got := ModelForCommand("ask"); got != "openai/gpt-global-ask" {
			t.Fatalf("ModelForCommand(ask) = %q, want %q", got, "openai/gpt-global-ask")
		}
	})

	t.Run("full precedence with local and global present", func(t *testing.T) {
		writeGlobalConfig(t,
			`{"default_model":"openai/gpt-global-default","models":{"ask":"openai/gpt-global-ask"}}`)
		writeLocalConfig(t,
			`{"default_model":"openai/gpt-local-default","models":{"ask":"openai/gpt-local-ask"}}`)

		// models.<command> (local) wins over everything.
		if got := ModelForCommand("ask"); got != "openai/gpt-local-ask" {
			t.Fatalf("ModelForCommand(ask) = %q, want %q", got, "openai/gpt-local-ask")
		}
		// default_model: local wins over global.
		if got := ModelForCommand("explain"); got != "openai/gpt-local-default" {
			t.Fatalf("ModelForCommand(explain) = %q, want %q", got, "openai/gpt-local-default")
		}
	})

	t.Run("invalid local json falls back to global model", func(t *testing.T) {
		writeGlobalConfig(t,
			`{"default_model":"openai/gpt-global-default","models":{"ask":"openai/gpt-global-ask"}}`)
		writeLocalConfig(t, `{invalid json`)

		if got := ModelForCommand("ask"); got != "openai/gpt-global-ask" {
			t.Fatalf("ModelForCommand(ask) = %q, want %q", got, "openai/gpt-global-ask")
		}
		if got := ModelForCommand("explain"); got != "openai/gpt-global-default" {
			t.Fatalf("ModelForCommand(explain) = %q, want %q", got, "openai/gpt-global-default")
		}
	})
}

// @req CORE-CLI-009
func TestFindLocalConfigFile(t *testing.T) {
	t.Run("returns the nearest file walking up from the start dir", func(t *testing.T) {
		repoDir := t.TempDir()
		want := filepath.Join(repoDir, localConfigFileName)
		writeLocalConfigFile(t, want, `{}`)

		nested := filepath.Join(repoDir, "a", "b")
		if err := os.MkdirAll(nested, 0o755); err != nil {
			t.Fatalf("mkdir nested: %v", err)
		}

		got, ok := findLocalConfigFile(nested)
		if !ok {
			t.Fatal("findLocalConfigFile() = false, want true")
		}
		if got != want {
			t.Fatalf("findLocalConfigFile() = %q, want %q", got, want)
		}
	})

	t.Run("nearest file wins over farther files", func(t *testing.T) {
		root := t.TempDir()
		writeLocalConfigFile(t, filepath.Join(root, localConfigFileName), `{}`)

		child := filepath.Join(root, "child")
		want := filepath.Join(child, localConfigFileName)
		writeLocalConfigFile(t, want, `{}`)

		got, ok := findLocalConfigFile(child)
		if !ok {
			t.Fatal("findLocalConfigFile() = false, want true")
		}
		if got != want {
			t.Fatalf("findLocalConfigFile() = %q, want %q", got, want)
		}
	})

	t.Run("skips non-regular files", func(t *testing.T) {
		dir := t.TempDir()
		nonRegular := filepath.Join(dir, localConfigFileName)
		if err := os.MkdirAll(nonRegular, 0o755); err != nil {
			t.Fatalf("mkdir non-regular: %v", err)
		}

		if got, ok := findLocalConfigFile(dir); ok {
			t.Fatalf("findLocalConfigFile() = (%q, true), want (_, false) for a directory", got)
		}
	})

	t.Run("returns false when no file exists up the tree", func(t *testing.T) {
		if _, ok := findLocalConfigFile(t.TempDir()); ok {
			t.Fatal("findLocalConfigFile() = true, want false")
		}
	})
}

// @req CORE-CLI-009
func TestLoadConfigFile(t *testing.T) {
	t.Run("parses a valid config file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), localConfigFileName)
		writeLocalConfigFile(t, path,
			`{"default_model":"openai/gpt-4.1","models":{"ask":"openai/gpt-ask"}}`)

		config, ok := loadConfigFile(path)
		if !ok {
			t.Fatal("loadConfigFile() = false, want true")
		}
		if config.DefaultModel != "openai/gpt-4.1" {
			t.Fatalf("config.DefaultModel = %q, want %q", config.DefaultModel, "openai/gpt-4.1")
		}
		if got := config.Models["ask"]; got != "openai/gpt-ask" {
			t.Fatalf("config.Models[ask] = %q, want %q", got, "openai/gpt-ask")
		}
	})

	t.Run("missing file returns not found", func(t *testing.T) {
		if _, ok := loadConfigFile(filepath.Join(t.TempDir(), "absent.json")); ok {
			t.Fatal("loadConfigFile() = true for a missing file, want false")
		}
	})

	t.Run("invalid json returns not found", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), localConfigFileName)
		writeLocalConfigFile(t, path, `{invalid json`)

		if _, ok := loadConfigFile(path); ok {
			t.Fatal("loadConfigFile() = true for invalid json, want false")
		}
	})
}

// @req CORE-CLI-009
func TestMergeConfig(t *testing.T) {
	global := spicyConfig{
		DefaultModel: "openai/gpt-global",
		Models: map[string]string{
			"ask":   "openai/gpt-global-ask",
			"tutor": "openai/gpt-global-tutor",
		},
	}

	t.Run("local default_model overrides global", func(t *testing.T) {
		merged := mergeConfig(global, spicyConfig{DefaultModel: "openai/gpt-local"})
		if merged.DefaultModel != "openai/gpt-local" {
			t.Fatalf("merged.DefaultModel = %q, want %q", merged.DefaultModel, "openai/gpt-local")
		}
	})

	t.Run("empty local default_model keeps global", func(t *testing.T) {
		merged := mergeConfig(global, spicyConfig{DefaultModel: "  "})
		if merged.DefaultModel != "openai/gpt-global" {
			t.Fatalf("merged.DefaultModel = %q, want %q", merged.DefaultModel, "openai/gpt-global")
		}
	})

	t.Run("local models entries override matching global entries", func(t *testing.T) {
		merged := mergeConfig(global, spicyConfig{Models: map[string]string{"ask": "openai/gpt-local-ask"}})
		if got := merged.Models["ask"]; got != "openai/gpt-local-ask" {
			t.Fatalf("merged.Models[ask] = %q, want %q", got, "openai/gpt-local-ask")
		}
	})

	t.Run("global only models keys are preserved", func(t *testing.T) {
		merged := mergeConfig(global, spicyConfig{Models: map[string]string{"ask": "openai/gpt-local-ask"}})
		if got := merged.Models["tutor"]; got != "openai/gpt-global-tutor" {
			t.Fatalf("merged.Models[tutor] = %q, want %q", got, "openai/gpt-global-tutor")
		}
	})

	t.Run("empty local models entry keeps global entry", func(t *testing.T) {
		merged := mergeConfig(global, spicyConfig{Models: map[string]string{"ask": ""}})
		if got := merged.Models["ask"]; got != "openai/gpt-global-ask" {
			t.Fatalf("merged.Models[ask] = %q, want %q", got, "openai/gpt-global-ask")
		}
	})

	t.Run("local config merges over zero global", func(t *testing.T) {
		local := spicyConfig{DefaultModel: "openai/gpt-local"}
		merged := mergeConfig(spicyConfig{}, local)
		if merged.DefaultModel != "openai/gpt-local" {
			t.Fatalf("merged.DefaultModel = %q, want %q", merged.DefaultModel, "openai/gpt-local")
		}
	})
}

// setTempHome points HOME at an empty temp dir so the global config is absent.
func setTempHome(t *testing.T) {
	t.Helper()

	t.Setenv("HOME", t.TempDir())
}

// writeGlobalConfig writes the global config under a temp HOME.
func writeGlobalConfig(t *testing.T, content string) string {
	t.Helper()

	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	configDir := filepath.Join(tempHome, ".config", "spicy")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	path := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config.json: %v", err)
	}

	return path
}

// writeLocalConfig writes a .spicy.config.json into a fresh repo root and
// moves the test process into it.
func writeLocalConfig(t *testing.T, content string) string {
	t.Helper()

	repoDir := t.TempDir()
	path := filepath.Join(repoDir, localConfigFileName)
	writeLocalConfigFile(t, path, content)

	chdir(t, repoDir)
	return path
}

func writeLocalConfigFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir local config dir: %v", err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write local config: %v", err)
	}
}

// chdir moves the test process into dir and restores the previous working
// directory on cleanup.
func chdir(t *testing.T, dir string) {
	t.Helper()

	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(prevWD); err != nil {
			t.Errorf("chdir back to %s: %v", prevWD, err)
		}
	})
}

// chdirClean moves the test process into a fresh empty directory so local
// config discovery cannot pick up a file from the checkout.
func chdirClean(t *testing.T) {
	t.Helper()

	chdir(t, t.TempDir())
}
