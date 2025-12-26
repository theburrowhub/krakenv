# CLI Interface Specification: Krakenv

**Date**: 2025-12-26  
**Version**: 1.0.0

## Overview

Krakenv provides a command-line interface for managing environment variable files. This document defines the contract for all commands, flags, and expected behaviors.

## Global Flags

Available on all commands:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--dist` | `-d` | string | `.env.dist` | Path to distributable file |
| `--non-interactive` | `-n` | bool | `false` | Disable TUI; fail on unresolved variables |
| `--quiet` | `-q` | bool | `false` | Suppress non-error output |
| `--verbose` | `-v` | bool | `false` | Enable detailed output |
| `--help` | `-h` | bool | - | Show help for command |
| `--version` | `-V` | bool | - | Show version information |

---

## Commands

### 1. `krakenv generate <target>`

Generate or update an environment file from the distributable.

**Synopsis**:

```bash
krakenv generate <target> [flags]
```

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `target` | Yes | Target file name (e.g., `.env.local`, `.env.testing`) |

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--force` | `-f` | bool | `false` | Overwrite existing file without confirmation |
| `--all` | `-a` | bool | `false` | Generate all environments defined in config |
| `--keep-annotations` | `-k` | bool | `false` | Preserve annotations in generated file |

**Behavior**:

1. Load distributable file (default: `.env.dist`)
2. Parse all variables and annotations
3. If target exists, load existing values
4. For each variable:
   - If valid value exists → skip prompt
   - If missing or invalid → prompt user (unless `--non-interactive`)
5. Write target file preserving structure

**Exit Codes**:

| Code | Meaning |
|------|---------|
| 0 | Success; file generated |
| 1 | Error; could not generate file |
| 2 | Non-interactive mode; unresolved variables exist |

**Examples**:

```bash
# Interactive generation
krakenv generate .env.local

# Generate with custom distributable
krakenv generate .env.local --dist config/env.template

# Generate all environments
krakenv generate --all

# CI/CD mode - fail if any variable unresolved
krakenv generate .env.testing --non-interactive
```

---

### 2. `krakenv validate <target>`

Validate an environment file against the distributable annotations.

**Synopsis**:

```bash
krakenv validate <target> [flags]
```

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `target` | Yes | Environment file to validate |

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--strict` | `-s` | bool | from config | Require all variables to have annotations |

**Behavior**:

1. Load distributable and target files
2. For each variable with annotation:
   - Validate value against type and constraints
   - Collect all errors (don't stop at first)
3. Report all validation errors with line numbers
4. Exit with appropriate code

**Exit Codes**:

| Code | Meaning |
|------|---------|
| 0 | All validations passed |
| 1 | Validation errors found |
| 2 | File not found or unreadable |

**Output Format** (errors):

```
VALIDATION FAILED: .env.local

  Line 5: DB_PORT
    ✗ Expected integer between 1-65535, got "abc"
    → Fix: Enter a valid port number (e.g., 5432)

  Line 12: APP_ENV
    ✗ Value "dev" not in allowed options: development, staging, production
    → Fix: Use one of: development, staging, production

Found 2 error(s) in 15 variables
```

**Examples**:

```bash
# Validate local environment
krakenv validate .env.local

# Validate with strict mode
krakenv validate .env.testing --strict

# CI/CD validation
krakenv validate .env.production --non-interactive
```

---

### 3. `krakenv inspect <target>`

Compare environment file with distributable; identify discrepancies.

**Synopsis**:

```bash
krakenv inspect <target> [flags]
```

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `target` | Yes | Environment file to inspect |

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--sync` | `-s` | bool | `false` | Interactively sync discrepancies |
| `--json` | `-j` | bool | `false` | Output as JSON (for scripting) |

**Behavior**:

1. Load both distributable and target files
2. Identify:
   - Variables in dist missing from target
   - Variables in target not in dist
   - Variables with invalid values
3. Display categorized report
4. If `--sync`: offer interactive resolution

**Exit Codes**:

| Code | Meaning |
|------|---------|
| 0 | No discrepancies found |
| 1 | Discrepancies found (report generated) |
| 2 | File not found or unreadable |

**Output Format** (default):

```
INSPECTION REPORT: .env.local vs .env.dist

MISSING IN .env.local (3):
  NEW_FEATURE_FLAG     "Enable new feature?" [boolean]
  CACHE_TTL            "Cache TTL in seconds" [int: 0-3600]
  LOG_LEVEL            "Logging level" [enum: debug|info|warn|error]

EXTRA IN .env.local (1):
  LEGACY_API_KEY       (not in distributable)

INVALID VALUES (1):
  DB_PORT              "not_a_number" → Expected int

Summary: 3 missing, 1 extra, 1 invalid
```

**JSON Output** (`--json`):

```json
{
  "missing": [
    {"name": "NEW_FEATURE_FLAG", "prompt": "Enable new feature?", "type": "boolean"}
  ],
  "extra": [
    {"name": "LEGACY_API_KEY", "value": "abc123"}
  ],
  "invalid": [
    {"name": "DB_PORT", "value": "not_a_number", "error": "Expected int"}
  ]
}
```

**Examples**:

```bash
# Basic inspection
krakenv inspect .env.local

# Interactive sync
krakenv inspect .env.local --sync

# JSON for scripting
krakenv inspect .env.testing --json | jq '.missing | length'
```

---

### 4. `krakenv add <name>`

Add a new annotated variable to the distributable.

**Synopsis**:

```bash
krakenv add <name> [flags]
```

