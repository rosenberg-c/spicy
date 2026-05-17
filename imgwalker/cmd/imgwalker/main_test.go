package main

import (
	"errors"
	"image/color"
	"io/fs"
	"os"
	"reflect"
	"testing"
	"time"

	"gioui.org/io/key"
)

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
	// @req IMGWALKER-020
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
	// @req IMGWALKER-027
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

func TestSplitPaneWidths_UsesLeftListAndRightPreviewWidths(t *testing.T) {
	// @req IMGWALKER-014
	left, right := splitPaneWidths(1000)
	if left != 350 {
		t.Fatalf("left = %d, want %d", left, 350)
	}
	if right != 650 {
		t.Fatalf("right = %d, want %d", right, 650)
	}
}

func TestListImageFiles_FiltersAndSortsSupportedImages(t *testing.T) {
	// @req IMGWALKER-015, IMGWALKER-028
	images, err := listImageFiles("/images", func(string) ([]os.DirEntry, error) {
		return []os.DirEntry{
			stubDirEntry{name: "z-last.JPG"},
			stubDirEntry{name: "nested", isDir: true},
			stubDirEntry{name: "a-first.png"},
			stubDirEntry{name: "notes.txt"},
			stubDirEntry{name: "mid.WeBp"},
		}, nil
	})
	if err != nil {
		t.Fatalf("listImageFiles error: %v", err)
	}

	want := []string{"a-first.png", "mid.WeBp", "z-last.JPG"}
	if !reflect.DeepEqual(images, want) {
		t.Fatalf("images = %#v, want %#v", images, want)
	}
}

func TestNoImagesText_IsExact(t *testing.T) {
	// @req IMGWALKER-016
	if uiText.NoImagesFound != "No images found" {
		t.Fatalf("uiText.NoImagesFound = %q, want %q", uiText.NoImagesFound, "No images found")
	}
}

func TestNextSelectionIndex_JAndKMoveWithinBounds(t *testing.T) {
	// @req IMGWALKER-017
	tests := []struct {
		name        string
		current     int
		total       int
		keyName     key.Name
		wantIndex   int
		wantChanged bool
	}{
		{name: "j increments", current: 0, total: 3, keyName: key.Name("j"), wantIndex: 1, wantChanged: true},
		{name: "k decrements", current: 1, total: 3, keyName: key.Name("k"), wantIndex: 0, wantChanged: true},
		{name: "j clamps at end", current: 2, total: 3, keyName: key.Name("j"), wantIndex: 2, wantChanged: true},
		{name: "k clamps at start", current: 0, total: 3, keyName: key.Name("k"), wantIndex: 0, wantChanged: true},
		{name: "down arrow increments", current: 0, total: 3, keyName: key.NameDownArrow, wantIndex: 1, wantChanged: true},
		{name: "up arrow decrements", current: 1, total: 3, keyName: key.NameUpArrow, wantIndex: 0, wantChanged: true},
		{name: "unknown key unchanged", current: 1, total: 3, keyName: key.Name("x"), wantIndex: 1, wantChanged: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, changed := nextSelectionIndex(tc.current, tc.total, tc.keyName)
			if got != tc.wantIndex || changed != tc.wantChanged {
				t.Fatalf("nextSelectionIndex(%d,%d,%q) = (%d,%t), want (%d,%t)", tc.current, tc.total, tc.keyName, got, changed, tc.wantIndex, tc.wantChanged)
			}
		})
	}
}

