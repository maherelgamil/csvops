package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	mergeInputDir   string
	mergeOutput     string
	mergeWithHeader bool
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge multiple CSV files into one",
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, err := os.ReadDir(mergeInputDir)
		if err != nil {
			return fmt.Errorf("failed to read input directory: %w", err)
		}

		csvFiles := []string{}
		for _, f := range entries {
			if f.IsDir() {
				continue
			}
			if strings.HasSuffix(strings.ToLower(f.Name()), ".csv") {
				csvFiles = append(csvFiles, filepath.Join(mergeInputDir, f.Name()))
			}
		}

		if len(csvFiles) == 0 {
			fmt.Println("⚠️  No CSV files found to merge.")
			return nil
		}

		sort.Strings(csvFiles)

		outFile, err := os.Create(mergeOutput)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outFile.Close()

		writer := csv.NewWriter(outFile)
		defer writer.Flush()

		writtenHeader := false
		fileCount := 0
		rowCount := 0

		bar := progressbar.Default(int64(len(csvFiles)), "Merging")

		for _, path := range csvFiles {
			n, err := streamMergeFile(path, writer, mergeWithHeader, &writtenHeader)
			if err != nil {
				fmt.Printf("⚠️  Skipping %s: %v\n", filepath.Base(path), err)
				_ = bar.Add(1)
				continue
			}
			rowCount += n
			fileCount++
			_ = bar.Add(1)
		}

		writer.Flush()
		if err := writer.Error(); err != nil {
			return fmt.Errorf("writer error: %w", err)
		}

		fmt.Printf("\n✅ Merged %d CSV files into %s (%d rows)\n", fileCount, mergeOutput, rowCount)
		return nil
	},
}

// streamMergeFile streams a single CSV file into the shared writer.
// writtenHeader is shared across calls so the header is emitted only once.
func streamMergeFile(path string, writer *csv.Writer, withHeader bool, writtenHeader *bool) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1

	first := true
	count := 0
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, err
		}
		if first {
			first = false
			if withHeader {
				if !*writtenHeader {
					if err := writer.Write(row); err != nil {
						return count, err
					}
					*writtenHeader = true
				}
				continue
			}
		}
		if err := writer.Write(row); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func init() {
	rootCmd.AddCommand(mergeCmd)

	mergeCmd.Flags().StringVar(&mergeInputDir, "input-dir", "", "Directory containing CSV files to merge")
	mergeCmd.Flags().StringVar(&mergeOutput, "output", "merged.csv", "Path for the output CSV file")
	mergeCmd.Flags().BoolVar(&mergeWithHeader, "with-header", true, "Include headers from the first file")
	_ = mergeCmd.MarkFlagRequired("input-dir")
}
