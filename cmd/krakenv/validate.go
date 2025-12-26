package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/theburrowhub/krakenv/internal/parser"
	"github.com/theburrowhub/krakenv/internal/validator"
)

var (
	validateStrict bool
)

var validateCmd = &cobra.Command{
	Use:   "validate <target>",
	Short: "Validate an environment file against the distributable annotations",
	Long: `Validate that all values in an environment file comply with the
annotations defined in the distributable.

Useful for CI/CD pipelines or pre-commit hooks to catch configuration errors early.

Exit codes:
  0 - All validations passed
  1 - Validation errors found
  2 - File not found or unreadable

Examples:
  krakenv validate .env.local
  krakenv validate .env.testing --strict
  krakenv validate .env.production --non-interactive`,
	Args: cobra.ExactArgs(1),
	RunE: runValidate,
}

func init() {
	validateCmd.Flags().BoolVarP(&validateStrict, "strict", "s", false,
		"Require all variables to have annotations")

	rootCmd.AddCommand(validateCmd)
}

func runValidate(_ *cobra.Command, args []string) error {
	targetPath := args[0]

	// Check target exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "ERROR: File not found: %s\n", targetPath)
		os.Exit(2)
	}

	// Parse distributable
	distFile, err := parser.ParseEnvFile(distPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to parse distributable %s: %v\n", distPath, err)
		os.Exit(2)
	}

	// Parse target file
	targetFile, err := parser.ParseEnvFile(targetPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to parse target %s: %v\n", targetPath, err)
		os.Exit(2)
	}

	// Override strict from config if set
	strictMode := validateStrict
	if !strictMode && distFile.Config != nil {
		strictMode = distFile.Config.Strict
	}

	// Validate
	result := validateFile(distFile, targetFile, strictMode)

	// Output results
	if !quiet {
		fmt.Print(result.FormatErrors(targetPath))
	}

	if !result.Valid {
		os.Exit(1)
	}

	return nil
}

func validateFile(distFile, targetFile *parser.EnvFile, strict bool) *validator.ValidationResult {
	result := validator.NewValidationResult()

	// Validate each variable in target against dist annotations
	for _, distVar := range distFile.Variables {
		targetVar := targetFile.GetVariable(distVar.Name)

		// Check if variable exists in target
		if targetVar == nil {
			// Missing variable - only error if required (has annotation and not optional)
			if distVar.Annotation != nil && !distVar.Annotation.IsOptional {
				result.AddError(validator.NewMissingRequiredError(
					distVar.Name,
					0,
					distVar.Annotation.PromptText,
				))
			}
			continue
		}

		// Skip validation if no annotation
		if distVar.Annotation == nil {
			if strict {
				result.AddError(validator.ValidationError{
					Variable:   distVar.Name,
					LineNumber: targetVar.LineNumber,
					Message:    "Variable has no annotation (strict mode)",
					Suggestion: "Add an annotation to the distributable",
					Type:       validator.ErrorAnnotationSyntax,
				})
			}
			continue
		}

		// Validate value
		if err := validator.ValidateValue(targetVar.Value, distVar.Annotation); err != nil {
			result.AddError(validator.ValidationError{
				Variable:   distVar.Name,
				LineNumber: targetVar.LineNumber,
				Message:    err.Error(),
				Suggestion: validator.GetSuggestion(distVar.Annotation),
				Example:    validator.GetExample(distVar.Annotation),
				Type:       validator.ErrorInvalidType,
			})
		}
	}

	return result
}
