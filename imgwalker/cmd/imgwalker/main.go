package main

import (
	"image/color"
	"log"

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
	return windowConfig{
		title:  windowTitle,
		width:  windowWidth,
		height: windowHeight,
	}
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
