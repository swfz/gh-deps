package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/swfz/gh-deps/internal/models"
)

var (
	// Dependabot format: "from 1.0.0 to 1.1.0"
	dependabotRegex = regexp.MustCompile(`from\s+([^\s]+)\s+to\s+([^\s]+)`)

	// Renovate format: "1.0.0 -> 1.1.0" or "`1.0.0` -> `1.1.0`"
	renovateRegex = regexp.MustCompile("`?([^`\\s]+)`?\\s+->\\s+`?([^`\\s]+)`?")
)

// ExtractVersion extracts version information from PR body based on bot type
// Returns formatted string "X -> Y" or "-" if not found
func ExtractVersion(body string, botType models.BotType) string {
	var regex *regexp.Regexp

	switch botType {
	case models.BotDependabot:
		regex = dependabotRegex
	case models.BotRenovate:
		regex = renovateRegex
	case models.BotGitHubActions:
		// GitHub Actions typically uses similar format to Renovate
		regex = renovateRegex
	default:
		return "-"
	}

	matches := regex.FindStringSubmatch(body)
	if len(matches) >= 3 {
		from := strings.TrimSpace(matches[1])
		to := strings.TrimSpace(matches[2])
		return fmt.Sprintf("%s -> %s", from, to)
	}

	return "-"
}
