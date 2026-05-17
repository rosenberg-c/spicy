package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
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
	NoImagesFound string
	ListLoadError string
}{
	NoImagesFound: "No images found",
	ListLoadError: "Unable to load images from configured directory.",
}

var listNavigationFilters = []event.Filter{
	key.Filter{Name: "j"},
	key.Filter{Name: "J"},
	key.Filter{Name: "k"},
	key.Filter{Name: "K"},
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
}

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

		nextIndex, changed := nextSelectionIndex(state.selectedIndex, len(state.images), ke.Name)
		if !changed {
			continue
		}

		state.selectedIndex = nextIndex
		state.list.ScrollTo(state.selectedIndex)
		state.list.Position.OffsetLast = 0
	}
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
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				if len(state.images) == 0 {
					label := material.Body1(theme, uiText.NoImagesFound)
					label.Color = theme.Palette.ContrastFg
					return label.Layout(gtx)
				}

				state.selectedIndex = selectedIndexAfterListClicks(gtx, state.selectedIndex, state.itemClicks, len(state.images))

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
		label := material.Body1(theme, previewPaneText(state.imageDir, state.images, state.selectedIndex))
		label.Color = theme.Palette.Fg
		return label.Layout(gtx)
	})
}

func previewPaneText(imageDir string, images []string, selectedIndex int) string {
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
