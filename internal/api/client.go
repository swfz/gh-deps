package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/shurcooL/graphql"
	"golang.org/x/time/rate"

	"github.com/swfz/gh-deps/internal/models"
	"github.com/swfz/gh-deps/internal/parser"
)

// Client wraps GitHub API client with rate limiting
type Client struct {
	graphqlClient *graphql.Client
	httpClient    *http.Client
	rateLimiter   *rate.Limiter
	verbose       bool
	skipChecks    bool
}

// NewClient creates a new GitHub API client using gh CLI authentication
func NewClient(verbose bool, skipChecks bool) (*Client, error) {
	// Use gh CLI's authentication
	httpClient, err := api.DefaultHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Create GraphQL client manually for more control
	graphqlClient := graphql.NewClient("https://api.github.com/graphql", httpClient)

	// Rate limiter: 5000 requests per hour = ~1.4 per second, use 1 per second to be safe
	rateLimiter := rate.NewLimiter(rate.Every(time.Second), 10)

	return &Client{
		graphqlClient: graphqlClient,
		httpClient:    httpClient,
		rateLimiter:   rateLimiter,
		verbose:       verbose,
		skipChecks:    skipChecks,
	}, nil
}

// FetchOrgPullRequests fetches all dependency update PRs from an organization
func (c *Client) FetchOrgPullRequests(ctx context.Context, orgName string, limit int) ([]models.PullRequest, error) {
	var allPRs []models.PullRequest
	var cursor *string

	for {
		// Wait for rate limiter
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter error: %w", err)
		}

		var query OrgRepositoriesQuery

		variables := map[string]interface{}{
			"orgName": graphql.String(orgName),
			"cursor":  (*graphql.String)(cursor),
		}

		if err := c.graphqlClient.Query(ctx, &query, variables); err != nil {
			return nil, fmt.Errorf("GraphQL query failed: %w", err)
		}

		// Process repositories and PRs
		for _, repo := range query.Organization.Repositories.Nodes {
			prs := c.processPRsFromRepo(ctx, repo, c.verbose)
			allPRs = append(allPRs, prs...)

			// Check PR limit after adding PRs from this repo
			if limit > 0 && len(allPRs) >= limit {
				if c.verbose {
					fmt.Fprintf(os.Stderr, "[DEBUG] Reached PR limit (%d), stopping\n", limit)
				}
				return allPRs[:limit], nil
			}
		}

		// Check if there are more pages
		if !query.Organization.Repositories.PageInfo.HasNextPage {
			break
		}
		cursor = &query.Organization.Repositories.PageInfo.EndCursor
	}

	return allPRs, nil
}

// FetchUserPullRequests fetches all dependency update PRs from a user's repositories
func (c *Client) FetchUserPullRequests(ctx context.Context, userName string, limit int) ([]models.PullRequest, error) {
	var allPRs []models.PullRequest
	var cursor *string

	for {
		// Wait for rate limiter
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter error: %w", err)
		}

		var query UserRepositoriesQuery

		variables := map[string]interface{}{
			"userName": graphql.String(userName),
			"cursor":   (*graphql.String)(cursor),
		}

		if err := c.graphqlClient.Query(ctx, &query, variables); err != nil {
			return nil, fmt.Errorf("GraphQL query failed: %w", err)
		}

		// Process repositories and PRs
		for _, repo := range query.User.Repositories.Nodes {
			prs := c.processPRsFromRepo(ctx, repo, c.verbose)
			allPRs = append(allPRs, prs...)

			// Check PR limit after adding PRs from this repo
			if limit > 0 && len(allPRs) >= limit {
				if c.verbose {
					fmt.Fprintf(os.Stderr, "[DEBUG] Reached PR limit (%d), stopping\n", limit)
				}
				return allPRs[:limit], nil
			}
		}

		// Check if there are more pages
		if !query.User.Repositories.PageInfo.HasNextPage {
			break
		}
		cursor = &query.User.Repositories.PageInfo.EndCursor
	}

	return allPRs, nil
}

// processPRsFromRepo extracts and filters PRs from a repository
func (c *Client) processPRsFromRepo(ctx context.Context, repo RepositoryNode, verbose bool) []models.PullRequest {
	var prs []models.PullRequest

	for _, pr := range repo.PullRequests.Nodes {
		if verbose {
			fmt.Fprintf(os.Stderr, "[DEBUG] Repo: %s, PR #%d, Author: %s\n",
				repo.NameWithOwner, pr.Number, pr.Author.Login)
		}

		// Detect if this is a bot PR
		botType, isBot := models.DetectBot(pr.Author.Login)
		if !isBot {
			if verbose {
				fmt.Fprintf(os.Stderr, "[DEBUG] PR #%d not a bot (author: %s)\n",
					pr.Number, pr.Author.Login)
			}
			continue
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "[DEBUG] PR #%d is %s bot\n", pr.Number, botType)
		}

		// Get check status from statusCheckRollup (efficient - no extra API call)
		var checkSummary models.CheckSummary
		if !c.skipChecks && len(pr.Commits.Nodes) > 0 && pr.Commits.Nodes[0].Commit.StatusCheckRollup != nil {
			state := pr.Commits.Nodes[0].Commit.StatusCheckRollup.State
			checkSummary = models.StatusCheckRollupToSummary(state)
			if verbose {
				fmt.Fprintf(os.Stderr, "[DEBUG] PR #%d status: %s\n", pr.Number, state)
			}
		} else {
			// No status check rollup available (or checks skipped)
			checkSummary = models.CheckSummary{Status: models.StatusNone, Total: 0}
			if verbose && !c.skipChecks {
				fmt.Fprintf(os.Stderr, "[DEBUG] PR #%d has no status check rollup\n", pr.Number)
			}
		}

		// Extract label names
		var labels []string
		for _, label := range pr.Labels.Nodes {
			labels = append(labels, label.Name)
		}

		// Create PR model
		prs = append(prs, models.PullRequest{
			Repository:     repo.NameWithOwner,
			Number:         pr.Number,
			Title:          pr.Title,
			Body:           pr.Body,
			Author:         pr.Author.Login,
			CreatedAt:      pr.CreatedAt,
			URL:            pr.URL,
			HeadSHA:        pr.HeadRefOid,
			BotType:        botType,
			CheckSummary:   checkSummary,
			Version:        parser.ExtractVersion(pr.Body, botType),
			MergeableState: models.MergeableState(pr.Mergeable),
			Labels:         labels,
		})
	}

	return prs
}

// Note: fetchCheckRuns function removed - now using statusCheckRollup from GraphQL
// which is much more efficient (no extra API calls per PR)
