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
	client, err := api.NewClient(config.Verbose, config.SkipChecks, config.ExcludeRepositories, config.Target, config.IsOrganization)
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

	// Fetch PRs based on mode
	if len(a.config.Repositories) > 0 {
		// Fetch PRs from specific repositories
		if a.config.Verbose {
			fmt.Printf("Fetching dependency PRs from specific repositories: %v\n", a.config.Repositories)
		}
		prs, err = a.fetchSpecificRepositories(ctx)
	} else if a.config.IsOrganization {
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
		if err := interactive.RunTUI(ctx, sortedPRs, a.client, a.config.Target, a.config.IsOrganization, a.config.Limit, a.config.Verbose); err != nil {
			return fmt.Errorf("interactive mode failed: %w", err)
		}
	}

	return nil
}

// fetchSpecificRepositories fetches PRs from the repositories specified by --repo.
// Note: archived repositories are not filtered here because explicitly specifying
// a repository via --repo is treated as an intentional user choice.
func (a *App) fetchSpecificRepositories(ctx context.Context) ([]models.PullRequest, error) {
	var allPRs []models.PullRequest

	for _, repo := range a.config.Repositories {
		owner, name, err := api.ParseRepository(repo)
		if err != nil {
			// Short format (reponame): use target as owner
			owner = a.config.Target
			name = repo
		}

		if a.config.Verbose {
			fmt.Printf("Fetching dependency PRs from repository: %s/%s\n", owner, name)
		}

		prs, err := a.client.FetchRepositoryPullRequests(ctx, owner, name)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch PRs from %s/%s: %w", owner, name, err)
		}
		allPRs = append(allPRs, prs...)

		if a.config.Limit > 0 && len(allPRs) >= a.config.Limit {
			return allPRs[:a.config.Limit], nil
		}
	}

	return allPRs, nil
}
