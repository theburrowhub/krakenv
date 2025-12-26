# Tasks: Krakenv CLI

**Input**: Design documents from `/specs/001-krakenv-cli/`  
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/cli-spec.md ✅

**Tests**: TDD is REQUIRED per project constitution. Tests written FIRST, must FAIL before implementation.

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story (US1, US2, US3, US4, US5)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Project initialization and basic structure

- [x] T001 Initialize Go module with `go mod init github.com/user/krakenv` in project root
- [x] T002 Create project directory structure per plan.md (`cmd/`, `internal/`, `pkg/`, `tests/`)
- [x] T003 [P] Create Makefile with targets: build, test, lint, fmt, clean, install in `Makefile`
- [x] T004 [P] Create `.goreleaser.yml` with multi-platform build configuration
- [x] T005 [P] Create `.github/workflows/ci.yml` for test + lint on PR
- [x] T006 [P] Create `.github/workflows/release.yml` for GoReleaser on push to main
- [x] T007 [P] Create `.golangci.yml` with strict linting rules
- [x] T008 [P] Create README.md with project overview and installation instructions
- [x] T009 [P] Create LICENSE file (MIT or Apache 2.0)
- [x] T010 Add dependencies to go.mod: bubbletea, lipgloss, bubbles, cobra, yaml.v3, testify

**Checkpoint**: Project compiles with `go build ./...`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST complete before ANY user story

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Core Types & Models

- [x] T011 Create VariableType enum in `internal/parser/types.go`
- [x] T012 [P] Create Constraint struct in `internal/parser/types.go`
- [x] T013 [P] Create Annotation struct in `internal/parser/types.go`
- [x] T014 [P] Create Variable struct in `internal/parser/types.go`
- [x] T015 [P] Create EnvFile struct in `internal/parser/types.go`
- [x] T016 [P] Create KrakenvConfig struct in `internal/config/config.go`
- [x] T017 [P] Create ValidationError struct in `internal/validator/errors.go`
- [x] T018 [P] Create ErrorType enum in `internal/validator/errors.go`

### Parser (Required by US1, US2, US3, US4, US5)

