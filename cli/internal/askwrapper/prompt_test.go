package askwrapper

import (
	"strings"
	"testing"
)

func TestBuildFollowUpPrompt_IncludesContextAndQuestion(t *testing.T) {
	// @req CLI-ASKWRAPPER-006
	prompt := BuildFollowUpPrompt(" previous q ", " previous a ", " follow up ")

	checks := []string{
		"Previous question:\nprevious q",
		"Previous answer:\nprevious a",
		"Follow-up question:\nfollow up",
	}
	for _, want := range checks {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q\nfull prompt:\n%s", want, prompt)
		}
	}
}
