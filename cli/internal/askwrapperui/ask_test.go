//go:build gio

package askwrapperui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"module/lib/internal/askwrapper"

	"gioui.org/layout"
	"gioui.org/op"
)

func TestNewAskUI_DefaultsAskModeAndFocusInput(t *testing.T) {
	// @req CLI-ASKWRAPPER-028
	ui := newAskUI(nil, 5*time.Second, false)
	if !ui.focusInputOnNextFrame {
		t.Fatal("focusInputOnNextFrame = false, want true")
	}
	if ui.modeEnum.Value != modeAskValue {
		t.Fatalf("modeEnum.Value = %q, want %q", ui.modeEnum.Value, modeAskValue)
	}
}

func TestSwitchMode_UsesRadioValuesAndStatus(t *testing.T) {
	// @req CLI-ASKWRAPPER-013
	// @req CLI-ASKWRAPPER-013A
	history := []askwrapper.HistoryEntry{{Question: "q", Answer: "a"}}
	ui := newAskUI(history, time.Second, false)
	ui.selected = 0

	ui.switchMode(true)
	if ui.modeEnum.Value != modeFollowUpValue {
		t.Fatalf("follow-up modeEnum = %q, want %q", ui.modeEnum.Value, modeFollowUpValue)
	}
	if ui.status != copyFollowUpContextSet {
		t.Fatalf("follow-up status = %q, want %q", ui.status, copyFollowUpContextSet)
	}

	ui.switchMode(false)
	if ui.modeEnum.Value != modeAskValue {
		t.Fatalf("ask modeEnum = %q, want %q", ui.modeEnum.Value, modeAskValue)
	}
	if ui.status != copyAskMode {
		t.Fatalf("ask status = %q, want %q", ui.status, copyAskMode)
	}
}

func TestDeleteSelected_UpdatesPreviewToCurrentSelection(t *testing.T) {
	// @req CLI-ASKWRAPPER-020
	// @req CLI-ASKWRAPPER-021
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := askwrapper.AppendHistory("older", "old answer"); err != nil {
		t.Fatalf("append older: %v", err)
	}
	if err := askwrapper.AppendHistory("newer", "new answer"); err != nil {
		t.Fatalf("append newer: %v", err)
	}

	history, err := askwrapper.LoadHistory()
	if err != nil {
		t.Fatalf("load history: %v", err)
	}

	ui := newAskUI(history, time.Second, false)
	ui.selected = 0
	ui.setPreview(previewForHistory(history[0]))

	ui.deleteSelected()

	if ui.selected != 0 {
		t.Fatalf("selected = %d, want 0", ui.selected)
	}
	want := previewForHistory(ui.history[0])
	if ui.preview != want {
		t.Fatalf("preview = %q, want %q", ui.preview, want)
	}
}

func TestStartAsk_FollowUpRequiresSelectedContext(t *testing.T) {
	// @req CLI-ASKWRAPPER-014
	history := []askwrapper.HistoryEntry{{Question: "q", Answer: "a"}}
	ui := newAskUI(history, time.Second, true)
	ui.question.SetText("follow up question")
	ui.selected = -1

	ui.startAsk(context.Background())

	if ui.running {
		t.Fatal("running = true, want false when no follow-up context selected")
	}
	if ui.status != copyPickHistoryContext {
		t.Fatalf("status = %q, want %q", ui.status, copyPickHistoryContext)
	}
}

func TestCanSubmit_FollowUpRequiresContextSelection(t *testing.T) {
	// @req CLI-ASKWRAPPER-014
	history := []askwrapper.HistoryEntry{{Question: "q", Answer: "a"}}
	ui := newAskUI(history, time.Second, true)

	if ui.canSubmit() {
		t.Fatal("canSubmit = true, want false without selected follow-up context")
	}

	ui.selected = 0
	if !ui.canSubmit() {
		t.Fatal("canSubmit = false, want true with selected follow-up context")
	}
}

