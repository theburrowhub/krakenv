package validator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/theburrowhub/krakenv/internal/parser"
)

// validBooleans lists all accepted boolean values (case-insensitive).
var validBooleans = map[string]bool{
	"true":  true,
	"false": true,
	"yes":   true,
	"no":    true,
	"1":     true,
	"0":     true,
	"on":    true,
	"off":   true,
}

// ValidateValue validates a value against an annotation's type and constraints.
// Returns nil if valid, or an error describing the validation failure.
func ValidateValue(value string, ann *parser.Annotation) error {
	if ann == nil {
		return nil
	}

	// Handle optional empty values
	if value == "" && ann.IsOptional {
		return nil
	}

	switch ann.Type {
	case parser.TypeInt:
		return validateInt(value, ann)
	case parser.TypeNumeric:
		return validateNumeric(value, ann)
	case parser.TypeString:
		return validateString(value, ann)
	case parser.TypeEnum:
		return validateEnum(value, ann)
	case parser.TypeBoolean:
		return validateBoolean(value)
	case parser.TypeObject:
		return validateObject(value, ann)
	default:
		return nil
	}
}

func validateInt(value string, ann *parser.Annotation) error {
	if value == "" {
		return fmt.Errorf("value is required")
	}

	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("expected integer, got %q", value)
	}

	// Check min constraint
	if minStr := ann.GetConstraint("min"); minStr != "" {
		min, err := strconv.ParseInt(minStr, 10, 64)
		if err == nil && n < min {
			return fmt.Errorf("value %d is below minimum %d", n, min)
		}
	}

	// Check max constraint
	if maxStr := ann.GetConstraint("max"); maxStr != "" {
		max, err := strconv.ParseInt(maxStr, 10, 64)
		if err == nil && n > max {
			return fmt.Errorf("value %d exceeds maximum %d", n, max)
		}
	}

	return nil
}

func validateNumeric(value string, ann *parser.Annotation) error {
	if value == "" {
		return fmt.Errorf("value is required")
	}

	n, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("expected numeric, got %q", value)
	}

	// Check min constraint
	if minStr := ann.GetConstraint("min"); minStr != "" {
		min, err := strconv.ParseFloat(minStr, 64)
		if err == nil && n < min {
			return fmt.Errorf("value %v is below minimum %v", n, min)
		}
	}

	// Check max constraint
	if maxStr := ann.GetConstraint("max"); maxStr != "" {
		max, err := strconv.ParseFloat(maxStr, 64)
		if err == nil && n > max {
			return fmt.Errorf("value %v exceeds maximum %v", n, max)
		}
	}

	return nil
}

func validateString(value string, ann *parser.Annotation) error {
	// Check minlen constraint
	if minlenStr := ann.GetConstraint("minlen"); minlenStr != "" {
		minlen, err := strconv.Atoi(minlenStr)
		if err == nil && len(value) < minlen {
			return fmt.Errorf("length %d is below minimum %d", len(value), minlen)
		}
	}

	// Check maxlen constraint
	if maxlenStr := ann.GetConstraint("maxlen"); maxlenStr != "" {
		maxlen, err := strconv.Atoi(maxlenStr)
		if err == nil && len(value) > maxlen {
			return fmt.Errorf("length %d exceeds maximum %d", len(value), maxlen)
		}
	}

	// Check pattern constraint
	if pattern := ann.GetConstraint("pattern"); pattern != "" {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid pattern: %v", err)
		}
		if !re.MatchString(value) {
			return fmt.Errorf("value %q does not match pattern %s", value, pattern)
		}
	}

	return nil
}

func validateEnum(value string, ann *parser.Annotation) error {
	if value == "" {
		return fmt.Errorf("value is required for enum")
	}

	optionsStr := ann.GetConstraint("options")
	if optionsStr == "" {
		return fmt.Errorf("enum has no options defined")
	}

	options := strings.Split(optionsStr, ",")
	for _, opt := range options {
		if strings.TrimSpace(opt) == value {
			return nil
		}
	}

	return fmt.Errorf("value %q not in allowed options: %s", value, optionsStr)
}

func validateBoolean(value string) error {
	if value == "" {
		return fmt.Errorf("value is required for boolean")
	}

	if !validBooleans[strings.ToLower(value)] {
		return fmt.Errorf("invalid boolean value %q (use: true/false, yes/no, 1/0, on/off)", value)
	}

	return nil
}

