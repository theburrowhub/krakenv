# Data Model: Krakenv CLI

**Date**: 2025-12-26  
**Branch**: `001-krakenv-cli`

## Core Entities

### 1. EnvFile

Represents a parsed `.env` or `.env.dist` file.

```go
type EnvFile struct {
    Path      string
    Variables []Variable
    Config    *KrakenvConfig  // nil if not a distributable
    Comments  []Comment       // Standalone comments (not attached to variables)
}
```

**Relationships**:
- Contains 0..n `Variable`
- Contains 0..1 `KrakenvConfig` (only in distributable files)
- Contains 0..n `Comment` (standalone comments)

**Validation Rules**:
- Path MUST exist and be readable
- Variables are ordered by appearance in file
- Duplicate variable names trigger warning (last wins)

---

### 2. Variable

Represents a single environment variable with optional annotation.

```go
type Variable struct {
    Name       string
    Value      string       // May be empty string
    Annotation *Annotation  // nil if no annotation
    LineNumber int
    IsSet      bool         // true if value was explicitly set (vs undefined)
}
```

**Validation Rules**:
- Name MUST match pattern `^[A-Z][A-Z0-9_]*$` (uppercase with underscores)
- Value MAY be empty string (different from undefined)
- LineNumber is 1-indexed

---

### 3. Annotation

Metadata extracted from inline comment on a variable line.

```go
type Annotation struct {
    PromptText  string
    Type        VariableType
    Constraints []Constraint
    IsOptional  bool
    IsSecret    bool
}
```

**Syntax**: `#prompt:MESSAGE|TYPE;CONSTRAINT:VALUE;...`

**Example**: `#prompt:How many workers?|int;min:1;max:16;optional`

---

### 4. VariableType

Enumeration of supported variable types.

```go
type VariableType int

const (
    TypeString VariableType = iota
    TypeInt
    TypeNumeric
    TypeBoolean
    TypeEnum
    TypeObject
)
```

**Type Behaviors**:

| Type | Go Validation | Example Values |
|------|---------------|----------------|
| `string` | len, regex | `"hello"`, `""` |
| `int` | ParseInt, range | `42`, `-1`, `0` |
| `numeric` | ParseFloat, range | `3.14`, `-0.5`, `100` |
| `boolean` | Case-insensitive match | `true`, `false`, `yes`, `no`, `1`, `0`, `on`, `off` |
| `enum` | Exact match from options | One of defined values |
| `object` | JSON/YAML parse | `{"key":"value"}` |

---

### 5. Constraint

A single validation constraint attached to an annotation.

```go
type Constraint struct {
    Name  string  // "min", "max", "minlen", "maxlen", "pattern", "options", "format"
    Value string  // Raw string value; parsed per constraint type
}
```

**Supported Constraints by Type**:

| Constraint | Applies To | Value Format |
|------------|------------|--------------|
| `min` | int, numeric | Number |
| `max` | int, numeric | Number |
| `minlen` | string | Integer |
| `maxlen` | string | Integer |
| `pattern` | string | Regex |
| `options` | enum | Comma-separated values |
| `format` | object | `json` or `yaml` |

---

### 6. KrakenvConfig

Project-level configuration extracted from distributable.

```go
type KrakenvConfig struct {
    Environments []string  // e.g., ["local", "testing", "production"]
    Strict       bool      // If true, unannotated variables are errors
    DistPath     string    // Override default .env.dist path
}
```

**Syntax**: `#krakenv:KEY=VALUE` comments in distributable

**Default Values**:
- `environments`: `["local"]`
- `strict`: `false`
- `distPath`: `.env.dist`

---

### 7. ValidationError

Represents a single validation failure.

```go
type ValidationError struct {
    Variable   string
    LineNumber int
    Message    string      // User-friendly description
    Suggestion string      // How to fix
    Type       ErrorType
}
```

**Error Types**:

```go
type ErrorType int

const (
    ErrorMissingRequired ErrorType = iota  // Required variable has no value
    ErrorInvalidType                        // Value doesn't match type
    ErrorConstraintViolation               // Value fails constraint
    ErrorAnnotationSyntax                  // Malformed annotation
    ErrorDuplicateVariable                 // Variable defined twice
)
```

---

### 8. InspectionResult

Result of comparing distributable with environment file.

