package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/maherelgamil/csvops/pkg/csvops"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	dedupeInput         string
	dedupeOutput        string
	dedupeKeyColumns    string
	dedupeKeepLast      bool
	caseSensitiveDedupe bool
)

var dedupeCmd = &cobra.Command{
	Use:   "dedupe",
	Short: "Remove duplicate rows from a CSV file based on key column(s)",
	RunE: func(cmd *cobra.Command, args []string) error {
		var bar *progressbar.ProgressBar
		res, err := csvops.Dedupe(context.Background(), csvops.DedupeOptions{
			Input:         dedupeInput,
			Output:        dedupeOutput,
			KeyColumns:    strings.Split(dedupeKeyColumns, ","),
			KeepLast:      dedupeKeepLast,
			CaseSensitive: caseSensitiveDedupe,
			Delimiter:     ',',
			Progress: func(done, total int64) {
				if bar == nil {
					bar = progressbar.Default(total, "Deduplicating")
				}
				_ = bar.Set64(done)
			},
		})
		if err != nil {
			return err
		}

		fmt.Printf("\n✅ Duplicates removed. Output written to %s\n", dedupeOutput)
		fmt.Printf("📊 Total rows: %d | Unique: %d | Duplicates removed: %d\n", res.TotalRows, res.UniqueRows, res.Duplicates)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dedupeCmd)

	dedupeCmd.Flags().StringVar(&dedupeInput, "input", "", "Input CSV file path (required)")
	dedupeCmd.Flags().StringVar(&dedupeOutput, "output", "", "Output CSV file path (required)")
	dedupeCmd.Flags().StringVar(&dedupeKeyColumns, "key", "", "Comma-separated key column(s) for deduplication (required)")
	dedupeCmd.Flags().BoolVar(&dedupeKeepLast, "keep-last", false, "Keep the last occurrence instead of the first")
	dedupeCmd.Flags().BoolVar(&caseSensitiveDedupe, "case-sensitive", false, "Case sensitive comparison for key columns")

	_ = dedupeCmd.MarkFlagRequired("input")
	_ = dedupeCmd.MarkFlagRequired("output")
	_ = dedupeCmd.MarkFlagRequired("key")
}
