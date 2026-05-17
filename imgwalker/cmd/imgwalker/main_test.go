package main

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
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
	// @req IMGWALKER-010-A
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
	// @req IMGWALKER-015, IMGWALKER-015-A
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
	got := selectedImagePath("/home/alex/screenshots", []string{"a.png", "b.jpg"}, 1)
	if got != "/home/alex/screenshots/b.jpg" {
		t.Fatalf("selectedImagePath = %q, want %q", got, "/home/alex/screenshots/b.jpg")
	}
}

func TestPreviewSourcePath_UsesEmptySourceWhenNoImages(t *testing.T) {
	// @req IMGWALKER-021
	got := previewSourcePath("/home/alex/screenshots", nil, -1)
	if got != "" {
		t.Fatalf("previewSourcePath = %q, want empty string", got)
	}
}

func TestPreviewSourcePath_DefaultsToFirstImageWhenSelectionInvalid(t *testing.T) {
	// @req IMGWALKER-025
	got := previewSourcePath("/home/alex/screenshots", []string{"a.png", "b.jpg"}, -1)
	if got != "/home/alex/screenshots/a.png" {
		t.Fatalf("previewSourcePath = %q, want %q", got, "/home/alex/screenshots/a.png")
	}
}

func TestLoadPreviewImage_DecodesValidImage(t *testing.T) {
	// @req IMGWALKER-019
	var encoded bytes.Buffer
	src := image.NewNRGBA(image.Rect(0, 0, 2, 3))
	if err := png.Encode(&encoded, src); err != nil {
		t.Fatalf("encode png: %v", err)
	}

	decoded, err := loadPreviewImage("/tmp/sample.png", func(string) ([]byte, error) {
		return encoded.Bytes(), nil
	})
	if err != nil {
		t.Fatalf("loadPreviewImage returned unexpected error: %v", err)
	}
	if decoded == nil {
		t.Fatal("loadPreviewImage returned nil for valid image")
	}
	b := decoded.Bounds()
	if b.Dx() != 2 || b.Dy() != 3 {
		t.Fatalf("decoded bounds = %v, want 2x3", b)
	}
}

