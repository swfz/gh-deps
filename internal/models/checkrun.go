package models

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
		// Check if any run is not completed
		if check.Status != "completed" {
			hasPending = true
			continue
		}

		// For completed checks, check conclusion
		// Success conclusions: success, neutral, skipped
		// Failure conclusions: failure, cancelled, timed_out, action_required
		switch check.Conclusion {
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
