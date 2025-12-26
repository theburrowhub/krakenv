package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theburrowhub/krakenv/internal/parser"
)

var (
	addType     string
	addPrompt   string
	addDefault  string
	addMin      string
	addMax      string
	addMinlen   string
	addMaxlen   string
	addPattern  string
	addOptions  string
	addFormat   string
	addOptional bool
	addSecret   bool
)

var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new annotated variable to the distributable",
	Long: `Add a new annotated variable to the distributable file.

The variable name must be uppercase letters, numbers, and underscores,
starting with a letter.

Examples:
  krakenv add API_URL --type string --prompt "API base URL?"
  krakenv add MAX_CONNECTIONS --type int --min 1 --max 100 --default 10
  krakenv add LOG_LEVEL --type enum --options "debug,info,warn,error" --default info
  krakenv add DB_PASSWORD --type string --prompt "Database password?" --secret
  krakenv add ENABLE_METRICS --type boolean --optional --default false`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

var variableNameRegex = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

func init() {
	addCmd.Flags().StringVarP(&addType, "type", "t", "string",
		"Variable type (string, int, numeric, boolean, enum, object)")
	addCmd.Flags().StringVarP(&addPrompt, "prompt", "p", "",
		"Prompt message for the wizard")
	addCmd.Flags().StringVarP(&addDefault, "default", "D", "",
		"Default value")
	addCmd.Flags().StringVar(&addMin, "min", "",
		"Minimum value (int/numeric)")
	addCmd.Flags().StringVar(&addMax, "max", "",
		"Maximum value (int/numeric)")
	addCmd.Flags().StringVar(&addMinlen, "minlen", "",
		"Minimum length (string)")
	addCmd.Flags().StringVar(&addMaxlen, "maxlen", "",
		"Maximum length (string)")
	addCmd.Flags().StringVar(&addPattern, "pattern", "",
		"Regex pattern (string)")
	addCmd.Flags().StringVarP(&addOptions, "options", "o", "",
		"Comma-separated options (enum)")
	addCmd.Flags().StringVar(&addFormat, "format", "",
		"Object format: json or yaml")
	addCmd.Flags().BoolVar(&addOptional, "optional", false,
		"Mark as optional")
	addCmd.Flags().BoolVar(&addSecret, "secret", false,
		"Mark as secret (hides input)")

	rootCmd.AddCommand(addCmd)
}

func runAdd(_ *cobra.Command, args []string) error {
	varName := args[0]

	// Validate variable name
	if !variableNameRegex.MatchString(varName) {
		return fmt.Errorf("invalid variable name %q: must be uppercase letters, numbers, and underscores, starting with a letter", varName)
	}

	// Check distributable exists
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		return fmt.Errorf("distributable not found: %s\nRun 'krakenv init' to create one", distPath)
	}

	// Parse existing distributable
	distFile, err := parser.ParseEnvFile(distPath)
	if err != nil {
		return fmt.Errorf("failed to parse distributable: %w", err)
	}

	// Check for duplicate
	if distFile.HasVariable(varName) {
		fmt.Fprintf(os.Stderr, "ERROR: Variable %s already exists in %s\n", varName, distPath)
		os.Exit(2)
	}

	// Build annotation
	annotation := buildAnnotation()

	// Build line
	line := buildVariableLine(varName, annotation)

	// Check if file ends with newline
	needsNewline := false
	if stat, err := os.Stat(distPath); err == nil && stat.Size() > 0 {
		checkFile, err := os.Open(distPath)
		if err == nil {
			checkFile.Seek(-1, 2)
			buf := make([]byte, 1)
			checkFile.Read(buf)
			checkFile.Close()
			if buf[0] != '\n' {
				needsNewline = true
			}
		}
	}

	// Append to file
	f, err := os.OpenFile(distPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Add newline if file doesn't end with one
	if needsNewline {
		fmt.Fprintln(f)
	}

	if _, err := fmt.Fprintln(f, line); err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}

	if !quiet {
		fmt.Printf("✓ Added: %s\n", line)
	}

	return nil
}

func buildAnnotation() string {
	// Default prompt if not provided
	prompt := addPrompt
	if prompt == "" {
		prompt = "Enter " + strings.ToLower(addType) + " value"
	}

	// Build parts
	parts := []string{addType}

	// Add constraints based on type
	switch addType {
	case "int", "numeric":
		if addMin != "" {
			parts = append(parts, "min:"+addMin)
		}
		if addMax != "" {
			parts = append(parts, "max:"+addMax)
		}
	case "string":
		if addMinlen != "" {
			parts = append(parts, "minlen:"+addMinlen)
		}
		if addMaxlen != "" {
			parts = append(parts, "maxlen:"+addMaxlen)
		}
		if addPattern != "" {
			parts = append(parts, "pattern:"+addPattern)
		}
	case "enum":
		if addOptions != "" {
			parts = append(parts, "options:"+addOptions)
		}
	case "object":
		if addFormat != "" {
			parts = append(parts, "format:"+addFormat)
		}
	}

	// Add modifiers
	if addOptional {
		parts = append(parts, "optional")
	}
	if addSecret {
		parts = append(parts, "secret")
	}

	return "#prompt:" + prompt + "|" + strings.Join(parts, ";")
}

func buildVariableLine(name, annotation string) string {
	line := name + "=" + addDefault
	if annotation != "" {
		line += " " + annotation
	}
	return line
}