```go
type InspectionResult struct {
    MissingInEnv   []Variable  // In dist, not in env
    ExtraInEnv     []Variable  // In env, not in dist
    InvalidValues  []ValidationError
    ValidCount     int
}
```

---

## State Transitions

### Variable Lifecycle

```
                ┌─────────────┐
                │  Undefined  │  (not in env file)
                └──────┬──────┘
                       │ generate/add
                       ▼
                ┌─────────────┐
                │   Pending   │  (in env but no valid value)
                └──────┬──────┘
                       │ user input + validation
                       ▼
                ┌─────────────┐
                │    Valid    │  (value passes constraints)
                └──────┬──────┘
                       │ value changed/file edited
                       ▼
                ┌─────────────┐
                │   Invalid   │  (value fails constraints)
                └─────────────┘
```

### TUI Wizard State Machine

```
            ┌────────────────┐
            │     Init       │
            └───────┬────────┘
                    │ Load files
                    ▼
            ┌────────────────┐
            │  Collecting    │◄────────┐
            └───────┬────────┘         │
                    │ Next variable    │
                    ▼                  │
            ┌────────────────┐         │
            │   Prompting    │─────────┤ Validation failed
            └───────┬────────┘         │
                    │ Valid input      │
                    ▼                  │
            ┌────────────────┐         │
            │   Validated    │─────────┘ More variables
            └───────┬────────┘
                    │ All complete
                    ▼
            ┌────────────────┐
            │    Writing     │
            └───────┬────────┘
                    │ File written
                    ▼
            ┌────────────────┐
            │    Complete    │
            └────────────────┘
```

---

## Entity Relationships Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         EnvFile                             │
│  ┌─────────┬──────────────┬───────────────┬──────────────┐  │
│  │ Path    │ Variables[]  │ Config?       │ Comments[]   │  │
│  └─────────┴──────┬───────┴───────┬───────┴──────────────┘  │
│                   │               │                          │
└───────────────────┼───────────────┼──────────────────────────┘
                    │               │
                    │  0..1         │ 0..1
                    ▼               ▼
            ┌───────────────┐  ┌──────────────────┐
            │   Variable    │  │  KrakenvConfig   │
            │  ┌──────────┐ │  │  ┌────────────┐  │
            │  │ Name     │ │  │  │ Environments│ │
            │  │ Value    │ │  │  │ Strict     │  │
            │  │ Annotation│ │  │  │ DistPath   │  │
            │  │ LineNum  │ │  │  └────────────┘  │
            │  └────┬─────┘ │  └──────────────────┘
            └───────┼───────┘
                    │ 0..1
                    ▼
            ┌───────────────┐
            │  Annotation   │
            │  ┌──────────┐ │
            │  │ PromptText│ │
            │  │ Type     │ │
            │  │ Constraints││
            │  │ IsOptional││
            │  │ IsSecret │ │
            │  └────┬─────┘ │
            └───────┼───────┘
                    │ 0..n
                    ▼
            ┌───────────────┐
            │  Constraint   │
            │  ┌──────────┐ │
            │  │ Name     │ │
            │  │ Value    │ │
            │  └──────────┘ │
            └───────────────┘
```

---

## File Format Examples

### Distributable (.env.dist)

```env
#krakenv:environments=local,testing,production
#krakenv:strict=false

# Database Configuration
DB_HOST=localhost #prompt:Database host?|string
DB_PORT=5432 #prompt:Database port?|int;min:1;max:65535
DB_NAME= #prompt:Database name?|string;minlen:1
DB_PASSWORD= #prompt:Database password?|string;secret

# Application Settings
APP_ENV=development #prompt:Environment?|enum;options:development,staging,production
DEBUG=true #prompt:Enable debug mode?|boolean;optional
WORKERS=4 #prompt:Number of worker processes?|int;min:1;max:16

# Feature Flags (no annotation = not prompted, just copied)
ENABLE_CACHE=true
```

### Generated Environment File (.env.local)

```env
#krakenv:environments=local,testing,production
#krakenv:strict=false

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp_dev
DB_PASSWORD=supersecret123

# Application Settings
APP_ENV=development
DEBUG=true
WORKERS=4

# Feature Flags (no annotation = not prompted, just copied)
ENABLE_CACHE=true
```

