package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
	"module/lib/internal/agent"
	"module/lib/internal/constants"
)

type selection struct {
	filePath string
	start    int
	end      int
}

func main() {
	cmd := &cli.Command{
		Name:  "ctx-edit",
		Usage: "Update a selected code context using a prompt",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "prompt",
				Aliases: []string{"p"},
				Usage:   "Instruction for how to update the selection",
			},
			&cli.StringFlag{
				Name:    "context",
				Aliases: []string{"c"},
				Usage:   "Context selection (use '-' to read from stdin)",
			},
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "File path to select context from",
			},
			&cli.IntFlag{
				Name:  "start",
				Usage: "Start line for selection (1-based)",
			},
			&cli.IntFlag{
				Name:  "end",
				Usage: "End line for selection (1-based)",
			},
			&cli.BoolFlag{
				Name:  "write",
				Usage: "Apply the update in-place (requires --file)",
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Output result as JSON",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Show verbose agent output",
			},
			&cli.StringFlag{
				Name:    "model",
				Aliases: []string{"m"},
				Value:   constants.DefaultModel,
				Usage:   "Model to use",
			},
		},
		ArgsUsage: "[prompt...]",
		UsageText: `ctx-edit [options] [prompt...]

EXAMPLES:
   ctx-edit -p "rename foo to bar" -c "const foo = 1"
   ctx-edit -p "add error handling" -f main.go --start 12 --end 24
   pbpaste | ctx-edit -p "make this more concise" -c -
   ctx-edit -p "convert to for-range" -f main.go --start 10 --end 18 --write`,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			runCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
			defer cancel()

			return run(runCtx, cmd)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	verbose := cmd.Bool("verbose")
	model := cmd.String("model")
	write := cmd.Bool("write")
	jsonOutput := cmd.Bool("json")

	prompt, err := getPromptInput(cmd)
	if err != nil {
		return fmt.Errorf("get prompt input: %w", err)
	}

	contextInput, selectionInfo, err := getContextInput(cmd)
	if err != nil {
		return fmt.Errorf("get context input: %w", err)
	}

	if write && selectionInfo.filePath == "" {
		return fmt.Errorf("--write requires --file")
	}

	if err := agent.ValidateAuth(model); err != nil {
		return fmt.Errorf("auth error: %w", err)
	}

	promptText := buildPrompt(prompt, contextInput)

	fmt.Fprintln(os.Stderr, "Generating update...")
	agentRunner := agent.New(verbose)
	updated, err := agentRunner.Run(ctx, model, promptText)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	if strings.TrimSpace(updated) == "" {
		return fmt.Errorf("agent returned empty content")
	}

	if !write {
		if jsonOutput {
			return printJSON(result{
				UpdatedText: updated,
				Applied:     false,
			})
		}
		fmt.Println(updated)
		return nil
	}

	updatedFile, err := applyUpdate(selectionInfo, updated)
	if err != nil {
		return fmt.Errorf("apply update: %w", err)
	}

	if jsonOutput {
		return printJSON(result{
			UpdatedText: updated,
			Applied:     true,
			FilePath:    updatedFile,
			StartLine:   selectionInfo.start,
			EndLine:     selectionInfo.end,
		})
	}

	fmt.Printf("Updated %s (lines %d-%d)\n", updatedFile, selectionInfo.start, selectionInfo.end)
	return nil
}

func getPromptInput(cmd *cli.Command) (string, error) {
	if prompt := strings.TrimSpace(cmd.String("prompt")); prompt != "" {
		return prompt, nil
	}

	if cmd.Args().Len() > 0 {
		return strings.TrimSpace(strings.Join(cmd.Args().Slice(), " ")), nil
	}

	fmt.Print("Prompt: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("empty input provided")
	}

	return input, nil
}

func getContextInput(cmd *cli.Command) (string, selection, error) {
	contextFlag := cmd.String("context")
	fileFlag := strings.TrimSpace(cmd.String("file"))
	start := cmd.Int("start")
	end := cmd.Int("end")

	if contextFlag != "" && fileFlag != "" {
		return "", selection{}, fmt.Errorf("use either --context or --file, not both")
	}

	if strings.TrimSpace(contextFlag) == "" && fileFlag == "" {
		return "", selection{}, fmt.Errorf("provide --context or --file")
	}

	if strings.TrimSpace(contextFlag) != "" {
		if contextFlag == "-" {
			content, err := readStdin()
			if err != nil {
				return "", selection{}, err
			}
			return content, selection{}, nil
		}

		return contextFlag, selection{}, nil
	}

	contentBytes, err := os.ReadFile(fileFlag)
	if err != nil {
		return "", selection{}, fmt.Errorf("read file: %w", err)
	}

	content := string(contentBytes)
	lines := strings.Split(content, "\n")

	if start == 0 && end == 0 {
		start = 1
		end = len(lines)
	} else {
		if start == 0 {
			start = 1
		}
		if end == 0 {
			end = start
		}
	}

	if start < 1 || end < 1 || end < start {
		return "", selection{}, fmt.Errorf("invalid line range %d-%d", start, end)
	}
	if start > len(lines) || end > len(lines) {
		return "", selection{}, fmt.Errorf("line range %d-%d exceeds file length %d", start, end, len(lines))
	}

	selectionLines := lines[start-1 : end]
	return strings.Join(selectionLines, "\n"), selection{filePath: fileFlag, start: start, end: end}, nil
}

func readStdin() (string, error) {
	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(string(content)) == "" {
		return "", fmt.Errorf("empty context provided")
	}

	return string(content), nil
}

func buildPrompt(prompt, contextInput string) string {
	return fmt.Sprintf(`You are a senior software engineer.
Update the selected code context based on the instruction.
Rules:
- Only output the updated code for the selection.
- Do not include markdown fences or explanations.
- Preserve formatting and indentation.
- If no changes are needed, output the original selection.

Instruction:
%s

Selected context:
%s`, prompt, contextInput)
}

func applyUpdate(sel selection, updated string) (string, error) {
	if sel.filePath == "" {
		return "", fmt.Errorf("missing file path")
	}

	contentBytes, err := os.ReadFile(sel.filePath)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}

	content := string(contentBytes)
	lines := strings.Split(content, "\n")
	if sel.start < 1 || sel.end < sel.start || sel.end > len(lines) {
		return "", fmt.Errorf("invalid selection range %d-%d", sel.start, sel.end)
	}

	replacementLines := strings.Split(updated, "\n")
	newLines := append([]string{}, lines[:sel.start-1]...)
	newLines = append(newLines, replacementLines...)
	newLines = append(newLines, lines[sel.end:]...)

	trailingNewline := strings.HasSuffix(content, "\n")
	updatedContent := strings.Join(newLines, "\n")

	if trailingNewline && !strings.HasSuffix(updatedContent, "\n") {
		updatedContent += "\n"
	}
	if !trailingNewline && strings.HasSuffix(updatedContent, "\n") {
		updatedContent = strings.TrimSuffix(updatedContent, "\n")
	}

	info, err := os.Stat(sel.filePath)
	if err != nil {
		return "", fmt.Errorf("stat file: %w", err)
	}

	if err := os.WriteFile(sel.filePath, []byte(updatedContent), info.Mode().Perm()); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return sel.filePath, nil
}

type result struct {
	UpdatedText string `json:"updated_text"`
	Applied     bool   `json:"applied"`
	FilePath    string `json:"file_path,omitempty"`
	StartLine   int    `json:"start_line,omitempty"`
	EndLine     int    `json:"end_line,omitempty"`
}

func printJSON(payload result) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	fmt.Println(string(encoded))
	return nil
}
