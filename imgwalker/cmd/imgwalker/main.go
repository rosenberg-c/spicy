package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"image/color"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

const (
	windowTitle        = "ImgWalker"
	windowWidth        = 640
	windowHeight       = 360
	leftPaneRatio      = 0.35
	paneDividerWidthDp = 1
)

var uiText = struct {
	NoImagesFound  string
	ListLoadError  string
	DeleteConfirm  string
	DeleteDone     string
	DeleteError    string
	PathCopied     string
	PathCopyError  string
	ImageCopied    string
	ImageCopyError string
	Reloaded       string
	ReloadError    string
	MoveDone       string
	MoveError      string
	MoveCancelled  string
	OpenDone       string
	OpenError      string
	RevealDone     string
	RevealError    string
	NoSelection    string
}{
	NoImagesFound:  "No images found",
	ListLoadError:  "Unable to load images from configured directory.",
	DeleteConfirm:  "Delete selected image? Press Enter to confirm or Esc to cancel.",
	DeleteDone:     "Deleted selected image.",
	DeleteError:    "Unable to delete selected image.",
	PathCopied:     "Copied selected image path.",
	PathCopyError:  "Unable to copy image path to clipboard.",
	ImageCopied:    "Copied selected image to clipboard.",
	ImageCopyError: "Unable to copy selected image to clipboard.",
	Reloaded:       "Reloaded image list.",
	ReloadError:    "Unable to reload image list.",
	MoveDone:       "Moved selected image.",
	MoveError:      "Unable to move selected image.",
	MoveCancelled:  "Move cancelled.",
	OpenDone:       "Opened selected image.",
	OpenError:      "Unable to open selected image.",
	RevealDone:     "Revealed selected image.",
	RevealError:    "Unable to reveal selected image.",
	NoSelection:    "No image selected.",
}

var listNavigationFilters = []event.Filter{
	key.Filter{Name: "j"},
	key.Filter{Name: "J"},
	key.Filter{Name: "k"},
	key.Filter{Name: "K"},
	key.Filter{Name: "p"},
	key.Filter{Name: "P"},
	key.Filter{Name: "c"},
	key.Filter{Name: "C"},
	key.Filter{Name: "f"},
	key.Filter{Name: "F"},
	key.Filter{Name: "o"},
	key.Filter{Name: "O"},
	key.Filter{Name: "r"},
	key.Filter{Name: "R"},
	key.Filter{Name: "d"},
	key.Filter{Name: "D"},
	key.Filter{Name: "m"},
	key.Filter{Name: "M"},
	key.Filter{Name: key.NameEscape},
	key.Filter{Name: key.NameReturn},
	key.Filter{Name: key.NameEnter},
	key.Filter{Name: key.NameDownArrow},
	key.Filter{Name: key.NameUpArrow},
}

type windowConfig struct {
	title    string
	width    int
	height   int
	imageDir string
}

type startupFileConfig struct {
	ImageDir string `json:"imageDir"`
}

type ConfigLoadErrorCode string

const (
	ConfigLoadErrorNotFound        ConfigLoadErrorCode = "not_found"
	ConfigLoadErrorInvalidConfig   ConfigLoadErrorCode = "invalid_config"
	ConfigLoadErrorInvalidImageDir ConfigLoadErrorCode = "invalid_image_dir"
	ConfigLoadErrorIOError         ConfigLoadErrorCode = "io_error"
)

type ConfigLoadError struct {
	Code ConfigLoadErrorCode
	Err  error
}

type appState struct {
	imageDir      string
	images        []string
	listLoadError string
	itemClicks    []widget.Clickable
	list          widget.List
	selectedIndex int
	deleteArmed   bool
	actionMessage string
	previewPath   string
	previewImage  image.Image
}

type shortcutAction int

const (
	shortcutActionNone shortcutAction = iota
	shortcutActionCopyPath
	shortcutActionCopyImage
	shortcutActionOpen
	shortcutActionReveal
	shortcutActionReload
	shortcutActionMove
	shortcutActionArmDelete
	shortcutActionConfirmDelete
	shortcutActionCancelDelete
)

var errPickerCancelled = errors.New("picker cancelled")

func (e *ConfigLoadError) Error() string {
	return fmt.Sprintf("config load error (%s): %v", e.Code, e.Err)
}

