package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theburrowhub/krakenv/internal/parser"
)

var (
	initPath         string
	initEnvironments string
	initForce        bool
	initTemplate     bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new distributable file with optional interactive wizard",
	Long: `Initialize a new distributable file for krakenv.

By default, runs an interactive wizard to add variables.
Use --template to create a file with example comments only.

Examples:
  krakenv init
  krakenv init --path config/.env.template
  krakenv init --template
  krakenv init --force`,
	Args: cobra.NoArgs,
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVarP(&initPath, "path", "p", ".env.dist",
		"Output path for the distributable")
	initCmd.Flags().StringVarP(&initEnvironments, "environments", "e", "local",
		"Comma-separated environments")
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false,
		"Overwrite existing file")
	initCmd.Flags().BoolVarP(&initTemplate, "template", "t", false,
		"Create with example comments only (skip wizard)")

	rootCmd.AddCommand(initCmd)
}

func runInit(_ *cobra.Command, _ []string) error {
	// Check if file exists
	if _, err := os.Stat(initPath); err == nil && !initForce {
		return fmt.Errorf("file %s already exists (use --force to overwrite)", initPath)
	}

	// Create file
	f, err := os.Create(initPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	writer := bufio.NewWriter(f)

	// Write config block
	envList := strings.Split(initEnvironments, ",")
	for i := range envList {
		envList[i] = strings.TrimSpace(envList[i])
	}
	fmt.Fprintf(writer, "#krakenv:environments=%s\n", strings.Join(envList, ","))
	fmt.Fprintln(writer, "#krakenv:strict=false")
	fmt.Fprintln(writer)

	// Write header comments
	fmt.Fprintln(writer, "# ========================================")
	fmt.Fprintln(writer, "# Krakenv Environment Template")
	fmt.Fprintln(writer, "# ========================================")
	fmt.Fprintln(writer, "#")
	fmt.Fprintln(writer, "# Annotation syntax:")
	fmt.Fprintln(writer, "#   VAR=default #prompt:Question?|type;constraint:value")
	fmt.Fprintln(writer, "#")
	fmt.Fprintln(writer, "# Types: string, int, numeric, boolean, enum, object")
	fmt.Fprintln(writer, "# Modifiers: optional, secret")
	fmt.Fprintln(writer, "#")
	fmt.Fprintln(writer, "# Examples:")
	fmt.Fprintln(writer, "#   PORT=3000 #prompt:Server port?|int;min:1;max:65535")
	fmt.Fprintln(writer, "#   ENV= #prompt:Environment?|enum;options:dev,staging,prod")
	fmt.Fprintln(writer, "#   API_KEY= #prompt:API key?|string;secret")
	fmt.Fprintln(writer, "#")
	fmt.Fprintln(writer, "# Run: krakenv generate .env.local")
	fmt.Fprintln(writer, "# ========================================")
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "# Add your variables below:")
	fmt.Fprintln(writer)

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}

	if !quiet {
		fmt.Printf("✓ Created %s\n", initPath)
		if !initTemplate {
			fmt.Println("\nTo add variables interactively:")
			fmt.Println("  krakenv add VAR_NAME --type string --prompt \"Question?\"")
			fmt.Println("\nTo generate environment files:")
			fmt.Println("  krakenv generate .env.local")
		}
	}

	// Run wizard if not template mode
	if !initTemplate && !nonInteractive {
		return runInitWizard(initPath)
	}

	return nil
}

func runInitWizard(path string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\nStarting variable wizard (empty name to finish):")

	for {
		// Variable name
		fmt.Print("Variable name? ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)

		if name == "" {
			break
		}

		// Validate name
		if !variableNameRegex.MatchString(name) {
			fmt.Println("  Invalid name. Use uppercase letters, numbers, underscores.")
			continue
		}

		// Type
		fmt.Print("Type? [string/int/numeric/boolean/enum/object]: ")
		typeStr, _ := reader.ReadString('\n')
		typeStr = strings.TrimSpace(typeStr)
		if typeStr == "" {
			typeStr = "string"
		}

		// Constraints
		fmt.Print("Constraints? (e.g., min:1;max:100): ")
		constraints, _ := reader.ReadString('\n')
		constraints = strings.TrimSpace(constraints)

		// Default
		fmt.Print("Default value? (optional): ")
		defaultVal, _ := reader.ReadString('\n')
		defaultVal = strings.TrimSpace(defaultVal)

		// Prompt
		fmt.Print("Prompt message: ")
		prompt, _ := reader.ReadString('\n')
		prompt = strings.TrimSpace(prompt)
		if prompt == "" {
			prompt = "Enter " + name
		}

		// Optional
		fmt.Print("Is it optional? [y/n]: ")
		optionalStr, _ := reader.ReadString('\n')
		optional := strings.ToLower(strings.TrimSpace(optionalStr)) == "y"

		// Secret
		fmt.Print("Is it secret? [y/n]: ")
		secretStr, _ := reader.ReadString('\n')
		secret := strings.ToLower(strings.TrimSpace(secretStr)) == "y"

		// Build and append
		annotation := buildWizardAnnotation(typeStr, prompt, constraints, optional, secret)
		line := name + "=" + defaultVal + " " + annotation

		// Append to file
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		fmt.Fprintln(f, line)
		f.Close()

		fmt.Printf("\n✓ Added: %s\n\n", line)
	}

	// Count variables
	distFile, err := parser.ParseEnvFile(path)
	if err != nil {
		return err
	}

	fmt.Printf("\n✓ Created %s with %d variables\n", path, len(distFile.Variables))
	fmt.Printf("  Run: krakenv generate .env.local\n")

	return nil
}

func buildWizardAnnotation(typeStr, prompt, constraints string, optional, secret bool) string {
	parts := []string{typeStr}

	if constraints != "" {
		for _, c := range strings.Split(constraints, ";") {
			c = strings.TrimSpace(c)
			if c != "" {
				parts = append(parts, c)
			}
		}
	}

	if optional {
		parts = append(parts, "optional")
	}
	if secret {
		parts = append(parts, "secret")
	}

	return "#prompt:" + prompt + "|" + strings.Join(parts, ";")
}
