# Feature Specification: Krakenv CLI

**Feature Branch**: `001-krakenv-cli`  
**Created**: 2025-12-26  
**Status**: Draft  
**Tagline**: "When envs get complex, release the krakenv"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Generate Environment File with Interactive Wizard (Priority: P1)

A developer wants to create a new `.env.local` file from the project's `.env.dist` template. The distributable contains annotated variables that guide the configuration process. When a variable lacks a value or has an invalid one, krakenv prompts the user with a contextual question based on the annotation, validates the input, and writes the final file.

**Why this priority**: This is the core value proposition - transforming a tedious manual copy-paste process into a guided, error-proof configuration experience.

**Independent Test**: Can be fully tested by running `krakenv generate .env.local` on a `.env.dist` with annotated variables, completing the wizard prompts, and verifying the output file contains valid values.

**Acceptance Scenarios**:

1. **Given** a `.env.dist` file with annotated variables, **When** the user runs `krakenv generate .env.local`, **Then** the system prompts for each undefined or invalid variable using the annotation's prompt text
2. **Given** a variable with `#prompt:Question?|int;min:1;max:12`, **When** the user enters a value outside the range, **Then** the system displays a friendly error message explaining the valid range and re-prompts
3. **Given** a variable with a preset value in `.env.dist`, **When** the user is prompted, **Then** the preset value is offered as the default (user can press Enter to accept)
4. **Given** an existing `.env.local` with some valid values, **When** the user runs `krakenv generate .env.local`, **Then** only missing or invalid variables trigger prompts

---

### User Story 2 - Inspect Missing and Extra Variables (Priority: P2)

A developer wants to identify discrepancies between the distributable and environment files. They need to find: (a) variables defined in `.env.dist` but missing from `.env.local`, and (b) variables present in `.env.local` that don't exist in `.env.dist`. The tool provides a clear report and offers to fix discrepancies interactively.

**Why this priority**: After initial setup, ongoing maintenance requires keeping environment files in sync with evolving requirements. This prevents "works on my machine" issues.

**Independent Test**: Can be tested by creating deliberate mismatches between `.env.dist` and `.env.local`, running inspection commands, and verifying the report accuracy.

**Acceptance Scenarios**:

1. **Given** a `.env.dist` with `VAR_A, VAR_B, VAR_C` and a `.env.local` with only `VAR_A, VAR_B`, **When** the user runs `krakenv inspect .env.local`, **Then** the output shows `VAR_C` as missing with its annotation description
2. **Given** a `.env.local` with an extra `LEGACY_VAR` not in `.env.dist`, **When** the user runs `krakenv inspect .env.local`, **Then** the output shows `LEGACY_VAR` as an ad-hoc variable with option to add it to distributable
3. **Given** inspection results showing discrepancies, **When** the user chooses to sync a variable, **Then** the system either adds it to the target file (with wizard if needed) or to the distributable

---

### User Story 3 - Validate Existing Environment File (Priority: P2)

A developer wants to validate that all values in an existing `.env.local` comply with the annotations in `.env.dist`. This is useful for CI/CD pipelines or pre-commit hooks to catch configuration errors early.

**Why this priority**: Validation prevents runtime errors caused by misconfigured environment variables.

**Independent Test**: Can be tested by running `krakenv validate .env.local` against various valid and invalid configurations and checking exit codes and error messages.

**Acceptance Scenarios**:

1. **Given** a `.env.local` where all values match their annotation constraints, **When** the user runs `krakenv validate .env.local`, **Then** the system exits with code 0 and displays a success message
2. **Given** a `.env.local` with `PORT=abc` but annotation `#prompt:Port?|int;min:1;max:65535`, **When** the user runs `krakenv validate .env.local`, **Then** the system shows a clear error: "PORT: Expected integer between 1-65535, got 'abc'" and exits with non-zero code
3. **Given** multiple validation errors, **When** the user runs validation, **Then** all errors are listed together with line numbers and fix suggestions

---

### User Story 4 - Add New Variable to Distributable (Priority: P3)

A developer needs to add a new environment variable to the project. They want to define it with proper annotations so all team members get guided configuration.

