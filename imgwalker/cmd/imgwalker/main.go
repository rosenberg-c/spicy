package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

const (
	windowTitle  = "ImgWalker"
	helloText    = "Hello, World!"
	windowWidth  = 640
	windowHeight = 360
)

type windowConfig struct {
	title  string
	width  int
	height int
	imageDir string
}

type startupFileConfig struct {
	ImageDir string `json:"imageDir"`
}

type ConfigLoadErrorCode string

const (
	ConfigLoadErrorNotFound      ConfigLoadErrorCode = "not_found"
	ConfigLoadErrorInvalidConfig ConfigLoadErrorCode = "invalid_config"
	ConfigLoadErrorInvalidImageDir ConfigLoadErrorCode = "invalid_image_dir"
	ConfigLoadErrorIOError       ConfigLoadErrorCode = "io_error"
)

type ConfigLoadError struct {
	Code ConfigLoadErrorCode
	Err  error
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
		window := newWindow(startupConfig())

		if err := run(window); err != nil {
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

func run(window *app.Window) error {
	var ops op.Ops
	theme := newTheme()

	for {
		event := window.Event()
		switch event := event.(type) {
		case app.DestroyEvent:
			return event.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, event)
			paint.FillShape(
				gtx.Ops,
				theme.Palette.Bg,
				clip.Rect{Max: gtx.Constraints.Max}.Op(),
			)
			layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				label := material.H3(theme, helloText)
				return label.Layout(gtx)
			})
			event.Frame(gtx.Ops)
		}
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
