package api

import (
	"context"
	"time"
)

// RateLimitInfo contains information about GitHub API rate limits
type RateLimitInfo struct {
	Limit     int
	Remaining int
	ResetAt   time.Time
}

// CheckRateLimit queries the current rate limit status
// This can be used for debugging or displaying remaining quota to users
func (c *Client) CheckRateLimit(ctx context.Context) (*RateLimitInfo, error) {
	// GraphQL query for rate limit info
	var query struct {
		RateLimit struct {
			Limit     int
			Remaining int
			ResetAt   time.Time
		}
	}

	if err := c.graphqlClient.Query(ctx, &query, nil); err != nil {
		return nil, err
	}

	return &RateLimitInfo{
		Limit:     query.RateLimit.Limit,
		Remaining: query.RateLimit.Remaining,
		ResetAt:   query.RateLimit.ResetAt,
	}, nil
}
