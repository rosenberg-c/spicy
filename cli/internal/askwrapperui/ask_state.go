//go:build gio

package askwrapperui

import (
	"context"
	"strings"
	"time"

	"module/lib/internal/askwrapper"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
)

func (u *askUI) startAsk(parent context.Context) {
	if u.running {
		return
	}
	q := strings.TrimSpace(u.question.Text())
	if q == "" {
		u.setStatus(copyQuestionEmpty)
		return
	}

	selected := u.selected
	if u.modeFollowUp {
		if selected < 0 || selected >= len(u.history) {
			u.setStatus(copyPickHistoryContext)
			return
		}
	}

	u.running = true
	u.setStatus(copyRunningAsk)
	runCtx, runCancel := context.WithCancel(parent)
	u.runCancel = runCancel

	go func(question string) {
		prompt := question
		if u.modeFollowUp {
			ctxItem := u.history[selected]
			prompt = askwrapper.BuildFollowUpPrompt(ctxItem.Question, ctxItem.Answer, question)
		}

		answer, runErr := askwrapper.RunAsk(runCtx, prompt, u.timeout)
		var historyErr error
		if runErr == nil {
			historyErr = askwrapper.AppendHistory(question, answer)
		}

		u.resultCh <- askResult{question: question, answer: answer, runErr: runErr, historyErr: historyErr}
	}(q)
}

func (u *askUI) applyAskResult(res askResult) {
	u.running = false
	u.runCancel = nil
	if res.answer != "" {
		u.history = append([]askwrapper.HistoryEntry{{
			Question: res.question,
			Answer:   res.answer,
			At:       time.Now().Unix(),
		}}, u.history...)
		u.historyClicks = make([]widget.Clickable, len(u.history))
		u.historyDelete = make([]widget.Clickable, len(u.history))
		u.selected = 0
		u.setPreview(res.answer)
		u.question.SetText("")
	}
	if res.runErr != nil {
		u.setStatus(copyAskFailed(res.runErr))
		return
	}
	if res.historyErr != nil {
		u.setStatus(copyAppendHistoryFailed(res.historyErr))
		return
	}
	u.setStatus(copyAskComplete)
}

func (u *askUI) deleteSelected() {
	if u.selected < 0 || u.selected >= len(u.history) {
		u.setStatus(copyNoHistorySelected)
		return
	}
	deletedIndex := u.selected
	if err := askwrapper.DeleteHistoryAt(u.selected); err != nil {
		u.setStatus(copyDeleteFailed(err))
		return
	}
	updated, err := askwrapper.LoadHistory()
	if err != nil {
		u.setStatus(copyReloadFailed(err))
		return
	}
	u.history = updated
	u.historyClicks = make([]widget.Clickable, len(updated))
	u.historyDelete = make([]widget.Clickable, len(updated))
	u.focusDeleteIndex = nextDeleteFocusIndex(deletedIndex, len(updated))
	if len(updated) == 0 {
		u.selected = -1
		u.setPreview(copyNoHistory)
	} else if u.selected >= len(updated) {
		u.selected = len(updated) - 1
	}
	if u.selected >= 0 && u.selected < len(updated) {
		u.setPreview(previewForHistory(updated[u.selected]))
	}
	u.setStatus(copyDeletedHistory)
}

func (u *askUI) deleteAt(index int) {
	if index < 0 || index >= len(u.history) {
		u.setStatus(copyNoHistorySelected)
		return
	}
	u.selected = index
	u.deleteSelected()
}

func (u *askUI) askLabel() string {
	if u.modeFollowUp {
		return copyModeFollowUpLabel
	}
	return copyModeAskLabel
}

func (u *askUI) switchMode(followUp bool) {
	u.modeFollowUp = followUp
	if followUp {
		u.modeEnum.Value = modeFollowUpValue
	} else {
		u.modeEnum.Value = modeAskValue
	}
	u.question.SetText("")
	if followUp {
		if u.selected >= 0 && u.selected < len(u.history) {
			u.setStatus(copyFollowUpContextSet)
		} else {
			u.setStatus(copyFollowUpNeedsContext)
		}
		return
	}
	u.setStatus(copyAskMode)
}

func (u *askUI) cancelAsk() {
	if !u.running {
		return
	}
	if u.runCancel != nil {
		u.runCancel()
	}
	u.setStatus(copyCancellingAsk)
}

func (u *askUI) handleDestroyEvent(err error) error {
	if u.running && u.runCancel != nil {
		u.runCancel()
	}
	return err
}

func (u *askUI) setPreview(text string) {
	u.previewText = text
	u.preview.SetText(text)
	u.preview.SetCaret(0, 0)
}

func (u *askUI) setStatus(text string) {
	u.status = strings.TrimSpace(text)
}

func (u *askUI) statusLine() string {
	if u.running {
		idx := int((time.Now().UnixNano() / int64(100*time.Millisecond)) % int64(len(spinnerFrames)))
		frame := spinnerFrames[idx]
		prefix := "[" + frame + "]"
		if u.status != "" {
			return prefix + " " + u.status
		}
		return prefix + " " + copyRunningAsk
	}
	if u.status != "" {
		return u.status
	}
	return u.helpText()
}

func (u *askUI) helpText() string {
	if u.running {
		return copyHelpRunning
	}
	if u.modeFollowUp {
		if u.selected >= 0 && u.selected < len(u.history) {
			return copyHelpFollowUpReady
		}
		return copyHelpFollowUpNeedsSelect
	}
	return copyHelpAsk
}

func (u *askUI) primaryLabel() string {
	if u.running {
		return copyPrimaryCancel
	}
	return copyPrimaryIdle
}

func (u *askUI) canSubmit() bool {
	if u.running {
		return true
	}
	if !u.modeFollowUp {
		return true
	}
	return u.selected >= 0 && u.selected < len(u.history)
}

func nextDeleteFocusIndex(deletedIndex, updatedLen int) int {
	if updatedLen <= 0 {
		return -1
	}
	if deletedIndex < 0 {
		return 0
	}
	if deletedIndex >= updatedLen {
		return updatedLen - 1
	}
	return deletedIndex
}

func (u *askUI) applyPendingDeleteFocus(gtx layout.Context) {
	if u.focusDeleteIndex < 0 || u.focusDeleteIndex >= len(u.historyDelete) {
		return
	}
	gtx.Execute(key.FocusCmd{Tag: &u.historyDelete[u.focusDeleteIndex]})
	u.focusDeleteIndex = -1
}

func (u *askUI) isQuestionReadOnly() bool { return u.running }

func (u *askUI) isHistoryDisabled() bool { return u.running }

func (u *askUI) inputRowAxis() layout.Axis { return layout.Horizontal }

func previewForHistory(entry askwrapper.HistoryEntry) string {
	q := strings.TrimSpace(entry.Question)
	a := strings.TrimSpace(entry.Answer)
	if q == "" {
		q = copyEmptyQuestion
	}
	if a == "" {
		a = copyEmptyAnswer
	}
	return copyPreviewHistory(q, a)
}
