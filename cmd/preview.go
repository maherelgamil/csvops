package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/maherelgamil/csvops/pkg/csvops"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	previewInput    string
	previewRows     int
	previewNoHeader bool
)

var previewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Preview the first N rows of a CSV file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if previewInput == "" {
			return fmt.Errorf("please provide an input file using --input")
		}

		res, err := csvops.Preview(context.Background(), csvops.PreviewOptions{
			Input:     previewInput,
			Rows:      previewRows,
			NoHeader:  previewNoHeader,
			Delimiter: ',',
		})
		if err != nil {
			return err
		}

		for _, e := range res.SkipErrors {
			fmt.Printf("⚠️  Skipping row due to error: %v\n", e)
		}

		if len(res.Rows) == 0 {
			fmt.Println("⚠️  No data rows found")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoWrapText(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetRowLine(true)

		if len(res.Headers) > 0 {
			table.SetHeader(res.Headers)
		}
		for _, row := range res.Rows {
			table.Append(row)
		}
		table.Render()

		fmt.Printf("\n📄 Showing %d row(s) from '%s'\n", len(res.Rows), previewInput)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(previewCmd)

	previewCmd.Flags().StringVar(&previewInput, "input", "", "Input CSV file path")
	previewCmd.Flags().IntVar(&previewRows, "rows", 5, "Number of rows to preview")
	previewCmd.Flags().BoolVar(&previewNoHeader, "no-header", false, "Do not treat first row as header")
}
