package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/maherelgamil/csvops/pkg/csvops"
	"github.com/olekukonko/tablewriter"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	statsInput     string
	statsMaxUnique int
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Display basic statistics about a CSV file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if statsInput == "" {
			return fmt.Errorf("please provide an input file using --input")
		}

		var bar *progressbar.ProgressBar
		res, err := csvops.Stats(context.Background(), csvops.StatsOptions{
			Input:     statsInput,
			MaxUnique: statsMaxUnique,
			Delimiter: ',',
			Progress: func(done, total int64) {
				if bar == nil {
					bar = progressbar.Default(total, "Analyzing")
				}
				_ = bar.Set64(done)
			},
		})
		if err != nil {
			return err
		}

		fmt.Printf("\n📊 Stats for: %s\n", statsInput)
		fmt.Printf("Total Rows (excluding header): %d\n", res.TotalRows)
		fmt.Printf("Columns: %d\n\n", len(res.Columns))

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Column", "Unique Values", "Empty Fields", "Top 3 Values"})
		table.SetAutoWrapText(false)
		table.SetRowLine(true)
		table.SetAlignment(tablewriter.ALIGN_LEFT)

		for _, col := range res.Columns {
			unique := fmt.Sprintf("%d", col.Unique)
			if col.UniqueCapped {
				unique = fmt.Sprintf(">=%d (capped)", col.Unique)
			}
			top := make([]string, 0, len(col.Top))
			for _, v := range col.Top {
				top = append(top, fmt.Sprintf("%s (%d)", v.Value, v.Count))
			}
			table.Append([]string{
				col.Name,
				unique,
				fmt.Sprintf("%d", col.Empty),
				strings.Join(top, ", "),
			})
		}
		table.Render()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
	statsCmd.Flags().StringVar(&statsInput, "input", "", "Input CSV file path")
	statsCmd.Flags().IntVar(&statsMaxUnique, "max-unique", 100000, "Max unique values tracked per column (0 = unlimited)")
}
