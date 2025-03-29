package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	filterInput      string
	filterOutput     string
	filterColumn     string
	eqValue          string
	containsValue    string
	gtValue          float64
	ltValue          float64
	enableGt         bool
	enableLt         bool
	filterWithHeader bool
)

var filterCmd = &cobra.Command{
	Use:   "filter",
	Short: "Filter rows based on a column condition",
	Run: func(cmd *cobra.Command, args []string) {
		if filterInput == "" || filterColumn == "" {
			fmt.Println("❌ Please provide --input and --column")
			return
		}

		file, err := os.Open(filterInput)
		if err != nil {
			fmt.Printf("❌ Failed to open input file: %v\n", err)
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			fmt.Printf("❌ Failed to read CSV: %v\n", err)
			return
		}

		if len(records) < 1 {
			fmt.Println("⚠️  Empty CSV file")
			return
		}

		headers := records[0]
		colIndex := -1
		for i, col := range headers {
			if col == filterColumn {
				colIndex = i
				break
			}
		}

		if colIndex == -1 {
			fmt.Printf("❌ Column '%s' not found\n", filterColumn)
			return
		}

		bar := progressbar.Default(int64(len(records)-1), "Filtering")

		filtered := [][]string{}
		if filterWithHeader {
			filtered = append(filtered, headers)
		}

		for _, row := range records[1:] {
			val := row[colIndex]
			match := false

			if eqValue != "" && val == eqValue {
				match = true
			}
			if containsValue != "" && strings.Contains(strings.ToLower(val), strings.ToLower(containsValue)) {
				match = true
			}
			if enableGt {
				num, err := strconv.ParseFloat(val, 64)
				if err == nil && num > gtValue {
					match = true
				}
			}
			if enableLt {
				num, err := strconv.ParseFloat(val, 64)
				if err == nil && num < ltValue {
					match = true
				}
			}

			if match {
				filtered = append(filtered, row)
			}
			bar.Add(1)
		}

		out := os.Stdout
		if filterOutput != "" {
			out, err = os.Create(filterOutput)
			if err != nil {
				fmt.Printf("❌ Failed to create output file: %v\n", err)
				return
			}
			defer out.Close()
		}

		writer := csv.NewWriter(out)
		err = writer.WriteAll(filtered)
		if err != nil {
			fmt.Printf("❌ Failed to write output: %v\n", err)
			return
		}

		fmt.Printf("✅ Filter complete. %d rows matched.\n", len(filtered)-1)
	},
}

func init() {
	rootCmd.AddCommand(filterCmd)

	filterCmd.Flags().StringVar(&filterInput, "input", "", "Input CSV file path (required)")
	filterCmd.Flags().StringVar(&filterOutput, "output", "", "Output CSV file path (optional, default is stdout)")
	filterCmd.Flags().StringVar(&filterColumn, "column", "", "Column to filter on (required)")
	filterCmd.Flags().StringVar(&eqValue, "eq", "", "Equals value")
	filterCmd.Flags().StringVar(&containsValue, "contains", "", "Substring match (case-insensitive)")
	filterCmd.Flags().Float64Var(&gtValue, "gt", 0, "Greater than (number)")
	filterCmd.Flags().Float64Var(&ltValue, "lt", 0, "Less than (number)")
	filterCmd.Flags().BoolVar(&enableGt, "enable-gt", false, "Enable --gt filter")
	filterCmd.Flags().BoolVar(&enableLt, "enable-lt", false, "Enable --lt filter")
	filterCmd.Flags().BoolVar(&filterWithHeader, "with-header", true, "Include header in output")
}
