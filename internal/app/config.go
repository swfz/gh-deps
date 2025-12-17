package app

import (
	"errors"
	"flag"
)

// Config holds the application configuration
type Config struct {
	Target         string // Organization or user name
	IsOrganization bool   // True if targeting an organization, false for user
	Verbose        bool   // Enable verbose output
	Limit          int    // Maximum PRs to display (0 = unlimited)
	SkipChecks     bool   // Skip fetching check runs
	Interactive    bool   // Enable interactive PR merge mode
}

// ParseConfig parses command-line flags and validates configuration
func ParseConfig() (*Config, error) {
	var org, user string

	flag.StringVar(&org, "org", "", "GitHub organization name")
	flag.StringVar(&user, "user", "", "GitHub user name")

	config := &Config{}
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&config.Verbose, "v", false, "Enable verbose output (shorthand)")
	flag.IntVar(&config.Limit, "limit", 50, "Limit number of PRs to display (0 = unlimited)")
	flag.IntVar(&config.Limit, "l", 50, "Limit number of PRs (shorthand)")
	flag.BoolVar(&config.SkipChecks, "skip-checks", false, "Skip fetching CI check runs")
	flag.BoolVar(&config.Interactive, "interactive", false, "Enable interactive PR merge mode")
	flag.BoolVar(&config.Interactive, "i", false, "Enable interactive mode (shorthand)")

	flag.Parse()

	// Validate that exactly one of org or user is specified
	if org == "" && user == "" {
		return nil, errors.New("either --org or --user must be specified")
	}

	if org != "" && user != "" {
		return nil, errors.New("cannot specify both --org and --user")
	}

	// Validate limit
	if config.Limit < 0 {
		return nil, errors.New("--limit must be >= 0")
	}

	// Set target and type
	if org != "" {
		config.Target = org
		config.IsOrganization = true
	} else {
		config.Target = user
		config.IsOrganization = false
	}

	return config, nil
}