func (e *ConfigLoadError) Unwrap() error {
	return e.Err
}

var (
	colorBg         = color.NRGBA{R: 16, G: 21, B: 28, A: 255}
	colorFg         = color.NRGBA{R: 222, G: 228, B: 236, A: 255}
	colorContrastBg = color.NRGBA{R: 41, G: 109, B: 196, A: 255}
	colorContrastFg = color.NRGBA{R: 248, G: 251, B: 255, A: 255}
)

func main() {
	go func() {
		cfg := startupConfig()
		window := newWindow(cfg)

		if err := run(window, cfg); err != nil {
			log.Fatal(err)
		}
	}()

	app.Main()
}

func startupConfig() windowConfig {
	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		log.Printf("imgwalker: failed to resolve user home: %v", homeErr)
	}

	workingDir, workingDirErr := os.Getwd()
	if workingDirErr != nil {
		log.Printf("imgwalker: failed to resolve working directory: %v", workingDirErr)
	}

	return buildStartupWindowConfig(func() (startupFileConfig, error) {
		return resolveStartupFileConfig(homeDir, workingDir, os.ReadFile, os.Stat)
	})
}

func buildStartupWindowConfig(resolve func() (startupFileConfig, error)) windowConfig {
	appCfg, err := resolve()
	if err != nil {
		log.Printf("imgwalker: %v", err)
	}

	return windowConfig{
		title:    windowTitle,
		width:    windowWidth,
		height:   windowHeight,
		imageDir: appCfg.ImageDir,
	}
}

func resolveStartupFileConfig(
	homeDir string,
	workingDir string,
	readFile func(string) ([]byte, error),
	stat func(string) (os.FileInfo, error),
) (startupFileConfig, error) {
	if homeDir == "" {
		return startupFileConfig{}, &ConfigLoadError{Code: ConfigLoadErrorInvalidConfig, Err: errors.New("home directory is required to locate config file")}
	}

	appCfg, err := loadStartupFileConfig(homeDir, readFile)
	if err != nil {
		return startupFileConfig{}, err
	}

	return validateStartupFileConfig(appCfg, homeDir, workingDir, stat)
}

func configFilePath(homeDir string) string {
	return filepath.Join(homeDir, ".config", "spicy", "imgwalker.json")
}

func loadStartupFileConfig(homeDir string, readFile func(string) ([]byte, error)) (startupFileConfig, error) {
	path := configFilePath(homeDir)
	raw, err := readFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return startupFileConfig{}, &ConfigLoadError{Code: ConfigLoadErrorNotFound, Err: err}
		}
		return startupFileConfig{}, &ConfigLoadError{Code: ConfigLoadErrorIOError, Err: err}
	}

	var cfg startupFileConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return startupFileConfig{}, &ConfigLoadError{Code: ConfigLoadErrorInvalidConfig, Err: err}
	}

	return cfg, nil
}

func validateStartupFileConfig(cfg startupFileConfig, homeDir string, workingDir string, stat func(string) (os.FileInfo, error)) (startupFileConfig, error) {
	if cfg.ImageDir == "" {
		return startupFileConfig{}, &ConfigLoadError{Code: ConfigLoadErrorInvalidConfig, Err: errors.New("imageDir must not be empty")}
	}

	imageDir, err := normalizeImageDir(cfg.ImageDir, homeDir, workingDir)
	if err != nil {
		return startupFileConfig{}, err
	}
	cfg.ImageDir = imageDir

	info, err := stat(cfg.ImageDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return startupFileConfig{}, &ConfigLoadError{Code: ConfigLoadErrorInvalidImageDir, Err: err}
		}
		return startupFileConfig{}, &ConfigLoadError{Code: ConfigLoadErrorInvalidImageDir, Err: err}
	}

	if !info.IsDir() {
		return startupFileConfig{}, &ConfigLoadError{Code: ConfigLoadErrorInvalidImageDir, Err: errors.New("imageDir is not a directory")}
	}

	return cfg, nil
}

func normalizeImageDir(imageDir string, homeDir string, workingDir string) (string, error) {
	if strings.HasPrefix(imageDir, "~/") {
		if homeDir == "" {
			return "", &ConfigLoadError{Code: ConfigLoadErrorInvalidConfig, Err: errors.New("home directory is required to expand ~/ paths")}
		}
		return filepath.Join(homeDir, imageDir[2:]), nil
	}
	if !filepath.IsAbs(imageDir) {
		if workingDir == "" {
			return "", &ConfigLoadError{Code: ConfigLoadErrorInvalidConfig, Err: errors.New("working directory is required to resolve relative paths")}
		}
		return filepath.Join(workingDir, imageDir), nil
	}
	return imageDir, nil
}

