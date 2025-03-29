package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var statsInput string

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Display basic statistics about a CSV file",
	Run: func(cmd *cobra.Command, args []string) {
		if statsInput == "" {
			fmt.Println("❌ Please provide an input file using --input")
			return
		}

		file, err := os.Open(statsInput)
		if err != nil {
			fmt.Printf("❌ Failed to open file: %v\n", err)
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil || len(records) < 1 {
			fmt.Printf("❌ Failed to read CSV or empty file\n")
			return
		}

		headers := records[0]
		columnCount := len(headers)
		rowCount := len(records) - 1

		fmt.Printf("\n📊 Stats for: %s\n", statsInput)
		fmt.Printf("Total Rows (excluding header): %d\n", rowCount)
		fmt.Printf("Columns: %d\n\n", columnCount)

		type columnStats struct {
			empty   int
			uniques map[string]int
		}

		stats := make([]columnStats, columnCount)
		for i := 0; i < columnCount; i++ {
			stats[i].uniques = make(map[string]int)
		}

		for _, row := range records[1:] {
			for i, cell := range row {
				if strings.TrimSpace(cell) == "" {
					stats[i].empty++
				} else {
					stats[i].uniques[cell]++
				}
			}
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Column", "Unique Values", "Empty Fields", "Top 3 Values"})
		table.SetAutoWrapText(false)
		table.SetRowLine(true)

		for i, name := range headers {
			uniqueCount := len(stats[i].uniques)
			emptyCount := stats[i].empty

			type kv struct {
				Key   string
				Count int
			}
			var sorted []kv
			for k, v := range stats[i].uniques {
				sorted = append(sorted, kv{k, v})
			}
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].Count > sorted[j].Count
			})

			topValues := []string{}
			for j := 0; j < len(sorted) && j < 3; j++ {
				topValues = append(topValues, fmt.Sprintf("%s (%d)", sorted[j].Key, sorted[j].Count))
			}

			table.Append([]string{
				name,
				fmt.Sprintf("%d", uniqueCount),
				fmt.Sprintf("%d", emptyCount),
				strings.Join(topValues, ", "),
			})
		}

		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
	statsCmd.Flags().StringVar(&statsInput, "input", "", "Input CSV file path")
}
