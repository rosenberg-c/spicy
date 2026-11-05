package validator

import "fmt"

// BuildValidationPrompt creates a prompt that asks the AI to validate input specificity.
// Returns a prompt instructing the AI to respond with JSON containing action and reason.
func BuildValidationPrompt(input string) string {
	return fmt.Sprintf(`You are a senior technical writer and educator.
Determine if this request is specific and clear enough to create a useful tutorial, 
or if it's too ambiguous and needs clarification.

Respond ONLY with a valid JSON object (no markdown, no extra text) in this exact format:

If the request is specific enough:
{
  "action": "continue",
  "reason": "brief explanation of your decision",
  "suggested_filename": "short-descriptive-name.md"
}

If the request is too ambiguous:
{
  "action": "exit",
  "reason": "brief explanation of your decision",
  "suggestions": ["clarifying question 1", "clarifying question 2"]
}

Guidelines for suggested_filename:
- Use lowercase with hyphens (kebab-case)
- Keep it short but descriptive (2-5 words max)
- Must end with .md
- Examples: "ffmpeg-video-conversion.md", "pandas-csv-guide.md", "echo-command-basics.md"

Examples of decisions:
- "how does ffmpeg work" -> too broad, suggest: "how to convert video formats", "how to extract audio", etc.
- "how to convert mp4 to webm using ffmpeg" -> specific enough, continue, suggest: "ffmpeg-mp4-to-webm.md"
- "explain python" -> too vague, suggest: "which aspect of Python", "what's your experience level", etc.
- "how to read a CSV file in Python using pandas" -> specific enough, continue, suggest: "pandas-csv-reading.md"
- "docker" -> too vague, suggest: "docker basics", "docker compose", "dockerfile best practices", etc.
- "echo command" -> specific enough, continue, suggest: "echo-command-guide.md"

Think carefully about whether the request has enough context and specificity to create a useful, focused tutorial.

Analyze the following user request for a tutorial:
"%s"`, input)
}