**Why this priority**: Enables the team to maintain the distributable as the single source of truth for environment configuration.

**Independent Test**: Can be tested by running `krakenv add DB_HOST` with annotation options and verifying the `.env.dist` is updated correctly.

**Acceptance Scenarios**:

1. **Given** an existing `.env.dist`, **When** the user runs `krakenv add DB_HOST --type string --prompt "Database hostname?"`, **Then** the variable is appended to `.env.dist` with the proper annotation format
2. **Given** the user specifies `--default localhost`, **When** the variable is added, **Then** it appears as `DB_HOST=localhost #prompt:Database hostname?|string`
3. **Given** the user adds a variable with `--type enum --options "dev,staging,prod"`, **When** added, **Then** the annotation includes enum validation: `#prompt:Environment?|enum;options:dev,staging,prod`

---

### User Story 5 - Configure Distributable Settings (Priority: P3)

A project maintainer wants to configure krakenv behavior at the project level, such as which environment files to check, naming conventions, and default validation strictness. These settings are embedded in the distributable file itself.

**Why this priority**: Project-level configuration ensures consistent behavior across the team without requiring per-developer setup.

**Independent Test**: Can be tested by adding a configuration block to `.env.dist` and verifying krakenv respects those settings.

**Acceptance Scenarios**:

1. **Given** a `.env.dist` with a config block `#krakenv:environments=local,testing,production`, **When** the user runs `krakenv generate --all`, **Then** the system generates `.env.local`, `.env.testing`, and `.env.production`
2. **Given** config `#krakenv:strict=true`, **When** validation runs, **Then** any variable without annotation is flagged as an error
3. **Given** no config block exists, **When** krakenv runs, **Then** sensible defaults are used (single environment, non-strict mode)

---

### Edge Cases

- What happens when `.env.dist` doesn't exist? → Clear error: "No distributable found. Create .env.dist or specify path with --dist"
- What happens when a variable annotation has syntax errors? → Warning message pointing to the line, variable treated as untyped string
- What happens when user presses Ctrl+C during wizard? → Graceful exit, no partial file written, clear message about incomplete operation
- What happens with multiline values (e.g., certificates)? → Use `encoding:base64` or `encoding:heredoc` constraint; annotation determines format
- What happens with empty values vs undefined variables? → Empty string is valid for optional strings; undefined triggers prompt
- What happens when distributable has duplicate variable names? → Warning about duplicate, last definition wins

## Requirements *(mandatory)*

### Functional Requirements

#### Core Parsing & Generation

- **FR-001**: System MUST parse `.env.dist` files preserving comments and whitespace structure
- **FR-002**: System MUST extract annotations from inline comments following the formal syntax below

**Annotation Syntax (EBNF)**:
```ebnf
annotation       = "#prompt:" message "|" type { ";" modifier } ;
message          = { character - "|" } ;
type             = "int" | "numeric" | "string" | "enum" | "boolean" | "object" ;
modifier         = constraint | flag ;
flag             = "optional" | "secret" ;
constraint       = constraint_name ":" constraint_value ;
constraint_name  = "min" | "max" | "minlen" | "maxlen" | "pattern" | "options" | "format" | "encoding" ;
constraint_value = { character - ";" } ;
```

**Examples**:
- `#prompt:How many workers?|int;min:1;max:16`
- `#prompt:API Key?|string;secret;minlen:32`
- `#prompt:Environment?|enum;options:dev,staging,prod`
- `#prompt:Config?|object;format:json;encoding:base64`
- **FR-003**: System MUST generate environment files (`.env.<name>`) from the distributable template
- **FR-004**: System MUST preserve non-annotated comments and blank lines in generated files

#### Type System & Validation

