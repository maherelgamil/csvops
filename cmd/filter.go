package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
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

		// Count total rows for progress bar
		totalLines := int64(0)
		lineCounter, err := os.Open(filterInput)
		if err != nil {
			fmt.Printf("❌ Failed to open input file: %v\n", err)
			return
		}
		lcReader := csv.NewReader(lineCounter)
		for {
			_, err := lcReader.Read()
			if err == io.EOF {
				break
			}
			if err == nil {
				totalLines++
			}
		}
		lineCounter.Close()
		if totalLines > 0 {
			totalLines--
		}

		file, err := os.Open(filterInput)
		if err != nil {
			fmt.Printf("❌ Failed to open input file: %v\n", err)
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		headers, err := reader.Read()
		if err != nil {
			fmt.Printf("❌ Failed to read headers: %v\n", err)
			return
		}

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

		bar := progressbar.Default(totalLines, "Filtering")
		filtered := [][]string{}
		if filterWithHeader {
			filtered = append(filtered, headers)
		}

		matchCount := 0
		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil || len(row) <= colIndex {
				bar.Add(1)
				continue
			}

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
				matchCount++
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

		fmt.Printf("\n✅ Filter complete. %d rows matched out of %d total.\n", matchCount, totalLines)
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

	filterCmd.MarkFlagRequired("input")
	filterCmd.MarkFlagRequired("column")
}
