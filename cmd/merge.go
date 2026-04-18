package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/maherelgamil/csvops/pkg/csvops"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	mergeInputDir   string
	mergeOutput     string
	mergeWithHeader bool
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge multiple CSV files into one",
	RunE: func(cmd *cobra.Command, args []string) error {
		outFile, err := os.Create(mergeOutput)
		if err != nil {
			return fmt.Errorf("create output: %w", err)
		}
		defer outFile.Close()

		var bar *progressbar.ProgressBar
		res, err := csvops.Merge(context.Background(), csvops.MergeOptions{
			InputDir:   mergeInputDir,
			Output:     outFile,
			WithHeader: mergeWithHeader,
			Delimiter:  ',',
			SkipErrors: true,
			OnWarn: func(path string, e error) {
				fmt.Printf("⚠️  Skipping %s: %v\n", filepath.Base(path), e)
			},
			Progress: func(done, total int64) {
				if bar == nil {
					bar = progressbar.Default(total, "Merging")
				}
				_ = bar.Set64(done)
			},
		})
		if err != nil {
			return err
		}

		if res.FilesProcessed == 0 {
			fmt.Println("⚠️  No CSV files found to merge.")
			return nil
		}
		fmt.Printf("\n✅ Merged %d CSV files into %s (%d rows)\n", res.FilesProcessed, mergeOutput, res.RowsWritten)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mergeCmd)

	mergeCmd.Flags().StringVar(&mergeInputDir, "input-dir", "", "Directory containing CSV files to merge")
	mergeCmd.Flags().StringVar(&mergeOutput, "output", "merged.csv", "Path for the output CSV file")
	mergeCmd.Flags().BoolVar(&mergeWithHeader, "with-header", true, "Include headers from the first file")
	_ = mergeCmd.MarkFlagRequired("input-dir")
}
