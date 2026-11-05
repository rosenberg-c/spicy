package askwrapper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const DefaultTimeout = 120 * time.Second

func RunAsk(parent context.Context, question string, timeout time.Duration) (string, error) {
	q := strings.TrimSpace(question)
	if q == "" {
		return "", fmt.Errorf("empty question")
	}

	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ask", q)
	cmd.Env = os.Environ()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return "", fmt.Errorf("ask cancelled")
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("ask timed out after %s", timeout)
		}
		if strings.TrimSpace(stderr.String()) != "" {
			return "", fmt.Errorf("ask failed: %s", strings.TrimSpace(stderr.String()))
		}
		return "", fmt.Errorf("ask failed: %w", err)
	}

	out := strings.TrimSpace(stdout.String())
	if out == "" {
		out = "(ask returned no output)"
	}

	return out, nil
}
