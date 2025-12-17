package interactive

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/swfz/gh-deps/internal/api"
	"github.com/swfz/gh-deps/internal/models"
)

// Styles
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Background(lipgloss.Color("235")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)
)

// model represents the TUI state
type model struct {
	prs          []models.PullRequest  // All PRs
	filtered     []models.PullRequest  // Filtered PRs based on search
	cursor       int                   // Current cursor position
	query        string                // Search query
	searchMode   bool                  // Whether in search mode
	confirmMode  bool                  // Whether in confirmation mode
	confirmingPR *models.PullRequest   // PR being confirmed for merge
	client       *api.Client           // API client for merging
	ctx          context.Context       // Context for API calls
	verbose      bool                  // Verbose mode
	message      string                // Status message
	messageType  string                // "error", "success", or ""
	width        int                   // Terminal width
	height       int                   // Terminal height
	merging      bool                  // Whether currently merging
	done         bool                  // Whether to quit
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case mergeResultMsg:
		m.merging = false
		m.message = msg.message
		if msg.success {
			m.messageType = "success"
			// Remove merged PR from list
			if m.cursor < len(m.filtered) {
				// Find and remove from both filtered and original list
				mergedPR := m.filtered[m.cursor]
				m.filtered = append(m.filtered[:m.cursor], m.filtered[m.cursor+1:]...)
				// Also remove from original prs list
				for i, pr := range m.prs {
					if pr.Number == mergedPR.Number && pr.Repository == mergedPR.Repository {
						m.prs = append(m.prs[:i], m.prs[i+1:]...)
						break
					}
				}
				// Adjust cursor
				if m.cursor >= len(m.filtered) && m.cursor > 0 {
					m.cursor--
				}
			}
		} else {
			m.messageType = "error"
		}
		return m, nil

	case tea.KeyMsg:
		// Clear message on any key press (except in confirm mode)
		if m.message != "" && !m.confirmMode {
			m.message = ""
			m.messageType = ""
		}

		switch msg.String() {
		case "ctrl+c":
			m.done = true
			return m, tea.Quit

		case "q", "esc":
			if m.confirmMode {
				// Cancel confirmation
				m.confirmMode = false
				m.confirmingPR = nil
				return m, nil
			}
			if m.searchMode {
				m.searchMode = false
				m.query = ""
				m.filterPRs()
				return m, nil
			}
			m.done = true
			return m, tea.Quit

		case "/":
			if !m.confirmMode {
				m.searchMode = true
			}
			return m, nil

		case "enter", "y":
			if m.confirmMode {
				// Confirm merge
				if m.confirmingPR != nil && !m.merging {
					m.merging = true
					m.message = "Merging..."
					m.messageType = ""
					pr := *m.confirmingPR
					m.confirmMode = false
					m.confirmingPR = nil
					return m, m.mergePR(pr)
				}
				return m, nil
			}
			if m.searchMode {
				m.searchMode = false
				return m, nil
			}
			// Show confirmation modal
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				m.confirmMode = true
				pr := m.filtered[m.cursor]
				m.confirmingPR = &pr
			}
			return m, nil

		case "n":
			if m.confirmMode {
				// Cancel confirmation
				m.confirmMode = false
				m.confirmingPR = nil
			}
			return m, nil

		case "up", "k":
			if !m.searchMode && !m.confirmMode && m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if !m.searchMode && !m.confirmMode && m.cursor < len(m.filtered)-1 {
				m.cursor++
			}

		case "backspace":
			if m.searchMode && len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.filterPRs()
			}

		default:
			if m.searchMode && !m.confirmMode && len(msg.String()) == 1 {
				m.query += msg.String()
				m.filterPRs()
			}
		}
	}

	return m, nil
}

