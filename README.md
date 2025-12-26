<p align="center">
  <img src="docs/assets/krakenv-banner.png" alt="Krakenv Banner" width="800">
</p>

<p align="center">
  <a href="https://github.com/theburrowhub/krakenv/releases"><img src="https://img.shields.io/github/v/release/theburrowhub/krakenv?style=flat-square&color=00B894" alt="Release"></a>
  <a href="https://github.com/theburrowhub/krakenv/actions"><img src="https://img.shields.io/github/actions/workflow/status/theburrowhub/krakenv/ci.yml?style=flat-square" alt="Build"></a>
  <a href="https://goreportcard.com/report/github.com/theburrowhub/krakenv"><img src="https://goreportcard.com/badge/github.com/theburrowhub/krakenv?style=flat-square" alt="Go Report Card"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue?style=flat-square" alt="License"></a>
</p>

<p align="center">
  <strong>Environment variable management with annotation-based wizards</strong>
</p>

---

Krakenv is a CLI tool for managing environment variable files (`.env`) with an annotation-based configuration wizard system. It transforms the tedious process of configuring environment files into a guided, validated experience.

## ✨ Features

- **Interactive Wizard**: Guided configuration with type validation and constraints
- **Annotation System**: Define prompts, types, and validation rules inline in your `.env.dist`
- **Multi-Environment Support**: Generate `.env.local`, `.env.testing`, `.env.production` from a single template
- **Validation**: Type checking (int, string, enum, boolean, object) with constraints (min, max, pattern, etc.)
- **Inspection**: Compare distributable and environment files, sync discrepancies
- **CI/CD Ready**: Non-interactive mode with exit codes for pipeline integration
- **Cross-Platform**: Runs on Linux, macOS, and Windows

## 📦 Installation

### Quick Install (Linux/macOS)

```bash
curl -sSL https://raw.githubusercontent.com/theburrowhub/krakenv/main/install.sh | bash
```

### Homebrew (macOS/Linux)

```bash
brew install theburrowhub/tap/krakenv
```

### Go Install

```bash
go install github.com/theburrowhub/krakenv/cmd/krakenv@latest
```

### Binary Downloads

Download pre-built binaries from the [Releases](https://github.com/theburrowhub/krakenv/releases) page.

## 🚀 Quick Start

### 1. Create a Distributable Template

Create a `.env.dist` file with annotations:

```env
#krakenv:environments=local,testing,production

# Database Configuration
DB_HOST=localhost #prompt:Database host?|string
DB_PORT=5432 #prompt:Database port?|int;min:1;max:65535
DB_NAME= #prompt:Database name?|string;minlen:1
DB_PASSWORD= #prompt:Database password?|string;secret

# Application Settings
APP_ENV=development #prompt:Environment?|enum;options:development,staging,production
DEBUG=true #prompt:Enable debug mode?|boolean;optional
WORKERS=4 #prompt:Number of workers?|int;min:1;max:16
```

### 2. Generate Environment File

```bash
krakenv generate .env.local
```

The wizard will prompt you for each undefined or invalid variable.

### 3. Validate Configuration

```bash
krakenv validate .env.local
```

### 4. Inspect Discrepancies

```bash
krakenv inspect .env.local
```

## 📖 Annotation Syntax

```
VARIABLE=default #prompt:Question?|type;constraint:value;modifier
```

### Supported Types

| Type | Description | Example |
|------|-------------|---------|
| `string` | Text value | `#prompt:Name?\|string;minlen:1;maxlen:100` |
| `int` | Integer | `#prompt:Port?\|int;min:1;max:65535` |
| `numeric` | Float/decimal | `#prompt:Rate?\|numeric;min:0;max:1` |
| `boolean` | true/false | `#prompt:Enable?\|boolean` |
| `enum` | One of options | `#prompt:Env?\|enum;options:dev,staging,prod` |
| `object` | JSON/YAML | `#prompt:Config?\|object;format:json` |

### Constraints

| Constraint | Applies To | Description |
|------------|------------|-------------|
| `min` | int, numeric | Minimum value |
| `max` | int, numeric | Maximum value |
| `minlen` | string | Minimum length |
| `maxlen` | string | Maximum length |
| `pattern` | string | Regex pattern |
| `options` | enum | Allowed values |
| `format` | object | `json` or `yaml` |

### Modifiers

| Modifier | Description |
|----------|-------------|
| `optional` | Variable can be empty |
| `secret` | Hide input in wizard |

## 🔧 Commands

```bash
krakenv generate <target>   # Generate environment file from distributable
krakenv validate <target>   # Validate environment file against annotations
krakenv inspect <target>    # Compare distributable and environment files
krakenv add <name>          # Add new annotated variable to distributable
krakenv init                # Initialize new distributable with wizard
krakenv version             # Show version information
```

### Global Flags

| Flag | Description |
|------|-------------|
| `--dist, -d` | Path to distributable file (default: `.env.dist`) |
| `--non-interactive, -n` | Disable TUI; fail on unresolved variables |
| `--quiet, -q` | Suppress non-error output |
| `--verbose, -v` | Enable detailed output |

## 🔄 CI/CD Integration

### Pre-commit Hook

```bash
#!/bin/bash
krakenv validate .env.local --non-interactive || exit 1
```

### GitHub Actions

```yaml
- name: Validate environment
  run: krakenv validate .env.production --non-interactive
```

### Makefile

```makefile
env-setup:
	krakenv generate .env.local

env-validate:
	krakenv validate .env.local --non-interactive
```

## 🏗️ Development

```bash
# Clone repository
git clone https://github.com/theburrowhub/krakenv.git
cd krakenv

# Install dependencies
make deps

# Run tests
make test

# Build binary
make build

# Run linter
make lint

# Run all checks
make check
```

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🤝 Contributing

Contributions are welcome! Please read our contributing guidelines before submitting a PR.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

<p align="center">
  <strong>Built with 💜 by <a href="https://github.com/theburrowhub">The Burrow Hub</a></strong>
  <br>
  <em>Crafting effective solutions for developers who deserve better tools</em>
</p>
