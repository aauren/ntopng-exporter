# Development Guide

This guide will help you get started with developing ntopng-exporter, whether you're fixing a bug, adding a feature, or
just exploring the codebase.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Building the Project](#building-the-project)
- [Running the Application](#running-the-application)
- [Code Quality](#code-quality)
- [Testing](#testing)
- [Dependency Management](#dependency-management)
- [Code Style Guidelines](#code-style-guidelines)
- [Project Structure](#project-structure)
- [Continuous Integration](#continuous-integration)
- [Making Changes](#making-changes)
- [Common Development Tasks](#common-development-tasks)
- [Getting Help](#getting-help)
- [Contributing](#contributing)

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.25.4 or later** - [Download Go](https://golang.org/dl/)
- **golangci-lint v2** - [Installation instructions](https://golangci-lint.run/welcome/install/)
- **Git** - For version control
- **Docker** (optional) - For building container images

## Getting Started

### 1. Fork and Clone the Repository

First, fork the repository on GitHub by clicking the "Fork" button at [github.com/aauren/ntopng-exporter](https://github.com/aauren/ntopng-exporter).

Then clone your fork:

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/ntopng-exporter.git
cd ntopng-exporter

# Add upstream remote to track the original repository
git remote add upstream https://github.com/aauren/ntopng-exporter.git

# Verify remotes
git remote -v
# Should show:
# origin    https://github.com/YOUR_USERNAME/ntopng-exporter.git (fetch)
# origin    https://github.com/YOUR_USERNAME/ntopng-exporter.git (push)
# upstream  https://github.com/aauren/ntopng-exporter.git (fetch)
# upstream  https://github.com/aauren/ntopng-exporter.git (push)
```

### 2. Install Dependencies

Download all Go module dependencies:

```bash
go mod download
```

### 3. Verify Your Setup

Run a quick build to ensure everything is working:

```bash
go build -v ./...
```

If this completes without errors, you're ready to start developing!

## Building the Project

### Standard Build

For development and testing:

```bash
go build -o ntopng-exporter ntopng-exporter.go
```

This creates an `ntopng-exporter` binary in the current directory.

### Production Build

For optimized production builds (same as CI):

```bash
CGO_ENABLED=0 go build -v -ldflags '-s -w' -o ntopng-exporter ntopng-exporter.go
```

The flags do the following:

- `CGO_ENABLED=0` - Disables CGO for static binary
- `-ldflags '-s -w'` - Strips debug information to reduce binary size

### Building All Packages

To build all packages in the project:

```bash
go build -v ./...
```

## Running the Application

### Running from Source

```bash
go run ntopng-exporter.go
```

### Running the Built Binary

```bash
./ntopng-exporter
```

**Note:** The application requires a configuration file. See the [main README](../README.md#configuring) for
configuration details.

### Quick Development Setup

For development, create a config file at `./config/ntopng-exporter.yaml`:

```bash
cp config/ntopng-exporter.yaml.example config/ntopng-exporter.yaml
# Edit the config file with your ntopng instance details
```

## Code Quality

### Linting

This project enforces code quality standards using golangci-lint for Go code and markdownlint for documentation.
**All code and documentation must pass linting before it can be merged.**

#### Go Linting

Run all Go linters:

```bash
golangci-lint run
```

Run specific linter:

```bash
# Check only import dependencies
golangci-lint run --enable-only=depguard

# Check only line length
golangci-lint run --enable-only=lll
```

Auto-fix issues where possible:

```bash
golangci-lint run --fix
```

#### Markdown Linting

Install markdownlint-cli (first time only):

```bash
npm install -g markdownlint-cli
```

Run markdown linter:

```bash
# Lint all markdown files
markdownlint .

# Lint specific file
markdownlint README.md docs/development.md

# Fix auto-fixable issues
markdownlint --fix .
```

### What Gets Checked

#### Go Code Checks

- **Import restrictions** (depguard) - Only approved dependencies allowed
- **Code complexity** - Functions shouldn't be too complex
- **Security issues** (gosec) - Common security problems
- **Line length** (lll) - Maximum 140 characters per line
- **Unused code** - Dead code and unused variables
- **Style issues** - Inconsistent naming, formatting, etc.
- **Common mistakes** - See `.golangci.yaml` for full list

#### Markdown Documentation Checks

- **Line length** - Maximum 120 characters (code blocks: 200)
- **Code fence style** - Must use fenced code blocks (```)
- **Heading structure** - Proper heading hierarchy
- **Link validity** - Proper link formatting
- See `.markdownlint.yaml` for configuration

### Pre-commit Checklist

Before committing code, always run:

```bash
# Format your Go code
gofmt -w .
goimports -w .

# Run Go linters
golangci-lint run

# Run markdown linter
markdownlint .

# Build to ensure it compiles
go build -v ./...

# Ensure all tests pass
go test ./...
```

## Testing

**Note:** This project currently has no automated tests. When adding tests:

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests in a specific package
go test -v ./internal/config

# Run a single test function
go test -v ./internal/config -run TestValidation
```

### Running Tests with Coverage

```bash
# Run tests and display coverage
go test -cover ./...

# Generate detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

This opens an HTML page showing exactly which lines are covered by tests.

### Writing Tests

When adding new functionality, please include tests:

1. Create a `*_test.go` file in the same package
2. Write test functions starting with `Test`
3. Use table-driven tests for multiple scenarios
4. Aim for at least 70% code coverage

Example:

```go
package config

import "testing"

func TestValidateConfig(t *testing.T) {
    tests := []struct {
        name    string
        config  Config
        wantErr bool
    }{
        {
            name: "valid config",
            config: Config{/* valid config */},
            wantErr: false,
        },
        // Add more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Dependency Management

### Adding a New Dependency

**Important:** This project has strict dependency controls via depguard.

Currently allowed dependencies:

- Go standard library
- `github.com/aauren/ntopng-exporter/*` (internal packages)
- `github.com/prometheus/client_golang`
- `github.com/spf13/viper`

To add a new dependency:

1. Add the dependency:

```bash
go get github.com/example/package
```

<!-- markdownlint-disable-next-line MD029 -->
2. Update `.golangci.yaml` to allow the new dependency:

```yaml
linters-settings:
 depguard:
   rules:
     main:
       allow:
         - "$gostd"
         - github.com/aauren/ntopng-exporter
         - github.com/prometheus/client_golang
         - github.com/spf13/viper
         - github.com/example/package  # Add new dependency here
```

<!-- markdownlint-disable-next-line MD029 -->
3. Run `go mod tidy` to clean up:

```bash
go mod tidy
```

### Updating Dependencies

```bash
# Update all dependencies to latest minor/patch versions
go get -u ./...

# Update a specific dependency
go get -u github.com/prometheus/client_golang

# Update to a specific version
go get github.com/prometheus/client_golang@v1.20.0

# Clean up unused dependencies
go mod tidy
```

## Code Style Guidelines

### Imports

Organize imports in three groups, separated by blank lines:

```go
import (
    // 1. Standard library
    "context"
    "fmt"
    "net/http"

    // 2. External dependencies
    "github.com/prometheus/client_golang/prometheus"
    "github.com/spf13/viper"

    // 3. Internal packages
    "github.com/aauren/ntopng-exporter/internal/config"
    "github.com/aauren/ntopng-exporter/internal/ntopng"
)
```

### Naming Conventions

- **Packages**: lowercase, single word (`config`, `ntopng`)
- **Exported types/functions**: PascalCase (`Controller`, `ParseConfig`)
- **Unexported types/functions**: camelCase (`ntopHost`, `parseEndpoint`)
- **Constants**: PascalCase for exported, camelCase for unexported

### Error Handling

Always handle errors explicitly:

```go
// Good
if err != nil {
    return fmt.Errorf("failed to parse config: %w", err)
}

// For fatal errors in main
if err != nil {
    fmt.Printf("fatal error: %v\n", err)
    os.Exit(1)
}
```

### Comments

- Document all exported functions and types
- Start comments with the name of what you're documenting
- Use complete sentences

```go
// Controller manages the ntopng scraping lifecycle and stores
// the current state of hosts and interfaces.
type Controller struct {
    // ...
}

// CacheInterfaceIds retrieves and caches the interface IDs from ntopng
// to avoid repeated lookups during scraping operations.
func (c *Controller) CacheInterfaceIds() error {
    // ...
}
```

## Project Structure

Understanding the codebase layout:

<!-- markdownlint-disable-next-line MD040 -->
```
ntopng-exporter/
├── config/                    # Sample configuration files
│   └── ntopng-exporter.yaml
├── docs/                      # Documentation
│   ├── development.md         # This file
│   └── grafana_example.png
├── internal/                  # Internal packages (not importable externally)
│   ├── config/               # Configuration parsing and validation
│   ├── metrics/
│   │   └── prometheus/       # Prometheus metric collectors
│   └── ntopng/               # ntopng API client
├── resources/                 # Supporting files
│   ├── grafana-dashboard.json
│   └── ntopng-exporter.service
├── .github/
│   └── workflows/
│       └── ci.yml            # CI/CD pipeline
├── .golangci.yaml            # Linter configuration
├── ntopng-exporter.go        # Main entry point
├── go.mod                     # Go module dependencies
└── README.md                  # User-facing documentation
```

### Key Packages

- **`internal/config`**: Parses YAML config files using Viper, validates settings
- **`internal/ntopng`**: HTTP client for ntopng REST API, handles auth and scraping
- **`internal/metrics/prometheus`**: Prometheus collectors for hosts and interfaces
- **`ntopng-exporter.go`**: Main entry point, wires everything together

## Continuous Integration

This project uses GitHub Actions for CI. On every push and pull request:

1. **Markdown Linting** - Documentation must pass markdownlint checks
2. **Go Linting** - Code must pass all golangci-lint checks
3. **Tests** - All tests must pass (with race detection and coverage)
4. **Build** - Code must compile successfully
5. **Release** (tags only) - Creates multi-platform binaries and Docker images

### CI Pipeline

View the CI configuration: [`.github/workflows/ci.yml`](../.github/workflows/ci.yml)

The pipeline runs these jobs:

**Quality Checks (run in parallel):**

- `ci-lint-markdown`: Runs markdownlint on all documentation
- `ci-lint-go`: Runs golangci-lint with all enabled linters
- `ci-test`: Runs tests with race detection and generates coverage report

**Build and Release (run sequentially after quality checks pass):**

- `ci-build-ntopng-exporter`: Builds the binary
- `ci-build-container`: Builds and pushes Docker images (tags only)
- `ci-goreleaser-tag`: Creates GitHub release with binaries (tags only)

**Important:** PRs cannot be merged if any CI checks fail.

### Test Coverage

The CI pipeline generates test coverage reports and uploads them to Codecov (if configured). You can view coverage locally:

```bash
# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report in terminal
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
# Open coverage.html in your browser
```

## Making Changes

### Development Workflow

This project follows a **forked workflow**. All contributions should be made through pull requests from your personal fork.

#### 1. Sync Your Fork

Before starting new work, ensure your fork is up to date with the upstream repository:

```bash
# Fetch the latest changes from upstream
git fetch upstream

# Switch to your main branch
git checkout main

# Merge upstream changes
git merge upstream/main

# Push updates to your fork
git push origin main
```

#### 2. Create a Feature Branch

Create a branch for your changes from an up-to-date main branch:

```bash
# Create and switch to a new branch
git checkout -b feat/add-dns-metrics

# Or for a bugfix
git checkout -b fix/interface-parsing-error
```

**Branch naming conventions:**

- `feat/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation changes
- `refactor/description` - Code refactoring
- `test/description` - Adding or updating tests
- `chore/description` - Maintenance tasks

#### 3. Make Your Changes

Make your changes following the code style guidelines and best practices:

```bash
# Edit files...
vim internal/ntopng/controller.go

# Format code
gofmt -w .
goimports -w .
```

#### 4. Test Your Changes

Before committing, verify everything works:

```bash
# Build to ensure it compiles
go build -v ./...

# Run linters
golangci-lint run

# Run tests (when tests exist)
go test ./...

# Test the application (if applicable)
./ntopng-exporter
```

#### 5. Commit Your Changes

Follow [Conventional Commits](https://www.conventionalcommits.org/) format:

```bash
git add .
git commit -m "feat: add DNS query type metrics to host collector"
```

See the [Commit Message Guidelines](#commit-message-guidelines) section below for detailed format.

#### 6. Push to Your Fork

Push your branch to your fork on GitHub:

```bash
git push origin feat/add-dns-metrics
```

#### 7. Create a Pull Request

1. Go to your fork on GitHub
2. Click "Compare & pull request" for your branch
3. Fill in the PR template with:
   - Clear description of changes
   - Motivation and context
   - How changes were tested
   - Related issue numbers
4. Submit the pull request to `aauren/ntopng-exporter:main`

#### 8. Address Review Comments

If maintainers request changes:

```bash
# Make the requested changes
vim internal/ntopng/controller.go

# Commit with a descriptive message
git add .
git commit -m "fix: address review comments on error handling"

# Push updates to the same branch
git push origin feat/add-dns-metrics
```

The pull request will automatically update with your new commits.

#### 9. Rebase if Needed

If the upstream main branch has advanced since you started:

```bash
# Fetch latest upstream changes
git fetch upstream

# Rebase your branch onto upstream/main
git rebase upstream/main

# If conflicts occur, resolve them and continue
git add <resolved-files>
git rebase --continue

# Force push to your fork (only for your own branches!)
git push origin feat/add-dns-metrics --force-with-lease
```

### Commit Message Guidelines

This project follows [Conventional Commits](https://www.conventionalcommits.org/) specification for commit messages.
This leads to more readable messages that are easy to follow when looking through the project history.

#### Format

Each commit message consists of a **header**, **body**, and **footer**:

```markdown
<type>(<scope>): <subject>

<body>

<footer>
```

The **header** is mandatory, while the **body** and **footer** are optional.

#### Type

Must be one of the following:

- **feat**: A new feature
- **fix**: A bug fix
- **docs**: Documentation only changes
- **style**: Changes that don't affect code meaning (formatting, missing semi-colons, etc.)
- **refactor**: Code change that neither fixes a bug nor adds a feature
- **perf**: Code change that improves performance
- **test**: Adding missing tests or correcting existing tests
- **build**: Changes that affect the build system or external dependencies
- **ci**: Changes to CI configuration files and scripts
- **chore**: Other changes that don't modify src or test files
- **revert**: Reverts a previous commit

#### Scope

The scope should specify the place of the commit change. For example:

- `config` - Configuration parsing and validation
- `ntopng` - ntopng API client
- `prometheus` - Prometheus metrics collectors
- `ci` - CI/CD pipeline
- `deps` - Dependencies

Example: `feat(prometheus): add interface throughput metrics`

#### Subject

The subject contains a succinct description of the change:

- Use the imperative, present tense: "add" not "added" nor "adds"
- Don't capitalize the first letter
- No period (.) at the end
- Limit to 72 characters

#### Body

The body should include the motivation for the change and contrast this with previous behavior.

- Use the imperative, present tense
- Wrap at 72 characters
- Explain the "what" and "why" vs. the "how"

#### Footer

The footer should contain any information about **Breaking Changes** and reference **GitHub issues** that this commit closes.

**Breaking Changes** should start with the word `BREAKING CHANGE:` followed by a description.

**Referencing issues:** Use `Fixes #123` or `Closes #456` to link commits to issues.

#### Examples

**Feature with body:**

```markdown
feat(ntopng): add support for L7 protocol metrics

Implement scraping of L7 protocol statistics from the ntopng API.
This allows users to monitor application-layer traffic patterns.

- Add L7Protocol struct to data structures
- Implement scrapeL7ProtocolEndpoint method
- Register L7 protocol collector

Closes #45
```

**Bug fix:**

```markdown
fix(config): validate interface names are not empty

Previously, blank interface names in the config would cause a panic
at runtime. Add validation to reject empty interface names during
config parsing.

Fixes #78
```

**Breaking change:**

```markdown
feat(config): rename authentication method field

BREAKING CHANGE: The config field `auth` has been renamed to `authMethod`
to better reflect its purpose. Users must update their configuration files.

Migration: Change `auth: "cookie"` to `authMethod: "cookie"`
```

**Simple fixes:**

```markdown
fix: correct typo in error message
```

```markdown
docs: update development guide with fork workflow
```

```markdown
chore(deps): update prometheus client to v1.20.0
```

#### Tips for Good Commit Messages

1. **Keep commits atomic** - One logical change per commit
2. **Commit early and often** - Smaller commits are easier to review
3. **Test before committing** - Ensure code builds and lints pass
4. **Reference issues** - Link commits to issues they address
5. **Separate formatting** - Don't mix style changes with logic changes

#### Amending Commits

If you need to fix your last commit:

```bash
# Make your changes
git add .

# Amend the previous commit
git commit --amend

# Force push to your branch (only for unpushed or your own branches!)
git push origin feat/add-dns-metrics --force-with-lease
```

**Warning:** Never amend commits that have been pushed to a shared branch or that others might have based work on.

## Common Development Tasks

### Adding a New Metric

1. Update the data structures in `internal/ntopng/ntop_data_structures.go`
2. Modify the API request to include the new field
3. Add the metric descriptor in the appropriate collector
4. Update the `Describe()` and `Collect()` methods
5. Test the new metric appears in `/metrics` endpoint

### Adding a New Scrape Target

1. Define a constant in `internal/config/config.go`
2. Add to `AvailableScrapeTargets` map
3. Implement scraping in `internal/ntopng/controller.go`
4. Create a new collector in `internal/metrics/prometheus/`
5. Register in `ntopng-exporter.go`

### Debugging

```bash
# Build with debug symbols
go build -v -o ntopng-exporter ntopng-exporter.go

# Run with delve debugger
dlv exec ./ntopng-exporter

# Or use your IDE's debugger (VSCode, GoLand, etc.)
```

## Getting Help

- **Issues**: Check [GitHub Issues](https://github.com/aauren/ntopng-exporter/issues)
- **Documentation**: See the main [README](../README.md)
- **Code Questions**: Open a discussion on GitHub

## Contributing

We welcome contributions! Before submitting a PR:

1. Ensure all tests pass (when tests exist)
2. Run `golangci-lint run` and fix all issues
3. Follow the code style guidelines
4. Update documentation if needed
5. Keep PRs focused on a single change

Thank you for contributing to ntopng-exporter!
