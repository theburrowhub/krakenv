package main

import (
	"bufio"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/theburrowhub/krakenv/internal/inspector"
	"github.com/theburrowhub/krakenv/internal/parser"
	"github.com/theburrowhub/krakenv/internal/tui/sync"
)

var (
	inspectSync bool
	inspectJSON bool
)

var inspectCmd = &cobra.Command{
	Use:   "inspect <target>",
	Short: "Compare environment file with distributable; identify discrepancies",
	Long: `Compare an environment file with the distributable to identify:
  - Variables in distributable missing from environment file
  - Variables in environment file not present in distributable
  - Variables with invalid values

Exit codes:
  0 - No discrepancies found
  1 - Discrepancies found (report generated)
  2 - File not found or unreadable

Examples:
  krakenv inspect .env.local
  krakenv inspect .env.local --sync
  krakenv inspect .env.testing --json | jq '.missing | length'`,
	Args: cobra.ExactArgs(1),
	RunE: runInspect,
}

func init() {
	inspectCmd.Flags().BoolVarP(&inspectSync, "sync", "s", false,
		"Interactively sync discrepancies")
	inspectCmd.Flags().BoolVarP(&inspectJSON, "json", "j", false,
		"Output as JSON (for scripting)")

	rootCmd.AddCommand(inspectCmd)
}

func runInspect(_ *cobra.Command, args []string) error {
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

	// Run inspection
	result := inspector.Inspect(distFile, targetFile)

	// Handle sync mode
	if inspectSync && result.HasDiscrepancies() {
		if nonInteractive {
			return handleNonInteractiveSync(result, distFile, targetFile, targetPath)
		}
		// Interactive sync
		return runInteractiveSync(result, distFile, targetFile, targetPath)
	}

	// Output results
	if inspectJSON {
		jsonOutput, err := result.FormatJSON()
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(jsonOutput)
	} else if !quiet {
		fmt.Print(result.FormatReport())
	}

	if result.HasDiscrepancies() {
		os.Exit(1)
	}

	return nil
}

func runInteractiveSync(result *inspector.InspectionResult, distFile, targetFile *parser.EnvFile, targetPath string) error {
	m := sync.New(result, distFile, targetFile)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("sync wizard error: %w", err)
	}

	syncModel := finalModel.(sync.Model)

	if syncModel.IsAborted() {
		fmt.Println("Sync aborted by user.")
		return nil
	}

	// Apply resolutions
	resolutions := syncModel.GetResolutions()
	return applyResolutions(resolutions, distFile, targetFile, targetPath)
}

func applyResolutions(resolutions []sync.Resolution, _, targetFile *parser.EnvFile, targetPath string) error {
	// Build updated variables map
	updates := make(map[string]string)
	removes := make(map[string]bool)
	addToDist := make([]sync.Resolution, 0)

	for _, r := range resolutions {
		switch r.Action {
		case sync.ActionAdd:
			updates[r.Variable.Name] = r.NewValue
		case sync.ActionRemove:
			removes[r.Variable.Name] = true
		case sync.ActionAddToDist:
			addToDist = append(addToDist, r)
		}
	}

	// Update target file
	if len(updates) > 0 || len(removes) > 0 {
		if err := updateTargetFile(targetFile, targetPath, updates, removes); err != nil {
			return fmt.Errorf("failed to update target file: %w", err)
		}
		if !quiet {
			fmt.Printf("✓ Updated %s\n", targetPath)
		}
	}

	// Add to distributable
	if len(addToDist) > 0 {
		if err := addToDistributable(distPath, addToDist); err != nil {
			return fmt.Errorf("failed to update distributable: %w", err)
		}
		if !quiet {
			fmt.Printf("✓ Updated %s with %d new variable(s)\n", distPath, len(addToDist))
		}
	}

	return nil
}

func updateTargetFile(targetFile *parser.EnvFile, targetPath string, updates map[string]string, removes map[string]bool) error {
	file, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Write existing variables (updated or kept)
	for _, v := range targetFile.Variables {
		if removes[v.Name] {
			continue // Skip removed variables
		}

		value := v.Value
		if newVal, ok := updates[v.Name]; ok {
			value = newVal
			delete(updates, v.Name) // Mark as written
		}

		fmt.Fprintf(writer, "%s=%s\n", v.Name, value)
	}

	// Write new variables (from updates that weren't in target)
	for name, value := range updates {
		fmt.Fprintf(writer, "%s=%s\n", name, value)
	}

	return writer.Flush()
}

func addToDistributable(distPath string, resolutions []sync.Resolution) error {
	// First, check if file ends with newline
	needsNewline := false
	if stat, err := os.Stat(distPath); err == nil && stat.Size() > 0 {
		f, err := os.Open(distPath)
		if err == nil {
			// Seek to last byte
			f.Seek(-1, 2)
			buf := make([]byte, 1)
			f.Read(buf)
			f.Close()
			if buf[0] != '\n' {
				needsNewline = true
			}
		}
	}

	file, err := os.OpenFile(distPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Add newline if file doesn't end with one
	if needsNewline {
		fmt.Fprintln(writer)
	}

	for _, r := range resolutions {
		v := r.Variable
		if r.Annotation != nil {
			// Add with annotation
			ann := parser.FormatAnnotation(r.Annotation)
			fmt.Fprintf(writer, "%s=%s %s\n", v.Name, v.Value, ann)
		} else {
			// Add as simple variable without annotation
			fmt.Fprintf(writer, "%s=%s\n", v.Name, v.Value)
		}
	}

	return writer.Flush()
}

func handleNonInteractiveSync(result *inspector.InspectionResult, _, targetFile *parser.EnvFile, targetPath string) error {
	// Per FR-032a/b: Try to auto-resolve with defaults for optional variables
	var unresolvable []string
	updates := make(map[string]string)

	for _, v := range result.MissingInEnv {
		// Check if optional or has default
		if v.Annotation != nil && v.Annotation.IsOptional {
			updates[v.Name] = "" // Empty value for optional
			continue
		}
		if v.Value != "" {
			updates[v.Name] = v.Value // Use default value
			continue
		}
		unresolvable = append(unresolvable, v.Name)
	}

	if len(unresolvable) > 0 {
		fmt.Fprintf(os.Stderr, "ERROR: Cannot sync in non-interactive mode\n\n")
		fmt.Fprintf(os.Stderr, "The following required variables cannot be resolved:\n")
		for _, name := range unresolvable {
			fmt.Fprintf(os.Stderr, "  - %s\n", name)
		}
		fmt.Fprintf(os.Stderr, "\nRun without --non-interactive to resolve interactively.\n")
		os.Exit(2)
	}

	// Apply updates
	if len(updates) > 0 {
		if err := updateTargetFile(targetFile, targetPath, updates, nil); err != nil {
			return fmt.Errorf("failed to update target file: %w", err)
		}
		if !quiet {
			fmt.Printf("✓ Auto-synced %d variable(s) in %s\n", len(updates), targetPath)
		}
	}

	return nil
}
