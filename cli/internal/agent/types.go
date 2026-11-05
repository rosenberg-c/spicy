// Package agent
package agent

import "context"

type Config struct {
	Model   string
	Verbose bool
}

type Runner interface {
	Run(ctx context.Context, model, prompt string) (string, error)
}
