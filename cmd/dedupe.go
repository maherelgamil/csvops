package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
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

		// Count lines for progress bar
		totalLines := int64(0)
		counterFile, _ := os.Open(dedupeInput)
		defer counterFile.Close()
		counterReader := csv.NewReader(counterFile)
		for {
			_, err := counterReader.Read()
			if err == io.EOF {
				break
			}
			if err == nil {
				totalLines++
			}
		}
		if totalLines > 0 {
			totalLines-- // exclude header
		}

		inFile, err := os.Open(dedupeInput)
		if err != nil {
			fmt.Printf("❌ Failed to open input file: %v\n", err)
			return
		}
		defer inFile.Close()

		reader := csv.NewReader(inFile)
		headers, err := reader.Read()
		if err != nil {
			fmt.Printf("❌ Failed to read headers: %v\n", err)
			return
		}

		keyCols := strings.Split(dedupeKeyColumns, ",")
		keyIndexes := make([]int, 0, len(keyCols))
		for _, key := range keyCols {
			found := false
			for i, h := range headers {
				colName := h
				if !caseSensitiveDedupe {
					colName = strings.ToLower(h)
					key = strings.ToLower(key)
				}
				if colName == key {
					keyIndexes = append(keyIndexes, i)
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("❌ Column '%s' not found in headers\n", key)
				return
			}
		}

		seen := make(map[string]int)
		tempPath := dedupeOutput + ".tmp"
		outFile, err := os.Create(tempPath)
		if err != nil {
			fmt.Printf("❌ Failed to create temp output file: %v\n", err)
			return
		}
		defer outFile.Close()

		writer := csv.NewWriter(outFile)
		_ = writer.Write(headers)

		rows := [][]string{}
		bar := progressbar.Default(totalLines, "Deduplicating")
		rowIndex := 1
		duplicates := 0
		for {
			row, err := reader.Read()
			if err != nil {
				break
			}
			if len(row) < len(headers) {
				bar.Add(1)
				continue
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
				if _, exists := seen[key]; exists {
					duplicates++
				}
				seen[key] = len(rows)
				rows = append(rows, row)
			} else if _, exists := seen[key]; !exists {
				seen[key] = len(rows)
				rows = append(rows, row)
			} else {
				duplicates++
			}

			bar.Add(1)
			rowIndex++
		}

		for _, idx := range seen {
			_ = writer.Write(rows[idx])
		}

		writer.Flush()

		// Handle in-place overwrite
		if dedupeOutput == dedupeInput {
			os.Remove(dedupeInput)
		}
		err = os.Rename(tempPath, dedupeOutput)
		if err != nil {
			fmt.Printf("❌ Failed to rename temp file: %v\n", err)
			return
		}

		fmt.Printf("\n✅ Duplicates removed. Output written to %s\n", dedupeOutput)
		fmt.Printf("📊 Total rows: %d | Unique: %d | Duplicates removed: %d\n", totalLines, len(seen), duplicates)
	},
}

func init() {
	rootCmd.AddCommand(dedupeCmd)

	dedupeCmd.Flags().StringVar(&dedupeInput, "input", "", "Input CSV file path (required)")
	dedupeCmd.Flags().StringVar(&dedupeOutput, "output", "", "Output CSV file path (required)")
	dedupeCmd.Flags().StringVar(&dedupeKeyColumns, "key", "", "Comma-separated key column(s) for deduplication (required)")
	dedupeCmd.Flags().BoolVar(&dedupeKeepLast, "keep-last", false, "Keep the last occurrence instead of the first")
	dedupeCmd.Flags().BoolVar(&caseSensitiveDedupe, "case-sensitive", false, "Case sensitive comparison for key columns")

	dedupeCmd.MarkFlagRequired("input")
	dedupeCmd.MarkFlagRequired("output")
	dedupeCmd.MarkFlagRequired("key")
}