func TestLoadPreviewImage_ReturnsNilOnReadOrDecodeError(t *testing.T) {
	// @req IMGWALKER-027
	readImg, readErr := loadPreviewImage("/tmp/missing.png", func(string) ([]byte, error) {
		return nil, errors.New("read failed")
	})
	if readImg != nil {
		t.Fatalf("loadPreviewImage on read error returned image = %#v, want nil", readImg)
	}
	if readErr == nil {
		t.Fatal("loadPreviewImage on read error returned nil error")
	}

	decodeImg, decodeErr := loadPreviewImage("/tmp/bad.png", func(string) ([]byte, error) {
		return []byte("not-an-image"), nil
	})
	if decodeImg != nil {
		t.Fatalf("loadPreviewImage on decode error returned image = %#v, want nil", decodeImg)
	}
	if decodeErr == nil {
		t.Fatal("loadPreviewImage on decode error returned nil error")
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

func TestShouldScrollSelectionIntoView(t *testing.T) {
	// @req IMGWALKER-042
	tests := []struct {
		name     string
		pos      layout.Position
		selected int
		total    int
		want     bool
	}{
		{name: "no items does not scroll", pos: layout.Position{First: 0, Count: 0}, selected: 0, total: 0, want: false},
		{name: "within visible range does not scroll", pos: layout.Position{First: 3, Count: 5}, selected: 6, total: 20, want: false},
		{name: "above visible range scrolls", pos: layout.Position{First: 3, Count: 5}, selected: 2, total: 20, want: true},
		{name: "below visible range scrolls", pos: layout.Position{First: 3, Count: 5}, selected: 8, total: 20, want: true},
		{name: "partially hidden first row scrolls", pos: layout.Position{First: 3, Count: 5, Offset: 4}, selected: 3, total: 20, want: true},
		{name: "partially hidden last row scrolls", pos: layout.Position{First: 3, Count: 5, OffsetLast: -6}, selected: 7, total: 20, want: true},
		{name: "unknown viewport count scrolls", pos: layout.Position{First: 0, Count: 0}, selected: 1, total: 20, want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldScrollSelectionIntoView(tc.pos, tc.selected, tc.total)
			if got != tc.want {
				t.Fatalf("shouldScrollSelectionIntoView(%+v, %d, %d) = %t, want %t", tc.pos, tc.selected, tc.total, got, tc.want)
			}
		})
	}
}

func TestApplyShortcutKey_JKDoNotSkipMiddleInThreeItems(t *testing.T) {
	// @req IMGWALKER-017
	start := 0
	next, action := applyShortcutKey(start, 3, false, key.Event{Name: key.Name("j"), State: key.Press})
	if action != shortcutActionNone {
		t.Fatalf("applyShortcutKey(j) action = %v, want %v", action, shortcutActionNone)
	}
	if next != 1 {
		t.Fatalf("applyShortcutKey(j) index = %d, want 1", next)
	}

	back, backAction := applyShortcutKey(2, 3, false, key.Event{Name: key.Name("k"), State: key.Press})
	if backAction != shortcutActionNone {
		t.Fatalf("applyShortcutKey(k) action = %v, want %v", backAction, shortcutActionNone)
	}
	if back != 1 {
		t.Fatalf("applyShortcutKey(k) index = %d, want 1", back)
	}
}

func TestApplyShortcutKey_DeleteArmConfirmAndEscape(t *testing.T) {
	// @req IMGWALKER-029, IMGWALKER-030, IMGWALKER-031, IMGWALKER-032
	tests := []struct {
		name          string
		event         key.Event
		deleteArmed   bool
		wantAction    shortcutAction
		wantSameIndex bool
	}{
		{name: "d arms delete", event: key.Event{Name: key.Name("d"), State: key.Press}, deleteArmed: false, wantAction: shortcutActionArmDelete, wantSameIndex: true},
		{name: "D arms delete", event: key.Event{Name: key.Name("D"), State: key.Press}, deleteArmed: false, wantAction: shortcutActionArmDelete, wantSameIndex: true},
		{name: "enter confirms only when armed", event: key.Event{Name: key.NameEnter, State: key.Press}, deleteArmed: true, wantAction: shortcutActionConfirmDelete, wantSameIndex: true},
		{name: "enter is no-op when not armed", event: key.Event{Name: key.NameEnter, State: key.Press}, deleteArmed: false, wantAction: shortcutActionNone, wantSameIndex: true},
		{name: "escape cancels when armed", event: key.Event{Name: key.NameEscape, State: key.Press}, deleteArmed: true, wantAction: shortcutActionCancelDelete, wantSameIndex: true},
		{name: "escape no-op when not armed", event: key.Event{Name: key.NameEscape, State: key.Press}, deleteArmed: false, wantAction: shortcutActionNone, wantSameIndex: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotIndex, gotAction := applyShortcutKey(2, 5, tc.deleteArmed, tc.event)
			if tc.wantSameIndex && gotIndex != 2 {
				t.Fatalf("applyShortcutKey() index = %d, want 2", gotIndex)
			}
			if gotAction != tc.wantAction {
				t.Fatalf("applyShortcutKey() action = %v, want %v", gotAction, tc.wantAction)
			}
		})
	}
}

func TestApplyShortcutKey_ActionKeys(t *testing.T) {
	// @req IMGWALKER-033, IMGWALKER-034, IMGWALKER-035, IMGWALKER-036, IMGWALKER-037, IMGWALKER-040
	tests := []struct {
		name       string
		keyName    key.Name
		wantAction shortcutAction
	}{
		{name: "p triggers copy path", keyName: key.Name("p"), wantAction: shortcutActionCopyPath},
		{name: "c triggers copy image", keyName: key.Name("c"), wantAction: shortcutActionCopyImage},
		{name: "o triggers open", keyName: key.Name("o"), wantAction: shortcutActionOpen},
		{name: "f triggers reveal", keyName: key.Name("f"), wantAction: shortcutActionReveal},
		{name: "r triggers reload", keyName: key.Name("r"), wantAction: shortcutActionReload},
		{name: "m triggers move", keyName: key.Name("m"), wantAction: shortcutActionMove},
		{name: "M triggers move", keyName: key.Name("M"), wantAction: shortcutActionMove},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, gotAction := applyShortcutKey(0, 3, false, key.Event{Name: tc.keyName, State: key.Press})
			if gotAction != tc.wantAction {
				t.Fatalf("applyShortcutKey() action = %v, want %v", gotAction, tc.wantAction)
			}
		})
	}
}

