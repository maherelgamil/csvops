package cmd

import (
	"encoding/csv"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
)

var (
	previewInput    string
	previewRows     int
	previewNoHeader bool
)

var previewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Preview the first N rows of a CSV file",
	Run: func(cmd *cobra.Command, args []string) {
		if previewInput == "" {
			fmt.Println("❌ Please provide an input file using --input")
			return
		}

		file, err := os.Open(previewInput)
		if err != nil {
			fmt.Printf("❌ Failed to open file: %v\n", err)
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil || len(records) == 0 {
			fmt.Printf("❌ Failed to read CSV or file is empty\n")
			return
		}

		start := 0
		if !previewNoHeader {
			start = 1
		}

		limit := previewRows
		if limit > len(records)-start {
			limit = len(records) - start
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoWrapText(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)

		if !previewNoHeader {
			table.SetHeader(records[0])
		}

		for _, row := range records[start : start+limit] {
			table.Append(row)
		}

		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(previewCmd)

	previewCmd.Flags().StringVar(&previewInput, "input", "", "Input CSV file path")
	previewCmd.Flags().IntVar(&previewRows, "rows", 5, "Number of rows to preview")
	previewCmd.Flags().BoolVar(&previewNoHeader, "no-header", false, "Do not treat first row as header")
}
