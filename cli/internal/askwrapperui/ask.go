//go:build gio

package askwrapperui

import (
	"context"
	"fmt"
	"image/color"
	"runtime"
	"time"

	"module/lib/internal/askwrapper"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type askResult struct {
	question   string
	answer     string
	runErr     error
	historyErr error
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func init() {
	if runtime.GOOS == "darwin" {
		spinnerFrames = []string{"-", "\\", "|", "/"}
	}
}

type askUI struct {
	timeout time.Duration

	history       []askwrapper.HistoryEntry
	historyClicks []widget.Clickable
	historyDelete []widget.Clickable
	historyList   widget.List
	previewList   widget.List
	selected      int
	modeFollowUp  bool

	question widget.Editor
	submit   widget.Clickable
	modeEnum widget.Enum

	running   bool
	preview   string
	status    string
	runCancel context.CancelFunc

	resultCh              chan askResult
	focusInputOnNextFrame bool
	focusDeleteIndex       int
}

func RunAskMode(ctx context.Context, timeout time.Duration) error {
	return runUIMode(ctx, timeout, false)
}

func RunFollowUpMode(ctx context.Context, timeout time.Duration) error {
	return runUIMode(ctx, timeout, true)
}

func runUIMode(ctx context.Context, timeout time.Duration, followUp bool) error {
	history, err := askwrapper.LoadHistory()
	if err != nil {
		return fmt.Errorf("load history: %w", err)
	}
	if followUp && len(history) == 0 {
		return fmt.Errorf("no askwrapper history yet")
	}

	ui := newAskUI(history, timeout, followUp)

	goErr := make(chan error, 1)
	go func() {
		w := new(app.Window)
		w.Option(
			app.Title(ui.windowTitle()),
			app.Size(unit.Dp(760), unit.Dp(500)),
			app.MinSize(unit.Dp(680), unit.Dp(440)),
		)
		goErr <- ui.loop(w, ctx)
	}()

	app.Main()
	return <-goErr
}

func newAskUI(history []askwrapper.HistoryEntry, timeout time.Duration, followUp bool) *askUI {
	ui := &askUI{
		timeout:               timeout,
		history:               history,
		historyClicks:         make([]widget.Clickable, len(history)),
		historyList:           widget.List{List: layout.List{Axis: layout.Vertical}},
		selected:              -1,
		modeFollowUp:          followUp,
		preview:               "",
		status:                "",
		resultCh:              make(chan askResult, 1),
		focusInputOnNextFrame: true,
		focusDeleteIndex:      -1,
	}
	ui.historyDelete = make([]widget.Clickable, len(history))
	if followUp {
		ui.status = copyFollowUpNeedsContext
		ui.modeEnum.Value = modeFollowUpValue
	} else {
		ui.modeEnum.Value = modeAskValue
	}
	ui.question.SingleLine = true
	ui.question.Submit = true
	ui.previewList.List.Axis = layout.Vertical
	return ui
}

func (u *askUI) windowTitle() string {
	if u.modeFollowUp {
		return copyWindowTitleFollowUp
	}
	return copyWindowTitleAsk
}

func (u *askUI) loop(w *app.Window, parent context.Context) error {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	th.Palette.Bg = color.NRGBA{R: 16, G: 21, B: 28, A: 255}
	th.Palette.Fg = color.NRGBA{R: 222, G: 228, B: 236, A: 255}
	th.Palette.ContrastBg = color.NRGBA{R: 41, G: 109, B: 196, A: 255}
	th.Palette.ContrastFg = color.NRGBA{R: 248, G: 251, B: 255, A: 255}

	var ops op.Ops
	for {
		select {
		case res := <-u.resultCh:
			u.applyAskResult(res)
			w.Invalidate()
		default:
		}

		e := w.Event()
		switch e := e.(type) {
		case app.DestroyEvent:
			return u.handleDestroyEvent(e.Err)
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			if u.focusInputOnNextFrame {
				gtx.Execute(key.FocusCmd{Tag: &u.question})
				u.focusInputOnNextFrame = false
			}
			u.applyPendingDeleteFocus(gtx)
			if u.running {
				w.Invalidate()
			}

			for {
				ev, ok := u.question.Update(gtx)
				if !ok {
					break
				}
				if _, isSubmit := ev.(widget.SubmitEvent); isSubmit {
					u.startAsk(parent)
				}
			}

			if !u.running && u.modeEnum.Update(gtx) {
				u.switchMode(u.modeEnum.Value == modeFollowUpValue)
			}

			if !u.running {
				for {
					ev, ok := gtx.Source.Event(
						key.Filter{Name: key.NameDeleteForward},
						key.Filter{Name: key.NameDeleteBackward, Required: key.ModShortcut},
						key.Filter{Name: "D", Required: key.ModShortcut},
					)
					if !ok {
						break
					}
					if ke, isKey := ev.(key.Event); isKey {
						switch ke.Name {
						case key.NameDeleteForward, key.NameDeleteBackward, "D":
							u.deleteSelected()
						}
					}
				}
			}

			u.draw(gtx, th, parent)
			e.Frame(gtx.Ops)
		}
	}
}