// View renders the UI
func (m model) View() string {
	if m.done {
		return "Exiting...\n"
	}

	var b strings.Builder

	// Header
	header := headerStyle.Render(" gh-deps Interactive Mode ")
	b.WriteString(header + "\n")
	b.WriteString(dimStyle.Render("  Use ↑/↓ or j/k to navigate, / to search, Enter to merge, q to quit") + "\n\n")

	// Search bar
	if m.searchMode {
		b.WriteString(fmt.Sprintf("Search: %s█\n\n", m.query))
	} else if m.query != "" {
		b.WriteString(dimStyle.Render(fmt.Sprintf("Filter: %s (press / to edit, Esc to clear)", m.query)) + "\n\n")
	}

	// Status message
	if m.message != "" {
		switch m.messageType {
		case "error":
			b.WriteString(errorStyle.Render("✗ "+m.message) + "\n\n")
		case "success":
			b.WriteString(successStyle.Render("✓ "+m.message) + "\n\n")
		default:
			b.WriteString(m.message + "\n\n")
		}
	}

	// PR list header
	listHeader := fmt.Sprintf("%-4s %-20s %-12s %-4s %-6s %-15s %-12s %s",
		"#", "REPO", "BOT", "CI", "MERGE", "LABELS", "VERSION", "TITLE")
	b.WriteString(dimStyle.Render(listHeader) + "\n")
	b.WriteString(strings.Repeat("─", m.width) + "\n")

	// PR list (limited to visible area)
	maxVisible := m.height - 10 // Reserve space for header, footer, etc.
	if maxVisible < 5 {
		maxVisible = 5
	}

	startIdx := m.cursor - maxVisible/2
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(m.filtered) {
		endIdx = len(m.filtered)
		startIdx = endIdx - maxVisible
		if startIdx < 0 {
			startIdx = 0
		}
	}

	for i := startIdx; i < endIdx; i++ {
		pr := m.filtered[i]
		line := m.formatPRLine(i+1, pr)

		if i == m.cursor {
			b.WriteString(selectedStyle.Render("❯ " + line) + "\n")
		} else {
			b.WriteString(normalStyle.Render("  " + line) + "\n")
		}
	}

	// Footer
	if len(m.filtered) == 0 {
		b.WriteString("\n" + dimStyle.Render("  No PRs match your filter") + "\n")
	} else {
		b.WriteString("\n" + dimStyle.Render(fmt.Sprintf("  %d/%d PRs", m.cursor+1, len(m.filtered))) + "\n")
	}

	// Confirmation modal overlay
	if m.confirmMode && m.confirmingPR != nil {
		pr := *m.confirmingPR

		// Build modal content
		var modal strings.Builder
		modal.WriteString("\n")
		modal.WriteString("╔═══════════════════════════════════════════════════════════════╗\n")
		modal.WriteString("║               CONFIRM MERGE                                   ║\n")
		modal.WriteString("╠═══════════════════════════════════════════════════════════════╣\n")
		modal.WriteString(fmt.Sprintf("║ Repository: %-49s ║\n", pr.Repository))
		modal.WriteString(fmt.Sprintf("║ PR Number:  #%-47d ║\n", pr.Number))
		modal.WriteString(fmt.Sprintf("║ Title:      %-49s ║\n", truncate(pr.Title, 49)))
		modal.WriteString(fmt.Sprintf("║ Bot:        %-49s ║\n", pr.BotType.DisplayName()))
		modal.WriteString(fmt.Sprintf("║ Version:    %-49s ║\n", pr.Version))
		modal.WriteString(fmt.Sprintf("║ CI Status:  %-49s ║\n", string(pr.CheckSummary.Status)))
		modal.WriteString(fmt.Sprintf("║ Mergeable:  %-49s ║\n", formatMergeableState(pr.MergeableState)))
		modal.WriteString("╠═══════════════════════════════════════════════════════════════╣\n")

		// Show warnings
		if pr.MergeableState == models.MergeableStateConflicting {
			modal.WriteString("║ " + errorStyle.Render("⚠ WARNING: This PR has conflicts!") + strings.Repeat(" ", 29) + "║\n")
		} else if pr.CheckSummary.Status == models.StatusFailure {
			modal.WriteString("║ " + errorStyle.Render("⚠ WARNING: CI checks are failing!") + strings.Repeat(" ", 27) + "║\n")
		} else if pr.CheckSummary.Status == models.StatusPending {
			modal.WriteString("║ ⚠ WARNING: CI checks are pending" + strings.Repeat(" ", 29) + "║\n")
		}

		modal.WriteString("║                                                               ║\n")
		modal.WriteString("║ Merge this PR? (y/n or Esc to cancel)                        ║\n")
		modal.WriteString("╚═══════════════════════════════════════════════════════════════╝\n")

		// Center the modal and overlay it on the screen
		modalContent := selectedStyle.Render(modal.String())
		b.WriteString("\n" + modalContent)
	}

	return b.String()
}