func TestApplyAskResult_HistoryAppendFailureKeepsPreviewWithWarning(t *testing.T) {
	// @req CLI-ASKWRAPPER-030
	ui := newAskUI(nil, time.Second, false)
	ui.running = true

	ui.applyAskResult(askResult{
		question:   "what",
		answer:     "answer text",
		historyErr: errors.New("disk full"),
	})

	if ui.preview != "answer text" {
		t.Fatalf("preview = %q, want answer text", ui.preview)
	}
	if !strings.Contains(ui.status, "history append failed") {
		t.Fatalf("status = %q, want append warning", ui.status)
	}
	if ui.running {
		t.Fatal("running = true, want false")
	}
}

func TestCancelAsk_CallsCancelAndUpdatesStatus(t *testing.T) {
	// @req CLI-ASKWRAPPER-017
	ui := newAskUI(nil, time.Second, false)
	ui.running = true

	cancelled := false
	ui.runCancel = func() {
		cancelled = true
	}

	ui.cancelAsk()

	if !cancelled {
		t.Fatal("expected runCancel to be called")
	}
	if ui.status != copyCancellingAsk {
		t.Fatalf("status = %q, want %q", ui.status, copyCancellingAsk)
	}
}

func TestHandleDestroyEvent_CancelsInFlightAsk(t *testing.T) {
	// @req CLI-ASKWRAPPER-029
	ui := newAskUI(nil, time.Second, false)
	ui.running = true

	cancelled := false
	ui.runCancel = func() {
		cancelled = true
	}

	err := ui.handleDestroyEvent(nil)
	if err != nil {
		t.Fatalf("handleDestroyEvent returned error: %v", err)
	}
	if !cancelled {
		t.Fatal("expected runCancel to be called")
	}
}

func TestPrimaryLabel_SwitchesBetweenArrowAndCancel(t *testing.T) {
	// @req CLI-ASKWRAPPER-019
	// @req CLI-ASKWRAPPER-024
	ui := newAskUI(nil, time.Second, false)

	if got := ui.primaryLabel(); got != copyPrimaryIdle {
		t.Fatalf("idle primary label = %q, want %q", got, copyPrimaryIdle)
	}

	ui.running = true
	if got := ui.primaryLabel(); got != copyPrimaryCancel {
		t.Fatalf("running primary label = %q, want %q", got, copyPrimaryCancel)
	}
}

func TestRunningState_DisablesQuestionInputAndHistory(t *testing.T) {
	// @req CLI-ASKWRAPPER-010
	// @req CLI-ASKWRAPPER-011
	ui := newAskUI(nil, time.Second, false)

	if ui.isQuestionReadOnly() {
		t.Fatal("question should be editable when idle")
	}
	if ui.isHistoryDisabled() {
		t.Fatal("history should be interactive when idle")
	}

	ui.running = true
	if !ui.isQuestionReadOnly() {
		t.Fatal("question should be read-only while running")
	}
	if !ui.isHistoryDisabled() {
		t.Fatal("history should be disabled while running")
	}
}

func TestInputAndPrimaryShareHorizontalRow(t *testing.T) {
	// @req CLI-ASKWRAPPER-023
	ui := newAskUI(nil, time.Second, false)
	if got := ui.inputRowAxis(); got != layout.Horizontal {
		t.Fatalf("input row axis = %v, want %v", got, layout.Horizontal)
	}
}

func TestDeleteSelected_RemovesEntryImmediatelyWithoutUndo(t *testing.T) {
	// @req CLI-ASKWRAPPER-012
	// @req CLI-ASKWRAPPER-016
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := askwrapper.AppendHistory("first", "a1"); err != nil {
		t.Fatalf("append first: %v", err)
	}
	if err := askwrapper.AppendHistory("second", "a2"); err != nil {
		t.Fatalf("append second: %v", err)
	}

	history, err := askwrapper.LoadHistory()
	if err != nil {
		t.Fatalf("load history: %v", err)
	}
	ui := newAskUI(history, time.Second, false)
	ui.selected = 0

	ui.deleteSelected()

	if len(ui.history) != 1 {
		t.Fatalf("history length = %d, want 1", len(ui.history))
	}
	if strings.Contains(strings.ToLower(ui.status), "undo") {
		t.Fatalf("status should not offer undo, got %q", ui.status)
	}
}

