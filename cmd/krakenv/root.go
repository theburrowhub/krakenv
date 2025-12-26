// Package main contains the CLI commands for krakenv.
package main

import (
	"github.com/spf13/cobra"
)

var (
	// Global flags.
	distPath       string
	nonInteractive bool
	quiet          bool
	verbose        bool
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "krakenv",
	Short: "Environment variable management with annotation-based wizards",
	Long: `Krakenv - When envs get complex, release the krakenv

Krakenv is a CLI tool for managing environment variable files (.env) with an
annotation-based configuration wizard system. It transforms the tedious process
of configuring environment files into a guided, validated experience.

Features:
  • Interactive wizard for environment configuration
  • Type validation (int, string, enum, boolean, object)
  • Constraints (min, max, pattern, options, etc.)
  • Multi-environment support
  • CI/CD integration with non-interactive mode

Example annotation:
  DB_PORT=5432 #prompt:Database port?|int;min:1;max:65535`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags available on all commands.
	rootCmd.PersistentFlags().StringVarP(&distPath, "dist", "d", ".env.dist",
		"Path to distributable file")
	rootCmd.PersistentFlags().BoolVarP(&nonInteractive, "non-interactive", "n", false,
		"Disable TUI; fail on unresolved variables")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false,
		"Suppress non-error output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"Enable detailed output")
}
