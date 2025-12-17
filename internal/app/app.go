package app

import (
	"context"
	"fmt"

	"github.com/swfz/gh-deps/internal/api"
	"github.com/swfz/gh-deps/internal/formatter"
	"github.com/swfz/gh-deps/internal/interactive"
	"github.com/swfz/gh-deps/internal/models"
)

// App encapsulates the application logic
type App struct {
	client *api.Client
	config *Config
}

// New creates a new application instance
func New(config *Config) (*App, error) {
	client, err := api.NewClient(config.Verbose, config.SkipChecks)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	return &App{
		client: client,
		config: config,
	}, nil
}

// Run executes the main application logic
func (a *App) Run(ctx context.Context) error {
	var prs []models.PullRequest
	var err error

	// Fetch PRs based on target type (org vs user)
	if a.config.IsOrganization {
		if a.config.Verbose {
			limitMsg := "all PRs"
			if a.config.Limit > 0 {
				limitMsg = fmt.Sprintf("up to %d PRs", a.config.Limit)
			}
			fmt.Printf("Fetching dependency PRs from organization: %s (%s)\n",
				a.config.Target, limitMsg)
		}
		prs, err = a.client.FetchOrgPullRequests(ctx, a.config.Target, a.config.Limit)
	} else {
		if a.config.Verbose {
			limitMsg := "all PRs"
			if a.config.Limit > 0 {
				limitMsg = fmt.Sprintf("up to %d PRs", a.config.Limit)
			}
			fmt.Printf("Fetching dependency PRs from user: %s (%s)\n",
				a.config.Target, limitMsg)
		}
		prs, err = a.client.FetchUserPullRequests(ctx, a.config.Target, a.config.Limit)
	}

	if err != nil {
		return fmt.Errorf("failed to fetch pull requests: %w", err)
	}

	// Handle empty results
	if len(prs) == 0 {
		fmt.Println("No dependency update PRs found.")
		return nil
	}

	// Render table (with row numbers if interactive mode)
	sortedPRs := formatter.RenderTable(prs, a.config.Interactive)

	// Print summary with indicators
	fmt.Printf("\nTotal: %d dependency update PRs", len(prs))
	if a.config.Limit > 0 && len(prs) >= a.config.Limit {
		fmt.Printf(" (limited to %d PRs)", a.config.Limit)
	}
	if a.config.SkipChecks {
		fmt.Printf(" [check runs skipped]")
	}
	fmt.Println()

	// Enter interactive mode if flag is set
	if a.config.Interactive {
		if err := interactive.RunTUI(ctx, sortedPRs, a.client, a.config.Verbose); err != nil {
			return fmt.Errorf("interactive mode failed: %w", err)
		}
	}

	return nil
}
