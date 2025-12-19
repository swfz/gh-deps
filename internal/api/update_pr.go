package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
)

// UpdatePRRequest represents the request body for updating a PR
type UpdatePRRequest struct {
	Body string `json:"body,omitempty"`
}

// UpdatePRResponse represents the response from GitHub's PR update API
type UpdatePRResponse struct {
	Number int    `json:"number"`
	Body   string `json:"body"`
}

// UpdatePullRequestBody updates the body of a PR
func (c *Client) UpdatePullRequestBody(ctx context.Context, owner, repo string, prNumber int, body string) (*UpdatePRResponse, error) {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, prNumber)

	reqBody := UpdatePRRequest{
		Body: body,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if c.verbose {
		fmt.Fprintf(os.Stderr, "[DEBUG] Updating PR body for %s/%s#%d\n", owner, repo, prNumber)
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
		return nil, fmt.Errorf("PR update failed (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var updateResp UpdatePRResponse
	if err := json.Unmarshal(respBody, &updateResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &updateResp, nil
}

// CheckRenovateRebaseCheckbox checks the rebase checkbox in Renovate PR body
// Returns the updated body with the checkbox checked, or error if checkbox not found
func CheckRenovateRebaseCheckbox(body string) (string, error) {
	// Renovate rebase checkbox patterns:
	// - [ ] <!-- rebase-check -->If you want to rebase/retry this PR, check this box
	// or similar variations

	// Pattern to match unchecked checkbox with "rebase" keyword
	pattern := regexp.MustCompile(`(?m)^(\s*-\s*\[)\s(\]\s*(?:<!--[^>]*-->)?\s*.*?rebase.*?)$`)

	if !pattern.MatchString(body) {
		return "", fmt.Errorf("rebase checkbox not found in PR body")
	}

	// Replace [ ] with [x]
	updatedBody := pattern.ReplaceAllString(body, "${1}x${2}")

	return updatedBody, nil
}

// TriggerRenovateRebase updates the Renovate PR body to check the rebase checkbox
func (c *Client) TriggerRenovateRebase(ctx context.Context, owner, repo string, prNumber int, currentBody string) error {
	updatedBody, err := CheckRenovateRebaseCheckbox(currentBody)
	if err != nil {
		return fmt.Errorf("failed to process rebase checkbox: %w", err)
	}

	_, err = c.UpdatePullRequestBody(ctx, owner, repo, prNumber, updatedBody)
	if err != nil {
		return fmt.Errorf("failed to update PR body: %w", err)
	}

	return nil
}
