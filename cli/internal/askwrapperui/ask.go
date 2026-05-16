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
		fmt.Print("Question (or :N to preview history, blank to cancel): ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("read input: %w", err)
		}

		input := strings.TrimSpace(line)
		if input == "" {
			fmt.Fprintln(os.Stderr, "Cancelled")
			return nil
		}

		if strings.HasPrefix(input, ":") {
			if err := previewHistory(history, strings.TrimPrefix(input, ":")); err != nil {
				fmt.Fprintf(os.Stderr, "Preview error: %v\n", err)
			}
			continue
		}

		fmt.Fprintln(os.Stderr, "Running ask...")
		answer, err := askwrapper.RunAsk(ctx, input, timeout)
		if err != nil {
			return err
		}

		fmt.Println(answer)
		if err := askwrapper.AppendHistory(input, answer); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to append history: %v\n", err)
		}
		return nil
	}
}

func printHistory(entries []askwrapper.HistoryEntry) {
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No askwrapper history yet.")
		return
	}

	limit := 8
	if len(entries) < limit {
		limit = len(entries)
	}

	fmt.Fprintln(os.Stderr, "Recent askwrapper history:")
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
			fmt.Fprintf(os.Stderr, "  %d. %s (%s)\n", i+1, strings.TrimSpace(entry.Question), when)
		} else {
			fmt.Fprintf(os.Stderr, "  %d. %s\n", i+1, strings.TrimSpace(entry.Question))
		}
		if preview != "" {
			fmt.Fprintf(os.Stderr, "     %s\n", preview)
		}
	}
	fmt.Fprintln(os.Stderr, "Tip: use :N to preview full answer for entry N.")
}

func previewHistory(entries []askwrapper.HistoryEntry, indexText string) error {
	idx, err := strconv.Atoi(strings.TrimSpace(indexText))
	if err != nil || idx < 1 || idx > len(entries) {
		return fmt.Errorf("invalid history index")
	}
	entry := entries[idx-1]
	fmt.Fprintf(os.Stderr, "\n[%d] %s\n", idx, strings.TrimSpace(entry.Question))
	fmt.Fprintf(os.Stderr, "%s\n\n", strings.TrimSpace(entry.Answer))
	return nil
}

func oneLine(input string) string {
	out := strings.ReplaceAll(input, "\r\n", "\n")
	out = strings.ReplaceAll(out, "\n", " ")
	out = strings.Join(strings.Fields(out), " ")
	return out
}
