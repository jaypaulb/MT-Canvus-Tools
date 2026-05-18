package atom

import "strings"

// MaskKey masks an API key for logging by replacing all but the last 4
// characters with asterisks.
func MaskKey(key string) string {
	if len(key) <= 4 {
		return key
	}
	return strings.Repeat("*", len(key)-4) + key[len(key)-4:]
}

// StripMarkdownCodeBlock removes a leading/trailing markdown code fence
// (```json … ``` or ``` … ```) so the inner JSON can be parsed.
func StripMarkdownCodeBlock(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(text, "```")
		text = strings.TrimSpace(text)
	}
	return text
}
