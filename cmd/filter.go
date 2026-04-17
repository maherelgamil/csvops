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
	filterWithHeader bool
	filterMatchAll   bool
)

var filterCmd = &cobra.Command{
	Use:   "filter",
	Short: "Filter rows based on a column condition",
	Long: `Filter rows based on a column condition.

By default rows match if ANY provided condition matches (OR).
Use --all to require ALL provided conditions to match (AND).
Flags --eq, --contains, --gt, --lt are only applied when explicitly set,
so matching empty strings via --eq="" works correctly.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		eqSet := cmd.Flags().Changed("eq")
		containsSet := cmd.Flags().Changed("contains")
		gtSet := cmd.Flags().Changed("gt")
		ltSet := cmd.Flags().Changed("lt")

		if !eqSet && !containsSet && !gtSet && !ltSet {
			return fmt.Errorf("at least one of --eq, --contains, --gt, --lt must be provided")
		}

		totalLines, err := countDataRows(filterInput, ',')
		if err != nil {
			return err
		}

		file, err := os.Open(filterInput)
		if err != nil {
			return fmt.Errorf("failed to open input file: %w", err)
		}
		defer file.Close()

		reader := csv.NewReader(file)
		headers, err := reader.Read()
		if err != nil {
			return fmt.Errorf("failed to read headers: %w", err)
		}

		colIndex := -1
		for i, col := range headers {
			if col == filterColumn {
				colIndex = i
				break
			}
		}
		if colIndex == -1 {
			return fmt.Errorf("column %q not found", filterColumn)
		}

		out := os.Stdout
		if filterOutput != "" {
			out, err = os.Create(filterOutput)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer out.Close()
		}

		writer := csv.NewWriter(out)
		if filterWithHeader {
			if err := writer.Write(headers); err != nil {
				return fmt.Errorf("failed to write header: %w", err)
			}
		}

		bar := progressbar.Default(totalLines, "Filtering")
		matchCount := 0

		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil || len(row) <= colIndex {
				_ = bar.Add(1)
				continue
			}

			val := row[colIndex]
			match := rowMatches(val, eqSet, containsSet, gtSet, ltSet, filterMatchAll)

			if match {
				if err := writer.Write(row); err != nil {
					return fmt.Errorf("failed to write row: %w", err)
				}
				matchCount++
			}
			_ = bar.Add(1)
		}

		writer.Flush()
		if err := writer.Error(); err != nil {
			return fmt.Errorf("writer error: %w", err)
		}

		fmt.Fprintf(os.Stderr, "\n✅ Filter complete. %d rows matched out of %d total.\n", matchCount, totalLines)
		return nil
	},
}

func rowMatches(val string, eqSet, containsSet, gtSet, ltSet, all bool) bool {
	checks := []bool{}
	if eqSet {
		checks = append(checks, val == eqValue)
	}
	if containsSet {
		checks = append(checks, strings.Contains(strings.ToLower(val), strings.ToLower(containsValue)))
	}
	if gtSet {
		num, err := strconv.ParseFloat(val, 64)
		checks = append(checks, err == nil && num > gtValue)
	}
	if ltSet {
		num, err := strconv.ParseFloat(val, 64)
		checks = append(checks, err == nil && num < ltValue)
	}
	if len(checks) == 0 {
		return false
	}
	if all {
		for _, c := range checks {
			if !c {
				return false
			}
		}
		return true
	}
	for _, c := range checks {
		if c {
			return true
		}
	}
	return false
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
	filterCmd.Flags().BoolVar(&filterWithHeader, "with-header", true, "Include header in output")
	filterCmd.Flags().BoolVar(&filterMatchAll, "all", false, "Require ALL conditions to match (AND) instead of ANY (OR)")

	_ = filterCmd.MarkFlagRequired("input")
	_ = filterCmd.MarkFlagRequired("column")
}