func newWindow(cfg windowConfig) *app.Window {
	window := new(app.Window)
	window.Option(
		app.Title(cfg.title),
		app.Size(unit.Dp(cfg.width), unit.Dp(cfg.height)),
	)
	return window
}

func run(window *app.Window, cfg windowConfig) error {
	var ops op.Ops
	theme := newTheme()
	state := newAppState(cfg.imageDir, os.ReadDir)
	setInitialFocus := false

	for {
		event := window.Event()
		switch event := event.(type) {
		case app.DestroyEvent:
			return event.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, event)
			if !setInitialFocus {
				gtx.Execute(key.FocusCmd{Tag: &state.list})
				setInitialFocus = true
			}
			handleListKeyInput(gtx, &state)
			paint.FillShape(
				gtx.Ops,
				theme.Palette.Bg,
				clip.Rect{Max: gtx.Constraints.Max}.Op(),
			)
			renderMainLayout(gtx, theme, &state)
			if state.deleteArmed {
				renderDeleteConfirmDialog(gtx, theme)
			}
			event.Frame(gtx.Ops)
		}
	}
}

func newAppState(imageDir string, readDir func(string) ([]os.DirEntry, error)) appState {
	images, err := listImageFiles(imageDir, readDir)
	listLoadError := ""
	if err != nil {
		log.Printf("imgwalker: failed to load image list: %v", err)
		listLoadError = listLoadErrorMessage(err)
	}

	return appState{
		imageDir:      imageDir,
		images:        images,
		listLoadError: listLoadError,
		itemClicks:    make([]widget.Clickable, len(images)),
		list: widget.List{
			List: layout.List{Axis: layout.Vertical},
		},
	}
}

func listLoadErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	return uiText.ListLoadError
}

func handleListKeyInput(gtx layout.Context, state *appState) {
	for {
		ev, ok := gtx.Source.Event(listNavigationFilters...)
		if !ok {
			break
		}

		ke, isKey := ev.(key.Event)
		if !isKey || ke.State != key.Press {
			continue
		}

		prevIndex := state.selectedIndex
		next, action := applyShortcutKey(state.selectedIndex, len(state.images), state.deleteArmed, ke)
		state.selectedIndex = next
		state.deleteArmed = clearDeleteArmOnSelectionChange(state.deleteArmed, prevIndex, state.selectedIndex)

		executeShortcutAction(state, action)

		if state.selectedIndex != prevIndex {
			state.list.ScrollTo(state.selectedIndex)
			state.list.Position.OffsetLast = 0
		}
	}
}

// Shortcut action handlers keep key event flow separate from side effects.
func executeShortcutAction(state *appState, action shortcutAction) {
	switch action {
	case shortcutActionArmDelete:
		state.deleteArmed = true
	case shortcutActionCancelDelete:
		state.deleteArmed = false
	case shortcutActionConfirmDelete:
		handleConfirmDeleteAction(state)
	case shortcutActionCopyPath:
		handleCopyPathAction(state)
	case shortcutActionCopyImage:
		handleCopyImageAction(state)
	case shortcutActionOpen:
		handleOpenAction(state)
	case shortcutActionReveal:
		handleRevealAction(state)
	case shortcutActionReload:
		handleReloadAction(state)
	case shortcutActionMove:
		handleMoveAction(state)
	}
}

func handleConfirmDeleteAction(state *appState) {
	state.deleteArmed = false
	deleted, err := deleteSelectedImage(state, os.Remove)
	if err != nil {
		log.Printf("imgwalker: delete selected image failed: %v", err)
		state.actionMessage = uiText.DeleteError
		return
	}
	if deleted {
		state.actionMessage = uiText.DeleteDone
		state.selectedIndex = clampIndex(state.selectedIndex, len(state.images))
		return
	}
	state.actionMessage = uiText.NoSelection
}

