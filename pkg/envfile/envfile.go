// Package envfile provides a public API for parsing and generating environment files.
// This package can be imported by other Go projects that need to work with .env files.
package envfile

import (
	"github.com/theburrowhub/krakenv/internal/parser"
	"github.com/theburrowhub/krakenv/internal/validator"
)

// EnvFile represents a parsed environment file.
type EnvFile = parser.EnvFile

// Variable represents an environment variable with optional annotation.
type Variable = parser.Variable

// Annotation represents variable metadata from inline comments.
type Annotation = parser.Annotation

// Constraint represents a validation constraint.
type Constraint = parser.Constraint

// VariableType represents the type of a variable.
type VariableType = parser.VariableType

// Type constants.
const (
	TypeString  = parser.TypeString
	TypeInt     = parser.TypeInt
	TypeNumeric = parser.TypeNumeric
	TypeBoolean = parser.TypeBoolean
	TypeEnum    = parser.TypeEnum
	TypeObject  = parser.TypeObject
)

// Parse parses an environment file from disk.
func Parse(path string) (*EnvFile, error) {
	return parser.ParseEnvFile(path)
}

// ParseContent parses environment content from a string.
func ParseContent(content, path string) (*EnvFile, error) {
	return parser.ParseEnvFileContent(content, path)
}

// Validate validates a variable value against an annotation.
func Validate(value string, ann *Annotation) error {
	return validator.ValidateValue(value, ann)
}

// ValidateFile validates all variables in an environment file against a distributable.
func ValidateFile(distPath, targetPath string) (*ValidationResult, error) {
	distFile, err := Parse(distPath)
	if err != nil {
		return nil, err
	}

	targetFile, err := Parse(targetPath)
	if err != nil {
		return nil, err
	}

	result := validator.NewValidationResult()

	for _, distVar := range distFile.Variables {
		if distVar.Annotation == nil {
			continue
		}

		targetVar := targetFile.GetVariable(distVar.Name)
		if targetVar == nil {
			if !distVar.Annotation.IsOptional {
				result.AddError(validator.NewMissingRequiredError(
					distVar.Name,
					0,
					distVar.Annotation.PromptText,
				))
			}
			continue
		}

		if err := Validate(targetVar.Value, distVar.Annotation); err != nil {
			result.AddError(validator.ValidationError{
				Variable:   distVar.Name,
				LineNumber: targetVar.LineNumber,
				Message:    err.Error(),
			})
		}
	}

	return &ValidationResult{
		Valid:  result.Valid,
		Errors: result.Errors,
	}, nil
}

// ValidationResult holds the results of validating an environment file.
type ValidationResult struct {
	Valid  bool
	Errors []validator.ValidationError
}

// FormatAnnotation formats an Annotation as a string.
func FormatAnnotation(ann *Annotation) string {
	return parser.FormatAnnotation(ann)
}

// FormatVariable formats a Variable as a line for an .env file.
func FormatVariable(v Variable, includeAnnotation bool) string {
	return parser.FormatVariable(v, includeAnnotation)
}
