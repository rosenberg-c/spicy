package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

// @req CORE-CLI-006
func TestAskDefersAuthenticationToOpenCode(t *testing.T) {
	for _, tc := range []struct {
		name, auth, script, wantError string
	}{
		{"missing auth", "", "#!/bin/sh\nprintf 'answer\n'\n", ""},
		{"expired auth", `{"openai":{"expires":1}}`, "#!/bin/sh\nprintf 'answer\n'\n", ""},
		{"opencode failure", "", "#!/bin/sh\nexit 9\n", "generation failed: agent command failed: exit status 9"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			if tc.auth != "" {
				dir := filepath.Join(home, ".local", "share", "opencode")
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "auth.json"), []byte(tc.auth), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			bin := t.TempDir()
			if err := os.WriteFile(filepath.Join(bin, "opencode"), []byte(tc.script), 0o700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			cmd := &cli.Command{Name: "ask", Flags: []cli.Flag{&cli.StringFlag{Name: "model"}}, Action: run}
			err := cmd.Run(context.Background(), []string{"ask", "--model", "openai/gpt-test", "question"})
			if tc.wantError == "" {
				if err != nil {
					t.Fatalf("OpenCode succeeded but ask failed: %v", err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("ask error = %v, want %q", err, tc.wantError)
			}
		})
	}
}
