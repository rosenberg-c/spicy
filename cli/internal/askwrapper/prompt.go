package askwrapper

import "strings"

func BuildFollowUpPrompt(contextQuestion, contextAnswer, followUpQuestion string) string {
	parts := []string{
		"Use this previous conversation context to answer the follow-up question.",
		"",
		"Previous question:",
		strings.TrimSpace(contextQuestion),
		"",
		"Previous answer:",
		strings.TrimSpace(contextAnswer),
		"",
		"Follow-up question:",
		strings.TrimSpace(followUpQuestion),
		"",
		"Respond directly to the follow-up question using the context above.",
	}
	return strings.Join(parts, "\n")
}