// formatPRLine formats a single PR line for display
func (m model) formatPRLine(num int, pr models.PullRequest) string {
	repo := truncate(pr.RepoName(), 20)
	bot := truncate(pr.BotType.DisplayName(), 12)
	ci := string(pr.CheckSummary.Status)
	merge := formatMergeableState(pr.MergeableState)
	labels := formatLabels(pr.Labels, 15)
	version := truncate(pr.Version, 12)

	// Calculate title width dynamically based on terminal width
	// Fixed columns: # (4) + REPO (20) + BOT (12) + CI (4) + MERGE (6) + LABELS (15) + VERSION (12) = 73
	// Add spaces between columns (~7) and margins (~10) = 90
	fixedWidth := 90
	titleWidth := m.width - fixedWidth
	if titleWidth < 30 {
		titleWidth = 30 // Minimum width for narrow terminals
	}
	// No maximum limit - use full terminal width
	title := truncate(pr.Title, titleWidth)

	return fmt.Sprintf("%-4d %-20s %-12s %-4s %-6s %-15s %-12s %s",
		num, repo, bot, ci, merge, labels, version, title)
}

// filterPRs filters PRs based on query
func (m *model) filterPRs() {
	if m.query == "" {
		m.filtered = m.prs
		m.cursor = 0
		return
	}

	m.filtered = []models.PullRequest{}
	query := strings.ToLower(m.query)

	for _, pr := range m.prs {
		// Search in repo name, title, bot type, labels, version
		searchText := strings.ToLower(fmt.Sprintf("%s %s %s %s %s",
			pr.RepoName(), pr.Title, pr.BotType.DisplayName(),
			strings.Join(pr.Labels, " "), pr.Version))

		if strings.Contains(searchText, query) {
			m.filtered = append(m.filtered, pr)
		}
	}

	// Reset cursor if out of bounds
	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// mergePR creates a command to merge the selected PR
func (m *model) mergePR(pr models.PullRequest) tea.Cmd {
	return func() tea.Msg {
		// Check for conflicts
		if pr.MergeableState == models.MergeableStateConflicting {
			return mergeResultMsg{
				success: false,
				message: fmt.Sprintf("PR #%d has conflicts and cannot be merged", pr.Number),
			}
		}

		// Parse repository
		owner, repo, err := api.ParseRepository(pr.Repository)
		if err != nil {
			return mergeResultMsg{
				success: false,
				message: fmt.Sprintf("Invalid repository format: %v", err),
			}
		}

		// Execute merge
		resp, err := m.client.MergePullRequest(m.ctx, owner, repo, pr.Number)
		if err != nil {
			return mergeResultMsg{
				success: false,
				message: fmt.Sprintf("Merge failed: %v", err),
			}
		}

		if !resp.Merged {
			return mergeResultMsg{
				success: false,
				message: fmt.Sprintf("Merge unsuccessful: %s", resp.Message),
			}
		}

		return mergeResultMsg{
			success: true,
			message: fmt.Sprintf("Successfully merged PR #%d in %s", pr.Number, pr.Repository),
		}
	}
}

// mergeResultMsg represents the result of a merge operation
type mergeResultMsg struct {
	success bool
	message string
}

// Helper functions
func formatMergeableState(state models.MergeableState) string {
	switch state {
	case models.MergeableStateMergeable:
		return "✓"
	case models.MergeableStateConflicting:
		return "✗"
	case models.MergeableStateUnknown:
		return "?"
	default:
		return "-"
	}
}

func formatLabels(labels []string, maxLen int) string {
	if len(labels) == 0 {
		return "-"
	}
	joined := strings.Join(labels, ",")
	return truncate(joined, maxLen)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// RunTUI starts the interactive TUI
func RunTUI(ctx context.Context, prs []models.PullRequest, client *api.Client, verbose bool) error {
	m := model{
		prs:      prs,
		filtered: prs,
		cursor:   0,
		client:   client,
		ctx:      ctx,
		verbose:  verbose,
		width:    80,
		height:   24,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