func handleCopyPathAction(state *appState) {
	path, ok := selectedPathOrNoSelection(state)
	if !ok {
		return
	}
	if err := copyTextToClipboard(path, runClipboardCommandBytes); err != nil {
		log.Printf("imgwalker: copy path failed: %v", err)
		state.actionMessage = uiText.PathCopyError
		return
	}
	log.Printf("imgwalker: copied path to clipboard: %q", path)
	state.actionMessage = uiText.PathCopied
}

func handleCopyImageAction(state *appState) {
	path, ok := selectedPathOrNoSelection(state)
	if !ok {
		return
	}
	if err := copyImageToClipboard(path, os.ReadFile, runClipboardCommandBytes); err != nil {
		log.Printf("imgwalker: copy image failed: %v", err)
		state.actionMessage = uiText.ImageCopyError
		return
	}
	log.Printf("imgwalker: copied image to clipboard: %q", path)
	state.actionMessage = uiText.ImageCopied
}

func handleOpenAction(state *appState) {
	path, ok := selectedPathOrNoSelection(state)
	if !ok {
		return
	}
	if err := openFileWithDefaultApp(path, runCommandNoInput); err != nil {
		log.Printf("imgwalker: open failed: %v", err)
		state.actionMessage = uiText.OpenError
		return
	}
	state.actionMessage = uiText.OpenDone
}

func handleRevealAction(state *appState) {
	path, ok := selectedPathOrNoSelection(state)
	if !ok {
		return
	}
	if err := revealFileInManager(path, runCommandNoInput); err != nil {
		log.Printf("imgwalker: reveal failed: %v", err)
		state.actionMessage = uiText.RevealError
		return
	}
	state.actionMessage = uiText.RevealDone
}

func handleReloadAction(state *appState) {
	if err := reloadImages(state, os.ReadDir); err != nil {
		log.Printf("imgwalker: reload failed: %v", err)
		state.actionMessage = uiText.ReloadError
		return
	}
	state.actionMessage = uiText.Reloaded
}

func handleMoveAction(state *appState) {
	moved, err := moveSelectedImage(state, selectDestinationDirectory, os.Rename)
	if err == nil && moved {
		state.actionMessage = uiText.MoveDone
		state.selectedIndex = clampIndex(state.selectedIndex, len(state.images))
		log.Printf("imgwalker: moved selected image")
		return
	}
	if errors.Is(err, errPickerCancelled) {
		state.actionMessage = uiText.MoveCancelled
		return
	}
	if err != nil {
		log.Printf("imgwalker: move selected image failed: %v", err)
		state.actionMessage = uiText.MoveError
		return
	}
	state.actionMessage = uiText.NoSelection
}

func applyShortcutKey(current int, total int, deleteArmed bool, ev key.Event) (int, shortcutAction) {
	next, changed := nextSelectionIndex(current, total, ev.Name)
	if changed {
		return next, shortcutActionNone
	}

	switch ev.Name {
	case "p", "P":
		return clampIndex(current, total), shortcutActionCopyPath
	case "c", "C":
		return clampIndex(current, total), shortcutActionCopyImage
	case "o", "O":
		return clampIndex(current, total), shortcutActionOpen
	case "f", "F":
		return clampIndex(current, total), shortcutActionReveal
	case "r", "R":
		return clampIndex(current, total), shortcutActionReload
	case "m", "M":
		return clampIndex(current, total), shortcutActionMove
	case "d", "D":
		return clampIndex(current, total), shortcutActionArmDelete
	case key.NameReturn, key.NameEnter:
		if deleteArmed {
			return clampIndex(current, total), shortcutActionConfirmDelete
		}
		return clampIndex(current, total), shortcutActionNone
	case key.NameEscape:
		if deleteArmed {
			return clampIndex(current, total), shortcutActionCancelDelete
		}
		return clampIndex(current, total), shortcutActionNone
	default:
		return clampIndex(current, total), shortcutActionNone
	}
}

func clearDeleteArmOnSelectionChange(deleteArmed bool, previousSelection int, nextSelection int) bool {
	return deleteArmed && previousSelection == nextSelection
}

func deleteSelectedImage(state *appState, remove func(string) error) (bool, error) {
	path := selectedImagePath(state.imageDir, state.images, state.selectedIndex)
	if path == "" {
		return false, nil
	}
	if err := remove(path); err != nil {
		return false, err
	}

	return removeSelectedRow(state), nil
}

