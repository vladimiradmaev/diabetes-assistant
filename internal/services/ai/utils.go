package ai

import (
	"fmt"
)

// extractJSONFromText attempts to extract JSON content from a text block
// This is helpful when the AI response contains explanatory text along with JSON
func extractJSONFromText(text string) (string, error) {
	// Look for opening and closing braces
	start := -1
	end := -1
	braceCount := 0
	inString := false
	escape := false

	for i, char := range text {
		// Handle string detection
		if char == '"' && !escape {
			inString = !inString
		}
		escape = (char == '\\' && !escape)

		// Only count braces outside of strings
		if !inString {
			if char == '{' {
				if braceCount == 0 {
					start = i
				}
				braceCount++
			} else if char == '}' {
				braceCount--
				if braceCount == 0 {
					end = i
					break
				}
			}
		}
	}

	if start >= 0 && end > start {
		return text[start : end+1], nil
	}

	return "", fmt.Errorf("no JSON object found in text")
}
