# Quickstart: Krakenv Development

**Date**: 2025-12-26  
**Branch**: `001-krakenv-cli`

## Prerequisites

- **Go 1.23+**: [Download](https://go.dev/dl/)
- **Make**: Pre-installed on macOS/Linux; Windows via chocolatey or WSL
- **Git**: For version control

Verify installation:

```bash
go version    # go1.23.x
make --version # GNU Make 4.x
```

## Project Setup

### 1. Clone and Initialize

```bash
git clone https://github.com/youruser/krakenv.git
cd krakenv
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Development Tools

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install air for live reload (optional)
go install github.com/cosmtrek/air@latest
```

## Build & Run

### Build Binary

```bash
make build
# Output: ./bin/krakenv
```

### Run Directly

```bash
go run ./cmd/krakenv generate .env.local
```

### Development Mode (Live Reload)

```bash
make dev
# Uses air to rebuild on file changes
```

## Testing

### Run All Tests

```bash
make test
```

### Run Specific Test Suites

```bash
# Unit tests only
make test-unit

# Integration tests only
make test-integration

# With coverage report
make test-coverage
# Open coverage.html in browser
```

### Run Single Test

```bash
go test -v -run TestParseAnnotation ./internal/parser/
```

## Code Quality

### Lint

```bash
make lint
# Runs golangci-lint with project config
```

### Format

```bash
make fmt
# Applies gofmt to all Go files
```

### Pre-commit Check (All)

```bash
make check
# Runs: fmt, lint, test
```

## Project Structure Overview

```
cmd/krakenv/main.go     # Entry point - start here
internal/
  parser/               # .env file parsing
  validator/            # Type validation logic
  generator/            # File generation
  inspector/            # Diff analysis
  config/               # Krakenv config parsing
  tui/                  # Bubbletea UI
    wizard/             # Interactive prompts
    inspect/            # Inspection view
    components/         # Reusable UI pieces
pkg/envfile/            # Public library API
tests/
  integration/          # End-to-end tests
  testdata/             # Sample .env files
  contract/             # CLI contract tests
```

## Development Workflow

### TDD Cycle

1. **Write failing test** in `*_test.go`

```go
func TestParseAnnotation_IntType(t *testing.T) {
    input := "#prompt:How many?|int;min:1;max:10"
    ann, err := ParseAnnotation(input)
    
    require.NoError(t, err)
    assert.Equal(t, "How many?", ann.PromptText)
    assert.Equal(t, TypeInt, ann.Type)
}
```

2. **Run test** (should fail):

```bash
go test -v -run TestParseAnnotation_IntType ./internal/parser/
```

3. **Implement** minimal code to pass

4. **Run test** (should pass)

5. **Refactor** if needed

### Adding a New Command

1. Create command in `cmd/krakenv/`:

```go
// cmd/krakenv/cmd_newcmd.go
var newcmdCmd = &cobra.Command{
    Use:   "newcmd",
    Short: "Description",
    RunE:  runNewCmd,
}

func init() {
    rootCmd.AddCommand(newcmdCmd)
}

func runNewCmd(cmd *cobra.Command, args []string) error {
    // Implementation
    return nil
}
```

2. Add tests in `tests/contract/`

3. Update CLI spec in `specs/001-krakenv-cli/contracts/cli-spec.md`

### Adding a New Validator Type

1. Add type constant in `internal/validator/types.go`
2. Implement validation in `internal/validator/validator.go`
3. Add TUI input component in `internal/tui/components/`
4. Write tests for all edge cases

## Debugging

### Enable Debug Logging

```bash
KRAKENV_DEBUG=1 go run ./cmd/krakenv generate .env.local
```

### Debug TUI

Bubbletea can be tricky to debug. Use logging to file:

```go
import "log"

func init() {
    f, _ := os.Create("debug.log")
    log.SetOutput(f)
}

// Then in your code:
log.Printf("State: %+v", m)
```

### Run with Delve

```bash
dlv debug ./cmd/krakenv -- generate .env.local
```

## Common Tasks

### Create Test Fixtures

Add sample files to `tests/testdata/`:

```
tests/testdata/
  valid/
    simple.env.dist      # Basic variables
    all-types.env.dist   # All type examples
    with-config.env.dist # Has krakenv config block
  invalid/
    bad-annotation.env.dist
    duplicate-vars.env.dist
```

### Update Dependencies

```bash
go get -u ./...
go mod tidy
```

### Generate Mocks (if needed)

```bash
go install github.com/vektra/mockery/v2@latest
mockery --all --dir ./internal --output ./internal/mocks
```

## Release Process

Releases are automated via GitHub Actions:

1. Push to `main` branch
2. CI runs tests and linting
3. GoReleaser creates:
   - Git tag (semantic version from commits)
   - GitHub Release with binaries
   - Changelog from conventional commits

### Manual Release (local testing)

```bash
# Dry run
goreleaser release --snapshot --clean

# Check output in dist/
ls -la dist/
```

## Useful Make Targets

```bash
make help           # Show all targets
make build          # Build binary
make test           # Run all tests
make lint           # Run linter
make fmt            # Format code
make check          # Pre-commit checks
make dev            # Live reload development
make clean          # Remove build artifacts
make install        # Install to $GOPATH/bin
make release-dry    # Test release locally
```

## Getting Help

- **Code questions**: Check existing tests for patterns
- **Bubbletea**: [Charm docs](https://charm.sh/docs/)
- **Cobra**: [Cobra docs](https://cobra.dev/)
- **Go testing**: [Go testing guide](https://go.dev/doc/tutorial/add-a-test)

## Next Steps

After setup, start with:

1. Read `specs/001-krakenv-cli/spec.md` for requirements
2. Review `data-model.md` for entity structure
3. Check `contracts/cli-spec.md` for command interface
4. Run `make test` to ensure everything works
5. Pick a task from `tasks.md` (after it's generated)

