package parser

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/theburrowhub/krakenv/internal/config"
)

// ErrInvalidAnnotation indicates the annotation syntax is invalid.
var ErrInvalidAnnotation = errors.New("invalid annotation syntax")

// knownConstraints lists all valid constraint names.
var knownConstraints = map[string]bool{
	"min":      true,
	"max":      true,
	"minlen":   true,
	"maxlen":   true,
	"pattern":  true,
	"options":  true,
	"format":   true,
	"encoding": true,
}

// ParseAnnotation parses an annotation string into an Annotation struct.
// Annotation format: #prompt:MESSAGE|TYPE;CONSTRAINT:VALUE;...
func ParseAnnotation(s string) (*Annotation, error) {
	s = strings.TrimSpace(s)

	// Must start with #prompt:
	if !strings.HasPrefix(s, "#prompt:") {
		return nil, fmt.Errorf("%w: must start with #prompt", ErrInvalidAnnotation)
	}

	// Remove prefix
	content := strings.TrimPrefix(s, "#prompt:")

	// Split by | to separate message from type+constraints
	pipeIdx := strings.Index(content, "|")
	if pipeIdx == -1 {
		return nil, fmt.Errorf("%w: missing | separator", ErrInvalidAnnotation)
	}

	ann := &Annotation{
		PromptText:  strings.TrimSpace(content[:pipeIdx]),
		Constraints: make([]Constraint, 0),
	}

	// Parse type and constraints
	rest := content[pipeIdx+1:]
	parts := strings.Split(rest, ";")

	if len(parts) == 0 || parts[0] == "" {
		return nil, fmt.Errorf("%w: missing type", ErrInvalidAnnotation)
	}

	// First part is the type
	ann.Type = ParseVariableType(strings.TrimSpace(parts[0]))

	// Remaining parts are constraints/modifiers
	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check for modifiers (no colon)
		if part == "optional" {
			ann.IsOptional = true
			continue
		}
		if part == "secret" {
			ann.IsSecret = true
			continue
		}

		// Parse constraint with colon
		colonIdx := strings.Index(part, ":")
		if colonIdx == -1 {
			// Unknown modifier without colon - ignore (FR-041)
			continue
		}

		constraintName := strings.TrimSpace(part[:colonIdx])
		constraintValue := strings.TrimSpace(part[colonIdx+1:])

		// Check if it's a known constraint (FR-041: ignore unknown)
		if !knownConstraints[constraintName] {
			// Unknown constraint - ignore with warning (TODO: add warning system)
			continue
		}

		ann.Constraints = append(ann.Constraints, Constraint{
			Name:  constraintName,
			Value: constraintValue,
		})
	}

	// FR-042: Enum with empty options becomes string
	if ann.Type == TypeEnum {
		options := ann.GetConstraint("options")
		if options == "" {
			ann.Type = TypeString
		}
	}

	return ann, nil
}

// ParseEnvFile parses an .env file from disk.
func ParseEnvFile(path string) (*EnvFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return parseLines(lines, path)
}

// ParseEnvFileContent parses .env content from a string.
func ParseEnvFileContent(content string, path string) (*EnvFile, error) {
	lines := strings.Split(content, "\n")
	return parseLines(lines, path)
}

// parseLines parses a slice of lines into an EnvFile.
func parseLines(lines []string, path string) (*EnvFile, error) {
	envFile := &EnvFile{
		Path:      path,
		Variables: make([]Variable, 0),
		Comments:  make([]Comment, 0),
	}

	// Collect config lines
	var configLines []string

	// Track variable positions for duplicate detection
	varPositions := make(map[string]int)

	for lineNum, line := range lines {
		lineNumber := lineNum + 1 // 1-indexed

		// Handle krakenv config lines
		if config.IsConfigLine(line) {
			configLines = append(configLines, line)
			continue
		}

		// Handle standalone comments
		if IsComment(line) && !IsAnnotationLine(line) {
			text := ExtractCommentText(line)
			if text != "" {
				envFile.Comments = append(envFile.Comments, Comment{
					Text:       text,
					LineNumber: lineNumber,
				})
			}
			continue
		}

		// Handle empty lines
		if IsEmptyLine(line) {
			continue
		}

		// Parse variable line
		name, value, annotationStr, err := TokenizeLine(line)
		if err != nil {
			// Invalid variable name - skip with warning
			continue
		}

		if name == "" {
			continue
		}

		// Create variable
		variable := Variable{
			Name:       name,
			Value:      strings.TrimSpace(value), // FR-040: trim whitespace
			LineNumber: lineNumber,
			IsSet:      value != "" || strings.Contains(line, "="),
		}

		// Parse annotation if present
		if annotationStr != "" {
			ann, err := ParseAnnotation(annotationStr)
			if err != nil {
				// Invalid annotation syntax - treat as no annotation
				// TODO: add warning
			} else {
				variable.Annotation = ann
			}
		}

		// Handle duplicates: last wins, but track position
		if existingIdx, exists := varPositions[name]; exists {
			// Replace existing variable
			envFile.Variables[existingIdx] = variable
		} else {
			varPositions[name] = len(envFile.Variables)
			envFile.Variables = append(envFile.Variables, variable)
		}
	}

	// Parse config block
	if len(configLines) > 0 {
		cfg := config.ParseConfig(configLines)
		envFile.Config = &KrakenvConfig{
			Environments: cfg.Environments,
			Strict:       cfg.Strict,
			DistPath:     cfg.DistPath,
		}
	}

	return envFile, nil
}

// FormatVariable formats a variable as a line for an .env file.
func FormatVariable(v Variable, includeAnnotation bool) string {
	line := v.Name + "=" + v.Value
	if includeAnnotation && v.Annotation != nil {
		line += " " + FormatAnnotation(v.Annotation)
	}
	return line
}

// FormatAnnotation formats an Annotation back to string format.
func FormatAnnotation(a *Annotation) string {
	var parts []string

	// Add type
	parts = append(parts, a.Type.String())

	// Add constraints
	for _, c := range a.Constraints {
		parts = append(parts, c.Name+":"+c.Value)
	}

	// Add modifiers
	if a.IsOptional {
		parts = append(parts, "optional")
	}
	if a.IsSecret {
		parts = append(parts, "secret")
	}

	return "#prompt:" + a.PromptText + "|" + strings.Join(parts, ";")
}