func TestClearDeleteArmOnSelectionChange(t *testing.T) {
	// @req IMGWALKER-041
	tests := []struct {
		name              string
		deleteArmed       bool
		previousSelection int
		nextSelection     int
		wantDeleteArmed   bool
	}{
		{name: "not armed stays false", deleteArmed: false, previousSelection: 1, nextSelection: 2, wantDeleteArmed: false},
		{name: "armed clears when selection changes", deleteArmed: true, previousSelection: 1, nextSelection: 2, wantDeleteArmed: false},
		{name: "armed remains when selection unchanged", deleteArmed: true, previousSelection: 1, nextSelection: 1, wantDeleteArmed: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := clearDeleteArmOnSelectionChange(tc.deleteArmed, tc.previousSelection, tc.nextSelection)
			if got != tc.wantDeleteArmed {
				t.Fatalf("clearDeleteArmOnSelectionChange() = %t, want %t", got, tc.wantDeleteArmed)
			}
		})
	}
}

func TestReloadImages_RefreshesListAndSelection(t *testing.T) {
	// @req IMGWALKER-040
	state := &appState{
		imageDir:      "/images",
		images:        []string{"old.png", "z.png"},
		itemClicks:    make([]widget.Clickable, 2),
		selectedIndex: 1,
		listLoadError: "x",
		previewPath:   "/images/old.png",
		previewImage:  image.NewNRGBA(image.Rect(0, 0, 1, 1)),
	}
	err := reloadImages(state, func(string) ([]os.DirEntry, error) {
		return []os.DirEntry{stubDirEntry{name: "a.png"}}, nil
	})
	if err != nil {
		t.Fatalf("reloadImages error: %v", err)
	}
	if !reflect.DeepEqual(state.images, []string{"a.png"}) {
		t.Fatalf("state.images = %#v, want %#v", state.images, []string{"a.png"})
	}
	if len(state.itemClicks) != 1 {
		t.Fatalf("itemClicks len = %d, want 1", len(state.itemClicks))
	}
	if state.selectedIndex != 0 {
		t.Fatalf("selectedIndex = %d, want 0", state.selectedIndex)
	}
	if state.listLoadError != "" {
		t.Fatalf("listLoadError = %q, want empty", state.listLoadError)
	}
	if state.previewPath != "" {
		t.Fatalf("previewPath = %q, want empty", state.previewPath)
	}
	if state.previewImage != nil {
		t.Fatalf("previewImage = %#v, want nil", state.previewImage)
	}
}

func TestDeleteConfirmText_IsExact(t *testing.T) {
	// @req IMGWALKER-038
	want := "Delete selected image? Press Enter to confirm or Esc to cancel."
	if uiText.DeleteConfirm != want {
		t.Fatalf("uiText.DeleteConfirm = %q, want %q", uiText.DeleteConfirm, want)
	}
}

func TestCopyTextToClipboard_PathUsesClipboardCommand(t *testing.T) {
	// @req IMGWALKER-033
	called := 0
	lastName := ""
	lastInput := ""
	err := copyTextToClipboard("/tmp/image.png", func(name string, args []string, stdin []byte) error {
		called++
		lastName = name
		lastInput = string(stdin)
		return nil
	})
	if err != nil {
		t.Fatalf("copyTextToClipboard error: %v", err)
	}
	if called != 1 {
		t.Fatalf("clipboard command calls = %d, want 1", called)
	}
	if lastName == "" {
		t.Fatal("clipboard command name empty")
	}
	if lastInput != "/tmp/image.png" {
		t.Fatalf("clipboard input = %q, want %q", lastInput, "/tmp/image.png")
	}
}