func removeSelectedRow(state *appState) bool {
	idx := state.selectedIndex
	if idx < 0 || idx >= len(state.images) {
		return false
	}
	state.images = append(state.images[:idx], state.images[idx+1:]...)
	state.itemClicks = append(state.itemClicks[:idx], state.itemClicks[idx+1:]...)
	return true
}

func reloadImages(state *appState, readDir func(string) ([]os.DirEntry, error)) error {
	images, err := listImageFiles(state.imageDir, readDir)
	state.images = images
	state.itemClicks = make([]widget.Clickable, len(images))
	state.previewPath = ""
	state.previewImage = nil
	if err != nil {
		state.listLoadError = listLoadErrorMessage(err)
		state.selectedIndex = 0
		return err
	}
	state.listLoadError = ""
	state.selectedIndex = clampIndex(state.selectedIndex, len(images))
	return nil
}

func selectedPathOrNoSelection(state *appState) (string, bool) {
	path := selectedImagePath(state.imageDir, state.images, state.selectedIndex)
	if path == "" {
		state.actionMessage = uiText.NoSelection
		return "", false
	}
	return path, true
}

func copyTextToClipboard(text string, run func(name string, args []string, stdin []byte) error) error {
	if strings.TrimSpace(text) == "" {
		return errors.New("empty clipboard text")
	}

	type clipboardCommand struct {
		name string
		args []string
	}

	commands := []clipboardCommand{
		{name: "pbcopy"},
		{name: "wl-copy"},
		{name: "xclip", args: []string{"-selection", "clipboard"}},
	}

	var errs []string
	for _, cmd := range commands {
		if err := run(cmd.name, cmd.args, []byte(text)); err == nil {
			return nil
		} else {
			errs = append(errs, fmt.Sprintf("%s: %v", cmd.name, err))
		}
	}

	return fmt.Errorf("no clipboard command succeeded: %s", strings.Join(errs, "; "))
}

func runClipboardCommandBytes(name string, args []string, stdin []byte) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = bytes.NewReader(stdin)
	return cmd.Run()
}

func copyImageToClipboard(path string, readFile func(string) ([]byte, error), run func(name string, args []string, stdin []byte) error) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("empty image path")
	}

	mimeType, ok := imageMimeType(path)
	if !ok {
		return fmt.Errorf("unsupported image extension: %s", filepath.Ext(path))
	}

	data, err := readFile(path)
	if err != nil {
		return err
	}

	type clipboardCommand struct {
		name string
		args []string
	}

	commands := []clipboardCommand{
		{name: "wl-copy", args: []string{"--type", mimeType}},
		{name: "xclip", args: []string{"-selection", "clipboard", "-t", mimeType}},
	}

	var errs []string
	for _, cmd := range commands {
		if err := run(cmd.name, cmd.args, data); err == nil {
			return nil
		} else {
			errs = append(errs, fmt.Sprintf("%s: %v", cmd.name, err))
		}
	}

	return fmt.Errorf("no image clipboard command succeeded: %s", strings.Join(errs, "; "))
}

func imageMimeType(path string) (string, bool) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		return "image/png", true
	case ".jpg", ".jpeg":
		return "image/jpeg", true
	case ".gif":
		return "image/gif", true
	case ".webp":
		return "image/webp", true
	case ".bmp":
		return "image/bmp", true
	default:
		return "", false
	}
}

func moveSelectedImage(state *appState, pickDir func() (string, error), rename func(string, string) error) (bool, error) {
	src := selectedImagePath(state.imageDir, state.images, state.selectedIndex)
	if src == "" {
		return false, nil
	}

	dstDir, err := pickDir()
	if err != nil {
		return false, err
	}

	dst := filepath.Join(dstDir, filepath.Base(src))
	if err := rename(src, dst); err != nil {
		return false, err
	}

	idx := state.selectedIndex
	if idx < 0 || idx >= len(state.images) {
		return false, nil
	}
	return removeSelectedRow(state), nil
}

func selectDestinationDirectory() (string, error) {
	return selectDestinationDirectoryWithRunner(runOutputCommand)
}

