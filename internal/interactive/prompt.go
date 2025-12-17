package interactive

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/swfz/gh-deps/internal/api"
	"github.com/swfz/gh-deps/internal/models"
)

// Runner handles interactive PR selection and merging
type Runner struct {
	client  *api.Client
	scanner *bufio.Scanner
	verbose bool
}

// NewRunner creates a new interactive runner
func NewRunner(client *api.Client, verbose bool) *Runner {
	return &Runner{
		client:  client,
		scanner: bufio.NewScanner(os.Stdin),
		verbose: verbose,
	}
}

// Run starts the interactive merge loop
func (r *Runner) Run(ctx context.Context, prs []models.PullRequest) error {
	if len(prs) == 0 {
		return fmt.Errorf("no PRs available to merge")
	}

	for {
		// Display selection prompt
		fmt.Println("\nSelect a PR to merge (or 'q' to quit):")
		fmt.Print("Enter PR number: ")

		// Read user input
		input, err := r.readLine()
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		// Handle quit
		if strings.ToLower(strings.TrimSpace(input)) == "q" {
			fmt.Println("Exiting interactive mode.")
			return nil
		}

		// Parse selection
		selection, err := r.parseSelection(input, len(prs))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		// Get selected PR (convert 1-based to 0-based index)
		pr := prs[selection-1]

		// Show PR details and warnings
		if err := r.displayPRDetails(pr); err != nil {
			// Don't exit - just show error and continue to next selection
			fmt.Printf("Cannot proceed: %v\n", err)
			continue
		}

		// Confirm merge
		if !r.confirmMerge() {
			fmt.Println("Merge cancelled.")
			continue
		}

		// Execute merge
		if err := r.executeMerge(ctx, pr); err != nil {
			fmt.Printf("Failed to merge: %v\n", err)
			// Continue loop - allow user to try again or select different PR
			continue
		}

		fmt.Printf("✅ Successfully merged PR #%d in %s\n", pr.Number, pr.Repository)

		// Ask if user wants to continue
		if !r.askContinue() {
			fmt.Println("Exiting interactive mode.")
			return nil
		}
	}
}

// readLine reads a line from stdin
func (r *Runner) readLine() (string, error) {
	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("EOF")
	}
	return r.scanner.Text(), nil
}

// parseSelection validates and parses user selection
func (r *Runner) parseSelection(input string, maxPRs int) (int, error) {
	input = strings.TrimSpace(input)

	if input == "" {
		return 0, fmt.Errorf("empty input")
	}

	num, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", input)
	}

	if num < 1 || num > maxPRs {
		return 0, fmt.Errorf("number must be between 1 and %d", maxPRs)
	}

	return num, nil
}

// displayPRDetails shows PR information and warnings
func (r *Runner) displayPRDetails(pr models.PullRequest) error {
	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Printf("Repository:      %s\n", pr.Repository)
	fmt.Printf("PR Number:       #%d\n", pr.Number)
	fmt.Printf("Title:           %s\n", pr.Title)
	fmt.Printf("URL:             %s\n", pr.URL)
	fmt.Printf("Bot:             %s\n", pr.BotType.DisplayName())
	fmt.Printf("Version:         %s\n", pr.Version)
	fmt.Printf("CI Status:       %s\n", pr.CheckSummary.Status)
	fmt.Printf("Mergeable State: %s\n", pr.MergeableState)
	fmt.Println(strings.Repeat("-", 60))

	hasWarnings := false

	// Show conflict error (blocking)
	if pr.MergeableState == models.MergeableStateConflicting {
		fmt.Println()
		fmt.Println("❌ ERROR: This PR has MERGE CONFLICTS!")
		fmt.Println("   Cannot merge until conflicts are resolved.")
		fmt.Println("   Please resolve conflicts on GitHub first.")
		fmt.Println()
		return fmt.Errorf("PR has merge conflicts")
	}

	// Show unknown state warning
	if pr.MergeableState == models.MergeableStateUnknown {
		fmt.Println()
		fmt.Println("⚠️  WARNING: Mergeable state is UNKNOWN!")
		fmt.Println("   GitHub is still calculating whether this PR can be merged.")
		fmt.Println("   You may need to wait a moment and try again.")
		fmt.Println()
		hasWarnings = true
	}

	// Show warning if checks are not passing
	if pr.CheckSummary.Status != models.StatusSuccess {
		if !hasWarnings {
			fmt.Println()
		}
		switch pr.CheckSummary.Status {
		case models.StatusFailure:
			fmt.Println("⚠️  WARNING: CI checks are FAILING!")
		case models.StatusPending:
			fmt.Println("⚠️  WARNING: CI checks are still PENDING!")
		case models.StatusNone:
			fmt.Println("ℹ️  INFO: No CI checks configured for this PR.")
		}
		fmt.Println("    Merging may introduce broken code.")
		fmt.Println()
	}

	return nil
}

// confirmMerge asks user to confirm the merge
func (r *Runner) confirmMerge() bool {
	fmt.Print("\nProceed with merge? (y/N): ")

	input, err := r.readLine()
	if err != nil {
		return false
	}

	response := strings.ToLower(strings.TrimSpace(input))
	return response == "y" || response == "yes"
}

// executeMerge performs the actual merge operation
func (r *Runner) executeMerge(ctx context.Context, pr models.PullRequest) error {
	owner, repo, err := api.ParseRepository(pr.Repository)
	if err != nil {
		return err
	}

	fmt.Printf("\nMerging PR #%d in %s...\n", pr.Number, pr.Repository)

	resp, err := r.client.MergePullRequest(ctx, owner, repo, pr.Number)
	if err != nil {
		return err
	}

	if !resp.Merged {
		return fmt.Errorf("merge unsuccessful: %s", resp.Message)
	}

	if r.verbose {
		fmt.Printf("[DEBUG] Merge SHA: %s\n", resp.SHA)
	}

	return nil
}

// askContinue asks if user wants to merge another PR
func (r *Runner) askContinue() bool {
	fmt.Print("\nMerge another PR? (y/N): ")

	input, err := r.readLine()
	if err != nil {
		return false
	}

	response := strings.ToLower(strings.TrimSpace(input))
	return response == "y" || response == "yes"
}
