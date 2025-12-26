# Implementation Plan: Krakenv CLI

**Branch**: `001-krakenv-cli` | **Date**: 2025-12-26 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `/specs/001-krakenv-cli/spec.md`

## Summary

Krakenv is a CLI tool for managing environment variable files (`.env`) with an annotation-based configuration wizard system. Built as a full-screen TUI application using Go and Bubbletea, it transforms the tedious process of configuring environment files into a guided, validated experience. The tool parses `.env.dist` templates with inline annotations (`#prompt:...|type;constraints`) and generates environment-specific files (`.env.local`, `.env.testing`, etc.) through an interactive wizard or non-interactive validation mode for CI/CD pipelines.

## Technical Context

**Language/Version**: Go 1.23 (latest stable)  
**Primary Dependencies**:
- `github.com/charmbracelet/bubbletea` - TUI framework for full-screen interactive experience
- `github.com/charmbracelet/lipgloss` - Styling for TUI components
- `github.com/charmbracelet/bubbles` - Pre-built TUI components (text input, spinners, etc.)
- `github.com/spf13/cobra` - CLI command/subcommand management
- `gopkg.in/yaml.v3` - YAML parsing for object type validation
- `github.com/go-playground/validator/v10` - Struct validation (optional, for complex rules)

**Storage**: Local filesystem only (`.env` files)  
**Testing**: Go standard testing + `github.com/stretchr/testify` for assertions  
**Target Platform**: Cross-platform CLI (Linux, macOS, Windows)  
**Project Type**: Single project (CLI tool)  
**Performance Goals**: Parse 1000-line `.env.dist` in <100ms; wizard response <50ms per keystroke  
**Constraints**: Single binary distribution; no external runtime dependencies  
**Scale/Scope**: Files up to 500 variables; typical use 10-100 variables

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|----------|
| **I. Code Quality First** | ✅ PASS | Go enforces single responsibility via packages; pure functions for parsing/validation; gofmt for formatting |
| **II. Test-Driven Development** | ✅ PASS | Unit tests for parser/validator; integration tests for file operations; contract tests for CLI interface |
| **III. User Experience Consistency** | ✅ PASS | Bubbletea provides consistent TUI patterns; lipgloss for unified styling; keyboard navigation native |
| **IV. Performance Requirements** | ✅ PASS | Go compiles to native binary; Bubbletea optimized for terminal rendering; no network calls |

### Quality Standards Alignment

- **Linting**: `golangci-lint` with strict ruleset
- **Formatting**: `gofmt` (enforced in CI)
- **Type Safety**: Go is statically typed; no `interface{}` abuse
- **Documentation**: GoDoc for public APIs; README for users

### Security Standards Alignment

- **Input Validation**: All annotation syntax validated before processing
- **Secrets Management**: `secret` modifier hides input and values; no logging of secret values

## Project Structure

### Documentation (this feature)

```text
specs/001-krakenv-cli/
├── plan.md              # This file
├── research.md          # Phase 0: Technology decisions
├── data-model.md        # Phase 1: Entity definitions
├── quickstart.md        # Phase 1: Developer guide
├── contracts/           # Phase 1: CLI interface specification
│   └── cli-spec.md
└── tasks.md             # Phase 2: Implementation tasks
```

### Source Code (repository root)

```text
cmd/
└── krakenv/
    └── main.go              # Entry point

internal/
├── parser/
│   ├── lexer.go             # Tokenize .env files
│   ├── parser.go            # Parse variables and annotations
│   └── parser_test.go
├── validator/
│   ├── types.go             # Type definitions (int, string, enum, etc.)
│   ├── validator.go         # Validation logic
│   └── validator_test.go
├── generator/
│   ├── generator.go         # File generation logic
│   └── generator_test.go
├── inspector/
│   ├── inspector.go         # Diff between dist and env files
│   └── inspector_test.go
├── config/
│   ├── config.go            # Krakenv configuration block parser
│   └── config_test.go
└── tui/
    ├── app.go               # Main Bubbletea application
    ├── wizard/
    │   ├── model.go         # Wizard state machine
    │   ├── view.go          # Wizard rendering
    │   └── wizard_test.go
    ├── inspect/
    │   ├── model.go         # Inspection view state
    │   └── view.go          # Inspection rendering
    └── components/
        ├── input.go         # Text input with validation
        ├── select.go        # Enum selection
        └── styles.go        # Lipgloss styles

pkg/
└── envfile/
    ├── envfile.go           # Public API for env file operations
    └── envfile_test.go      # Can be imported by other tools

tests/
├── integration/
│   ├── generate_test.go     # End-to-end generate tests
│   ├── validate_test.go     # End-to-end validate tests
│   └── inspect_test.go      # End-to-end inspect tests
├── testdata/
│   ├── valid/               # Valid .env.dist samples
│   └── invalid/             # Invalid samples for error testing
└── contract/
    └── cli_test.go          # CLI interface contract tests

docs/                        # GitHub Pages static site
├── index.html
├── getting-started.md
├── annotation-syntax.md
├── commands.md
└── assets/
    └── styles.css

.github/
└── workflows/
    ├── ci.yml               # Test + lint on PR
    └── release.yml          # GoReleaser on push to main

Makefile                     # Development commands
.goreleaser.yml              # Release configuration
go.mod
go.sum
README.md
LICENSE
```

**Structure Decision**: Single project with `internal/` for private packages, `pkg/` for importable library, `cmd/` for entry point. This follows Go best practices for CLI tools that may expose a library API.

## Complexity Tracking

> No constitution violations. Structure follows Go standard project layout.

| Decision | Rationale |
|----------|-----------|
| `internal/` vs `pkg/` split | Core logic is internal; env file parsing exported for potential library use |
| Bubbletea full-screen TUI | User requested; provides superior UX over simple line prompts |
| Separate `tui/wizard/` and `tui/inspect/` | Different UI flows warrant isolation; easier testing |
