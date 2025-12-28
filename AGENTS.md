# AGENTS.md - Developer Guide for ntopng-exporter

This guide provides coding agents with essential information about building, testing, and contributing to ntopng-exporter.

## Project Overview

ntopng-exporter is a Prometheus metric exporter for ntopng. It scrapes the ntopng REST API (v2) and publishes metrics
for network monitoring. The project is written in Go and targets ntopng version 5.X+ (v1 API).

## Build Commands

### Building the Binary

```bash
# Standard build
go build -v ./...

# Production build with optimized flags (used in CI)
CGO_ENABLED=0 go build -v -ldflags '-s -w' ./...

# Build the main binary
go build -o ntopng-exporter ntopng-exporter.go
```

### Running the Application

```bash
# Run directly
go run ntopng-exporter.go

# Run with config in default location
./ntopng-exporter
```

### Docker Build

```bash
# Build Docker image
docker build .

# Build with custom base images
docker build --build-arg BUILDTIME_BASE=golang:1.25.4 \
             --build-arg RUNTIME_BASE=gcr.io/distroless/static .
```

### Dependencies

```bash
# Download dependencies
go get .

# Update dependencies
go mod download

# Tidy dependencies
go mod tidy
```

## Linting

The project uses golangci-lint for Go code quality checks and markdownlint for documentation. Linting is enforced in CI
on all PRs and commits.

### Go Linting

```bash
# Run all enabled linters
golangci-lint run

# Run only specific linter
golangci-lint run --enable-only=depguard

# Run with auto-fix where possible
golangci-lint run --fix

# List all enabled linters
golangci-lint linters
```

### Markdown Linting

```bash
# Install markdownlint-cli (if not already installed)
npm install -g markdownlint-cli

# Lint all markdown files
markdownlint .

# Lint specific file
markdownlint README.md

# Fix auto-fixable issues
markdownlint --fix .
```

**Note:** The CI workflow runs both `golangci-lint` and `markdownlint` automatically. Make sure your code and
documentation pass linting before pushing.

## Testing

**Note:** This project currently has no test files. However, the CI pipeline runs `go test` to ensure the framework is
in place. When adding tests:

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests in a specific package
go test -v ./internal/config

# Run a single test
go test -v ./internal/config -run TestParsing

# Run tests with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Code Style Guidelines

### Markdown Style Guidelines

All markdown files must comply with markdownlint rules configured in `.markdownlint.yaml`. Key requirements:

<!-- markdownlint-capture -->
<!-- markdownlint-disable -->
- **Line length**: Maximum 120 characters for text, 200 for code blocks
- **Blank lines around lists**: Always add a blank line before and after lists
- **Blank lines around code blocks**: Always add a blank line before and after fenced code blocks (```)
- **Code block language**: Always specify a language for fenced code blocks (e.g., ```bash, ```go, or ```text)
- **Heading hierarchy**: Use proper heading levels (no skipping levels)
- **List formatting**: Use `-` for unordered lists, numbers for ordered lists
- **Consistent formatting**: Use fenced code blocks (```) instead of indented code blocks
<!-- makrdownlint-restore -->

Example of properly formatted markdown:

```markdown
## Section Title

This is a paragraph with text that doesn't exceed 120 characters per line.
When text is longer, break it across multiple lines.

Here's a list with proper spacing:

- First item
- Second item
- Third item

And here's a code block with language specified:

` ``bash
echo "Hello, World!"
` ``

Another paragraph follows with proper blank lines around it.
```

**Before editing any markdown file:**

1. Check the existing style and match it
2. Keep lines under 120 characters (break long lines)
3. Add blank lines around lists and code blocks
4. Always specify code block languages
5. Run `markdownlint --fix .` to auto-fix issues
6. Verify with `markdownlint .` before committing

### Import Organization

Imports should be organized in three groups separated by blank lines:

1. Standard library imports
2. External dependencies
3. Internal project imports

Example:

```go
import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/spf13/viper"

    "github.com/aauren/ntopng-exporter/internal/config"
    "github.com/aauren/ntopng-exporter/internal/ntopng"
)
```

### Naming Conventions

- **Package names**: Lowercase, single word (e.g., `config`, `ntopng`, `prometheus`)
- **Exported types**: PascalCase (e.g., `Controller`, `Config`)
- **Unexported types**: camelCase (e.g., `ntopHost`, `metricServe`)
- **Constants**: PascalCase or ALL_CAPS for exported, camelCase for unexported
- **Variables**: camelCase (e.g., `myConfig`, `stopChan`)
- **Functions**: PascalCase for exported, camelCase for unexported

### Type Definitions

- Use struct types for configuration and data structures
- Embed JSON tags for API responses: `` `json:"field_name"` ``
- Group related types together in the same file
- Use pointer receivers for methods that modify state
- Use value receivers for methods that only read state

### Error Handling

Always check and handle errors explicitly:

```go
// Good
if err != nil {
    return fmt.Errorf("descriptive context: %v", err)
}

