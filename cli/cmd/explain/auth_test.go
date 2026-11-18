package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/urfave/cli/v3"
)

// @req CORE-CLI-006
func TestExplainDefersAuthenticationToOpenCode(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "opencode"), []byte("#!/bin/sh\nprintf 'explanation\n'\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	source := filepath.Join(t.TempDir(), "example.go")
	if err := os.WriteFile(source, []byte("package main\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := &cli.Command{Name: "explain", Flags: []cli.Flag{&cli.StringFlag{Name: "model"}}, Action: run}
	if err := cmd.Run(context.Background(), []string{"explain", "--model", "openai/gpt-test", source}); err != nil {
		t.Fatalf("OpenCode succeeded but explain failed: %v", err)
	}
}