func selectDestinationDirectoryWithRunner(run func(name string, args []string) (string, error)) (string, error) {
	type pickerCommand struct {
		name string
		args []string
	}

	commands := []pickerCommand{
		{name: "zenity", args: []string{"--file-selection", "--directory", "--title=Select destination folder"}},
		{name: "kdialog", args: []string{"--getexistingdirectory", os.Getenv("HOME")}},
	}

	var errs []string
	for _, cmd := range commands {
		out, err := run(cmd.name, cmd.args)
		if err == nil {
			dir := strings.TrimSpace(out)
			if dir == "" {
				return "", errPickerCancelled
			}
			return dir, nil
		}
		if errors.Is(err, errPickerCancelled) {
			return "", errPickerCancelled
		}
		if errors.Is(err, exec.ErrNotFound) {
			errs = append(errs, fmt.Sprintf("%s: %v", cmd.name, err))
			continue
		}
		errs = append(errs, fmt.Sprintf("%s: %v", cmd.name, err))
		return "", fmt.Errorf("%s failed: %w", cmd.name, err)
	}

	return "", fmt.Errorf("no file picker command succeeded: %s", strings.Join(errs, "; "))
}

func runOutputCommand(name string, args []string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr := new(exec.ExitError); errors.As(err, &exitErr) && isPickerCancelExitCode(name, exitErr.ExitCode()) {
			return "", errPickerCancelled
		}
		return "", err
	}
	return string(out), nil
}

func isPickerCancelExitCode(commandName string, exitCode int) bool {
	switch commandName {
	case "zenity", "kdialog":
		return exitCode == 1
	default:
		return false
	}
}

func openFileWithDefaultApp(path string, run func(name string, args []string) error) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("empty file path")
	}

	commands := []struct {
		name string
		args []string
	}{
		{name: "xdg-open", args: []string{path}},
		{name: "open", args: []string{path}},
	}

	var errs []string
	for _, cmd := range commands {
		if err := run(cmd.name, cmd.args); err == nil {
			return nil
		} else {
			errs = append(errs, fmt.Sprintf("%s: %v", cmd.name, err))
		}
	}

	return fmt.Errorf("no open command succeeded: %s", strings.Join(errs, "; "))
}

func revealFileInManager(path string, run func(name string, args []string) error) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("empty file path")
	}

	dir := filepath.Dir(path)
	commands := []struct {
		name string
		args []string
	}{
		{name: "nautilus", args: []string{"--select", path}},
		{name: "nemo", args: []string{"--no-desktop", path}},
		{name: "dolphin", args: []string{"--select", path}},
		{name: "thunar", args: []string{path}},
		{name: "xdg-open", args: []string{dir}},
		{name: "open", args: []string{"-R", path}},
	}

	var errs []string
	for _, cmd := range commands {
		if err := run(cmd.name, cmd.args); err == nil {
			return nil
		} else {
			errs = append(errs, fmt.Sprintf("%s: %v", cmd.name, err))
		}
	}

	return fmt.Errorf("no reveal command succeeded: %s", strings.Join(errs, "; "))
}

func runCommandNoInput(name string, args []string) error {
	cmd := exec.Command(name, args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		_ = cmd.Wait()
	}()
	return nil
}

func renderMainLayout(gtx layout.Context, theme *material.Theme, state *appState) layout.Dimensions {
	leftWidth, rightWidth := splitPaneWidths(gtx.Constraints.Max.X)
	dividerWidth := gtx.Dp(unit.Dp(paneDividerWidthDp))
	if dividerWidth < 1 {
		dividerWidth = 1
	}
	if rightWidth >= dividerWidth {
		rightWidth -= dividerWidth
	}

	return layout.Flex{Axis: layout.Horizontal}.Layout(
		gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = leftWidth
			gtx.Constraints.Max.X = leftWidth
			return renderImageListPane(gtx, theme, state)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = dividerWidth
			gtx.Constraints.Max.X = dividerWidth
			return renderPaneDivider(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = rightWidth
			gtx.Constraints.Max.X = rightWidth
			return renderPreviewPane(gtx, theme, state)
		}),
	)
}

func splitPaneWidths(totalWidth int) (int, int) {
	leftWidth := int(float32(totalWidth) * leftPaneRatio)
	if leftWidth < 0 {
		leftWidth = 0
	}
	if leftWidth > totalWidth {
		leftWidth = totalWidth
	}

	rightWidth := totalWidth - leftWidth
	if rightWidth < 0 {
		rightWidth = 0
	}

	return leftWidth, rightWidth
}