- **FR-005**: System MUST support these variable types: `int`, `numeric`, `string`, `enum`, `boolean`, `object`
- **FR-006**: System MUST validate `int` type with optional `min` and `max` constraints
- **FR-007**: System MUST validate `numeric` type (integers and decimals) with optional `min` and `max` constraints
- **FR-008**: System MUST validate `string` type with optional `minlen`, `maxlen`, and `pattern` (regex) constraints
- **FR-009**: System MUST validate `enum` type against a defined list of allowed values (`options:val1,val2,val3`)
- **FR-010**: System MUST validate `boolean` type accepting: true/false, yes/no, 1/0, on/off (case-insensitive)
- **FR-011**: System MUST support `object` type with `format:json` or `format:yaml` for structured data
- **FR-011a**: System MUST support `encoding:base64` constraint for multiline values (certificates, keys)
- **FR-011b**: System MUST support `encoding:heredoc` constraint for readable multiline values
- **FR-011c**: When `encoding:base64` is set, system MUST decode for validation and store encoded value
- **FR-011d**: When `encoding:heredoc` is set, system MUST use `"""` delimiters for multiline input/output
- **FR-012**: System MUST support `optional` modifier in annotations (e.g., `#prompt:...|string;optional`)
- **FR-013**: Optional variables MAY be left empty without triggering validation errors
- **FR-014**: Required variables (no `optional` modifier) MUST have a valid value or validation fails
- **FR-015a**: System MUST support `secret` modifier for sensitive variables (e.g., `#prompt:...|string;secret`)
- **FR-015b**: When prompting for `secret` variables, input MUST be hidden (like password fields)
- **FR-015c**: System MUST NOT display current values of `secret` variables in any output

#### Interactive Wizard

- **FR-016**: System MUST display annotation's prompt message when requesting user input
- **FR-017**: System MUST show type constraints in the prompt (e.g., `[int: 1-12]`, `[enum: dev|staging|prod]`)
- **FR-018**: System MUST use preset value from distributable as default when available
- **FR-019**: System MUST re-prompt with clear error message when validation fails
- **FR-020**: System MUST skip prompting for variables that already have valid values in target file

#### Inspection & Sync

- **FR-021**: System MUST identify variables in distributable missing from environment file
- **FR-022**: System MUST identify variables in environment file not present in distributable
- **FR-023**: System MUST offer interactive option to add ad-hoc variables back to distributable
- **FR-024**: System MUST offer interactive option to add missing variables to environment file

#### Configuration & Commands

- **FR-025**: System MUST support configuration block in distributable via `#krakenv:KEY=VALUE` comments
- **FR-026**: System MUST provide `generate` command to create/update environment files
- **FR-027**: System MUST provide `validate` command to check existing files without modification
- **FR-028**: System MUST provide `inspect` command to compare distributable and environment files
- **FR-029**: System MUST provide `add` command to append new annotated variables to distributable
- **FR-029a**: System MUST provide `init` command to create a new distributable file via interactive wizard
- **FR-029b**: The `init` wizard MUST iterate prompting user for variable name, type, constraints, default value, and prompt text
- **FR-029c**: The `init` command MUST convert wizard responses into properly formatted variable lines with annotations

#### Non-Interactive Mode

- **FR-030**: System MUST support `--non-interactive` flag for CI/CD pipeline execution
- **FR-031**: In non-interactive mode, system MUST fail with descriptive error listing all variables that lack valid values
- **FR-032**: In non-interactive mode, system MUST NOT prompt for user input; exit immediately with non-zero code if unresolved variables exist
- **FR-032a**: When `--non-interactive` is combined with `inspect --sync`, system MUST auto-resolve using default values for optional variables
- **FR-032b**: When `--non-interactive` + `--sync` cannot resolve required variables without user input, system MUST exit with error listing unresolvable items

#### Error Handling

- **FR-033**: System MUST display user-friendly error messages with ALL four components:
  - **Problem**: What failed exactly (e.g., "Expected integer between 1-65535")
  - **Location**: Line number and variable name affected
  - **Suggestion**: How to fix it (e.g., "Enter a valid port number")
  - **Example**: A valid reference value (e.g., "5432")
- **FR-034**: System MUST exit with appropriate status codes (0 for success, non-zero for errors)
- **FR-035**: System MUST handle graceful interruption (Ctrl+C) by prompting: "Discard changes or save progress?" before exiting
- **FR-035a**: On Ctrl+C, if user chooses "discard", no files are written and temp files are cleaned up
- **FR-035b**: On Ctrl+C, if user chooses "save", completed variables are written to target file

