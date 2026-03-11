package validator

import (
	"context"
	"reflect"
	"testing"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAgent := &mockAgentRunner{response: tt.agentResponse}
			v := New(mockAgent)

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

// Mock implementation
type mockAgentRunner struct {
	response string
	err      error
}

func (m *mockAgentRunner) Run(ctx context.Context, model, prompt string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}
