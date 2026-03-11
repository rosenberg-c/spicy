package generator

import "fmt"

// BuildTutorialPrompt creates a prompt for generating tutorial content.
func BuildTutorialPrompt(input string) string {
	return fmt.Sprintf(`You are a senior coder. Write a tutorial to answer the user question, as detailed as you can. The response must be valid markdown.

User input:
%s`, input)
}
