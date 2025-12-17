package api

import "time"

// GitObjectID represents a Git object ID scalar type in GitHub's GraphQL API
type GitObjectID string

// GraphQL query structures for fetching organization/user repositories with PRs and check runs

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
	Author     struct {
		Login string
	}
}

// CheckRunsQuery represents the GraphQL query for check runs on a specific commit
type CheckRunsQuery struct {
	Repository struct {
		Object struct {
			Commit struct {
				CheckSuites struct {
					Nodes []CheckSuiteNode
				} `graphql:"checkSuites(first: 10)"`
			} `graphql:"... on Commit"`
		} `graphql:"object(oid: $commitSHA)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

// CheckSuiteNode represents a check suite with its check runs
type CheckSuiteNode struct {
	CheckRuns struct {
		Nodes []CheckRunNode
	} `graphql:"checkRuns(first: 20)"`
}

// CheckRunNode represents a single check run
type CheckRunNode struct {
	Name       string
	Status     string
	Conclusion string
}

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
