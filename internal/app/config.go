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
}

// ParseConfig parses command-line flags and validates configuration
func ParseConfig() (*Config, error) {
	var org, user string

	flag.StringVar(&org, "org", "", "GitHub organization name")
	flag.StringVar(&user, "user", "", "GitHub user name")

	config := &Config{}
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&config.Verbose, "v", false, "Enable verbose output (shorthand)")

	flag.Parse()

	// Validate that exactly one of org or user is specified
	if org == "" && user == "" {
		return nil, errors.New("either --org or --user must be specified")
	}

	if org != "" && user != "" {
		return nil, errors.New("cannot specify both --org and --user")
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
