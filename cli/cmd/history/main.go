package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"
	"module/lib/internal/filewriter"
	"module/lib/internal/history"
)

func main() {
	cmd := &cli.Command{
		Name:  "history",
		Usage: "Manage and export command history",
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List history entries",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "command",
						Aliases: []string{"c"},
						Usage: "Filter by command " +
							"(ask, explain, tutor, gitmessage)",
					},
				},
				Action: listAction,
			},
			{
				Name:  "export",
				Usage: "Export history entries to markdown",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "command",
						Aliases: []string{"c"},
						Usage: "Filter by command " +
							"(ask, explain, tutor, gitmessage)",
					},
					&cli.StringFlag{
						Name:    "file",
						Aliases: []string{"f"},
						Usage:   "Specific history JSON file to export",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Output markdown file path",
					},
				},
				Action: exportAction,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func listAction(ctx context.Context, cmd *cli.Command) error {
	commandFilter := cmd.String("command")

	var allEntries map[string][]history.Entry
	var err error

	if commandFilter != "" {
		// List specific command
		entries, err := history.List(commandFilter)
		if err != nil {
			return fmt.Errorf("list history: %w", err)
		}

		if len(entries) == 0 {
			fmt.Printf("No history found for command: %s\n",
				commandFilter)
			return nil
		}

		allEntries = map[string][]history.Entry{commandFilter: entries}
	} else {
		// List all commands
		allEntries, err = history.ListAll()
		if err != nil {
			return fmt.Errorf("list all history: %w", err)
		}

		if len(allEntries) == 0 {
			fmt.Println("No history found")
			return nil
		}
	}

	// Print entries grouped by command
	commands := make([]string, 0, len(allEntries))
	for cmd := range allEntries {
		commands = append(commands, cmd)
	}
	sort.Strings(commands)

	for _, cmdName := range commands {
		entries := allEntries[cmdName]
		fmt.Printf("\n=== %s (%d entries) ===\n\n", cmdName, len(entries))

		for i := range entries {
			fmt.Printf("[%d] %s\n", i+1, entries[i].Date)
			fmt.Printf("    File: %s\n", entries[i].FilePath)

			// Print key data fields
			if question, ok := entries[i].Data["question"].(string); ok {
				fmt.Printf("    Question: %s\n", truncate(question, 60))
			}
			if source, ok := entries[i].Data["source"].(string); ok {
				fmt.Printf("    Source: %s\n", source)
			}
			if hint, ok := entries[i].Data["hint"].(string); ok && hint != "" {
				fmt.Printf("    Hint: %s\n", truncate(hint, 60))
			}
			if result, ok := entries[i].Data["result"].(string); ok {
				fmt.Printf("    Result: %s\n", truncate(result, 80))
			}

			fmt.Print("\n")
		}
	}

	return nil
}

func exportAction(ctx context.Context, cmd *cli.Command) error {
	commandFilter := cmd.String("command")
	fileFilter := cmd.String("file")
	outputPath := cmd.String("output")

	// If specific file is provided, export it directly
	if fileFilter != "" {
		return exportFile(fileFilter, outputPath)
	}

	// Otherwise, show interactive selection
	return exportInteractive(commandFilter, outputPath)
}

func exportFile(filePath string, outputPath string) error {
	// Load the entry
	entry, err := history.Load(filePath)
	if err != nil {
		return fmt.Errorf("load history file: %w", err)
	}

	// Generate markdown content
	markdown := formatEntryAsMarkdown(entry)

	// Determine output path
	if outputPath == "" {
		outputPath, err = suggestOutputPath(entry)
		if err != nil {
			return err
		}
	}

	// Write to file
	finalPath, err := filewriter.WriteAtomic(outputPath, markdown)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	fmt.Printf("Exported to: %s\n", finalPath)
	return nil
}

