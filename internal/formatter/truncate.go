package formatter

import "unicode/utf8"

// TruncateString truncates a string to maxLen characters
// Handles UTF-8 properly by counting runes, not bytes
func TruncateString(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}

	runes := []rune(s)
	return string(runes[:maxLen])
}

// TruncateWithEllipsis truncates a string and adds "..." if truncated
func TruncateWithEllipsis(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}

	runes := []rune(s)
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}

	return string(runes[:maxLen-3]) + "..."
}
