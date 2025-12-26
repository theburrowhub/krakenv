# Research: Krakenv CLI

**Date**: 2025-12-26  
**Branch**: `001-krakenv-cli`

## Technology Decisions

### 1. Go Version Selection

**Decision**: Go 1.23 (latest stable as of December 2025)

**Rationale**:
- Latest stable release with all modern features (generics, improved error handling)
- Excellent cross-compilation support for multi-platform distribution
- Single binary output with no runtime dependencies
- Strong standard library reduces external dependencies

**Alternatives Considered**:
- Go 1.22: Still supported but missing latest optimizations
- Go 1.21: Older, no significant benefit

### 2. TUI Framework: Bubbletea

**Decision**: `github.com/charmbracelet/bubbletea` v1.x

**Rationale**:
- The Elm Architecture (TEA) provides predictable state management
- Full-screen TUI with proper terminal handling
- Active maintenance by Charm (last release within 30 days)
- Rich ecosystem: lipgloss for styling, bubbles for components
- Built-in support for async operations (tea.Cmd)
- Excellent keyboard handling including Ctrl+C graceful shutdown

**Alternatives Considered**:
- `tview`: More traditional widget-based; less flexible styling
- `gocui`: Lower-level; more boilerplate required
- `termui`: Dashboard-focused; not ideal for wizards
- Simple `fmt.Scan`: Rejected per user requirement for full-screen TUI

### 3. CLI Framework: Cobra

**Decision**: `github.com/spf13/cobra` v1.x

**Rationale**:
- Industry standard for Go CLI applications (kubectl, docker, gh)
- Automatic help generation and shell completion
- Nested subcommands support
- Flag parsing with type validation
- Works seamlessly alongside Bubbletea (Cobra handles args, Bubbletea handles interaction)

**Alternatives Considered**:
- `urfave/cli`: Good but less adoption; Cobra has more examples
- `kong`: Struct-based; less flexible for complex scenarios
- Standard `flag`: Too basic for subcommands

### 4. Styling: Lipgloss

**Decision**: `github.com/charmbracelet/lipgloss` v1.x

**Rationale**:
- Native integration with Bubbletea
- CSS-like styling API (padding, borders, colors)
- Terminal capability detection (true color vs 256 color)
- Responsive layouts for different terminal sizes

**Alternatives Considered**:
- ANSI escape codes directly: Too low-level; error-prone
- `termenv`: Lower-level than lipgloss; lipgloss uses it internally

### 5. Build System: Makefile

**Decision**: GNU Make with standard targets

**Rationale**:
- Universal availability on Unix systems
- Simple, well-understood syntax
- Standard targets: `make build`, `make test`, `make lint`
- Easy integration with CI/CD

**Standard Targets**:

```makefile
.PHONY: build test lint clean install

build:           # Compile binary
test:            # Run all tests
test-unit:       # Run unit tests only
test-integration: # Run integration tests
lint:            # Run golangci-lint
fmt:             # Format code with gofmt
clean:           # Remove build artifacts
install:         # Install to GOPATH/bin
dev:             # Run with live reload (air)
```

### 6. CI/CD: GitHub Actions + GoReleaser

**Decision**: GitHub Actions for CI; GoReleaser for releases

**Rationale**:
- GitHub Actions: Native integration, free for open source
- GoReleaser: Standard tool for Go binary distribution
  - Automatic changelog from conventional commits
  - Multi-platform builds (Linux, macOS, Windows × amd64, arm64)
  - Checksums and signatures
  - GitHub Releases integration
  - Homebrew tap generation (optional)

**Pipeline Design**:

1. **CI (on PR/push to any branch)**:
   - `go test ./...`
   - `golangci-lint run`
   - Build verification

2. **Release (on push to main)**:
   - Generate semantic version tag from commits
   - Run GoReleaser
   - Upload binaries to GitHub Releases
   - Update GitHub Pages documentation

### 7. Testing Strategy

**Decision**: Go standard testing + testify

**Rationale**:
- `testing` package is built-in and well-supported
- `testify/assert` provides readable assertions
- `testify/require` for fatal assertions
- Table-driven tests for parser/validator coverage

**Test Structure**:

| Layer | Location | Focus |
|-------|----------|-------|
| Unit | `internal/*_test.go` | Pure functions, edge cases |
| Integration | `tests/integration/` | File I/O, end-to-end flows |
| Contract | `tests/contract/` | CLI interface stability |

### 8. Annotation Parser Design

**Decision**: Hand-written recursive descent parser

**Rationale**:
- Annotation syntax is simple enough to not require parser generators
- Full control over error messages (user-friendly feedback is critical)
- No external dependency for parsing
- Easy to extend with new modifiers

**Grammar** (informal):

```
annotation    := "#prompt:" prompt_text "|" type_spec (";" modifier)*
prompt_text   := [^|]+
type_spec     := "int" | "numeric" | "string" | "enum" | "boolean" | "object"
modifier      := modifier_name (":" modifier_value)?
modifier_name := "min" | "max" | "minlen" | "maxlen" | "pattern" | "options" | "format" | "optional" | "secret"
modifier_value := [^;]+
```

### 9. Documentation Site

**Decision**: Static HTML/CSS in `/docs` for GitHub Pages

**Rationale**:
- No build step required (Jekyll, Hugo, etc. add complexity)
- Simple markdown files with minimal styling
- GitHub Pages serves from `/docs` on main branch
- Fast load times; no JavaScript required

**Structure**:

```
docs/
├── index.html           # Landing page with tagline
├── getting-started.md   # Installation + first use
├── annotation-syntax.md # Complete annotation reference
├── commands.md          # CLI command reference
└── assets/
    └── styles.css       # Minimal dark theme
```

## Dependencies Summary

| Package | Version | Purpose | Last Activity |
|---------|---------|---------|---------------|
| `charmbracelet/bubbletea` | v1.2+ | TUI framework | Active (weekly) |
| `charmbracelet/lipgloss` | v1.0+ | TUI styling | Active (weekly) |
| `charmbracelet/bubbles` | v0.20+ | TUI components | Active (weekly) |
| `spf13/cobra` | v1.8+ | CLI framework | Active (monthly) |
| `stretchr/testify` | v1.9+ | Test assertions | Active (monthly) |
| `gopkg.in/yaml.v3` | v3.0+ | YAML parsing | Stable (quarterly) |

All dependencies are actively maintained with releases in the past 90 days.

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Bubbletea breaking changes | Pin to major version; review changelog before updating |
| Terminal compatibility (Windows) | Test on Windows Terminal, PowerShell; use Bubbletea's cross-platform support |
| Large file performance | Lazy loading; stream parsing for files >1000 lines |
| Complex regex patterns | Validate regex at parse time; provide clear error on invalid patterns |

