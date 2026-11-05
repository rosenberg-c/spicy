package validator

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestValidator_Validate(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		agentResponse string
		want          *ValidationResponse
		wantErr       bool
	}{
		{
			name:  "continue action with filename",
			input: "how to use ffmpeg",
			agentResponse: `{
                "action": "continue",
                "reason": "Specific enough",
                "suggested_filename": "ffmpeg-guide.md"
            }`,
			want: &ValidationResponse{
				Action:            ActionContinue,
				Reason:            "Specific enough",
				SuggestedFilename: "ffmpeg-guide.md",
			},
			wantErr: false,
		},
		{
			name:  "exit action with suggestions",
			input: "ffmpeg",
			agentResponse: `{
                "action": "exit",
                "reason": "Too vague",
                "suggestions": ["be more specific"]
            }`,
			want: &ValidationResponse{
				Action:      ActionExit,
				Reason:      "Too vague",
				Suggestions: []string{"be more specific"},
			},
			wantErr: false,
		},
		{
			name:  "exit with multiple suggestions",
			input: "python",
			agentResponse: `{
                "action": "exit",
                "reason": "Too broad",
                "suggestions": [
                    "specify which aspect",
                    "mention use case",
                    "provide context"
                ]
            }`,
			want: &ValidationResponse{
				Action: ActionExit,
				Reason: "Too broad",
				Suggestions: []string{
					"specify which aspect",
					"mention use case",
					"provide context",
				},
			},
			wantErr: false,
		},
		{
			name:  "continue with kebab-case filename",
			input: "docker compose tutorial",
			agentResponse: `{
                "action": "continue",
                "reason": "Clear and specific",
                "suggested_filename": "docker-compose-tutorial.md"
            }`,
			want: &ValidationResponse{
				Action:            ActionContinue,
				Reason:            "Clear and specific",
				SuggestedFilename: "docker-compose-tutorial.md",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAgent := &mockAgentRunner{response: tt.agentResponse}
			v := New(mockAgent, "test-model")

			got, err := v.Validate(context.Background(), tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

// @req CLI-TUTOR-002, CORE-CLI-004
func TestValidator_Validate_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		agentResponse string
		agentError    error
		wantErr       bool
		wantErrType   error
		errContains   string
	}{
		{
			name:          "invalid json",
			input:         "test",
			agentResponse: "not json at all",
			wantErr:       true,
			wantErrType:   ErrInvalidJSON,
		},
		{
			name:          "malformed json",
			input:         "test",
			agentResponse: `{"action": "continue", "reason": }`,
			wantErr:       true,
			wantErrType:   ErrInvalidJSON,
		},
		{
			name:          "missing action field",
			input:         "test",
			agentResponse: `{"reason": "test"}`,
			wantErr:       true,
			wantErrType:   ErrMissingField,
		},
		{
			name:          "missing reason field",
			input:         "test",
			agentResponse: `{"action": "continue"}`,
			wantErr:       true,
			wantErrType:   ErrMissingField,
		},
		{
			name:          "invalid action value",
			input:         "test",
			agentResponse: `{"action": "invalid", "reason": "test"}`,
			wantErr:       true,
			wantErrType:   ErrInvalidAction,
		},
		{
			name:          "empty action value",
			input:         "test",
			agentResponse: `{"action": "", "reason": "test"}`,
			wantErr:       true,
			wantErrType:   ErrMissingField,
		},
		{
			name:        "agent error",
			input:       "test",
			agentError:  fmt.Errorf("connection failed"),
			wantErr:     true,
			errContains: "agent call failed",
		},
		{
			name:  "continue without filename",
			input: "test",
			agentResponse: `{
                "action": "continue",
                "reason": "test"
            }`,
			wantErr:     true,
			wantErrType: ErrMissingField,
			errContains: "suggested_filename",
		},
		{
			name:  "exit without suggestions is valid",
			input: "test",
			agentResponse: `{
                "action": "exit",
                "reason": "test"
            }`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAgent := &mockAgentRunner{
				response: tt.agentResponse,
				err:      tt.agentError,
			}
			v := New(mockAgent, "test-model")

			_, err := v.Validate(context.Background(), tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				return
			}

			if tt.wantErrType != nil {
				var validationErr *ValidationError
				if errors.As(err, &validationErr) {
					if !errors.Is(validationErr.Err, tt.wantErrType) {
						t.Errorf("error type = %v, want %v",
							validationErr.Err, tt.wantErrType)
					}
				} else if !errors.Is(err, tt.wantErrType) {
					t.Errorf("error type = %v, want %v", err, tt.wantErrType)
				}
			}

			if tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want containing %q",
						err, tt.errContains)
				}
			}
		})
	}
}

