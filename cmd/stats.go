package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

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

		totalRows, err := countDataRows(statsInput, ',')
		if err != nil {
			return err
		}

		file, err := os.Open(statsInput)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		reader := csv.NewReader(file)
		reader.FieldsPerRecord = -1

		headers, err := reader.Read()
		if err != nil {
			return fmt.Errorf("failed to read headers: %w", err)
		}

		columnCount := len(headers)

		type columnStats struct {
			empty     int
			uniques   map[string]int
			capped    bool
			totalVals int
		}

		stats := make([]columnStats, columnCount)
		for i := 0; i < columnCount; i++ {
			stats[i].uniques = make(map[string]int)
		}

		bar := progressbar.Default(totalRows, "Analyzing")
		rowCount := 0

		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				continue
			}
			rowCount++
			for i := range headers {
				cell := ""
				if i < len(row) {
					cell = strings.TrimSpace(row[i])
				}
				if cell == "" {
					stats[i].empty++
					continue
				}
				stats[i].totalVals++
				if statsMaxUnique > 0 && len(stats[i].uniques) >= statsMaxUnique {
					if _, exists := stats[i].uniques[cell]; exists {
						stats[i].uniques[cell]++
					} else {
						stats[i].capped = true
					}
				} else {
					stats[i].uniques[cell]++
				}
			}
			_ = bar.Add(1)
		}

		fmt.Printf("\n📊 Stats for: %s\n", statsInput)
		fmt.Printf("Total Rows (excluding header): %d\n", rowCount)
		fmt.Printf("Columns: %d\n\n", columnCount)

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Column", "Unique Values", "Empty Fields", "Top 3 Values"})
		table.SetAutoWrapText(false)
		table.SetRowLine(true)
		table.SetAlignment(tablewriter.ALIGN_LEFT)

		type kv struct {
			Key   string
			Count int
		}

		for i, name := range headers {
			uniqueCount := fmt.Sprintf("%d", len(stats[i].uniques))
			if stats[i].capped {
				uniqueCount = fmt.Sprintf(">=%d (capped)", len(stats[i].uniques))
			}

			sorted := make([]kv, 0, len(stats[i].uniques))
			for k, v := range stats[i].uniques {
				sorted = append(sorted, kv{k, v})
			}
			sort.Slice(sorted, func(a, b int) bool {
				return sorted[a].Count > sorted[b].Count
			})

			topValues := []string{}
			for j := 0; j < len(sorted) && j < 3; j++ {
				topValues = append(topValues, fmt.Sprintf("%s (%d)", sorted[j].Key, sorted[j].Count))
			}

			table.Append([]string{
				name,
				uniqueCount,
				fmt.Sprintf("%d", stats[i].empty),
				strings.Join(topValues, ", "),
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
