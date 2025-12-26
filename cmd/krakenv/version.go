// Package main contains the CLI commands for krakenv.
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version and build information",
	Long:  `Display the version, commit hash, and build date of krakenv.`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("krakenv version %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
