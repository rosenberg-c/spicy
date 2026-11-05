//go:build !gio

package askwrapperui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"module/lib/internal/askwrapper"
)

func RunAskMode(ctx context.Context, timeout time.Duration) error {
	history, err := askwrapper.LoadHistory()
	if err != nil {
		return fmt.Errorf("load history: %w", err)
	}

	printHistory(history)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(copyTerminalPromptQuestion)
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("read input: %w", err)
		}

		input := strings.TrimSpace(line)
		if input == "" {
			fmt.Fprintln(os.Stderr, copyTerminalCancelled)
			return nil
		}

		if strings.HasPrefix(input, ":d") {
			if err := deleteHistory(&history, strings.TrimPrefix(input, ":d")); err != nil {
				fmt.Fprintf(os.Stderr, copyDeleteError(err))
			}
			continue
		}
		if strings.HasPrefix(input, ":") {
			if err := previewHistory(history, strings.TrimPrefix(input, ":")); err != nil {
				fmt.Fprintf(os.Stderr, copyPreviewError(err))
			}
			continue
		}

		fmt.Fprintln(os.Stderr, copyTerminalRunAsk)
		answer, err := askwrapper.RunAsk(ctx, input, timeout)
		if err != nil {
			return err
		}

		fmt.Println(answer)
		if err := askwrapper.AppendHistory(input, answer); err != nil {
			fmt.Fprintf(os.Stderr, copyAppendWarn(err))
		}
		return nil
	}
}

func RunFollowUpMode(ctx context.Context, timeout time.Duration) error {
	history, err := askwrapper.LoadHistory()
	if err != nil {
		return fmt.Errorf("load history: %w", err)
	}
	if len(history) == 0 {
		return fmt.Errorf("no askwrapper history yet")
	}

	printHistory(history)
	reader := bufio.NewReader(os.Stdin)

	var selected *askwrapper.HistoryEntry
	for selected == nil {
		fmt.Print(copyTerminalPromptSelect)
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("read input: %w", err)
		}
		input := strings.TrimSpace(line)
		if input == "" {
			fmt.Fprintln(os.Stderr, copyTerminalCancelled)
			return nil
		}
		if strings.HasPrefix(input, ":d") {
			if err := deleteHistory(&history, strings.TrimPrefix(input, ":d")); err != nil {
				fmt.Fprintf(os.Stderr, copyDeleteError(err))
			}
			continue
		}
		if strings.HasPrefix(input, ":") {
			idx, err := parseIndex(history, strings.TrimPrefix(input, ":"))
			if err != nil {
				fmt.Fprintf(os.Stderr, copySelectError(err))
				continue
			}
			selected = &history[idx]
			_ = previewHistory(history, strings.TrimPrefix(input, ":"))
		}
	}

	fmt.Print(copyTerminalPromptFollowUp)
	line, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}
	follow := strings.TrimSpace(line)
	if follow == "" {
		fmt.Fprintln(os.Stderr, copyTerminalCancelled)
		return nil
	}

	prompt := askwrapper.BuildFollowUpPrompt(selected.Question, selected.Answer, follow)
	fmt.Fprintln(os.Stderr, copyTerminalRunFollowUp)
	answer, err := askwrapper.RunAsk(ctx, prompt, timeout)
	if err != nil {
		return err
	}

	fmt.Println(answer)
	if err := askwrapper.AppendHistory(follow, answer); err != nil {
		fmt.Fprintf(os.Stderr, copyAppendWarn(err))
	}
	return nil
}

func printHistory(entries []askwrapper.HistoryEntry) {
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, copyTerminalNoAskHistory)
		return
	}

	limit := 8
	if len(entries) < limit {
		limit = len(entries)
	}

	fmt.Fprintln(os.Stderr, copyTerminalHistoryTitle)
	for i := 0; i < limit; i++ {
		entry := entries[i]
		when := ""
		if entry.At > 0 {
			when = time.Unix(entry.At, 0).Format("2006-01-02 15:04")
		}
		preview := oneLine(entry.Answer)
		if len(preview) > 72 {
			preview = preview[:72] + "..."
		}
		if when != "" {
			fmt.Fprintf(os.Stderr, copyHistoryLineWithWhen(i, strings.TrimSpace(entry.Question), when))
		} else {
			fmt.Fprintf(os.Stderr, copyHistoryLine(i, strings.TrimSpace(entry.Question)))
		}
		if preview != "" {
			fmt.Fprintf(os.Stderr, copyHistoryPreviewLine(preview))
		}
	}
	fmt.Fprintln(os.Stderr, copyTerminalTipPreview)
	fmt.Fprintln(os.Stderr, copyTerminalTipDelete)
}

func previewHistory(entries []askwrapper.HistoryEntry, indexText string) error {
	idx, err := parseIndex(entries, indexText)
	if err != nil {
		return err
	}
	entry := entries[idx]
	fmt.Fprintf(os.Stderr, copyPreviewHeader(idx, strings.TrimSpace(entry.Question)))
	fmt.Fprintf(os.Stderr, copyPreviewBody(strings.TrimSpace(entry.Answer)))
	return nil
}

func deleteHistory(entries *[]askwrapper.HistoryEntry, indexText string) error {
	idx, err := parseIndex(*entries, indexText)
	if err != nil {
		return err
	}
	if err := askwrapper.DeleteHistoryAt(idx); err != nil {
		return err
	}
	updated, err := askwrapper.LoadHistory()
	if err != nil {
		return err
	}
	*entries = updated
	fmt.Fprintf(os.Stderr, copyDeletedEntry(idx))
	return nil
}

func parseIndex(entries []askwrapper.HistoryEntry, indexText string) (int, error) {
	idx, err := strconv.Atoi(strings.TrimSpace(indexText))
	if err != nil || idx < 1 || idx > len(entries) {
		return 0, fmt.Errorf(copyTerminalInvalidIndex)
	}
	return idx - 1, nil
}

func oneLine(input string) string {
	out := strings.ReplaceAll(input, "\r\n", "\n")
	out = strings.ReplaceAll(out, "\n", " ")
	out = strings.Join(strings.Fields(out), " ")
	return out
}
