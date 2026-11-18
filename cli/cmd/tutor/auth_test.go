package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/urfave/cli/v3"
)

// @req CORE-CLI-006
func TestTutorDefersAuthenticationForBothModels(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	bin := t.TempDir()
	script := `#!/bin/sh
case "$*" in
  *"Respond ONLY with a valid JSON"*) printf '%s\n' '{"action":"continue","reason":"ready","suggested_filename":"tutorial.md"}' ;;
  *) printf 'tutorial\n' ;;
esac
`
	if err := os.WriteFile(filepath.Join(bin, "opencode"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	cmd := &cli.Command{
		Name: "tutor",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "validation-model"},
			&cli.StringFlag{Name: "generation-model"},
		},
		Action: run,
	}
	args := []string{"tutor", "--validation-model", "openai/gpt-test", "--generation-model", "anthropic/claude-test", "question"}
	if err := cmd.Run(context.Background(), args); err != nil {
		t.Fatalf("OpenCode succeeded for both models but tutor failed: %v", err)
	}
}
