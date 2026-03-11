package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"module/tutor/internal/agent"
)

type Validator struct {
	agent  agent.Runner
	logger *slog.Logger
}

// New creates a Validator that uses the given agent to validate user input.
func New(agent agent.Runner) *Validator {
	return &Validator{
		agent: agent,
		logger: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}
}

// Validate checks if the input is specific enough for tutorial generation.
// Returns ValidationResponse with action=continue or action=exit.
// Returns ValidationError if the agent response is invalid.
func (v *Validator) Validate(ctx context.Context, input string) (*ValidationResponse, error) {
	prompt := BuildValidationPrompt(input)

	v.logger.Debug("validating input", "input", input)

	// Call agent with hardcoded model
	response, err := v.agent.Run(ctx, "openai/gpt-5.2", prompt)
	if err != nil {
		return nil, fmt.Errorf("agent call failed: %w", err)
	}

	v.logger.Debug("received validation response", "response", response)

	// Parse JSON response
	var result ValidationResponse
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, &ValidationError{
			Input:    input,
			Response: response,
			Err:      fmt.Errorf("%w: %v", ErrInvalidJSON, err),
		}
	}

	// Validate required fields
	if result.Action == "" || result.Reason == "" {
		return nil, &ValidationError{
			Input:    input,
			Response: response,
			Err:      ErrMissingField,
		}
	}

	// Validate action value
	if !result.Action.IsValid() {
		return nil, &ValidationError{
			Input:    input,
			Response: response,
			Err:      ErrInvalidAction,
		}
	}

	// Validate action-specific fields
	if result.Action == ActionContinue && result.SuggestedFilename == "" {
		return nil, &ValidationError{
			Input:    input,
			Response: response,
			Err:      fmt.Errorf("%w: suggested_filename required for continue action", ErrMissingField),
		}
	}

	v.logger.Info("validation complete",
		"action", result.Action,
		"filename", result.SuggestedFilename)

	return &result, nil
}
