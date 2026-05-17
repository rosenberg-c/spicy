package main

import (
	"errors"
	"io/fs"
	"os"
	"testing"
	"time"
)

func TestHelloText_IsExact(t *testing.T) {
	// @req IMGWALKER-002
	if helloText != "Hello, World!" {
		t.Fatalf("helloText = %q, want %q", helloText, "Hello, World!")
	}
}

func TestStartupConfig_Defaults(t *testing.T) {
	// @req IMGWALKER-001
	cfg := startupConfig()

	if cfg.title != "ImgWalker" {
		t.Fatalf("cfg.title = %q, want %q", cfg.title, "ImgWalker")
	}
	if cfg.width != 640 {
		t.Fatalf("cfg.width = %d, want %d", cfg.width, 640)
	}
	if cfg.height != 360 {
		t.Fatalf("cfg.height = %d, want %d", cfg.height, 360)
	}
}

func TestTheme_UsesAskwrapperPalette(t *testing.T) {
	// @req IMGWALKER-003
	theme := newTheme()

	if theme.Palette.Bg != colorBg {
		t.Fatalf("bg color = %#v, want %#v", theme.Palette.Bg, colorBg)
	}
	if theme.Palette.Fg != colorFg {
		t.Fatalf("fg color = %#v, want %#v", theme.Palette.Fg, colorFg)
	}
	if theme.Palette.ContrastBg != colorContrastBg {
		t.Fatalf("contrast bg = %#v, want %#v", theme.Palette.ContrastBg, colorContrastBg)
	}
	if theme.Palette.ContrastFg != colorContrastFg {
		t.Fatalf("contrast fg = %#v, want %#v", theme.Palette.ContrastFg, colorContrastFg)
	}
}

func TestConfigFilePath_UsesXDGHomeLocation(t *testing.T) {
	// @req IMGWALKER-004
	path := configFilePath("/home/alex")
	if path != "/home/alex/.config/spicy/imgwalker.json" {
		t.Fatalf("config path = %q, want %q", path, "/home/alex/.config/spicy/imgwalker.json")
	}
}

func TestLoadStartupFileConfig_LoadsImageDir(t *testing.T) {
	// @req IMGWALKER-005
	cfg, err := loadStartupFileConfig(
		"/home/alex",
		func(string) ([]byte, error) { return []byte(`{"imageDir":"~/screenshots"}`), nil },
	)
	if err != nil {
		t.Fatalf("loadStartupFileConfig error: %v", err)
	}
	if cfg.ImageDir != "~/screenshots" {
		t.Fatalf("cfg.ImageDir = %q, want %q", cfg.ImageDir, "~/screenshots")
	}
}

func TestLoadStartupFileConfig_MissingFileReturnsNotFound(t *testing.T) {
	// @req IMGWALKER-006
	_, err := loadStartupFileConfig(
		"/home/alex",
		func(string) ([]byte, error) { return nil, os.ErrNotExist },
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var cfgErr *ConfigLoadError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("error type = %T, want *ConfigLoadError", err)
	}
	if cfgErr.Code != ConfigLoadErrorNotFound {
		t.Fatalf("cfgErr.Code = %q, want %q", cfgErr.Code, ConfigLoadErrorNotFound)
	}
}

func TestLoadStartupFileConfig_InvalidJSONReturnsInvalidConfig(t *testing.T) {
	// @req IMGWALKER-007
	_, err := loadStartupFileConfig(
		"/home/alex",
		func(string) ([]byte, error) { return []byte(`{"imageDir":`), nil },
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var cfgErr *ConfigLoadError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("error type = %T, want *ConfigLoadError", err)
	}
	if cfgErr.Code != ConfigLoadErrorInvalidConfig {
		t.Fatalf("cfgErr.Code = %q, want %q", cfgErr.Code, ConfigLoadErrorInvalidConfig)
	}
}

func TestLoadStartupFileConfig_ReadFailureReturnsIOError(t *testing.T) {
	_, err := loadStartupFileConfig(
		"/home/alex",
		func(string) ([]byte, error) { return nil, errors.New("permission denied") },
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var cfgErr *ConfigLoadError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("error type = %T, want *ConfigLoadError", err)
	}
	if cfgErr.Code != ConfigLoadErrorIOError {
		t.Fatalf("cfgErr.Code = %q, want %q", cfgErr.Code, ConfigLoadErrorIOError)
	}
}