**Arguments**:

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Variable name (uppercase with underscores) |

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--type` | `-t` | string | `string` | Variable type |
| `--prompt` | `-p` | string | `"Enter {name}"` | Prompt message |
| `--default` | `-D` | string | `""` | Default value |
| `--min` | | int | - | Minimum value (int/numeric) |
| `--max` | | int | - | Maximum value (int/numeric) |
| `--minlen` | | int | - | Minimum length (string) |
| `--maxlen` | | int | - | Maximum length (string) |
| `--pattern` | | string | - | Regex pattern (string) |
| `--options` | `-o` | string | - | Comma-separated options (enum) |
| `--format` | | string | - | Object format: json or yaml |
| `--optional` | | bool | `false` | Mark as optional |
| `--secret` | | bool | `false` | Mark as secret |

**Behavior**:

1. Validate variable name format
2. Build annotation from flags
3. Append to distributable file
4. Display confirmation

**Exit Codes**:

| Code | Meaning |
|------|---------|
| 0 | Variable added successfully |
| 1 | Invalid arguments or write error |
| 2 | Variable already exists |

**Examples**:

```bash
# Add simple string
krakenv add API_URL --type string --prompt "API base URL?"

# Add integer with constraints
krakenv add MAX_CONNECTIONS --type int --min 1 --max 100 --default 10

# Add enum
krakenv add LOG_LEVEL --type enum --options "debug,info,warn,error" --default info

# Add secret
krakenv add DB_PASSWORD --type string --prompt "Database password?" --secret

# Add optional boolean
krakenv add ENABLE_METRICS --type boolean --optional --default false
```

**Generated Output**:

```env
# After: krakenv add MAX_CONNECTIONS --type int --min 1 --max 100 --default 10
MAX_CONNECTIONS=10 #prompt:Enter MAX_CONNECTIONS|int;min:1;max:100
```

---

### 5. `krakenv init`

Initialize a new distributable file via interactive wizard.

**Synopsis**:

```bash
krakenv init [flags]
```

**Flags**:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--path` | `-p` | string | `.env.dist` | Output path |
| `--environments` | `-e` | string | `local` | Comma-separated environments |
| `--force` | `-f` | bool | `false` | Overwrite existing file |
| `--template` | `-t` | bool | `false` | Create with example comments only (skip wizard) |

**Behavior**:

1. Check if file exists (fail unless `--force`)
2. Create distributable with config block
3. **Interactive Wizard Loop** (unless `--template`):
   - Prompt: "Variable name? (empty to finish)"
   - Prompt: "Type? [string/int/numeric/boolean/enum/object]"
   - Prompt: "Constraints? (e.g., min:1;max:100)"
   - Prompt: "Default value? (optional)"
   - Prompt: "Prompt message for users?"
   - Prompt: "Is it optional? [y/n]"
   - Prompt: "Is it secret? [y/n]"
   - Write variable to file
   - Repeat until user enters empty variable name
4. Display summary and next steps

**Exit Codes**:

| Code | Meaning |
|------|---------|
| 0 | File created successfully |
| 1 | File already exists (use --force to overwrite) |
| 2 | Write error |

**Examples**:

```bash
# Interactive wizard to create new distributable
krakenv init

# Create at custom path
krakenv init --path config/.env.template

# Create with template comments only (no wizard)
krakenv init --template

# Overwrite existing file
krakenv init --force
```

**Wizard Session Example**:

```
$ krakenv init

Creating new distributable: .env.dist

Variable name? (empty to finish): DB_HOST
Type? [string/int/numeric/boolean/enum/object]: string
Constraints? (e.g., minlen:1): 
Default value? (optional): localhost
Prompt message: Database hostname?
Is it optional? [y/n]: n
Is it secret? [y/n]: n

✓ Added: DB_HOST=localhost #prompt:Database hostname?|string

Variable name? (empty to finish): DB_PORT
Type? [string/int/numeric/boolean/enum/object]: int
Constraints? (e.g., min:1;max:100): min:1;max:65535
Default value? (optional): 5432
Prompt message: Database port?
Is it optional? [y/n]: n
Is it secret? [y/n]: n

✓ Added: DB_PORT=5432 #prompt:Database port?|int;min:1;max:65535

Variable name? (empty to finish): 

✓ Created .env.dist with 2 variables
  Run: krakenv generate .env.local
```

---

### 6. `krakenv version`

Display version and build information.

**Synopsis**:

```bash
krakenv version
```

**Output**:

```
krakenv version 1.0.0
  commit: abc1234
  built:  2025-12-26T10:00:00Z
  go:     go1.23.0
```

---

## TUI Keybindings

When running in interactive mode (no `--non-interactive`):

| Key | Action |
|-----|--------|
| `Enter` | Submit current input / Accept default |
| `Tab` | Next field (in multi-field forms) |
| `Shift+Tab` | Previous field |
| `↑` / `↓` | Navigate options (enum), history |
| `Ctrl+C` | Cancel operation (no changes saved) |
| `Ctrl+D` | Skip optional variable |
| `Esc` | Back to previous screen |
| `?` | Show help overlay |

---

## Environment Variables

Krakenv respects these environment variables:

| Variable | Description |
|----------|-------------|
| `KRAKENV_DIST` | Default distributable path (overrides `.env.dist`) |
| `KRAKENV_NO_COLOR` | Disable colored output |
| `KRAKENV_DEBUG` | Enable debug logging |

---

## Integration Examples

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit
krakenv validate .env.local --non-interactive || exit 1
```

### GitHub Actions

```yaml
- name: Validate environment
  run: |
    krakenv validate .env.production --non-interactive
```

### Makefile

```makefile
.PHONY: env-setup env-validate

env-setup:
	krakenv generate .env.local

env-validate:
	krakenv validate .env.local --non-interactive
```

