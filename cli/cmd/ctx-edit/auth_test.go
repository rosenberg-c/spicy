package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/urfave/cli/v3"
)

// @req CORE-CLI-006
func TestContextEditDefersAuthenticationToOpenCode(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "opencode"), []byte("#!/bin/sh\nprintf 'updated\n'\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	cmd := &cli.Command{
		Name: "v-edit",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "model"},
			&cli.StringFlag{Name: "prompt"},
			&cli.StringFlag{Name: "context"},
		},
		Action: run,
	}
	if err := cmd.Run(context.Background(), []string{"v-edit", "--model", "openai/gpt-test", "--prompt", "edit", "--context", "old"}); err != nil {
		t.Fatalf("OpenCode succeeded but v-edit failed: %v", err)
	}
}
