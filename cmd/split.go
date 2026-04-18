package cmd

import (
	"context"
	"fmt"

	"github.com/maherelgamil/csvops/pkg/csvops"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	inputPath   string
	outputDir   string
	rowsPerFile int
	withHeader  bool
	delimiter   string
)

var splitCmd = &cobra.Command{
	Use:   "split",
	Short: "Split a large CSV file into smaller chunks",
	RunE: func(cmd *cobra.Command, args []string) error {
		if inputPath == "" {
			return fmt.Errorf("please provide an input CSV file using --input")
		}
		delim, err := parseDelimiter(delimiter)
		if err != nil {
			return err
		}

		var bar *progressbar.ProgressBar
		res, err := csvops.Split(context.Background(), csvops.SplitOptions{
			Input:       inputPath,
			OutputDir:   outputDir,
			RowsPerFile: rowsPerFile,
			WithHeader:  withHeader,
			Delimiter:   delim,
			Progress: func(done, total int64) {
				if bar == nil {
					bar = progressbar.Default(total, "Splitting")
				}
				_ = bar.Set64(done)
			},
		})
		if err != nil {
			return err
		}

		fmt.Printf("\n✅ Finished splitting %d rows into %d file(s).\n", res.RowsProcessed, res.FilesCreated)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(splitCmd)

	splitCmd.Flags().StringVar(&inputPath, "input", "", "Input CSV file path (required)")
	splitCmd.Flags().StringVar(&outputDir, "output-dir", "./output", "Directory to save split files")
	splitCmd.Flags().IntVar(&rowsPerFile, "rows", 1000, "Max rows per output file")
	splitCmd.Flags().BoolVar(&withHeader, "with-header", true, "Include header in each output file")
	splitCmd.Flags().StringVar(&delimiter, "delimiter", ",", "CSV delimiter character")
}
