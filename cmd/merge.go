package cmd

import (
	"encoding/csv"
	"fmt"
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
	Run: func(cmd *cobra.Command, args []string) {
		files, err := os.ReadDir(mergeInputDir)
		if err != nil {
			fmt.Printf("❌ Failed to read input directory: %v\n", err)
			return
		}

		csvFiles := []string{}
		for _, f := range files {
			if strings.HasSuffix(strings.ToLower(f.Name()), ".csv") {
				csvFiles = append(csvFiles, filepath.Join(mergeInputDir, f.Name()))
			}
		}

		if len(csvFiles) == 0 {
			fmt.Println("⚠️  No CSV files found to merge.")
			return
		}

		sort.Strings(csvFiles) // ensure consistent order

		outFile, err := os.Create(mergeOutput)
		if err != nil {
			fmt.Printf("❌ Failed to create output file: %v\n", err)
			return
		}
		defer outFile.Close()

		writer := csv.NewWriter(outFile)
		writtenHeader := false
		fileCount := 0
		rowCount := 0

		bar := progressbar.Default(int64(len(csvFiles)), "Merging")

		for _, path := range csvFiles {
			f, err := os.Open(path)
			if err != nil {
				fmt.Printf("⚠️  Skipping %s: %v\n", filepath.Base(path), err)
				bar.Add(1)
				continue
			}

			reader := csv.NewReader(f)
			records, err := reader.ReadAll()
			f.Close()
			if err != nil {
				fmt.Printf("⚠️  Skipping %s due to error: %v\n", filepath.Base(path), err)
				bar.Add(1)
				continue
			}

			if len(records) == 0 {
				bar.Add(1)
				continue
			}

			start := 0
			if mergeWithHeader && !writtenHeader {
				_ = writer.Write(records[0])
				writtenHeader = true
				start = 1
			} else if mergeWithHeader {
				start = 1
			}

			for _, row := range records[start:] {
				_ = writer.Write(row)
				rowCount++
			}

			fileCount++
			bar.Add(1)
		}

		writer.Flush()
		fmt.Printf("\n✅ Merged %d CSV files into %s (%d rows)\n", fileCount, mergeOutput, rowCount)
	},
}

func init() {
	rootCmd.AddCommand(mergeCmd)

	mergeCmd.Flags().StringVar(&mergeInputDir, "input-dir", "", "Directory containing CSV files to merge")
	mergeCmd.Flags().StringVar(&mergeOutput, "output", "merged.csv", "Path for the output CSV file")
	mergeCmd.Flags().BoolVar(&mergeWithHeader, "with-header", true, "Include headers from the first file")
	_ = mergeCmd.MarkFlagRequired("input-dir")
}
