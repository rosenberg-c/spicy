package askwrapperui

import "fmt"

const (
	modeAskValue      = "ask"
	modeFollowUpValue = "followup"

	copyWindowTitleAsk      = "askwrapper ui ask"
	copyWindowTitleFollowUp = "askwrapper ui followup"

	copyModeAskLabel      = "Ask"
	copyModeFollowUpLabel = "Follow-up"

	copyFollowUpNeedsContext = "Follow-up mode: select a context item first."
	copyFollowUpContextSet   = "Follow-up mode: context selected."
	copyFollowUpContextPick  = "Follow-up context selected."
	copyAskComplete          = "Ask complete."
	copyQuestionHint         = "Type your question..."
	copyPrimaryIdle          = "→"
	copyPrimaryCancel        = "Cancel"
	copyHistoryTitle         = "History"
	copyPreviewTitle         = "Preview"
	copyNoHistory            = "No history yet."
	copyRowDelete            = "Del"

	copyQuestionEmpty      = "Question cannot be empty."
	copyPickHistoryContext = "Pick a history context first."
	copyRunningAsk         = "Running ask..."
	copyNoHistorySelected  = "No history selected."
	copyDeletedHistory     = "Deleted selected history entry."
	copyAskMode            = "Ask mode."
	copyCancellingAsk      = "Cancelling ask..."

	copyHelpRunning             = "Running ask... input and history are temporarily disabled."
	copyHelpFollowUpReady       = "Follow-up mode: context selected. Enter question and press Ask. Delete/Ctrl+D removes selected history."
	copyHelpFollowUpNeedsSelect = "Follow-up mode: select a history item first, then enter a question. Delete/Ctrl+D removes selected history."
	copyHelpAsk                 = "Ask mode: enter a question and press Ask. Select history to preview. Delete/Ctrl+D removes selected history."

	copyEmptyQuestion = "(empty question)"
	copyEmptyAnswer   = "(empty answer)"
	copyPreviewFormat = "Question:\n%s\n\nAnswer:\n%s"

	copyTerminalPromptQuestion = "Question (or :N to preview history, blank to cancel): "
	copyTerminalPromptSelect   = "Select context with :N (or :dN delete, blank to cancel): "
	copyTerminalPromptFollowUp = "Follow-up question (blank to cancel): "
	copyTerminalCancelled      = "Cancelled"
	copyTerminalRunAsk         = "Running ask..."
	copyTerminalRunFollowUp    = "Running follow-up ask..."
	copyTerminalNoAskHistory   = "No askwrapper history yet."

	copyTerminalHistoryTitle = "Recent askwrapper history:"
	copyTerminalTipPreview   = "Tip: use :N to preview full answer for entry N."
	copyTerminalTipDelete    = "Tip: use :dN to delete history entry N."
	copyTerminalInvalidIndex = "invalid history index"
)

func copyAskFailed(err error) string { return fmt.Sprintf("ask failed: %v", err) }
func copyAppendHistoryFailed(err error) string {
	return fmt.Sprintf("ask complete, but history append failed: %v", err)
}
func copyDeleteFailed(err error) string { return fmt.Sprintf("delete failed: %v", err) }
func copyReloadFailed(err error) string { return fmt.Sprintf("reload history failed: %v", err) }

func copyPreviewHistory(q, a string) string { return fmt.Sprintf(copyPreviewFormat, q, a) }

func copyDeleteError(err error) string { return fmt.Sprintf("Delete error: %v\n", err) }
func copyPreviewError(err error) string { return fmt.Sprintf("Preview error: %v\n", err) }
func copySelectError(err error) string { return fmt.Sprintf("Select error: %v\n", err) }
func copyAppendWarn(err error) string {
	return fmt.Sprintf("Warning: failed to append history: %v\n", err)
}
func copyDeletedEntry(idx int) string { return fmt.Sprintf("Deleted history entry %d\n", idx+1) }
func copyPreviewHeader(idx int, q string) string { return fmt.Sprintf("\n[%d] %s\n", idx+1, q) }
func copyPreviewBody(a string) string { return fmt.Sprintf("%s\n\n", a) }
func copyHistoryLineWithWhen(i int, q, when string) string {
	return fmt.Sprintf("  %d. %s (%s)\n", i+1, q, when)
}
func copyHistoryLine(i int, q string) string { return fmt.Sprintf("  %d. %s\n", i+1, q) }
func copyHistoryPreviewLine(p string) string { return fmt.Sprintf("     %s\n", p) }
