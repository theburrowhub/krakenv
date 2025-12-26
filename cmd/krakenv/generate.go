package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/theburrowhub/krakenv/internal/generator"
	"github.com/theburrowhub/krakenv/internal/parser"
	"github.com/theburrowhub/krakenv/internal/tui/wizard"
)

var (
	generateForce           bool
	generateAll             bool
	generateKeepAnnotations bool
)

var generateCmd = &cobra.Command{
	Use:   "generate <target>",
	Short: "Generate or update an environment file from the distributable",
	Long: `Generate an environment file from the distributable template.

The wizard will prompt for each variable that needs a value.
Variables with existing valid values are skipped.

Examples:
  krakenv generate .env.local
  krakenv generate .env.testing --dist config/env.template
  krakenv generate --all
  krakenv generate .env.local --non-interactive`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGenerate,
}

func init() {
	generateCmd.Flags().BoolVarP(&generateForce, "force", "f", false,
		"Overwrite existing file without confirmation")
	generateCmd.Flags().BoolVarP(&generateAll, "all", "a", false,
		"Generate all environments defined in config")
	generateCmd.Flags().BoolVarP(&generateKeepAnnotations, "keep-annotations", "k", false,
		"Preserve annotations in generated file")

	rootCmd.AddCommand(generateCmd)
}

func runGenerate(_ *cobra.Command, args []string) error {
	// Parse distributable
	distFile, err := parser.ParseEnvFile(distPath)
	if err != nil {
		return fmt.Errorf("failed to parse distributable %s: %w", distPath, err)
	}

	// Determine target(s)
	var targets []string
	if generateAll {
		if distFile.Config != nil && len(distFile.Config.Environments) > 0 {
			for _, env := range distFile.Config.Environments {
				targets = append(targets, ".env."+env)
			}
		} else {
			targets = []string{".env.local"}
		}
	} else if len(args) > 0 {
		targets = []string{args[0]}
	} else {
		return fmt.Errorf("target file required (e.g., .env.local) or use --all")
	}

	// Process each target
	for _, target := range targets {
		if err := generateTarget(distFile, target); err != nil {
			return err
		}
	}

	return nil
}

func generateTarget(distFile *parser.EnvFile, targetPath string) error {
	// Check if target exists
	if _, err := os.Stat(targetPath); err == nil && !generateForce {
		if nonInteractive {
			// In non-interactive mode, just proceed with update
		} else if !quiet {
			fmt.Printf("Target file %s exists, will update...\n", targetPath)
		}
	}

	// Create generator
	gen := generator.NewGenerator(distFile, targetPath)
	gen.KeepAnnotations = generateKeepAnnotations

	// Load existing target
	if err := gen.LoadTarget(); err != nil {
		return fmt.Errorf("failed to load target: %w", err)
	}

	// Get variables that need prompting
	toPrompt := gen.GetVariablesToPrompt()

	var userValues map[string]string

	if len(toPrompt) > 0 {
		if nonInteractive {
			// Non-interactive mode: fail if any variables need values
			return handleNonInteractive(toPrompt, targetPath)
		}

		// Run interactive wizard
		values, err := runWizard(toPrompt)
		if err != nil {
			return err
		}
		if values == nil {
			// User aborted
			return fmt.Errorf("generation aborted by user")
		}
		userValues = values
	} else {
		userValues = make(map[string]string)
	}

	// Merge and write
	variables := gen.MergeVariables(userValues)
	if err := gen.WriteFile(variables); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if !quiet {
		fmt.Printf("✓ Generated %s with %d variables\n", targetPath, len(variables))
	}

	return nil
}

func handleNonInteractive(toPrompt []parser.Variable, targetPath string) error {
	// Check if we can resolve with defaults
	var unresolved []string

	for _, v := range toPrompt {
		// Check if optional or has default
		if v.Annotation != nil && v.Annotation.IsOptional {
			continue // Optional can be empty
		}
		if v.Value != "" {
			continue // Has default
		}
		unresolved = append(unresolved, v.Name)
	}

	if len(unresolved) == 0 {
		// All can be resolved with defaults/optional
		return nil
	}

	// Cannot resolve - fail with descriptive error
	fmt.Fprintf(os.Stderr, "ERROR: Cannot generate %s in non-interactive mode\n\n", targetPath)
	fmt.Fprintf(os.Stderr, "The following variables require values:\n")
	for _, name := range unresolved {
		fmt.Fprintf(os.Stderr, "  - %s\n", name)
	}
	fmt.Fprintf(os.Stderr, "\nTo fix this:\n")
	fmt.Fprintf(os.Stderr, "  1. Run interactively: krakenv generate %s\n", targetPath)
	fmt.Fprintf(os.Stderr, "  2. Or set values in environment before running\n")
	fmt.Fprintf(os.Stderr, "  3. Or add default values to your distributable\n")

	os.Exit(2)
	return nil // Won't reach here
}

func runWizard(variables []parser.Variable) (map[string]string, error) {
	m := wizard.New(variables)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("wizard error: %w", err)
	}

	wizardModel := finalModel.(wizard.Model)

	if wizardModel.IsAborted() {
		return nil, nil // User aborted
	}

	return wizardModel.GetValues(), nil
}