// For main/critical paths, print and exit
if err != nil {
    fmt.Printf("error message: %v\n", err)
    os.Exit(1)
}
```

### Formatting

- Use `gofmt` for automatic formatting (tabs for indentation)
- Maximum line length: No hard limit, but keep reasonable (~120 chars)
- Use blank lines to separate logical sections
- Add comments for exported functions and complex logic

### Comments

- Export functions/types must have doc comments starting with the name
- Doc comments should be complete sentences
- Use `//` for single-line comments
- Place comments above the code they describe

### Concurrency

- Use channels for communication between goroutines
- Use `sync.RWMutex` for protecting shared data structures
- Always defer `Unlock()` immediately after `Lock()`
- Use read locks (`RLock()`) when only reading shared data

Example:

```go
c.ListRWMutex.RLock()
defer c.ListRWMutex.RUnlock()
// ... read operations
```

### Configuration

- Use Viper for configuration management
- Support multiple config file locations (home dir, /etc, ./config)
- Validate all configuration in a `validate()` method
- Set sensible defaults using `viper.SetDefault()`

### HTTP Requests

- Use custom HTTP clients with proper timeouts
- Always close response bodies: `defer resp.Body.Close()`
- Support multiple authentication methods (cookie, basic, token, none)
- Handle TLS configuration via `AllowUnsafeTLS` flag

### Dependency Management

**Allowed dependencies** (per depguard.yaml):

- Go standard library
- `github.com/aauren/ntopng-exporter/*` (internal packages)
- `github.com/prometheus/client_golang`
- `github.com/spf13/viper`

Do not add new dependencies without updating `depguard.yaml`.

## Project Structure

```text
ntopng-exporter/
├── config/                    # Configuration files
│   └── ntopng-exporter.yaml  # Sample config
├── internal/                  # Internal packages
│   ├── config/               # Configuration parsing
│   ├── metrics/
│   │   └── prometheus/       # Prometheus metrics implementation
│   └── ntopng/               # ntopng API client
├── resources/                 # Supporting resources
│   ├── grafana-dashboard.json
│   └── ntopng-exporter.service
└── ntopng-exporter.go        # Main entry point
```

## Release Process

Releases are automated via GitHub Actions:

- Tag with `v*` pattern triggers GoReleaser
- Builds for multiple platforms: linux, darwin, windows
- Architectures: amd64, arm64, arm, 386, s390x, ppc64le, riscv64
- Docker images published to Docker Hub: `aauren/ntopng-exporter`

## Environment Variables

- `NTOPNG_TOKEN`: Override token authentication (takes precedence over config file)

## Common Tasks

### Adding a New Metric

1. Define the metric in the appropriate collector (`host.go` or `interface.go`)
2. Add Prometheus descriptor in the `New*Collector()` function
3. Register the descriptor in `Describe()` method
4. Collect the metric in `Collect()` method
5. Update documentation

### Adding a New Scrape Target

1. Define constant in `internal/config/config.go`
2. Add to `AvailableScrapeTargets` map
3. Implement scraping logic in `internal/ntopng/controller.go`
4. Create collector in `internal/metrics/prometheus/`
5. Register collector in `ntopng-exporter.go`

## Commit Message Guidelines

This project follows [Conventional Commits](https://www.conventionalcommits.org/) specification. All commit messages
must follow this format:

```text
<type>(<scope>): <subject>

<body>

<footer>
```

### Commit Types

- **feat**: A new feature
- **fix**: A bug fix
- **docs**: Documentation only changes
- **style**: Changes that don't affect code meaning (formatting, etc.)
- **refactor**: Code change that neither fixes a bug nor adds a feature
- **perf**: Code change that improves performance
- **test**: Adding missing tests or correcting existing tests
- **build**: Changes that affect the build system or external dependencies
- **ci**: Changes to CI configuration files and scripts
- **chore**: Other changes that don't modify src or test files

### Commit Message Rules

- **Subject line**: Use imperative, present tense (e.g., "add" not "added")
- **Subject line**: Don't capitalize first letter, no period at end
- **Subject line**: Limit to 72 characters
- **Body**: Explain the "what" and "why" vs. "how"
- **Body**: Wrap at 72 characters
- **Footer**: Reference issues with `Fixes #123` or `Closes #456`
- **Breaking changes**: Start footer with `BREAKING CHANGE:` followed by description

### Examples

**Simple feature:**

```text
feat(prometheus): add interface throughput metrics
```

**Bug fix with body:**

```text
fix(config): validate interface names are not empty

Previously, blank interface names would cause a panic at runtime.
Add validation to reject empty names during config parsing.

Fixes #78
```

**Breaking change:**

```text
feat(config): rename authentication method field

BREAKING CHANGE: The config field `auth` has been renamed to `authMethod`.
Users must update their configuration files.
```

For more details, see the [Commit Message Guidelines](docs/development.md#commit-message-guidelines) in the
development guide.

## Development Tips

- The exporter runs continuously, scraping at configured intervals
- Graceful shutdown is handled via signal catching (SIGTERM, SIGINT)
- Interface IDs are cached on startup to avoid repeated lookups
- Host and interface lists are replaced entirely on each scrape
- Locks minimize hold time by building temp maps before swapping