- [x] T019 Write unit tests for lexer tokenization in `internal/parser/lexer_test.go`
- [x] T020 Implement lexer to tokenize .env lines in `internal/parser/lexer.go`
- [x] T021 Write unit tests for annotation parsing in `internal/parser/parser_test.go`
- [x] T022 Implement annotation parser (#prompt:...|type;constraints) in `internal/parser/parser.go`
- [x] T023 Write unit tests for full env file parsing in `internal/parser/parser_test.go`
- [x] T024 Implement EnvFile parser preserving comments/whitespace in `internal/parser/parser.go`

### Validator (Required by US1, US2, US3)

- [x] T025 Write unit tests for int validation with min/max in `internal/validator/validator_test.go`
- [x] T026 [P] Write unit tests for numeric validation in `internal/validator/validator_test.go`
- [x] T027 [P] Write unit tests for string validation (minlen/maxlen/pattern) in `internal/validator/validator_test.go`
- [x] T028 [P] Write unit tests for enum validation in `internal/validator/validator_test.go`
- [x] T029 [P] Write unit tests for boolean validation in `internal/validator/validator_test.go`
- [x] T030 [P] Write unit tests for object validation (json/yaml) in `internal/validator/validator_test.go`
- [x] T031 Implement type validators (int, numeric, string, enum, boolean, object) in `internal/validator/validator.go`
- [x] T032 Implement optional and secret modifier handling in `internal/validator/validator.go`

### Config Parser (Required by US5, used by all)

- [x] T033 Write unit tests for krakenv config block parsing in `internal/config/config_test.go`
- [x] T034 Implement config block parser (#krakenv:key=value) in `internal/config/config.go`

### CLI Foundation (Cobra setup)

- [x] T035 Create root command with global flags in `cmd/krakenv/root.go`
- [x] T036 [P] Create version command in `cmd/krakenv/version.go`
- [x] T037 Create main.go entry point in `cmd/krakenv/main.go`

### TUI Foundation (Bubbletea setup)

- [x] T038 Create lipgloss styles (colors, borders, typography) in `internal/tui/components/styles.go`
- [x] T039 [P] Create base text input component with validation in `internal/tui/components/input.go`
- [x] T040 [P] Create enum select component in `internal/tui/components/select.go`
- [x] T041 [P] Create password input component (hidden) in `internal/tui/components/password.go`
- [x] T042 Create main TUI app wrapper in `internal/tui/app.go`

### Test Data

- [x] T043 [P] Create valid .env.dist samples in `tests/testdata/valid/`
- [x] T044 [P] Create invalid .env.dist samples in `tests/testdata/invalid/`

**Checkpoint**: Foundation ready - `go test ./internal/...` passes; user story implementation can begin

---

## Phase 3: User Story 1 - Generate Environment File (Priority: P1) 🎯 MVP

**Goal**: Developer creates `.env.local` from `.env.dist` via interactive wizard

**Independent Test**: Run `krakenv generate .env.local` with annotated `.env.dist`, complete prompts, verify output

### Tests for US1

- [ ] T045 [P] [US1] Write contract test for generate command flags in `tests/contract/cli_test.go`
- [ ] T046 [P] [US1] Write integration test for full generate flow in `tests/integration/generate_test.go`
- [x] T047 [P] [US1] Write unit test for generator file output in `internal/generator/generator_test.go`

### Implementation for US1

- [x] T048 [US1] Implement Generator struct with file writing logic in `internal/generator/generator.go`
- [x] T049 [US1] Implement wizard model (state machine) in `internal/tui/wizard/model.go`
- [x] T050 [US1] Implement wizard view (prompt rendering) in `internal/tui/wizard/view.go`
- [x] T051 [US1] Implement wizard update (input handling, validation) in `internal/tui/wizard/model.go`
- [x] T052 [US1] Create generate command with Cobra in `cmd/krakenv/generate.go`
- [x] T053 [US1] Wire generate command to wizard TUI in `cmd/krakenv/generate.go`
- [x] T054 [US1] Implement --non-interactive mode (fail on missing values) in `cmd/krakenv/generate.go`
- [x] T055 [US1] Implement --force flag (overwrite without confirmation) in `cmd/krakenv/generate.go`
- [x] T056 [US1] Implement default value display and acceptance in `internal/tui/wizard/model.go`
- [x] T057 [US1] Implement secret input hiding in wizard in `internal/tui/wizard/model.go`
- [x] T058 [US1] Implement graceful Ctrl+C handling (no partial writes) in `internal/tui/wizard/model.go`

**Checkpoint**: `krakenv generate .env.local` works interactively; MVP complete

---

## Phase 4: User Story 2 - Inspect Missing and Extra Variables (Priority: P2)

**Goal**: Developer identifies discrepancies between dist and env files

**Independent Test**: Run `krakenv inspect .env.local` with mismatched files, verify accurate report

### Tests for US2

- [ ] T059 [P] [US2] Write contract test for inspect command flags in `tests/contract/cli_test.go`
- [ ] T060 [P] [US2] Write integration test for inspect report in `tests/integration/inspect_test.go`
- [ ] T061 [P] [US2] Write unit test for diff algorithm in `internal/inspector/inspector_test.go`

### Implementation for US2

- [x] T062 [US2] Create InspectionResult struct in `internal/inspector/inspector.go`
- [x] T063 [US2] Implement diff logic (missing, extra, invalid) in `internal/inspector/inspector.go`
- [ ] T064 [US2] Implement inspection TUI model in `internal/tui/inspect/model.go`
- [ ] T065 [US2] Implement inspection TUI view (categorized report) in `internal/tui/inspect/view.go`
- [x] T066 [US2] Create inspect command with Cobra in `cmd/krakenv/inspect.go`
- [x] T067 [US2] Implement --json output format in `cmd/krakenv/inspect.go`
- [ ] T068 [US2] Implement --sync interactive resolution in `internal/tui/inspect/model.go`

**Checkpoint**: `krakenv inspect .env.local` shows accurate diff report

---

## Phase 5: User Story 3 - Validate Existing Environment File (Priority: P2)

**Goal**: Developer validates env file values against annotations for CI/CD

**Independent Test**: Run `krakenv validate .env.local` with valid/invalid files, check exit codes

### Tests for US3

- [ ] T069 [P] [US3] Write contract test for validate command flags in `tests/contract/cli_test.go`
- [ ] T070 [P] [US3] Write integration test for validation pass/fail in `tests/integration/validate_test.go`
- [ ] T071 [P] [US3] Write unit test for error message formatting in `internal/validator/errors_test.go`

### Implementation for US3

- [x] T072 [US3] Implement batch validation (collect all errors) in `internal/validator/validator.go`
- [x] T073 [US3] Implement user-friendly error formatting with suggestions in `internal/validator/errors.go`
- [x] T074 [US3] Create validate command with Cobra in `cmd/krakenv/validate.go`
- [x] T075 [US3] Implement --strict mode (require annotations) in `cmd/krakenv/validate.go`
- [x] T076 [US3] Implement exit codes (0=pass, 1=errors, 2=file missing) in `cmd/krakenv/validate.go`

**Checkpoint**: `krakenv validate .env.local` works in CI/CD pipelines

---

## Phase 6: User Story 4 - Add New Variable to Distributable (Priority: P3)

**Goal**: Developer adds annotated variable to .env.dist via CLI

**Independent Test**: Run `krakenv add VAR_NAME --type int --min 1`, verify .env.dist updated

### Tests for US4

- [ ] T077 [P] [US4] Write contract test for add command flags in `tests/contract/cli_test.go`
- [ ] T078 [P] [US4] Write integration test for add variable in `tests/integration/add_test.go`

### Implementation for US4

- [x] T079 [US4] Implement annotation builder from flags in `internal/parser/builder.go`
- [x] T080 [US4] Implement append to distributable file in `internal/generator/generator.go`
- [x] T081 [US4] Create add command with all type flags in `cmd/krakenv/add.go`
- [x] T082 [US4] Implement variable name validation (uppercase, underscores) in `cmd/krakenv/add.go`
- [x] T083 [US4] Implement duplicate detection (exit code 2) in `cmd/krakenv/add.go`

**Checkpoint**: `krakenv add DB_PORT --type int --min 1 --max 65535` works

---

## Phase 7: User Story 5 - Configure Distributable Settings (Priority: P3)

**Goal**: Project maintainer configures krakenv via embedded config block

**Independent Test**: Add `#krakenv:environments=local,prod` to .env.dist, run `krakenv generate --all`

### Tests for US5

- [ ] T084 [P] [US5] Write contract test for init command in `tests/contract/cli_test.go`
- [ ] T085 [P] [US5] Write integration test for --all flag with config in `tests/integration/generate_test.go`

### Implementation for US5

- [x] T086 [US5] Implement --all flag reading environments from config in `cmd/krakenv/generate.go`
- [x] T087 [US5] Create init command with Cobra in `cmd/krakenv/init.go`
- [x] T088 [US5] Implement init template generation in `cmd/krakenv/init.go`
- [x] T089 [US5] Implement --environments flag for init in `cmd/krakenv/init.go`

**Checkpoint**: Project configuration via #krakenv comments works

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, CI/CD, release preparation

### Public Library API

- [x] T090 [P] Create public EnvFile API in `pkg/envfile/envfile.go`
- [x] T091 [P] Write tests for public API in `pkg/envfile/envfile_test.go`

### Documentation Site (GitHub Pages)

- [x] T092 [P] Create landing page in `docs/index.html`
- [x] T093 [P] Create getting-started guide in `docs/getting-started.html`
- [x] T094 [P] Create annotation syntax reference in `docs/annotation-syntax.html`
- [x] T095 [P] Create commands reference in `docs/commands.html`
- [x] T096 [P] Create CSS styles in `docs/assets/styles.css`

### Final Polish

- [x] T097 Add shell completion generation to root command in `cmd/krakenv/root.go`
- [x] T098 [P] Run full test suite and fix any failures
- [ ] T099 [P] Run golangci-lint and fix all issues
- [x] T100 Update README with final usage examples in `README.md`

**Checkpoint**: `make check` passes; ready for release

---

## Dependencies & Execution Order

### Phase Dependencies

```
Setup (Phase 1)
     │
     ▼
Foundational (Phase 2) ──── BLOCKS ALL USER STORIES
     │
     ├──────────────────────────────────────────┐
     ▼                                          ▼
User Story 1 (P1) ◄─── MVP          User Stories 2-5 can start
     │                              after Foundational
     ▼
Polish (Phase 8) ◄─── After all stories complete
```

### User Story Dependencies

| Story | Can Start After | Dependencies on Other Stories |
|-------|-----------------|-------------------------------|
| US1 | Phase 2 | None (MVP) |
| US2 | Phase 2 | None (uses parser, validator) |
| US3 | Phase 2 | None (uses parser, validator) |
| US4 | Phase 2 | None (uses parser) |
| US5 | Phase 2 | Enhances US1 (--all flag) |

### Within Each User Story

1. Tests MUST be written and FAIL before implementation
2. Models before services
3. Services before commands
4. Core implementation before TUI integration

---

## Parallel Execution Examples

### Phase 2: Foundational

```bash
# Parallel: All type struct definitions
T011, T012, T013, T014, T015, T016, T017, T018

# Parallel: All validator test files
T025, T026, T027, T028, T029, T030

# Parallel: TUI components
T038, T039, T040, T041

# Parallel: Test data
T043, T044
```

### Phase 3: User Story 1

```bash
# Parallel: All US1 tests
T045, T046, T047

# Sequential: Implementation
T048 → T049 → T050 → T051 → T052 → T053 → T054...
```

### Multiple Stories in Parallel (if team capacity allows)

```bash
# Developer A: US1 (MVP)
# Developer B: US3 (simpler, validation only)
# Developer C: US4 (add command - independent)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup ✓
2. Complete Phase 2: Foundational ✓
3. Complete Phase 3: User Story 1 ✓
4. **STOP and VALIDATE**: Test `krakenv generate .env.local` end-to-end
5. Deploy/demo MVP

### Incremental Delivery

| Increment | Stories | Deliverable |
|-----------|---------|-------------|
| MVP | US1 | Interactive generation |
| v1.1 | US1 + US3 | + Validation for CI/CD |
| v1.2 | US1 + US2 + US3 | + Inspection/sync |
| v1.3 | All | + Add command + Init |
| v2.0 | All + Polish | Full release with docs |

### Suggested MVP Scope

**Include**: US1 only (40% of tasks, 100% of core value)

**Why**: Generate command delivers the complete wizard experience. Validate and inspect are useful but secondary. Users can manually diff files initially.

---

## Summary

| Metric | Value |
|--------|-------|
| **Total Tasks** | 100 |
| **Setup Phase** | 10 |
| **Foundational Phase** | 34 |
| **US1 Tasks** | 14 |
| **US2 Tasks** | 10 |
| **US3 Tasks** | 8 |
| **US4 Tasks** | 7 |
| **US5 Tasks** | 6 |
| **Polish Tasks** | 11 |
| **Parallelizable** | 48 (marked with [P]) |

---

## Notes

- All tests follow TDD: Write test → Verify fail → Implement → Verify pass
- Commit after each task or logical group
- Each checkpoint validates story independently
- [P] tasks can run in parallel (different files, no dependencies)
- [USn] labels track traceability to user stories

