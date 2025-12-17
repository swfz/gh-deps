package api

import "time"

// GraphQL query structures for fetching organization/user repositories with PRs

// RepositoryNode represents a repository with its pull requests
type RepositoryNode struct {
	NameWithOwner string
	PullRequests  struct {
		Nodes []PullRequestNode
	} `graphql:"pullRequests(first: 100, states: OPEN)"`
}

// PullRequestNode represents a pull request with its metadata
type PullRequestNode struct {
	Number     int
	Title      string
	Body       string
	CreatedAt  time.Time
	URL        string
	HeadRefOid string
	Mergeable  string // MERGEABLE, CONFLICTING, UNKNOWN
	Author     struct {
		Login string
	}
	Commits struct {
		Nodes []struct {
			Commit struct {
				StatusCheckRollup *struct {
					State string // SUCCESS, FAILURE, PENDING, ERROR, or null
				}
			}
		}
	} `graphql:"commits(last: 1)"`
	Labels struct {
		Nodes []struct {
			Name string
		}
	} `graphql:"labels(first: 10)"`
}

// Note: CheckRunsQuery, CheckSuiteNode, and CheckRunNode removed
// Now using statusCheckRollup in PullRequestNode for better efficiency

// OrgRepositoriesQuery represents the GraphQL query for organization repositories
type OrgRepositoriesQuery struct {
	Organization struct {
		Repositories struct {
			PageInfo struct {
				HasNextPage bool
				EndCursor   string
			}
			Nodes []RepositoryNode
		} `graphql:"repositories(first: 50, after: $cursor)"`
	} `graphql:"organization(login: $orgName)"`
}

// UserRepositoriesQuery represents the GraphQL query for user repositories
type UserRepositoriesQuery struct {
	User struct {
		Repositories struct {
			PageInfo struct {
				HasNextPage bool
				EndCursor   string
			}
			Nodes []RepositoryNode
		} `graphql:"repositories(first: 50, after: $cursor)"`
	} `graphql:"user(login: $userName)"`
}
