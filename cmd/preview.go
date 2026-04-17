package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

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

		file, err := os.Open(previewInput)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		reader := csv.NewReader(file)
		reader.FieldsPerRecord = -1

		var headers []string
		if !previewNoHeader {
			headers, err = reader.Read()
			if err != nil {
				return fmt.Errorf("failed to read header: %w", err)
			}
		}

		rows := [][]string{}
		for len(rows) < previewRows {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("⚠️  Skipping row due to error: %v\n", err)
				continue
			}
			rows = append(rows, record)
		}

		if len(rows) == 0 {
			fmt.Println("⚠️  No data rows found")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoWrapText(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetRowLine(true)

		if !previewNoHeader {
			table.SetHeader(headers)
		}
		for _, row := range rows {
			table.Append(row)
		}

		table.Render()
		fmt.Printf("\n📄 Showing %d row(s) from '%s'\n", len(rows), previewInput)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(previewCmd)

	previewCmd.Flags().StringVar(&previewInput, "input", "", "Input CSV file path")
	previewCmd.Flags().IntVar(&previewRows, "rows", 5, "Number of rows to preview")
	previewCmd.Flags().BoolVar(&previewNoHeader, "no-header", false, "Do not treat first row as header")
}
