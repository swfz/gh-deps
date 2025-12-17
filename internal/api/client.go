package api

import (
	"context"
	"fmt"
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
	rateLimiter   *rate.Limiter
	verbose       bool
}

// NewClient creates a new GitHub API client using gh CLI authentication
func NewClient(verbose bool) (*Client, error) {
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
		rateLimiter:   rateLimiter,
		verbose:       verbose,
	}, nil
}

// FetchOrgPullRequests fetches all dependency update PRs from an organization
func (c *Client) FetchOrgPullRequests(ctx context.Context, orgName string) ([]models.PullRequest, error) {
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
func (c *Client) FetchUserPullRequests(ctx context.Context, userName string) ([]models.PullRequest, error) {
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

		// Fetch check runs for this PR's head commit
		checkRuns := c.fetchCheckRuns(ctx, repo.NameWithOwner, pr.HeadRefOid)

		// Create PR model
		prs = append(prs, models.PullRequest{
			Repository:   repo.NameWithOwner,
			Number:       pr.Number,
			Title:        pr.Title,
			Body:         pr.Body,
			Author:       pr.Author.Login,
			CreatedAt:    pr.CreatedAt,
			URL:          pr.URL,
			HeadSHA:      pr.HeadRefOid,
			BotType:      botType,
			CheckSummary: models.AggregateCheckStatus(checkRuns),
			Version:      parser.ExtractVersion(pr.Body, botType),
		})
	}

	return prs
}

// fetchCheckRuns fetches check runs for a specific commit
// Returns empty slice if fetch fails (to avoid blocking the entire operation)
func (c *Client) fetchCheckRuns(ctx context.Context, repoFullName string, commitSHA string) []models.CheckRun {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return []models.CheckRun{}
	}

	// Parse owner and repo name
	parts := parseRepoFullName(repoFullName)
	if len(parts) != 2 {
		return []models.CheckRun{}
	}
	owner, name := parts[0], parts[1]

	var query CheckRunsQuery
	variables := map[string]interface{}{
		"owner":     graphql.String(owner),
		"name":      graphql.String(name),
		"commitSHA": graphql.String(commitSHA),
	}

	if err := c.graphqlClient.Query(ctx, &query, variables); err != nil {
		// Silently fail for individual PRs to avoid blocking entire operation
		return []models.CheckRun{}
	}

	// Extract check runs from response
	var checkRuns []models.CheckRun
	for _, suite := range query.Repository.Object.Commit.CheckSuites.Nodes {
		for _, run := range suite.CheckRuns.Nodes {
			checkRuns = append(checkRuns, models.CheckRun{
				Name:       run.Name,
				Status:     run.Status,
				Conclusion: run.Conclusion,
			})
		}
	}

	return checkRuns
}

// parseRepoFullName splits "owner/repo" into [owner, repo]
func parseRepoFullName(fullName string) []string {
	for i, ch := range fullName {
		if ch == '/' {
			return []string{fullName[:i], fullName[i+1:]}
		}
	}
	return []string{fullName}
}
