package models

import "strings"

// CheckStatus represents the visual status indicator for CI checks
type CheckStatus string

const (
	StatusSuccess CheckStatus = "✅"
	StatusFailure CheckStatus = "❌"
	StatusPending CheckStatus = "⏳"
	StatusNone    CheckStatus = "-"
)

// CheckRun represents a single check run from GitHub
type CheckRun struct {
	Name       string
	Status     string // queued, in_progress, completed
	Conclusion string // success, failure, neutral, cancelled, skipped, timed_out, action_required
}

// CheckSummary aggregates check run results
type CheckSummary struct {
	Status CheckStatus
	Total  int
}

// AggregateCheckStatus analyzes all check runs and returns overall status
// Rules:
// - If no checks: StatusNone
// - If any check is not completed (queued/in_progress): StatusPending
// - If any completed check has failure conclusion: StatusFailure
// - If all checks are completed successfully: StatusSuccess
func AggregateCheckStatus(checks []CheckRun) CheckSummary {
	if len(checks) == 0 {
		return CheckSummary{Status: StatusNone, Total: 0}
	}

	hasFailure := false
	hasPending := false

	for _, check := range checks {
		// Check if any run is not completed (case-insensitive)
		if !strings.EqualFold(check.Status, "completed") {
			hasPending = true
			continue
		}

		// For completed checks, check conclusion (case-insensitive)
		// Success conclusions: success, neutral, skipped
		// Failure conclusions: failure, cancelled, timed_out, action_required
		conclusion := strings.ToLower(check.Conclusion)
		switch conclusion {
		case "success", "neutral", "skipped":
			// These are considered passing
			continue
		default:
			// failure, cancelled, timed_out, action_required, or any other value
			hasFailure = true
		}
	}

	// Determine overall status
	status := StatusSuccess
	if hasFailure {
		status = StatusFailure
	} else if hasPending {
		status = StatusPending
	}

	return CheckSummary{Status: status, Total: len(checks)}
}

// StatusCheckRollupToSummary converts GitHub's statusCheckRollup state to CheckSummary
// This is more efficient as it uses the aggregated state from GraphQL
// States: SUCCESS, FAILURE, PENDING, ERROR, or null
func StatusCheckRollupToSummary(state string) CheckSummary {
	switch strings.ToUpper(state) {
	case "SUCCESS":
		return CheckSummary{Status: StatusSuccess, Total: 1}
	case "FAILURE", "ERROR":
		return CheckSummary{Status: StatusFailure, Total: 1}
	case "PENDING":
		return CheckSummary{Status: StatusPending, Total: 1}
	default:
		// Empty state or unknown - no checks configured
		return CheckSummary{Status: StatusNone, Total: 0}
	}
}
