// Package main is the entry point for the krakenv CLI.
package main

import (
	"fmt"
	"os"
)

// Version information set by ldflags during build.
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	if err := Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
