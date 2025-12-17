# gh-deps

A GitHub CLI extension to centrally manage dependency update PRs (Renovate, Dependabot, GitHub Actions) across multiple repositories.

## Overview

When managing multiple repositories, dependency update bots like Renovate and Dependabot create numerous PRs that are scattered across different repos. This tool aggregates all such PRs in a single view, showing CI status and version changes at a glance.

## Features

- Aggregates dependency update PRs from all repositories in an organization or user account
- Supports three bot types:
  - Renovate
  - Dependabot
  - GitHub Actions
- Displays CI/test status with visual indicators (✅ ❌ ⏳)
- Extracts and shows version changes (e.g., "1.0.0 -> 1.1.0")
- Clean table format optimized for terminal viewing
- Uses GitHub GraphQL API for efficient data fetching

## Installation

### Prerequisites

- [GitHub CLI](https://cli.github.com/) installed and authenticated
- Go 1.21 or higher

### Install as gh extension

```bash
gh extension install swfz/gh-deps
```

### Build from source

```bash
git clone https://github.com/swfz/gh-deps.git
cd gh-deps
make build
gh extension install .
```

## Usage

### List dependency PRs for an organization

```bash
gh deps --org <organization-name>
```

### List dependency PRs for a user

```bash
gh deps --user <username>
```

### Enable verbose output

```bash
gh deps --org <organization-name> --verbose
```

## Output Format

The tool displays results in a table with the following columns:

| Column | Description |
|--------|-------------|
| REPO | Repository name (truncated to 20 characters) |
| BOT | Bot type (renovate, dependabot, github-actions) |
| STATUS | CI status (✅ success, ❌ failure, ⏳ pending, - no checks) |
| DATE | PR creation date (YYYY-MM-DD) |
| VERSION | Version change extracted from PR body |
| TITLE | PR title (truncated to 60 characters with ellipsis) |

### Example Output

```
REPO                  BOT             STATUS  DATE        VERSION           TITLE
my-api                renovate        ✅      2025-12-15  1.2.0 -> 1.3.0   Update dependency express to v1.3.0
my-frontend           dependabot      ❌      2025-12-14  2.0.1 -> 2.1.0   Bump react from 2.0.1 to 2.1.0
my-backend            github-actions  ⏳      2025-12-13  v3 -> v4         Update actions/checkout to v4

Total: 3 dependency update PRs
```

## How It Works

### CI Status Detection

The tool examines all check runs for each PR:
- ✅ **Success**: All checks completed successfully
- ❌ **Failure**: One or more checks failed
- ⏳ **Pending**: One or more checks are still running
- **-**: No checks configured

### Version Extraction

Version information is extracted from PR body text using patterns specific to each bot:

- **Dependabot**: `from X to Y`
- **Renovate**: `X -> Y` (with or without backticks)
- **GitHub Actions**: `X -> Y`

If no version pattern is found, "-" is displayed.

## Development

### Prerequisites

- Go 1.21+
- GitHub CLI with authentication configured

### Setup

```bash
# Clone the repository
git clone https://github.com/swfz/gh-deps.git
cd gh-deps

# Install dependencies
make deps

# Build
make build

# Run tests
make test
```

### Project Structure

```
gh-deps/
├── cmd/gh-deps/        # Main entry point
├── internal/
│   ├── api/           # GitHub API client
│   ├── models/        # Data models
│   ├── parser/        # Version extraction logic
│   ├── formatter/     # Table formatting
│   └── app/           # Application logic
├── Makefile
├── go.mod
└── README.md
```

### Available Make Targets

- `make build` - Build the binary
- `make install` - Install to GOPATH
- `make test` - Run tests
- `make test-coverage` - Run tests with coverage report
- `make lint` - Run linter
- `make clean` - Remove build artifacts
- `make deps` - Download and tidy dependencies

## Limitations

- Only shows open (unmerged/unclosed) PRs
- Fetches up to 100 PRs per repository (configurable in code)
- Rate limited to GitHub API limits (typically 5000 req/hour)

## Future Enhancements

Potential features for future versions:

- Filter by bot type, repository, or status
- Interactive mode to select and view/merge PRs
- JSON/CSV output formats
- Auto-merge PRs that pass all checks
- Caching for faster repeat queries

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Author

[swfz](https://github.com/swfz)
