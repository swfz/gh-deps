package models

import (
	"strings"
	"time"
)

// MergeableState represents the mergeable state of a PR
type MergeableState string

const (
	MergeableStateMergeable   MergeableState = "MERGEABLE"   // PR can be merged
	MergeableStateConflicting MergeableState = "CONFLICTING" // PR has conflicts
	MergeableStateUnknown     MergeableState = "UNKNOWN"     // State is being calculated
)

// PullRequest represents a dependency update pull request
type PullRequest struct {
	Repository     string         // Full repository name (owner/repo)
	Number         int            // PR number
	Title          string         // PR title
	Body           string         // PR description body
	Author         string         // Author login
	CreatedAt      time.Time      // Creation timestamp
	URL            string         // PR URL
	HeadSHA        string         // Head commit SHA
	BotType        BotType        // Detected bot type
	CheckSummary   CheckSummary   // Aggregated check status
	Version        string         // Extracted version info (e.g., "1.0.0 -> 1.1.0")
	MergeableState MergeableState // Mergeable state (MERGEABLE, CONFLICTING, UNKNOWN)
}

// FormattedDate returns the creation date in YYYY-MM-DD format
func (pr *PullRequest) FormattedDate() string {
	return pr.CreatedAt.Format("2006-01-02")
}

// RepoName extracts just the repository name from the full name (owner/repo)
func (pr *PullRequest) RepoName() string {
	parts := strings.Split(pr.Repository, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return pr.Repository
}