func TestValidator_ContextCancellation(t *testing.T) {
	mockAgent := &mockAgentRunner{
		response: `{"action": "continue", "reason": "test",
                     "suggested_filename": "test.md"}`,
	}
	v := New(mockAgent, "test-model")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := v.Validate(ctx, "test input")
	if err == nil {
		t.Error("expected error from cancelled context, got nil")
	}
}

func TestValidator_ContextTimeout(t *testing.T) {
	mockAgent := &mockAgentRunner{
		response: `{"action": "continue", "reason": "test",
                     "suggested_filename": "test.md"}`,
	}
	v := New(mockAgent, "test-model")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(1 * time.Millisecond)

	_, err := v.Validate(ctx, "test input")
	if err == nil {
		t.Error("expected error from timeout context, got nil")
	}
}

func TestAction_IsValid(t *testing.T) {
	tests := []struct {
		action Action
		want   bool
	}{
		{ActionContinue, true},
		{ActionExit, true},
		{Action("invalid"), false},
		{Action(""), false},
		{Action("CONTINUE"), false},
		{Action("Exit"), false},
		{Action("continue "), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.action), func(t *testing.T) {
			if got := tt.action.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAction_String(t *testing.T) {
	tests := []struct {
		action Action
		want   string
	}{
		{ActionContinue, "continue"},
		{ActionExit, "exit"},
		{Action("custom"), "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.action.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Input:    "test input",
		Response: "bad response",
		Err:      ErrInvalidJSON,
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "test input") {
		t.Errorf("error message missing input: %s", errMsg)
	}
	if !strings.Contains(errMsg, "invalid JSON") {
		t.Errorf("error message missing wrapped error: %s", errMsg)
	}
}

func TestValidationError_Unwrap(t *testing.T) {
	wrappedErr := fmt.Errorf("wrapped error")
	err := &ValidationError{
		Input:    "test",
		Response: "response",
		Err:      wrappedErr,
	}

	if unwrapped := err.Unwrap(); unwrapped != wrappedErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, wrappedErr)
	}
}

func TestValidationResponse_NilSuggestions(t *testing.T) {
	mockAgent := &mockAgentRunner{
		response: `{
            "action": "exit",
            "reason": "test",
            "suggestions": null
        }`,
	}
	v := New(mockAgent, "test-model")

	got, err := v.Validate(context.Background(), "test")
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if got.Suggestions != nil {
		t.Errorf("Suggestions = %v, want nil", got.Suggestions)
	}
}

func TestValidationResponse_EmptySuggestions(t *testing.T) {
	mockAgent := &mockAgentRunner{
		response: `{
            "action": "exit",
            "reason": "test",
            "suggestions": []
        }`,
	}
	v := New(mockAgent, "test-model")

	got, err := v.Validate(context.Background(), "test")
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if got.Suggestions == nil {
		t.Error("Suggestions should not be nil for empty array")
	}
	if len(got.Suggestions) != 0 {
		t.Errorf("len(Suggestions) = %d, want 0", len(got.Suggestions))
	}
}

// Mock implementation
type mockAgentRunner struct {
	response string
	err      error
}

func (m *mockAgentRunner) Run(ctx context.Context, model, prompt string) (string, error) {
	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}
