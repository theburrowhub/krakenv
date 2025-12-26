// Package inspector provides functionality for comparing env files.
package inspector

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/theburrowhub/krakenv/internal/parser"
	"github.com/theburrowhub/krakenv/internal/tui/components"
	"github.com/theburrowhub/krakenv/internal/validator"
)

// InspectionResult holds the results of comparing distributable and target files.
type InspectionResult struct {
	DistPath      string                      // Distributable file path
	TargetPath    string                      // Target file path
	MissingInEnv  []parser.Variable           // Variables in dist but not in target
	ExtraInEnv    []parser.Variable           // Variables in target but not in dist
	InvalidValues []validator.ValidationError // Variables with invalid values
	ValidCount    int                         // Count of valid variables
}

// Inspect compares a target file against the distributable.
func Inspect(distFile, targetFile *parser.EnvFile) *InspectionResult {
	result := &InspectionResult{
		DistPath:      distFile.Path,
		TargetPath:    targetFile.Path,
		MissingInEnv:  make([]parser.Variable, 0),
		ExtraInEnv:    make([]parser.Variable, 0),
		InvalidValues: make([]validator.ValidationError, 0),
	}

	// Track variable names in dist
	distVarNames := make(map[string]bool)
	for _, v := range distFile.Variables {
		distVarNames[v.Name] = true
	}

	// Check each dist variable
	for _, distVar := range distFile.Variables {
		targetVar := targetFile.GetVariable(distVar.Name)

		if targetVar == nil {
			// Missing in target
			result.MissingInEnv = append(result.MissingInEnv, distVar)
			continue
		}

		// Validate if annotation exists
		if distVar.Annotation != nil {
			if err := validator.ValidateValue(targetVar.Value, distVar.Annotation); err != nil {
				result.InvalidValues = append(result.InvalidValues, validator.ValidationError{
					Variable:   distVar.Name,
					LineNumber: targetVar.LineNumber,
					Message:    err.Error(),
					Suggestion: validator.GetSuggestion(distVar.Annotation),
					Example:    validator.GetExample(distVar.Annotation),
				})
				continue
			}
		}

		result.ValidCount++
	}

	// Find extra variables in target
	for _, targetVar := range targetFile.Variables {
		if !distVarNames[targetVar.Name] {
			result.ExtraInEnv = append(result.ExtraInEnv, targetVar)
		}
	}

	return result
}

// HasDiscrepancies returns true if there are any discrepancies.
func (r *InspectionResult) HasDiscrepancies() bool {
	return len(r.MissingInEnv) > 0 || len(r.ExtraInEnv) > 0 || len(r.InvalidValues) > 0
}

// FormatReport returns a formatted text report.
func (r *InspectionResult) FormatReport() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("INSPECTION REPORT: %s vs %s\n\n", r.TargetPath, r.DistPath))

	// Missing variables
	if len(r.MissingInEnv) > 0 {
		b.WriteString(components.WarningStyle.Render(fmt.Sprintf("MISSING IN %s (%d):\n", r.TargetPath, len(r.MissingInEnv))))
		for _, v := range r.MissingInEnv {
			desc := ""
			typeStr := ""
			if v.Annotation != nil {
				desc = v.Annotation.PromptText
				typeStr = fmt.Sprintf("[%s]", v.Annotation.Type.String())
			}
			b.WriteString(fmt.Sprintf("  %-20s %q %s\n", v.Name, desc, typeStr))
		}
		b.WriteString("\n")
	}

	// Extra variables
	if len(r.ExtraInEnv) > 0 {
		b.WriteString(components.InfoStyle.Render(fmt.Sprintf("EXTRA IN %s (%d):\n", r.TargetPath, len(r.ExtraInEnv))))
		for _, v := range r.ExtraInEnv {
			b.WriteString(fmt.Sprintf("  %-20s (not in distributable)\n", v.Name))
		}
		b.WriteString("\n")
	}

	// Invalid values
	if len(r.InvalidValues) > 0 {
		b.WriteString(components.ErrorStyle.Render(fmt.Sprintf("INVALID VALUES (%d):\n", len(r.InvalidValues))))
		for _, err := range r.InvalidValues {
			b.WriteString(fmt.Sprintf("  %-20s %s\n", err.Variable, err.Message))
		}
		b.WriteString("\n")
	}

	// Summary
	b.WriteString(fmt.Sprintf("Summary: %d missing, %d extra, %d invalid, %d valid\n",
		len(r.MissingInEnv), len(r.ExtraInEnv), len(r.InvalidValues), r.ValidCount))

	return b.String()
}

// JSONReport represents the JSON output format.
type JSONReport struct {
	Missing []JSONVariable        `json:"missing"`
	Extra   []JSONVariable        `json:"extra"`
	Invalid []JSONValidationError `json:"invalid"`
}

// JSONVariable represents a variable in JSON output.
type JSONVariable struct {
	Name   string `json:"name"`
	Prompt string `json:"prompt,omitempty"`
	Type   string `json:"type,omitempty"`
	Value  string `json:"value,omitempty"`
}

// JSONValidationError represents a validation error in JSON output.
type JSONValidationError struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
	Error string `json:"error"`
}

// FormatJSON returns a JSON formatted report.
func (r *InspectionResult) FormatJSON() (string, error) {
	report := JSONReport{
		Missing: make([]JSONVariable, 0, len(r.MissingInEnv)),
		Extra:   make([]JSONVariable, 0, len(r.ExtraInEnv)),
		Invalid: make([]JSONValidationError, 0, len(r.InvalidValues)),
	}

	for _, v := range r.MissingInEnv {
		jv := JSONVariable{Name: v.Name}
		if v.Annotation != nil {
			jv.Prompt = v.Annotation.PromptText
			jv.Type = v.Annotation.Type.String()
		}
		report.Missing = append(report.Missing, jv)
	}

	for _, v := range r.ExtraInEnv {
		report.Extra = append(report.Extra, JSONVariable{
			Name:  v.Name,
			Value: v.Value,
		})
	}

	for _, err := range r.InvalidValues {
		report.Invalid = append(report.Invalid, JSONValidationError{
			Name:  err.Variable,
			Error: err.Message,
		})
	}

	jsonBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