#### File System Handling

- **FR-036**: System MUST display clear error when file permissions prevent read/write: "Cannot write to {path}: Permission denied. Check file permissions."
- **FR-037**: System MUST use streaming/pagination in TUI for files with >500 variables
- **FR-038**: Variable names MUST match ASCII pattern `^[A-Z][A-Z0-9_]*$` (no Unicode)
- **FR-039**: When `--dist` points to a directory, system MUST look for `.env.dist` inside that directory
- **FR-040**: Variable values with only whitespace MUST be treated as empty (trimmed)

#### Annotation Parsing Edge Cases

- **FR-041**: Unknown constraints in annotations MUST trigger a warning and be ignored (not error)
- **FR-042**: Enum type with empty options list (`options:`) MUST be treated as string without restrictions

#### Output Control

- **FR-043**: System MUST support `--keep-annotations` flag to preserve annotations in generated files
- **FR-043a**: By default (no flag), annotations are stripped from generated environment files
- **FR-043b**: With `--keep-annotations`, full annotation comments are preserved for reference

### Non-Functional Requirements

#### Performance

- **NFR-001**: File parsing MUST complete in <100ms for files with up to 100 variables
- **NFR-002**: TUI keystroke response MUST be <50ms

#### Platform Support

- **NFR-003**: System MUST support Linux (glibc-based distributions)
- **NFR-004**: System MUST support macOS 10.15 (Catalina) and later
- **NFR-005**: System MUST support Windows 10 and later

#### Accessibility

- **NFR-006**: TUI MUST be fully navigable via keyboard only
- **NFR-007**: TUI MUST support screen readers via proper ARIA-like semantics where possible
- **NFR-008**: TUI MUST respect `NO_COLOR` environment variable for colorblind users

#### Debugging

- **NFR-009**: System MUST support `--verbose` flag for detailed debug output
- **NFR-010**: Verbose output MUST include file paths, parse timings, and validation steps

### Key Entities

- **Distributable**: The template file (`.env.dist`) containing variable definitions with optional annotations and configuration block
- **Environment File**: A generated file (`.env.local`, `.env.testing`, etc.) containing resolved variable values for a specific environment
- **Annotation**: Metadata attached to a variable as an inline comment, defining prompt message, type, and constraints
- **Variable**: A key-value pair with optional annotation; key is the identifier, value may be preset or user-provided
- **Configuration Block**: Special comments in distributable that define krakenv behavior (environments, strictness, etc.)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can generate a complete environment file from distributable in under 5 minutes for files with up to 50 variables
- **SC-002**: 95% of validation errors are resolved on first attempt after reading the error message (message clarity)
- **SC-003**: Zero data loss when interrupted mid-operation - original files remain intact
- **SC-004**: Users identify and sync discrepancies between files in under 2 minutes per environment
- **SC-005**: New team members can configure their local environment without asking questions if distributable is properly annotated
- **SC-006**: All six supported types (int, numeric, string, enum, boolean, object) validate correctly with their constraints
- **SC-007**: Tool provides actionable feedback - every error message includes what went wrong and how to fix it

## Clarifications

### Session 2025-12-26

- Q: ¿Cómo debe comportarse krakenv en entornos no interactivos (CI/CD)? → A: Flag `--non-interactive` que falla con error si hay variables sin valor válido
- Q: ¿Debe existir distinción entre variables obligatorias y opcionales? → A: Sí, añadir modificador `optional` en anotación (ej: `#prompt:...|string;optional`)
- Q: ¿Debe krakenv tener consideraciones especiales para variables sensibles? → A: Sí, modificador `secret` que oculta input en wizard y no muestra valor actual

## Assumptions

- The `.env.dist` file follows standard `.env` format: `KEY=VALUE` per line, `#` for comments
- Users have terminal access and can respond to interactive prompts
- File system permissions allow reading distributable and writing environment files
- The annotation format `#prompt:...|type;constraint:value` is unique enough to not conflict with regular comments
- UTF-8 encoding is used for all files
