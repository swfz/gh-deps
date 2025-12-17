package app

import (
	"context"
	"fmt"

	"github.com/swfz/gh-deps/internal/api"
	"github.com/swfz/gh-deps/internal/formatter"
	"github.com/swfz/gh-deps/internal/models"
)

// App encapsulates the application logic
type App struct {
	client *api.Client
	config *Config
}

// New creates a new application instance
func New(config *Config) (*App, error) {
	client, err := api.NewClient(config.Verbose)
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
			fmt.Printf("Fetching dependency PRs from organization: %s\n", a.config.Target)
		}
		prs, err = a.client.FetchOrgPullRequests(ctx, a.config.Target)
	} else {
		if a.config.Verbose {
			fmt.Printf("Fetching dependency PRs from user: %s\n", a.config.Target)
		}
		prs, err = a.client.FetchUserPullRequests(ctx, a.config.Target)
	}

	if err != nil {
		return fmt.Errorf("failed to fetch pull requests: %w", err)
	}

	// Handle empty results
	if len(prs) == 0 {
		fmt.Println("No dependency update PRs found.")
		return nil
	}

	// Render table
	formatter.RenderTable(prs)

	// Print summary
	fmt.Printf("\nTotal: %d dependency update PRs\n", len(prs))

	return nil
}
