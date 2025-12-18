package models

import "strings"

// BotType represents the type of dependency update bot
type BotType string

const (
	BotRenovate      BotType = "renovate"
	BotDependabot    BotType = "dependabot"
	BotGitHubActions BotType = "github-actions"
)

// BotLogins maps bot types to their GitHub login patterns
var BotLogins = map[BotType][]string{
	BotRenovate:      {"app/renovate", "renovate[bot]", "renovate"},
	BotDependabot:    {"app/dependabot", "dependabot[bot]", "dependabot"},
	BotGitHubActions: {"app/github-actions", "github-actions[bot]", "github-actions"},
}

// DetectBot detects the bot type from the author login
// Returns the bot type and true if detected, empty string and false otherwise
func DetectBot(author string) (BotType, bool) {
	author = strings.ToLower(author)

	for botType, logins := range BotLogins {
		for _, login := range logins {
			if strings.Contains(author, strings.ToLower(login)) {
				return botType, true
			}
		}
	}

	return "", false
}

// DisplayName returns a clean display name for the bot (removes [bot] suffix)
func (b BotType) DisplayName() string {
	return string(b)
}

// RebaseCommand returns the comment command to trigger a rebase for dependabot
// Returns empty string if the bot doesn't support comment-based rebase
func (b BotType) RebaseCommand() string {
	switch b {
	case BotDependabot:
		return "@dependabot rebase"
	default:
		return ""
	}
}

// SupportsRebase returns true if the bot supports rebase functionality
func (b BotType) SupportsRebase() bool {
	return b == BotDependabot || b == BotRenovate
}

// UsesCheckboxRebase returns true if the bot uses checkbox-based rebase (Renovate)
func (b BotType) UsesCheckboxRebase() bool {
	return b == BotRenovate
}
