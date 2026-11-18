package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/urfave/cli/v3"
)

// @req CORE-CLI-006
func TestGitmessageDefersAuthenticationToOpenCode(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	bin := t.TempDir()
	for name, script := range map[string]string{
		"git":      "#!/bin/sh\nprintf 'staged diff\n'\n",
		"opencode": "#!/bin/sh\nprintf 'commit message\n'\n",
	} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte(script), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	cmd := &cli.Command{Name: "gitmessage", Flags: []cli.Flag{&cli.StringFlag{Name: "model"}}, Action: run}
	if err := cmd.Run(context.Background(), []string{"gitmessage", "--model", "openai/gpt-test"}); err != nil {
		t.Fatalf("OpenCode succeeded but gitmessage failed: %v", err)
	}
}