func TestHelpText_CommunicatesModeAndDeleteShortcut(t *testing.T) {
	// @req CLI-ASKWRAPPER-015
	ui := newAskUI(nil, time.Second, false)
	askHelp := ui.helpText()
	if !strings.Contains(askHelp, "Ask mode") {
		t.Fatalf("ask help missing mode context: %q", askHelp)
	}
	if !strings.Contains(askHelp, "Delete/Ctrl+D") {
		t.Fatalf("ask help missing delete shortcut: %q", askHelp)
	}

	follow := newAskUI(nil, time.Second, true)
	followHelp := follow.helpText()
	if !strings.Contains(followHelp, "Follow-up mode") {
		t.Fatalf("follow-up help missing mode context: %q", followHelp)
	}
	if !strings.Contains(followHelp, "Delete/Ctrl+D") {
		t.Fatalf("follow-up help missing delete shortcut: %q", followHelp)
	}
}

func TestStatusLine_ShowsSpinnerWhileRunning(t *testing.T) {
	// @req CLI-ASKWRAPPER-022
	ui := newAskUI(nil, time.Second, false)
	ui.running = true
	ui.setStatus(copyRunningAsk)

	line := ui.statusLine()
	hasSpinner := false
	for _, frame := range spinnerFrames {
		if strings.HasPrefix(line, "["+frame+"] ") {
			hasSpinner = true
			break
		}
	}
	if !hasSpinner {
		t.Fatalf("status line missing spinner prefix: %q", line)
	}
}

func TestPreviewPanel_UsesVerticalListForScrollableContent(t *testing.T) {
	// @req CLI-ASKWRAPPER-018
	ui := newAskUI(nil, time.Second, false)
	if ui.previewList.List.Axis != layout.Vertical {
		t.Fatalf("preview list axis = %v, want %v", ui.previewList.List.Axis, layout.Vertical)
	}
}

func TestHistoryRowDeleteAction_IsExplicitPerRow(t *testing.T) {
	// @req CLI-ASKWRAPPER-012
	// @req CLI-ASKWRAPPER-020
	if strings.TrimSpace(copyRowDelete) == "" {
		t.Fatal("row delete label must be explicit and non-empty")
	}
}

func TestNextDeleteFocusIndex_FollowsDeletedRowPosition(t *testing.T) {
	// @req CLI-ASKWRAPPER-031
	if got := nextDeleteFocusIndex(1, 4); got != 1 {
		t.Fatalf("nextDeleteFocusIndex(1,4) = %d, want 1", got)
	}
	if got := nextDeleteFocusIndex(3, 3); got != 2 {
		t.Fatalf("nextDeleteFocusIndex(3,3) = %d, want 2", got)
	}
	if got := nextDeleteFocusIndex(0, 0); got != -1 {
		t.Fatalf("nextDeleteFocusIndex(0,0) = %d, want -1", got)
	}
}

func TestApplyPendingDeleteFocus_ClearsPendingIndex(t *testing.T) {
	// @req CLI-ASKWRAPPER-031
	ui := newAskUI([]askwrapper.HistoryEntry{{Question: "q1", Answer: "a1"}}, time.Second, false)
	ui.focusDeleteIndex = 0

	var ops op.Ops
	gtx := layout.Context{Ops: &ops}
	ui.applyPendingDeleteFocus(gtx)

	if ui.focusDeleteIndex != -1 {
		t.Fatalf("focusDeleteIndex = %d, want -1", ui.focusDeleteIndex)
	}
}
