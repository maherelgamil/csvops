/*
Copyright © 2025 Maher El Gamil
*/

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

// SetVersion allows dynamic version injection during build
var SetVersion = func(v string) {
	version = v
	rootCmd.Version = v
}

var rootCmd = &cobra.Command{
	Use:   "csvops",
	Short: "A fast and modular CLI toolkit for working with CSV files",
	Long: `📊 csvops is a blazing fast, modular CLI tool for working with CSV files.

Built in Go, designed for humans — ideal for data engineers, analysts, and developers.
Supports operations like:
  • split      - break large files into chunks
  • merge      - combine CSV files
  • dedupe     - remove duplicates by column
  • filter     - filter rows by condition
  • stats      - get descriptive statistics
  • preview    - preview first N rows
  • to-sqlite  - convert to SQLite
... and more coming soon!`,

	Version:       version,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Optionally define global flags here
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Optional config file (default is $HOME/.csvops.yaml)")
}
