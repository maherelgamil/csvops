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
		headers, err := reader.Read()
		if err != nil {
			fmt.Printf("❌ Failed to read headers: %v\n", err)
			return
		}

		columnCount := len(headers)
		rowCount := 0

		type columnStats struct {
			empty   int
			uniques map[string]int
		}

		stats := make([]columnStats, columnCount)
		for i := 0; i < columnCount; i++ {
			stats[i].uniques = make(map[string]int)
		}

		// Pre-scan for progress bar
		totalRows := int64(0)
		counter, _ := os.Open(statsInput)
		defer counter.Close()
		countReader := csv.NewReader(counter)
		countReader.Read() // skip header
		for {
			_, err := countReader.Read()
			if err == io.EOF {
				break
			}
			if err == nil {
				totalRows++
			}
		}

		bar := progressbar.Default(totalRows, "Analyzing")

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
				if i < len(row) {
					cell := strings.TrimSpace(row[i])
					if cell == "" {
						stats[i].empty++
					} else {
						stats[i].uniques[cell]++
					}
				} else {
					stats[i].empty++
				}
			}

			bar.Add(1)
		}

		fmt.Printf("\n📊 Stats for: %s\n", statsInput)
		fmt.Printf("Total Rows (excluding header): %d\n", rowCount)
		fmt.Printf("Columns: %d\n\n", columnCount)

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Column", "Unique Values", "Empty Fields", "Top 3 Values"})
		table.SetAutoWrapText(false)
		table.SetRowLine(true)
		table.SetAlignment(tablewriter.ALIGN_LEFT)

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