func exportInteractive(commandFilter string, outputPath string) error {
	var entries []history.Entry
	var err error

	if commandFilter != "" {
		entries, err = history.List(commandFilter)
		if err != nil {
			return fmt.Errorf("list history: %w", err)
		}
	} else {
		// Get all entries from all commands
		allEntries, err := history.ListAll()
		if err != nil {
			return fmt.Errorf("list all history: %w", err)
		}

		// Flatten into a single slice
		for _, cmdEntries := range allEntries {
			entries = append(entries, cmdEntries...)
		}

		// Sort by timestamp
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Timestamp > entries[j].Timestamp
		})
	}

	if len(entries) == 0 {
		fmt.Println("No history entries found")
		return nil
	}

	// Display entries
	fmt.Print("\n=== Available History Entries ===\n\n")
	for i := range entries {
		fmt.Printf("[%d] %s - %s\n", i+1, entries[i].Command, entries[i].Date)

		// Show preview of content
		if question, ok := entries[i].Data["question"].(string); ok {
			fmt.Printf("    %s\n", truncate(question, 70))
		} else if source, ok := entries[i].Data["source"].(string); ok {
			fmt.Printf("    Source: %s\n", source)
		} else if result, ok := entries[i].Data["result"].(string); ok {
			fmt.Printf("    %s\n", truncate(result, 70))
		}
	}

	// Prompt for selection
	fmt.Print("\nSelect entry to export (number or 'q' to quit): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "q" || input == "quit" {
		return nil
	}

	// Parse selection
	selection, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("invalid selection: %w", err)
	}

	if selection < 1 || selection > len(entries) {
		return fmt.Errorf("selection out of range (1-%d)", len(entries))
	}

	selectedEntry := &entries[selection-1]

	// Generate markdown content
	markdown := formatEntryAsMarkdown(selectedEntry)

	// Determine output path
	if outputPath == "" {
		outputPath, err = suggestOutputPath(selectedEntry)
		if err != nil {
			return err
		}
	}

	// Write to file
	finalPath, err := filewriter.WriteAtomic(outputPath, markdown)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	fmt.Printf("Exported to: %s\n", finalPath)
	return nil
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func formatEntryAsMarkdown(entry *history.Entry) string {
	var sb strings.Builder

	// Title based on command type
	sb.WriteString(fmt.Sprintf("# %s History Entry\n\n",
		capitalize(entry.Command)))

	// Metadata
	sb.WriteString("## Metadata\n\n")
	sb.WriteString(fmt.Sprintf("- **Date**: %s\n", entry.Date))
	sb.WriteString(fmt.Sprintf("- **Command**: `%s`\n", entry.Command))
	sb.WriteString(fmt.Sprintf("- **Timestamp**: %d\n\n", entry.Timestamp))

	// Command-specific content
	switch entry.Command {
	case "ask":
		if q, ok := entry.Data["question"].(string); ok {
			sb.WriteString("## Question\n\n")
			sb.WriteString(fmt.Sprintf("%s\n\n", q))
		}
		if r, ok := entry.Data["result"].(string); ok {
			sb.WriteString("## Answer\n\n")
			sb.WriteString(fmt.Sprintf("%s\n\n", r))
		}

	case "explain":
		if src, ok := entry.Data["source"].(string); ok {
			sb.WriteString("## Source\n\n")
			sb.WriteString(fmt.Sprintf("`%s`\n\n", src))
		}
		if lang, ok := entry.Data["language"].(string); ok {
			sb.WriteString("## Language\n\n")
			sb.WriteString(fmt.Sprintf("%s\n\n", lang))
		}
		if out, ok := entry.Data["output"].(string); ok {
			sb.WriteString("## Output File\n\n")
			sb.WriteString(fmt.Sprintf("`%s`\n\n", out))
		}
		if r, ok := entry.Data["result"].(string); ok {
			sb.WriteString("## Explanation\n\n")
			sb.WriteString(fmt.Sprintf("%s\n\n", r))
		}

	case "tutor":
		if q, ok := entry.Data["question"].(string); ok {
			sb.WriteString("## Question\n\n")
			sb.WriteString(fmt.Sprintf("%s\n\n", q))
		}
		if out, ok := entry.Data["output"].(string); ok {
			sb.WriteString("## Output File\n\n")
			sb.WriteString(fmt.Sprintf("`%s`\n\n", out))
		}
		if r, ok := entry.Data["result"].(string); ok {
			sb.WriteString("## Tutorial\n\n")
			sb.WriteString(fmt.Sprintf("%s\n\n", r))
		}

	case "gitmessage":
		if hint, ok := entry.Data["hint"].(string); ok && hint != "" {
			sb.WriteString("## Hint\n\n")
			sb.WriteString(fmt.Sprintf("%s\n\n", hint))
		}
		if prefix, ok := entry.Data["prefix"].(string); ok && prefix != "" {
			sb.WriteString("## Prefix\n\n")
			sb.WriteString(fmt.Sprintf("%s\n\n", prefix))
		}
		if r, ok := entry.Data["result"].(string); ok {
			sb.WriteString("## Commit Message\n\n")
			sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", r))
		}

	default:
		// Generic handling for unknown commands
		sb.WriteString("## Data\n\n")
		for k, v := range entry.Data {
			sb.WriteString(fmt.Sprintf("### %s\n\n%v\n\n",
				capitalize(k), v))
		}
	}

	sb.WriteString("---\n\n")
	sb.WriteString(fmt.Sprintf("*Exported from %s*\n", entry.FilePath))

	return sb.String()
}

func suggestOutputPath(entry *history.Entry) (string, error) {
	// Generate suggestion based on command and timestamp
	basename := fmt.Sprintf("%s-history-%d.md",
		entry.Command,
		entry.Timestamp)

	fmt.Printf("Save to file (default: %s) => ", basename)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read input: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return basename, nil
	}

	return input, nil
}

func truncate(s string, maxLen int) string {
	// Remove newlines and extra whitespace
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.Join(strings.Fields(s), " ")

	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen-3] + "..."
}
