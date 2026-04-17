package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		totalLines, err := countDataRows(dedupeInput, ',')
		if err != nil {
			return err
		}

		inFile, err := os.Open(dedupeInput)
		if err != nil {
			return fmt.Errorf("failed to open input file: %w", err)
		}
		defer inFile.Close()

		reader := csv.NewReader(inFile)
		headers, err := reader.Read()
		if err != nil {
			return fmt.Errorf("failed to read headers: %w", err)
		}

		keyCols := strings.Split(dedupeKeyColumns, ",")
		keyIndexes := make([]int, 0, len(keyCols))
		for _, key := range keyCols {
			lookup := key
			if !caseSensitiveDedupe {
				lookup = strings.ToLower(lookup)
			}
			found := false
			for i, h := range headers {
				colName := h
				if !caseSensitiveDedupe {
					colName = strings.ToLower(h)
				}
				if colName == lookup {
					keyIndexes = append(keyIndexes, i)
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("column %q not found in headers", key)
			}
		}

		tempPath := dedupeOutput + ".tmp"
		outFile, err := os.Create(tempPath)
		if err != nil {
			return fmt.Errorf("failed to create temp output file: %w", err)
		}

		writer := csv.NewWriter(outFile)
		if err := writer.Write(headers); err != nil {
			outFile.Close()
			return fmt.Errorf("failed to write header: %w", err)
		}

		bar := progressbar.Default(totalLines, "Deduplicating")
		duplicates := 0
		unique := 0

		if dedupeKeepLast {
			// Must buffer everything: we only know "last" after seeing all rows.
			rows := [][]string{}
			seen := make(map[string]int) // key -> index into rows
			rowIdx := 0
			for {
				row, err := reader.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					_ = bar.Add(1)
					continue
				}
				if len(row) < len(headers) {
					_ = bar.Add(1)
					continue
				}
				key := buildDedupeKey(row, keyIndexes, caseSensitiveDedupe)
				if _, exists := seen[key]; exists {
					duplicates++
				}
				rows = append(rows, row)
				seen[key] = rowIdx
				rowIdx++
				_ = bar.Add(1)
			}
			// Emit kept rows in original file order.
			kept := make([]int, 0, len(seen))
			for _, idx := range seen {
				kept = append(kept, idx)
			}
			sort.Ints(kept)
			for _, idx := range kept {
				if err := writer.Write(rows[idx]); err != nil {
					outFile.Close()
					return fmt.Errorf("failed to write row: %w", err)
				}
			}
			unique = len(kept)
		} else {
			// keep-first: stream rows directly as we encounter them.
			seen := make(map[string]struct{})
			for {
				row, err := reader.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					_ = bar.Add(1)
					continue
				}
				if len(row) < len(headers) {
					_ = bar.Add(1)
					continue
				}
				key := buildDedupeKey(row, keyIndexes, caseSensitiveDedupe)
				if _, exists := seen[key]; exists {
					duplicates++
				} else {
					seen[key] = struct{}{}
					if err := writer.Write(row); err != nil {
						outFile.Close()
						return fmt.Errorf("failed to write row: %w", err)
					}
					unique++
				}
				_ = bar.Add(1)
			}
		}

		writer.Flush()
		if err := writer.Error(); err != nil {
			outFile.Close()
			return fmt.Errorf("writer error: %w", err)
		}
		if err := outFile.Close(); err != nil {
			return fmt.Errorf("failed to close output: %w", err)
		}

		// Close input before rename for Windows compatibility when overwriting in place.
		if dedupeOutput == dedupeInput {
			inFile.Close()
		}
		if err := os.Rename(tempPath, dedupeOutput); err != nil {
			return fmt.Errorf("failed to rename temp file: %w", err)
		}

		fmt.Printf("\n✅ Duplicates removed. Output written to %s\n", dedupeOutput)
		fmt.Printf("📊 Total rows: %d | Unique: %d | Duplicates removed: %d\n", totalLines, unique, duplicates)
		return nil
	},
}

func buildDedupeKey(row []string, indexes []int, caseSensitive bool) string {
	parts := make([]string, len(indexes))
	for i, idx := range indexes {
		v := row[idx]
		if !caseSensitive {
			v = strings.ToLower(v)
		}
		parts[i] = v
	}
	return strings.Join(parts, "||")
}

func init() {
	rootCmd.AddCommand(dedupeCmd)

	dedupeCmd.Flags().StringVar(&dedupeInput, "input", "", "Input CSV file path (required)")
	dedupeCmd.Flags().StringVar(&dedupeOutput, "output", "", "Output CSV file path (required)")
	dedupeCmd.Flags().StringVar(&dedupeKeyColumns, "key", "", "Comma-separated key column(s) for deduplication (required)")
	dedupeCmd.Flags().BoolVar(&dedupeKeepLast, "keep-last", false, "Keep the last occurrence instead of the first")
	dedupeCmd.Flags().BoolVar(&caseSensitiveDedupe, "case-sensitive", false, "Case sensitive comparison for key columns")

	_ = dedupeCmd.MarkFlagRequired("input")
	_ = dedupeCmd.MarkFlagRequired("output")
	_ = dedupeCmd.MarkFlagRequired("key")
}
