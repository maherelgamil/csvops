package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
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

		totalFiles := 0
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".csv") {
				totalFiles++
			}
		}

		bar := progressbar.Default(int64(totalFiles), "Merging")

		for _, file := range files {
			if !strings.HasSuffix(file.Name(), ".csv") {
				continue
			}

			path := filepath.Join(mergeInputDir, file.Name())
			f, err := os.Open(path)
			if err != nil {
				fmt.Printf("⚠️  Skipping %s: %v\n", file.Name(), err)
				continue
			}

			reader := csv.NewReader(f)
			head := false
			if mergeWithHeader {
				head = true
			}

			records, err := reader.ReadAll()
			f.Close()
			if err != nil {
				fmt.Printf("⚠️  Skipping %s due to error: %v\n", file.Name(), err)
				continue
			}

			if len(records) == 0 {
				continue
			}

			start := 0
			if head && !writtenHeader {
				_ = writer.Write(records[0])
				writtenHeader = true
				start = 1
			} else if head {
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
	err := mergeCmd.MarkFlagRequired("input-dir")
	if err != nil {
		return
	}
}
