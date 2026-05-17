package main

import "testing"

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
