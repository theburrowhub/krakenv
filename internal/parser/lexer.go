package parser

import (
	"errors"
	"regexp"
	"strings"
)

// variableNamePattern matches valid variable names: uppercase letters, numbers, and underscores.
// Must start with an uppercase letter.
var variableNamePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

// ErrInvalidVariableName indicates the variable name doesn't match the required pattern.
var ErrInvalidVariableName = errors.New("variable name must start with uppercase letter and contain only uppercase letters, numbers, and underscores")

// TokenizeLine parses a single line from an .env file.
// Returns the variable name, value, annotation string, and any error.
// For comments or empty lines, returns empty name with no error.
func TokenizeLine(line string) (name, value, annotation string, err error) {
	line = strings.TrimSpace(line)

	// Empty line
	if line == "" {
		return "", "", "", nil
	}

	// Full-line comment (including krakenv config lines)
	if strings.HasPrefix(line, "#") {
		return "", "", "", nil
	}

	// Find the first equals sign
	eqIdx := strings.Index(line, "=")
	if eqIdx == -1 {
		return "", "", "", nil
	}

	// Extract variable name
	name = strings.TrimSpace(line[:eqIdx])
	if name == "" {
		return "", "", "", nil
	}

	// Validate variable name
	if !variableNamePattern.MatchString(name) {
		return "", "", "", ErrInvalidVariableName
	}

	// Extract value and potential annotation
	rest := line[eqIdx+1:]

	// Check for annotation (#prompt:...)
	annotationIdx := strings.Index(rest, " #prompt:")
	if annotationIdx != -1 {
		annotation = strings.TrimSpace(rest[annotationIdx+1:])
		rest = rest[:annotationIdx]
	}

	// Parse the value
	value = parseValue(strings.TrimSpace(rest))

	return name, value, annotation, nil
}

// parseValue handles quoted and unquoted values.
func parseValue(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	// Handle double-quoted values
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}

	// Handle single-quoted values
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		return s[1 : len(s)-1]
	}

	return s
}

// IsComment checks if a line is a comment.
func IsComment(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "#")
}

// IsEmptyLine checks if a line is empty or whitespace only.
func IsEmptyLine(line string) bool {
	return strings.TrimSpace(line) == ""
}

// IsKrakenvConfigLine checks if a line is a krakenv configuration comment.
func IsKrakenvConfigLine(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "#krakenv:")
}

// IsAnnotationLine checks if a line contains a krakenv annotation.
func IsAnnotationLine(line string) bool {
	return strings.Contains(line, " #prompt:")
}

// ExtractCommentText extracts the text from a comment line (without the #).
func ExtractCommentText(line string) string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "#") {
		return strings.TrimSpace(line[1:])
	}
	return ""
}
