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
