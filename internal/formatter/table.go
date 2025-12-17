package formatter

import (
	"os"
	"sort"

	"github.com/olekukonko/tablewriter"
	"github.com/swfz/gh-deps/internal/models"
)

// RenderTable displays pull requests in a formatted table
// PRs are sorted by repository name (alphabetical)
func RenderTable(prs []models.PullRequest) {
	// Sort by repository name
	sort.Slice(prs, func(i, j int) bool {
		return prs[i].RepoName() < prs[j].RepoName()
	})

	table := tablewriter.NewWriter(os.Stdout)

	// Set header using new API
	table.Header("REPO", "BOT", "STATUS", "DATE", "VERSION", "TITLE", "URL")

	// Add rows
	for _, pr := range prs {
		row := []interface{}{
			TruncateString(pr.RepoName(), 20),
			pr.BotType.DisplayName(),
			string(pr.CheckSummary.Status),
			pr.FormattedDate(),
			pr.Version,
			TruncateWithEllipsis(pr.Title, 60),
			pr.URL,
		}
		table.Append(row...)
	}

	table.Render()
}
