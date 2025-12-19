package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// CommentRequest represents the request body for creating a PR comment
type CommentRequest struct {
	Body string `json:"body"`
}

// CommentResponse represents the response from GitHub's comment API
type CommentResponse struct {
	ID        int    `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	URL       string `json:"html_url"`
}

// CreateComment posts a comment on a PR
func (c *Client) CreateComment(ctx context.Context, owner, repo string, prNumber int, body string) (*CommentResponse, error) {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments", owner, repo, prNumber)

	reqBody := CommentRequest{
		Body: body,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if c.verbose {
		fmt.Fprintf(os.Stderr, "[DEBUG] Creating comment on PR %s/%s#%d: %s\n", owner, repo, prNumber, body)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("comment creation failed (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var commentResp CommentResponse
	if err := json.Unmarshal(respBody, &commentResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &commentResp, nil
}
