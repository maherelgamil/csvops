package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/maherelgamil/csvops/pkg/csvops"
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
		opts := csvops.FilterOptions{
			Input:      filterInput,
			Column:     filterColumn,
			All:        filterMatchAll,
			WithHeader: filterWithHeader,
			Delimiter:  ',',
		}
		if cmd.Flags().Changed("eq") {
			opts.Eq = &eqValue
		}
		if cmd.Flags().Changed("contains") {
			opts.Contains = &containsValue
		}
		if cmd.Flags().Changed("gt") {
			opts.Gt = &gtValue
		}
		if cmd.Flags().Changed("lt") {
			opts.Lt = &ltValue
		}

		out := os.Stdout
		if filterOutput != "" {
			f, err := os.Create(filterOutput)
			if err != nil {
				return fmt.Errorf("create output: %w", err)
			}
			defer f.Close()
			out = f
		}
		opts.Output = out

		var bar *progressbar.ProgressBar
		opts.Progress = func(done, total int64) {
			if bar == nil {
				bar = progressbar.Default(total, "Filtering")
			}
			_ = bar.Set64(done)
		}

		res, err := csvops.Filter(context.Background(), opts)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "\n✅ Filter complete. %d rows matched out of %d total.\n", res.Matched, res.TotalRows)
		return nil
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
	filterCmd.Flags().BoolVar(&filterWithHeader, "with-header", true, "Include header in output")
	filterCmd.Flags().BoolVar(&filterMatchAll, "all", false, "Require ALL conditions to match (AND) instead of ANY (OR)")

	_ = filterCmd.MarkFlagRequired("input")
	_ = filterCmd.MarkFlagRequired("column")
}
