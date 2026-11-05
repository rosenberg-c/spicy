package filename

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// GenerateTimestamped creates a filename with format: [year-month-day_hour-minute]_[cmd]_[filename-suggestion].md
func GenerateTimestamped(cmd, input string) string {
	now := time.Now()
	timestamp := now.Format("2006-01-02_15-04")

	// Create filename suggestion from input (sanitize and truncate)
	suggestion := Sanitize(input)
	if len(suggestion) > 40 {
		suggestion = suggestion[:40]
	}

	return fmt.Sprintf("%s_%s_%s.md", timestamp, cmd, suggestion)
}

// Sanitize removes invalid characters from filename
func Sanitize(s string) string {
	// Remove extension if present
	if ext := filepath.Ext(s); ext != "" {
		s = strings.TrimSuffix(s, ext)
	}

	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")

	// Remove any character that's not alphanumeric, hyphen, or underscore
	reg := regexp.MustCompile(`[^a-z0-9\-_]`)
	s = reg.ReplaceAllString(s, "")

	// Remove consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	s = reg.ReplaceAllString(s, "-")

	// Trim hyphens from start and end
	s = strings.Trim(s, "-")

	if s == "" {
		s = "output"
	}

	return s
}