func renderImageListPane(gtx layout.Context, theme *material.Theme, state *appState) layout.Dimensions {
	paint.FillShape(gtx.Ops, listPaneBackgroundColor(theme), clip.Rect{Max: gtx.Constraints.Max}.Op())

	inset := layout.UniformInset(unit.Dp(12))
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(
			gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if state.listLoadError == "" {
					return layout.Dimensions{}
				}
				errorLabel := material.Body2(theme, state.listLoadError)
				errorLabel.Color = color.NRGBA{R: 230, G: 146, B: 146, A: 255}
				return layout.Inset{Bottom: unit.Dp(10)}.Layout(gtx, errorLabel.Layout)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if state.actionMessage == "" {
					return layout.Dimensions{}
				}
				msgLabel := material.Caption(theme, state.actionMessage)
				msgLabel.Color = color.NRGBA{R: 189, G: 203, B: 225, A: 255}
				return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, msgLabel.Layout)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				if len(state.images) == 0 {
					label := material.Body1(theme, uiText.NoImagesFound)
					label.Color = theme.Palette.ContrastFg
					return label.Layout(gtx)
				}

				previousSelection := state.selectedIndex
				state.selectedIndex = selectedIndexAfterListClicks(gtx, state.selectedIndex, state.itemClicks, len(state.images))
				state.deleteArmed = clearDeleteArmOnSelectionChange(state.deleteArmed, previousSelection, state.selectedIndex)

				return material.List(theme, &state.list).Layout(gtx, len(state.images), func(gtx layout.Context, index int) layout.Dimensions {
					name := state.images[index]

					return state.itemClicks[index].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						bg, fg := listRowColors(index == state.selectedIndex)
						radius := gtx.Dp(unit.Dp(6))
						paint.FillShape(gtx.Ops, bg, clip.RRect{Rect: image.Rectangle{Max: gtx.Constraints.Max}, NW: radius, NE: radius, SW: radius, SE: radius}.Op(gtx.Ops))

						row := material.Body1(theme, name)
						row.Color = fg
						return layout.Inset{Left: unit.Dp(8), Right: unit.Dp(8), Top: unit.Dp(8), Bottom: unit.Dp(10)}.Layout(gtx, row.Layout)
					})
				})
			}),
		)
	})
}

func selectedIndexAfterListClicks(gtx layout.Context, current int, clicks []widget.Clickable, total int) int {
	clickedIndexes := make([]int, 0, len(clicks))
	for i := range clicks {
		if i >= total {
			break
		}
		if clicks[i].Clicked(gtx) {
			clickedIndexes = append(clickedIndexes, i)
		}
	}

	return selectedIndexAfterClickedIndexes(current, total, clickedIndexes)
}

func selectedIndexAfterClickedIndexes(current int, total int, clickedIndexes []int) int {
	next := clampIndex(current, total)
	for _, idx := range clickedIndexes {
		if idx < 0 || idx >= total {
			continue
		}
		next = idx
	}
	return next
}

func renderPaneDivider(gtx layout.Context) layout.Dimensions {
	paint.FillShape(gtx.Ops, paneDividerColor(), clip.Rect{Max: gtx.Constraints.Max}.Op())
	return layout.Dimensions{Size: gtx.Constraints.Max}
}

func renderDeleteConfirmDialog(gtx layout.Context, theme *material.Theme) {
	max := gtx.Constraints.Max
	paint.FillShape(gtx.Ops, color.NRGBA{A: 140}, clip.Rect{Max: max}.Op())

	w := max.X - gtx.Dp(unit.Dp(64))
	if w > gtx.Dp(unit.Dp(420)) {
		w = gtx.Dp(unit.Dp(420))
	}
	if w < gtx.Dp(unit.Dp(220)) {
		w = gtx.Dp(unit.Dp(220))
	}

	box := layout.Stack{Alignment: layout.Center}
	box.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: max}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = w
			gtx.Constraints.Max.X = w

			radius := gtx.Dp(unit.Dp(10))
			return widget.Border{Color: color.NRGBA{R: 72, G: 84, B: 101, A: 255}, CornerRadius: unit.Dp(10), Width: unit.Dp(1)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				paint.FillShape(gtx.Ops, color.NRGBA{R: 21, G: 28, B: 37, A: 250}, clip.RRect{Rect: image.Rectangle{Max: gtx.Constraints.Max}, NW: radius, NE: radius, SW: radius, SE: radius}.Op(gtx.Ops))
				label := material.Body1(theme, uiText.DeleteConfirm)
				label.Color = color.NRGBA{R: 235, G: 240, B: 247, A: 255}
				return layout.UniformInset(unit.Dp(14)).Layout(gtx, label.Layout)
			})
		}),
	)
}

