/*
Copyright © 2025 Maher El Gamil
*/

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var version = "v0.0.1"

var SetVersion = func(v string) {
	version = v
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Version: version,
	Use:     "csvops",
	Short:   "A fast and modular CLI toolkit for working with CSV files",
	Long: `csvops is a powerful command-line tool written in Go for working with CSV files.

It supports splitting, merging, filtering, validating, deduplication, statistics, and more.
Perfect for data engineers, analysts, and developers who want terminal-first CSV workflows.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// You can define persistent flags here
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.csvops.yaml)")
}
