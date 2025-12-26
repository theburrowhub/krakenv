// Package generator provides functionality for generating environment files.
package generator

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/theburrowhub/krakenv/internal/parser"
)

// Generator handles generation of environment files from distributables.
type Generator struct {
	DistFile        *parser.EnvFile
	TargetPath      string
	TargetFile      *parser.EnvFile
	KeepAnnotations bool
}

// NewGenerator creates a new Generator for the given distributable.
func NewGenerator(distFile *parser.EnvFile, targetPath string) *Generator {
	return &Generator{
		DistFile:   distFile,
		TargetPath: targetPath,
	}
}

// LoadTarget loads an existing target file if it exists.
func (g *Generator) LoadTarget() error {
	if _, err := os.Stat(g.TargetPath); os.IsNotExist(err) {
		g.TargetFile = nil
		return nil
	}

	target, err := parser.ParseEnvFile(g.TargetPath)
	if err != nil {
		return fmt.Errorf("failed to parse target file: %w", err)
	}

	g.TargetFile = target
	return nil
}

// GetVariablesToPrompt returns variables that need user input.
// A variable needs prompting if:
// - It has an annotation (interactive config)
// - AND has no value in dist AND has no value in target
func (g *Generator) GetVariablesToPrompt() []parser.Variable {
	var toPrompt []parser.Variable

	for _, v := range g.DistFile.Variables {
		if v.Annotation == nil {
			continue // No annotation = no prompting needed
		}

		// Check if dist has a default value
		if v.Value != "" {
			continue // Has default value, no prompt needed
		}

		// Check if target has a valid value
		if g.TargetFile != nil {
			if existing := g.TargetFile.GetVariable(v.Name); existing != nil {
				if existing.Value != "" {
					// Has a value - check if valid
					// For now, assume existing values are valid (validation done separately)
					continue
				}
			}
		}

		// No existing valid value - needs prompting
		toPrompt = append(toPrompt, v)
	}

	return toPrompt
}

// MergeVariables creates the final list of variables for output.
// Priority: Target values > User-provided values > Dist defaults.
func (g *Generator) MergeVariables(userValues map[string]string) []parser.Variable {
	result := make([]parser.Variable, len(g.DistFile.Variables))

	for i, v := range g.DistFile.Variables {
		result[i] = v

		// Check for user-provided value
		if userValue, ok := userValues[v.Name]; ok {
			result[i].Value = userValue
			result[i].IsSet = true
			continue
		}

		// Check for existing target value
		if g.TargetFile != nil {
			if existing := g.TargetFile.GetVariable(v.Name); existing != nil && existing.Value != "" {
				result[i].Value = existing.Value
				result[i].IsSet = true
				continue
			}
		}

		// Use dist default (already set)
		result[i].IsSet = v.Value != ""
	}

	return result
}

// WriteFile writes the generated environment file to disk.
func (g *Generator) WriteFile(variables []parser.Variable) error {
	file, err := os.Create(g.TargetPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Write config block if present
	if g.DistFile.Config != nil {
		for _, line := range formatConfigBlock(g.DistFile.Config) {
			fmt.Fprintln(writer, line)
		}
		fmt.Fprintln(writer)
	}

	// Track current comment section
	var lastCommentIdx int
	for i, comment := range g.DistFile.Comments {
		if comment.LineNumber > 0 {
			lastCommentIdx = i
		}
	}
	_ = lastCommentIdx // Suppress unused warning for now

	// Write variables with their comments
	for i, v := range variables {
		// Write any preceding comments
		for _, comment := range g.DistFile.Comments {
			// Find comments that belong before this variable
			if i > 0 && i < len(g.DistFile.Variables) {
				prevLine := g.DistFile.Variables[i-1].LineNumber
				currLine := v.LineNumber
				if comment.LineNumber > prevLine && comment.LineNumber < currLine {
					fmt.Fprintf(writer, "# %s\n", comment.Text)
				}
			} else if i == 0 && comment.LineNumber < v.LineNumber {
				fmt.Fprintf(writer, "# %s\n", comment.Text)
			}
		}

		// Write variable
		line := v.Name + "=" + v.Value
		if g.KeepAnnotations && v.Annotation != nil {
			line += " " + parser.FormatAnnotation(v.Annotation)
		}
		fmt.Fprintln(writer, line)
	}

	return writer.Flush()
}

// formatConfigBlock formats the config as comment lines.
func formatConfigBlock(config *parser.KrakenvConfig) []string {
	var lines []string

	if len(config.Environments) > 0 {
		lines = append(lines, "#krakenv:environments="+strings.Join(config.Environments, ","))
	}
	if config.Strict {
		lines = append(lines, "#krakenv:strict=true")
	}

	return lines
}

// GenerateResult holds the result of a generate operation.
type GenerateResult struct {
	Created   bool   // True if file was created (vs updated)
	Path      string // Output file path
	Variables int    // Number of variables written
	Prompted  int    // Number of variables that were prompted
	Skipped   int    // Number of variables with existing valid values
}

// Generate is a convenience function that generates a file in one call.
func Generate(distPath, targetPath string, values map[string]string, keepAnnotations bool) (*GenerateResult, error) {
	// Parse distributable
	distFile, err := parser.ParseEnvFile(distPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse distributable: %w", err)
	}

	// Create generator
	gen := NewGenerator(distFile, targetPath)
	gen.KeepAnnotations = keepAnnotations

	// Load existing target
	fileExists := true
	if err := gen.LoadTarget(); err != nil {
		return nil, err
	}
	if gen.TargetFile == nil {
		fileExists = false
	}

	// Merge values
	variables := gen.MergeVariables(values)

	// Write output
	if err := gen.WriteFile(variables); err != nil {
		return nil, err
	}

	return &GenerateResult{
		Created:   !fileExists,
		Path:      targetPath,
		Variables: len(variables),
		Prompted:  len(values),
		Skipped:   len(variables) - len(values),
	}, nil
}