func TestValidateStartupFileConfig_EmptyImageDirReturnsInvalidConfig(t *testing.T) {
	// @req IMGWALKER-008
	_, err := validateStartupFileConfig(startupFileConfig{ImageDir: ""}, "/home/alex", "/work", func(string) (os.FileInfo, error) {
		return nil, nil
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var cfgErr *ConfigLoadError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("error type = %T, want *ConfigLoadError", err)
	}
	if cfgErr.Code != ConfigLoadErrorInvalidConfig {
		t.Fatalf("cfgErr.Code = %q, want %q", cfgErr.Code, ConfigLoadErrorInvalidConfig)
	}
}

func TestValidateStartupFileConfig_ImageDirMissingReturnsInvalidImageDir(t *testing.T) {
	// @req IMGWALKER-009
	_, err := validateStartupFileConfig(startupFileConfig{ImageDir: "/tmp/missing"}, "/home/alex", "/work", func(string) (os.FileInfo, error) {
		return nil, os.ErrNotExist
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var cfgErr *ConfigLoadError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("error type = %T, want *ConfigLoadError", err)
	}
	if cfgErr.Code != ConfigLoadErrorInvalidImageDir {
		t.Fatalf("cfgErr.Code = %q, want %q", cfgErr.Code, ConfigLoadErrorInvalidImageDir)
	}
}

func TestValidateStartupFileConfig_ImageDirNotDirectoryReturnsInvalidImageDir(t *testing.T) {
	// @req IMGWALKER-010
	_, err := validateStartupFileConfig(startupFileConfig{ImageDir: "/tmp/file"}, "/home/alex", "/work", func(string) (os.FileInfo, error) {
		return stubFileInfo{isDir: false}, nil
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var cfgErr *ConfigLoadError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("error type = %T, want *ConfigLoadError", err)
	}
	if cfgErr.Code != ConfigLoadErrorInvalidImageDir {
		t.Fatalf("cfgErr.Code = %q, want %q", cfgErr.Code, ConfigLoadErrorInvalidImageDir)
	}
}

func TestValidateStartupFileConfig_StatFailureReturnsInvalidImageDir(t *testing.T) {
	_, err := validateStartupFileConfig(startupFileConfig{ImageDir: "/tmp/images"}, "/home/alex", "/work", func(string) (os.FileInfo, error) {
		return nil, errors.New("permission denied")
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var cfgErr *ConfigLoadError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("error type = %T, want *ConfigLoadError", err)
	}
	if cfgErr.Code != ConfigLoadErrorInvalidImageDir {
		t.Fatalf("cfgErr.Code = %q, want %q", cfgErr.Code, ConfigLoadErrorInvalidImageDir)
	}
}

func TestValidateStartupFileConfig_ValidImageDirPasses(t *testing.T) {
	_, err := validateStartupFileConfig(startupFileConfig{ImageDir: "/tmp/images"}, "/home/alex", "/work", func(string) (os.FileInfo, error) {
		return stubFileInfo{isDir: true}, nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestValidateStartupFileConfig_ExpandsTildeBeforeStat(t *testing.T) {
	// @req IMGWALKER-011
	calledPath := ""
	cfg, err := validateStartupFileConfig(startupFileConfig{ImageDir: "~/screenshots"}, "/home/alex", "/work", func(path string) (os.FileInfo, error) {
		calledPath = path
		return stubFileInfo{isDir: true}, nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calledPath != "/home/alex/screenshots" {
		t.Fatalf("stat path = %q, want %q", calledPath, "/home/alex/screenshots")
	}
	if cfg.ImageDir != "/home/alex/screenshots" {
		t.Fatalf("cfg.ImageDir = %q, want %q", cfg.ImageDir, "/home/alex/screenshots")
	}
}

func TestValidateStartupFileConfig_ResolvesRelativePathBeforeStat(t *testing.T) {
	// @req IMGWALKER-012
	calledPath := ""
	cfg, err := validateStartupFileConfig(startupFileConfig{ImageDir: "screenshots"}, "/home/alex", "/work/project", func(path string) (os.FileInfo, error) {
		calledPath = path
		return stubFileInfo{isDir: true}, nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calledPath != "/work/project/screenshots" {
		t.Fatalf("stat path = %q, want %q", calledPath, "/work/project/screenshots")
	}
	if cfg.ImageDir != "/work/project/screenshots" {
		t.Fatalf("cfg.ImageDir = %q, want %q", cfg.ImageDir, "/work/project/screenshots")
	}
}

func TestValidateStartupFileConfig_TildeWithoutHomeReturnsInvalidConfig(t *testing.T) {
	errCfg, err := validateStartupFileConfig(startupFileConfig{ImageDir: "~/screenshots"}, "", "/work", func(string) (os.FileInfo, error) {
		return stubFileInfo{isDir: true}, nil
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errCfg != (startupFileConfig{}) {
		t.Fatalf("cfg = %#v, want zero value", errCfg)
	}
	var cfgErr *ConfigLoadError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("error type = %T, want *ConfigLoadError", err)
	}
	if cfgErr.Code != ConfigLoadErrorInvalidConfig {
		t.Fatalf("cfgErr.Code = %q, want %q", cfgErr.Code, ConfigLoadErrorInvalidConfig)
	}
}

func TestValidateStartupFileConfig_RelativeWithoutWorkingDirReturnsInvalidConfig(t *testing.T) {
	errCfg, err := validateStartupFileConfig(startupFileConfig{ImageDir: "screenshots"}, "/home/alex", "", func(string) (os.FileInfo, error) {
		return stubFileInfo{isDir: true}, nil
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errCfg != (startupFileConfig{}) {
		t.Fatalf("cfg = %#v, want zero value", errCfg)
	}
	var cfgErr *ConfigLoadError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("error type = %T, want *ConfigLoadError", err)
	}
	if cfgErr.Code != ConfigLoadErrorInvalidConfig {
		t.Fatalf("cfgErr.Code = %q, want %q", cfgErr.Code, ConfigLoadErrorInvalidConfig)
	}
}

func TestBuildStartupWindowConfig_UsesEmptyImageDirOnResolveError(t *testing.T) {
	// @req IMGWALKER-013
	cfg := buildStartupWindowConfig(func() (startupFileConfig, error) {
		return startupFileConfig{}, &ConfigLoadError{Code: ConfigLoadErrorNotFound, Err: os.ErrNotExist}
	})
	if cfg.imageDir != "" {
		t.Fatalf("cfg.imageDir = %q, want empty", cfg.imageDir)
	}
}

type stubFileInfo struct {
	isDir bool
}

func (s stubFileInfo) Name() string       { return "stub" }
func (s stubFileInfo) Size() int64        { return 0 }
func (s stubFileInfo) Mode() fs.FileMode  { return 0 }
func (s stubFileInfo) ModTime() time.Time { return time.Time{} }
func (s stubFileInfo) IsDir() bool        { return s.isDir }
func (s stubFileInfo) Sys() any           { return nil }
