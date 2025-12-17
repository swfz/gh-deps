package formatter

import (
	"fmt"
	"os"
	"sort"

	"github.com/olekukonko/tablewriter"
	"github.com/swfz/gh-deps/internal/models"
)

// RenderTable displays pull requests in a formatted table
// PRs are sorted by repository name (alphabetical)
// Returns the sorted slice for consistent indexing when interactive mode is enabled
func RenderTable(prs []models.PullRequest, showRowNumbers bool) []models.PullRequest {
	// Sort by repository name
	sort.Slice(prs, func(i, j int) bool {
		return prs[i].RepoName() < prs[j].RepoName()
	})

	table := tablewriter.NewWriter(os.Stdout)

	// Set header - add # column if showing row numbers
	if showRowNumbers {
		table.Header("#", "REPO", "BOT", "CI", "MERGE", "LABELS", "DATE", "VERSION", "TITLE", "URL")
	} else {
		table.Header("REPO", "BOT", "CI", "MERGE", "LABELS", "DATE", "VERSION", "TITLE", "URL")
	}

	// Add rows
	for i, pr := range prs {
		var row []interface{}
		if showRowNumbers {
			row = []interface{}{
				fmt.Sprintf("%d", i+1), // 1-based row number
				TruncateString(pr.RepoName(), 20),
				pr.BotType.DisplayName(),
				string(pr.CheckSummary.Status),
				formatMergeableState(pr.MergeableState),
				formatLabels(pr.Labels),
				pr.FormattedDate(),
				pr.Version,
				TruncateWithEllipsis(pr.Title, 60),
				pr.URL,
			}
		} else {
			row = []interface{}{
				TruncateString(pr.RepoName(), 20),
				pr.BotType.DisplayName(),
				string(pr.CheckSummary.Status),
				formatMergeableState(pr.MergeableState),
				formatLabels(pr.Labels),
				pr.FormattedDate(),
				pr.Version,
				TruncateWithEllipsis(pr.Title, 60),
				pr.URL,
			}
		}
		table.Append(row...)
	}

	table.Render()
	return prs
}

// formatMergeableState returns a visual indicator for mergeable state
func formatMergeableState(state models.MergeableState) string {
	switch state {
	case models.MergeableStateMergeable:
		return "✓"
	case models.MergeableStateConflicting:
		return "✗"
	case models.MergeableStateUnknown:
		return "?"
	default:
		return "-"
	}
}

// formatLabels formats PR labels for display
func formatLabels(labels []string) string {
	if len(labels) == 0 {
		return "-"
	}
	// Join labels with comma and truncate if too long
	labelsStr := ""
	for i, label := range labels {
		if i > 0 {
			labelsStr += ","
		}
		labelsStr += label
	}
	return TruncateWithEllipsis(labelsStr, 30)
}
