// Package validator provides functionality for validating environment variable values.
package validator

import "fmt"

// ErrorType represents the type of validation error.
type ErrorType int

const (
	// ErrorMissingRequired indicates a required variable has no value.
	ErrorMissingRequired ErrorType = iota
	// ErrorInvalidType indicates the value doesn't match the expected type.
	ErrorInvalidType
	// ErrorConstraintViolation indicates the value fails a constraint.
	ErrorConstraintViolation
	// ErrorAnnotationSyntax indicates a malformed annotation.
	ErrorAnnotationSyntax
	// ErrorDuplicateVariable indicates a variable is defined more than once.
	ErrorDuplicateVariable
)

// String returns the string representation of an ErrorType.
func (e ErrorType) String() string {
	switch e {
	case ErrorMissingRequired:
		return "missing_required"
	case ErrorInvalidType:
		return "invalid_type"
	case ErrorConstraintViolation:
		return "constraint_violation"
	case ErrorAnnotationSyntax:
		return "annotation_syntax"
	case ErrorDuplicateVariable:
		return "duplicate_variable"
	default:
		return "unknown"
	}
}

// ValidationError represents a single validation failure.
type ValidationError struct {
	Variable   string    // Variable name
	LineNumber int       // Line number in source file (1-indexed)
	Message    string    // User-friendly problem description
	Suggestion string    // How to fix the error
	Example    string    // Example of a valid value
	Type       ErrorType // Type of error
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s (line %d): %s", e.Variable, e.LineNumber, e.Message)
}

// Format returns a formatted error message with all four required components.
// Per FR-033: Problem, Location, Suggestion, Example.
func (e *ValidationError) Format() string {
	result := fmt.Sprintf("  Line %d: %s\n", e.LineNumber, e.Variable)
	result += fmt.Sprintf("    ✗ %s\n", e.Message)
	if e.Suggestion != "" {
		result += fmt.Sprintf("    → Fix: %s\n", e.Suggestion)
	}
	if e.Example != "" {
		result += fmt.Sprintf("    → Example: %s\n", e.Example)
	}
	return result
}

// ValidationResult holds the results of validating an environment file.
type ValidationResult struct {
	Errors []ValidationError // All validation errors
	Valid  bool              // True if no errors
}

// NewValidationResult creates a new empty ValidationResult.
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Errors: make([]ValidationError, 0),
		Valid:  true,
	}
}

// AddError adds an error to the result and marks it as invalid.
func (r *ValidationResult) AddError(err ValidationError) {
	r.Errors = append(r.Errors, err)
	r.Valid = false
}

// ErrorCount returns the number of errors.
func (r *ValidationResult) ErrorCount() int {
	return len(r.Errors)
}

// FormatErrors returns a formatted string of all errors.
func (r *ValidationResult) FormatErrors(filePath string) string {
	if r.Valid {
		return fmt.Sprintf("✓ VALIDATION PASSED: %s\n", filePath)
	}

	result := fmt.Sprintf("✗ VALIDATION FAILED: %s\n\n", filePath)
	for _, err := range r.Errors {
		result += err.Format() + "\n"
	}
	result += fmt.Sprintf("Found %d error(s)\n", len(r.Errors))
	return result
}

// NewMissingRequiredError creates a ValidationError for a missing required variable.
func NewMissingRequiredError(variable string, lineNumber int, prompt string) ValidationError {
	return ValidationError{
		Variable:   variable,
		LineNumber: lineNumber,
		Message:    "Required variable has no value",
		Suggestion: fmt.Sprintf("Set a value for %s", variable),
		Example:    prompt,
		Type:       ErrorMissingRequired,
	}
}

// NewInvalidTypeError creates a ValidationError for an invalid type.
func NewInvalidTypeError(variable string, lineNumber int, expected, got string) ValidationError {
	return ValidationError{
		Variable:   variable,
		LineNumber: lineNumber,
		Message:    fmt.Sprintf("Expected %s, got %q", expected, got),
		Suggestion: fmt.Sprintf("Enter a valid %s value", expected),
		Example:    "",
		Type:       ErrorInvalidType,
	}
}

// NewConstraintError creates a ValidationError for a constraint violation.
func NewConstraintError(variable string, lineNumber int, message, suggestion, example string) ValidationError {
	return ValidationError{
		Variable:   variable,
		LineNumber: lineNumber,
		Message:    message,
		Suggestion: suggestion,
		Example:    example,
		Type:       ErrorConstraintViolation,
	}
}

// NewAnnotationSyntaxError creates a ValidationError for malformed annotation.
func NewAnnotationSyntaxError(variable string, lineNumber int, message string) ValidationError {
	return ValidationError{
		Variable:   variable,
		LineNumber: lineNumber,
		Message:    fmt.Sprintf("Malformed annotation: %s", message),
		Suggestion: "Check annotation syntax: #prompt:Message?|type;constraint:value",
		Example:    "#prompt:Enter value?|string;minlen:1",
		Type:       ErrorAnnotationSyntax,
	}
}

// NewDuplicateVariableError creates a ValidationError for duplicate variable.
func NewDuplicateVariableError(variable string, lineNumber, firstLine int) ValidationError {
	return ValidationError{
		Variable:   variable,
		LineNumber: lineNumber,
		Message:    fmt.Sprintf("Variable already defined on line %d", firstLine),
		Suggestion: "Remove duplicate definition",
		Example:    "",
		Type:       ErrorDuplicateVariable,
	}
}