func listPaneBackgroundColor(theme *material.Theme) color.NRGBA {
	return theme.Palette.Bg
}

func paneDividerColor() color.NRGBA {
	return color.NRGBA{R: 64, G: 74, B: 89, A: 255}
}

func listRowColors(selected bool) (color.NRGBA, color.NRGBA) {
	bg := color.NRGBA{R: 31, G: 41, B: 56, A: 255}
	fg := color.NRGBA{R: 224, G: 232, B: 245, A: 255}

	if selected {
		return color.NRGBA{R: 44, G: 95, B: 181, A: 255}, color.NRGBA{R: 248, G: 251, B: 255, A: 255}
	}

	return bg, fg
}

func nextSelectionIndex(current int, total int, keyName key.Name) (int, bool) {
	if total <= 0 {
		return 0, false
	}

	step, ok := selectionStep(keyName)
	if !ok {
		return clampIndex(current, total), false
	}

	return clampIndex(current+step, total), true
}

func selectionStep(name key.Name) (int, bool) {
	switch name {
	case "j", "J", key.NameDownArrow:
		return 1, true
	case "k", "K", key.NameUpArrow:
		return -1, true
	default:
		return 0, false
	}
}

func clampIndex(index int, total int) int {
	if total <= 0 {
		return 0
	}
	if index < 0 {
		return 0
	}
	last := total - 1
	if index > last {
		return last
	}
	return index
}

func renderPreviewPane(gtx layout.Context, theme *material.Theme, state *appState) layout.Dimensions {
	paint.FillShape(gtx.Ops, theme.Palette.Bg, clip.Rect{Max: gtx.Constraints.Max}.Op())

	return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		previewPath := previewSourcePath(state.imageDir, state.images, state.selectedIndex)
		if previewPath == "" {
			return layout.Dimensions{}
		}

		if previewPath != state.previewPath {
			state.previewPath = previewPath
			var err error
			state.previewImage, err = loadPreviewImage(previewPath, os.ReadFile)
			if err != nil {
				log.Printf("imgwalker: preview image unavailable %q: %v", previewPath, err)
			}
		}

		if state.previewImage == nil {
			return layout.Dimensions{}
		}

		img := widget.Image{Src: paint.NewImageOp(state.previewImage), Fit: widget.Contain}
		return img.Layout(gtx)
	})
}

func loadPreviewImage(path string, readFile func(string) ([]byte, error)) (image.Image, error) {
	data, err := readFile(path)
	if err != nil {
		return nil, fmt.Errorf("read preview image: %w", err)
	}

	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode preview image: %w", err)
	}

	return decoded, nil
}

func previewSourcePath(imageDir string, images []string, selectedIndex int) string {
	if len(images) == 0 {
		return ""
	}

	text := selectedImagePath(imageDir, images, selectedIndex)
	if text == "" {
		return selectedImagePath(imageDir, images, 0)
	}
	return text
}

func selectedImagePath(imageDir string, images []string, selectedIndex int) string {
	if imageDir == "" {
		return ""
	}
	if selectedIndex < 0 || selectedIndex >= len(images) {
		return ""
	}
	return filepath.Join(imageDir, images[selectedIndex])
}

func listImageFiles(imageDir string, readDir func(string) ([]os.DirEntry, error)) ([]string, error) {
	if imageDir == "" {
		return nil, nil
	}

	entries, err := readDir(imageDir)
	if err != nil {
		return nil, err
	}

	images := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if isSupportedImageFile(name) {
			images = append(images, name)
		}
	}

	sort.Strings(images)
	return images, nil
}

func isSupportedImageFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp":
		return true
	default:
		return false
	}
}

func newTheme() *material.Theme {
	theme := material.NewTheme()
	theme.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	theme.Palette.Bg = colorBg
	theme.Palette.Fg = colorFg
	theme.Palette.ContrastBg = colorContrastBg
	theme.Palette.ContrastFg = colorContrastFg
	return theme
}
