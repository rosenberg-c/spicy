package main

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestAskMode_TimeoutValidation(t *testing.T) {
	// @req CLI-ASKWRAPPER-005
	cmd := buildCommand()
	err := cmd.Run(context.Background(), []string{"askwrapper", "ui", "ask", "--timeout", "0"})
	if err == nil {
		t.Fatal("expected timeout validation error")
	}
	if !strings.Contains(err.Error(), "timeout must be greater than zero") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAskMode_WiresTimeoutToRunner(t *testing.T) {
	// @req CLI-ASKWRAPPER-001
	// @req CLI-ASKWRAPPER-005
	orig := runAskMode
	t.Cleanup(func() { runAskMode = orig })

	called := false
	var got time.Duration
	runAskMode = func(_ context.Context, timeout time.Duration) error {
		called = true
		got = timeout
		return nil
	}

	cmd := buildCommand()
	if err := cmd.Run(context.Background(), []string{"askwrapper", "ui", "ask", "--timeout", "42"}); err != nil {
		t.Fatalf("run command: %v", err)
	}
	if !called {
		t.Fatal("runAskMode was not called")
	}
	if got != 42*time.Second {
		t.Fatalf("timeout = %s, want %s", got, 42*time.Second)
	}
}

func TestFollowUpMode_WiresTimeoutToRunner(t *testing.T) {
	// @req CLI-ASKWRAPPER-002
	// @req CLI-ASKWRAPPER-005
	orig := runFollowUpMode
	t.Cleanup(func() { runFollowUpMode = orig })

	called := false
	var got time.Duration
	runFollowUpMode = func(_ context.Context, timeout time.Duration) error {
		called = true
		got = timeout
		return nil
	}

	cmd := buildCommand()
	if err := cmd.Run(context.Background(), []string{"askwrapper", "ui", "followup", "--timeout", "7"}); err != nil {
		t.Fatalf("run command: %v", err)
	}
	if !called {
		t.Fatal("runFollowUpMode was not called")
	}
	if got != 7*time.Second {
		t.Fatalf("timeout = %s, want %s", got, 7*time.Second)
	}
}
