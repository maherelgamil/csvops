package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	dedupeInput         string
	dedupeOutput        string
	dedupeKeyColumns    string
	dedupeKeepLast      bool
	caseSensitiveDedupe bool
)

var dedupeCmd = &cobra.Command{
	Use:   "dedupe",
	Short: "Remove duplicate rows from a CSV file based on key column(s)",
	Run: func(cmd *cobra.Command, args []string) {
		if dedupeInput == "" || dedupeOutput == "" || dedupeKeyColumns == "" {
			fmt.Println("❌ Please provide --input, --output, and --key flags")
			return
		}

		file, err := os.Open(dedupeInput)
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

		if len(records) == 0 {
			fmt.Println("⚠️  Empty CSV file")
			return
		}

		headers := records[0]
		keyCols := strings.Split(dedupeKeyColumns, ",")
		keyIndexes := make([]int, 0, len(keyCols))

		for _, key := range keyCols {
			found := false
			for i, h := range headers {
				match := h
				if !caseSensitiveDedupe {
					match = strings.ToLower(h)
					key = strings.ToLower(key)
				}
				if match == key {
					keyIndexes = append(keyIndexes, i)
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("❌ Column '%s' not found in header\n", key)
				return
			}
		}

		seen := make(map[string]int)
		unique := [][]string{headers}

		bar := progressbar.Default(int64(len(records)-1), "Deduplicating")

		for i, row := range records[1:] {
			if len(row) < len(headers) {
				bar.Add(1)
				continue // Skip incomplete rows
			}

			var keyParts []string
			for _, idx := range keyIndexes {
				val := row[idx]
				if !caseSensitiveDedupe {
					val = strings.ToLower(val)
				}
				keyParts = append(keyParts, val)
			}
			key := strings.Join(keyParts, "||")

			if dedupeKeepLast {
				seen[key] = i + 1
			} else if _, exists := seen[key]; !exists {
				seen[key] = i + 1
			}

			bar.Add(1)
		}

		for _, idx := range seen {
			unique = append(unique, records[idx])
		}

		outFile, err := os.Create(dedupeOutput)
		if err != nil {
			fmt.Printf("❌ Failed to create output file: %v\n", err)
			return
		}
		defer outFile.Close()

		writer := csv.NewWriter(outFile)
		err = writer.WriteAll(unique)
		if err != nil {
			fmt.Printf("❌ Failed to write CSV: %v\n", err)
			return
		}

		fmt.Printf("✅ Duplicates removed. Output written to %s\n", dedupeOutput)
	},
}

func init() {
	rootCmd.AddCommand(dedupeCmd)

	dedupeCmd.Flags().StringVar(&dedupeInput, "input", "", "Input CSV file path (required)")
	dedupeCmd.Flags().StringVar(&dedupeOutput, "output", "", "Output CSV file path (required)")
	dedupeCmd.Flags().StringVar(&dedupeKeyColumns, "key", "", "Comma-separated key column(s) for deduplication (required)")
	dedupeCmd.Flags().BoolVar(&dedupeKeepLast, "keep-last", false, "Keep the last occurrence instead of the first")
	dedupeCmd.Flags().BoolVar(&caseSensitiveDedupe, "case-sensitive", false, "Case sensitive comparison for key columns")
}