func validateObject(value string, ann *parser.Annotation) error {
	if value == "" {
		return fmt.Errorf("value is required for object")
	}

	format := ann.GetConstraint("format")
	if format == "" {
		format = "json" // Default to JSON
	}

	switch format {
	case "json":
		var js interface{}
		if err := json.Unmarshal([]byte(value), &js); err != nil {
			return fmt.Errorf("invalid JSON: %v", err)
		}
	case "yaml":
		var yml interface{}
		if err := yaml.Unmarshal([]byte(value), &yml); err != nil {
			return fmt.Errorf("invalid YAML: %v", err)
		}
	default:
		return fmt.Errorf("unknown object format: %s", format)
	}

	return nil
}

// ValidateVariable validates a single variable and returns a ValidationError if invalid.
func ValidateVariable(v *parser.Variable) *ValidationError {
	if v.Annotation == nil {
		return nil
	}

	err := ValidateValue(v.Value, v.Annotation)
	if err != nil {
		return &ValidationError{
			Variable:   v.Name,
			LineNumber: v.LineNumber,
			Message:    err.Error(),
			Suggestion: GetSuggestion(v.Annotation),
			Example:    GetExample(v.Annotation),
			Type:       getErrorType(err),
		}
	}

	return nil
}

// ValidateEnvFile validates all variables in an EnvFile and returns a ValidationResult.
func ValidateEnvFile(envFile *parser.EnvFile) *ValidationResult {
	result := NewValidationResult()

	for i := range envFile.Variables {
		if err := ValidateVariable(&envFile.Variables[i]); err != nil {
			result.AddError(*err)
		}
	}

	return result
}

// GetSuggestion generates a helpful suggestion based on the annotation.
func GetSuggestion(ann *parser.Annotation) string {
	switch ann.Type {
	case parser.TypeInt:
		if min := ann.GetConstraint("min"); min != "" {
			if max := ann.GetConstraint("max"); max != "" {
				return fmt.Sprintf("Enter an integer between %s and %s", min, max)
			}
			return fmt.Sprintf("Enter an integer >= %s", min)
		}
		if max := ann.GetConstraint("max"); max != "" {
			return fmt.Sprintf("Enter an integer <= %s", max)
		}
		return "Enter a valid integer"
	case parser.TypeNumeric:
		return "Enter a valid number"
	case parser.TypeString:
		if pattern := ann.GetConstraint("pattern"); pattern != "" {
			return fmt.Sprintf("Enter a value matching pattern: %s", pattern)
		}
		return "Enter a valid string"
	case parser.TypeEnum:
		return fmt.Sprintf("Choose one of: %s", ann.GetConstraint("options"))
	case parser.TypeBoolean:
		return "Enter true/false, yes/no, 1/0, or on/off"
	case parser.TypeObject:
		return fmt.Sprintf("Enter valid %s", ann.GetConstraint("format"))
	default:
		return "Enter a valid value"
	}
}

// GetExample generates an example value based on the annotation.
func GetExample(ann *parser.Annotation) string {
	switch ann.Type {
	case parser.TypeInt:
		if min := ann.GetConstraint("min"); min != "" {
			return min
		}
		return "42"
	case parser.TypeNumeric:
		return "3.14"
	case parser.TypeString:
		return "example_value"
	case parser.TypeEnum:
		options := ann.GetConstraint("options")
		if options != "" {
			parts := strings.Split(options, ",")
			if len(parts) > 0 {
				return strings.TrimSpace(parts[0])
			}
		}
		return ""
	case parser.TypeBoolean:
		return "true"
	case parser.TypeObject:
		if ann.GetConstraint("format") == "yaml" {
			return "key: value"
		}
		return `{"key": "value"}`
	default:
		return ""
	}
}

// getErrorType determines the error type from the validation error message.
func getErrorType(err error) ErrorType {
	msg := err.Error()
	if strings.Contains(msg, "required") {
		return ErrorMissingRequired
	}
	if strings.Contains(msg, "not in allowed") || strings.Contains(msg, "does not match") {
		return ErrorConstraintViolation
	}
	return ErrorInvalidType
}