func TestCopyTextToClipboard_EmptyTextReturnsError(t *testing.T) {
	err := copyTextToClipboard("", func(name string, args []string, stdin []byte) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCopyImageToClipboard_UsesMimeAndPayload(t *testing.T) {
	// @req IMGWALKER-034
	called := 0
	var gotArgs []string
	var gotInput []byte
	err := copyImageToClipboard(
		"/tmp/a.png",
		func(string) ([]byte, error) { return []byte("PNGDATA"), nil },
		func(name string, args []string, stdin []byte) error {
			called++
			gotArgs = append([]string(nil), args...)
			gotInput = append([]byte(nil), stdin...)
			return nil
		},
	)
	if err != nil {
		t.Fatalf("copyImageToClipboard error: %v", err)
	}
	if called != 1 {
		t.Fatalf("clipboard command calls = %d, want 1", called)
	}
	if !reflect.DeepEqual(gotArgs, []string{"--type", "image/png"}) {
		t.Fatalf("clipboard args = %#v, want %#v", gotArgs, []string{"--type", "image/png"})
	}
	if string(gotInput) != "PNGDATA" {
		t.Fatalf("clipboard input = %q, want %q", string(gotInput), "PNGDATA")
	}
}

func TestImageMimeType_SupportedExtensions(t *testing.T) {
	tests := []struct {
		path string
		want string
		ok   bool
	}{
		{path: "a.png", want: "image/png", ok: true},
		{path: "a.JPG", want: "image/jpeg", ok: true},
		{path: "a.webp", want: "image/webp", ok: true},
		{path: "a.txt", want: "", ok: false},
	}

	for _, tc := range tests {
		got, ok := imageMimeType(tc.path)
		if got != tc.want || ok != tc.ok {
			t.Fatalf("imageMimeType(%q) = (%q,%t), want (%q,%t)", tc.path, got, ok, tc.want, tc.ok)
		}
	}
}

func TestMoveSelectedImage_MovesFileAndRemovesRow(t *testing.T) {
	// @req IMGWALKER-039
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcFile := filepath.Join(srcDir, "a.png")
	if err := os.WriteFile(srcFile, []byte("A"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	state := &appState{
		imageDir:      srcDir,
		images:        []string{"a.png"},
		itemClicks:    make([]widget.Clickable, 1),
		selectedIndex: 0,
	}

	moved, err := moveSelectedImage(
		state,
		func() (string, error) { return dstDir, nil },
		os.Rename,
	)
	if err != nil {
		t.Fatalf("moveSelectedImage error: %v", err)
	}
	if !moved {
		t.Fatal("moveSelectedImage moved = false, want true")
	}
	if _, err := os.Stat(filepath.Join(dstDir, "a.png")); err != nil {
		t.Fatalf("dest stat error: %v", err)
	}
	if len(state.images) != 0 {
		t.Fatalf("state.images len = %d, want 0", len(state.images))
	}
}

func TestMoveSelectedImage_CancelledPickerDoesNotMutate(t *testing.T) {
	// @req IMGWALKER-039
	state := &appState{
		imageDir:      "/tmp/images",
		images:        []string{"a.png", "b.png"},
		itemClicks:    make([]widget.Clickable, 2),
		selectedIndex: 1,
	}

	moved, err := moveSelectedImage(
		state,
		func() (string, error) { return "", errPickerCancelled },
		func(string, string) error { return nil },
	)
	if moved {
		t.Fatal("moveSelectedImage moved = true, want false")
	}
	if !errors.Is(err, errPickerCancelled) {
		t.Fatalf("moveSelectedImage err = %v, want errPickerCancelled", err)
	}
	if !reflect.DeepEqual(state.images, []string{"a.png", "b.png"}) {
		t.Fatalf("state.images changed: %#v", state.images)
	}
}

func TestSelectDestinationDirectoryWithRunner_MissingFirstPickerFallsBack(t *testing.T) {
	called := []string{}
	dir, err := selectDestinationDirectoryWithRunner(func(name string, args []string) (string, error) {
		called = append(called, name)
		if name == "zenity" {
			return "", exec.ErrNotFound
		}
		if name == "kdialog" {
			return "/tmp/dst\n", nil
		}
		return "", errors.New("unexpected command")
	})
	if err != nil {
		t.Fatalf("selectDestinationDirectoryWithRunner error: %v", err)
	}
	if dir != "/tmp/dst" {
		t.Fatalf("dir = %q, want %q", dir, "/tmp/dst")
	}
	if !reflect.DeepEqual(called, []string{"zenity", "kdialog"}) {
		t.Fatalf("called = %#v, want %#v", called, []string{"zenity", "kdialog"})
	}
}

func TestSelectDestinationDirectoryWithRunner_CancelDoesNotFallback(t *testing.T) {
	called := []string{}
	dir, err := selectDestinationDirectoryWithRunner(func(name string, args []string) (string, error) {
		called = append(called, name)
		if name == "zenity" {
			return "", errPickerCancelled
		}
		return "/tmp/dst\n", nil
	})
	if dir != "" {
		t.Fatalf("dir = %q, want empty", dir)
	}
	if !errors.Is(err, errPickerCancelled) {
		t.Fatalf("err = %v, want errPickerCancelled", err)
	}
	if !reflect.DeepEqual(called, []string{"zenity"}) {
		t.Fatalf("called = %#v, want %#v", called, []string{"zenity"})
	}
}

func TestSelectDestinationDirectoryWithRunner_EmptyOutputIsCancelled(t *testing.T) {
	dir, err := selectDestinationDirectoryWithRunner(func(name string, args []string) (string, error) {
		return "   \n", nil
	})
	if dir != "" {
		t.Fatalf("dir = %q, want empty", dir)
	}
	if !errors.Is(err, errPickerCancelled) {
		t.Fatalf("err = %v, want errPickerCancelled", err)
	}
}

func TestIsPickerCancelExitCode(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		exitCode  int
		cancelled bool
	}{
		{name: "zenity cancel code", command: "zenity", exitCode: 1, cancelled: true},
		{name: "kdialog cancel code", command: "kdialog", exitCode: 1, cancelled: true},
		{name: "zenity failure code", command: "zenity", exitCode: 2, cancelled: false},
		{name: "unknown command", command: "other", exitCode: 1, cancelled: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isPickerCancelExitCode(tc.command, tc.exitCode)
			if got != tc.cancelled {
				t.Fatalf("isPickerCancelExitCode(%q, %d) = %t, want %t", tc.command, tc.exitCode, got, tc.cancelled)
			}
		})
	}
}

func TestSelectedPathOrNoSelection(t *testing.T) {
	t.Run("returns path when selected", func(t *testing.T) {
		state := &appState{imageDir: "/tmp", images: []string{"a.png"}, selectedIndex: 0}
		path, ok := selectedPathOrNoSelection(state)
		if !ok {
			t.Fatal("ok = false, want true")
		}
		if path != "/tmp/a.png" {
			t.Fatalf("path = %q, want %q", path, "/tmp/a.png")
		}
		if state.actionMessage != "" {
			t.Fatalf("actionMessage = %q, want empty", state.actionMessage)
		}
	})

	t.Run("sets no selection message", func(t *testing.T) {
		state := &appState{imageDir: "/tmp", images: nil, selectedIndex: 0}
		path, ok := selectedPathOrNoSelection(state)
		if ok {
			t.Fatal("ok = true, want false")
		}
		if path != "" {
			t.Fatalf("path = %q, want empty", path)
		}
		if state.actionMessage != uiText.NoSelection {
			t.Fatalf("actionMessage = %q, want %q", state.actionMessage, uiText.NoSelection)
		}
	})
}

func TestOpenFileWithDefaultApp_UsesRunner(t *testing.T) {
	// @req IMGWALKER-035
	called := 0
	var gotName string
	var gotArgs []string
	err := openFileWithDefaultApp("/tmp/a.png", func(name string, args []string) error {
		called++
		gotName = name
		gotArgs = append([]string(nil), args...)
		return nil
	})
	if err != nil {
		t.Fatalf("openFileWithDefaultApp error: %v", err)
	}
	if called != 1 {
		t.Fatalf("runner calls = %d, want 1", called)
	}
	if gotName == "" {
		t.Fatal("runner command name empty")
	}
	if !reflect.DeepEqual(gotArgs, []string{"/tmp/a.png"}) {
		t.Fatalf("runner args = %#v, want %#v", gotArgs, []string{"/tmp/a.png"})
	}
}

func TestRevealFileInManager_UsesRunner(t *testing.T) {
	// @req IMGWALKER-036
	called := 0
	var gotName string
	var gotArgs []string
	err := revealFileInManager("/tmp/a.png", func(name string, args []string) error {
		called++
		gotName = name
		gotArgs = append([]string(nil), args...)
		return nil
	})
	if err != nil {
		t.Fatalf("revealFileInManager error: %v", err)
	}
	if called != 1 {
		t.Fatalf("runner calls = %d, want 1", called)
	}
	if gotName == "" {
		t.Fatal("runner command name empty")
	}
	if len(gotArgs) == 0 {
		t.Fatal("runner args empty")
	}
}

func TestDeleteSelectedImage_RemovesSelectedFileAndRow(t *testing.T) {
	tmp := t.TempDir()
	first := filepath.Join(tmp, "a.png")
	second := filepath.Join(tmp, "b.png")
	if err := os.WriteFile(first, []byte("a"), 0o644); err != nil {
		t.Fatalf("write first: %v", err)
	}
	if err := os.WriteFile(second, []byte("b"), 0o644); err != nil {
		t.Fatalf("write second: %v", err)
	}

	state := &appState{
		imageDir:      tmp,
		images:        []string{"a.png", "b.png"},
		itemClicks:    make([]widget.Clickable, 2),
		selectedIndex: 0,
	}

	deleted, err := deleteSelectedImage(state, os.Remove)
	if err != nil {
		t.Fatalf("deleteSelectedImage() error = %v, want nil", err)
	}
	if !deleted {
		t.Fatal("deleteSelectedImage() = false, want true")
	}
	if _, err := os.Stat(first); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("deleted file stat err = %v, want not exist", err)
	}
	if got := state.images; !reflect.DeepEqual(got, []string{"b.png"}) {
		t.Fatalf("state.images = %#v, want %#v", got, []string{"b.png"})
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