func TestListLoadErrorMessage(t *testing.T) {
	// @req IMGWALKER-026
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "nil error returns empty", err: nil, want: ""},
		{name: "error returns safe static copy", err: errors.New("permission denied"), want: "Unable to load images from configured directory."},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := listLoadErrorMessage(tc.err)
			if got != tc.want {
				t.Fatalf("listLoadErrorMessage() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestListRowColors_SelectedDiffersFromDefault(t *testing.T) {
	// @req IMGWALKER-018
	defaultBg, defaultFg := listRowColors(false)
	selectedBg, selectedFg := listRowColors(true)
	if defaultBg == selectedBg {
		t.Fatalf("selected bg should differ from default bg: %#v", selectedBg)
	}
	if defaultFg == selectedFg {
		t.Fatalf("selected fg should differ from default fg: %#v", selectedFg)
	}
}

func TestSelectedImagePath_ReturnsFullPathForSelection(t *testing.T) {
	// @req IMGWALKER-019
	got := selectedImagePath("/home/alex/screenshots", []string{"a.png", "b.jpg"}, 1)
	if got != "/home/alex/screenshots/b.jpg" {
		t.Fatalf("selectedImagePath = %q, want %q", got, "/home/alex/screenshots/b.jpg")
	}
}

func TestPreviewPaneText_UsesEmptyTextWhenNoImages(t *testing.T) {
	// @req IMGWALKER-021
	got := previewPaneText("/home/alex/screenshots", nil, -1)
	if got != "" {
		t.Fatalf("previewPaneText = %q, want empty string", got)
	}
}

func TestPreviewPaneText_DefaultsToFirstImageWhenSelectionInvalid(t *testing.T) {
	// @req IMGWALKER-025
	got := previewPaneText("/home/alex/screenshots", []string{"a.png", "b.jpg"}, -1)
	if got != "/home/alex/screenshots/a.png" {
		t.Fatalf("previewPaneText = %q, want %q", got, "/home/alex/screenshots/a.png")
	}
}

func TestSelectedIndexAfterClickedIndexes_ClickSetsSelection(t *testing.T) {
	// @req IMGWALKER-022
	tests := []struct {
		name          string
		current       int
		total         int
		clicked       []int
		wantSelection int
	}{
		{name: "single click selects row", current: 0, total: 4, clicked: []int{2}, wantSelection: 2},
		{name: "no click keeps current", current: 1, total: 4, clicked: nil, wantSelection: 1},
		{name: "multiple clicks use latest", current: 0, total: 4, clicked: []int{1, 3}, wantSelection: 3},
		{name: "out of bounds clicks ignored", current: 2, total: 4, clicked: []int{-1, 10}, wantSelection: 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := selectedIndexAfterClickedIndexes(tc.current, tc.total, tc.clicked)
			if got != tc.wantSelection {
				t.Fatalf("selectedIndexAfterClickedIndexes(%d,%d,%v) = %d, want %d", tc.current, tc.total, tc.clicked, got, tc.wantSelection)
			}
		})
	}
}

func TestListPaneBackgroundColor_UsesThemeBg(t *testing.T) {
	// @req IMGWALKER-023
	th := newTheme()
	if got := listPaneBackgroundColor(th); got != th.Palette.Bg {
		t.Fatalf("listPaneBackgroundColor = %#v, want %#v", got, th.Palette.Bg)
	}
}

func TestPaneDividerColor_IsVisibleAndStable(t *testing.T) {
	// @req IMGWALKER-024
	if got := paneDividerColor(); got != (color.NRGBA{R: 64, G: 74, B: 89, A: 255}) {
		t.Fatalf("paneDividerColor = %#v, want %#v", got, color.NRGBA{R: 64, G: 74, B: 89, A: 255})
	}
}

type stubDirEntry struct {
	name  string
	isDir bool
}

func (s stubDirEntry) Name() string               { return s.name }
func (s stubDirEntry) IsDir() bool                { return s.isDir }
func (s stubDirEntry) Type() fs.FileMode          { return 0 }
func (s stubDirEntry) Info() (fs.FileInfo, error) { return stubFileInfo{isDir: s.isDir}, nil }

type stubFileInfo struct {
	isDir bool
}

func (s stubFileInfo) Name() string       { return "stub" }
func (s stubFileInfo) Size() int64        { return 0 }
func (s stubFileInfo) Mode() fs.FileMode  { return 0 }
func (s stubFileInfo) ModTime() time.Time { return time.Time{} }
func (s stubFileInfo) IsDir() bool        { return s.isDir }
func (s stubFileInfo) Sys() any           { return nil }
