package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	inputPath   string
	outputDir   string
	rowsPerFile int
	withHeader  bool
	delimiter   string
)

var splitCmd = &cobra.Command{
	Use:   "split",
	Short: "Split a large CSV file into smaller chunks",
	Run: func(cmd *cobra.Command, args []string) {
		if inputPath == "" {
			fmt.Println("❌ Please provide an input CSV file using --input")
			return
		}

		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			fmt.Printf("❌ Failed to create output directory: %v\n", err)
			return
		}

		file, err := os.Open(inputPath)
		if err != nil {
			fmt.Printf("❌ Failed to open input file: %v\n", err)
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		reader.Comma = ([]rune(delimiter))[0]

		rows, err := reader.ReadAll()
		if err != nil {
			fmt.Printf("❌ Failed to read CSV: %v\n", err)
			return
		}

		var header []string
		startIndex := 0
		if withHeader {
			header = rows[0]
			startIndex = 1
		}

		totalRows := len(rows) - startIndex
		bar := progressbar.Default(int64(totalRows), "Splitting")

		part := 1
		for i := startIndex; i < len(rows); i += rowsPerFile {
			end := i + rowsPerFile
			if end > len(rows) {
				end = len(rows)
			}

			filename := filepath.Join(outputDir, fmt.Sprintf("part_%d.csv", part))
			outFile, err := os.Create(filename)
			if err != nil {
				fmt.Printf("❌ Failed to create file %s: %v\n", filename, err)
				continue
			}

			writer := csv.NewWriter(outFile)
			writer.Comma = ([]rune(delimiter))[0]

			if withHeader {
				_ = writer.Write(header)
			}

			for j := i; j < end; j++ {
				_ = writer.Write(rows[j])
				bar.Add(1)
			}

			writer.Flush()
			outFile.Close()

			part++
		}
	},
}

func init() {
	rootCmd.AddCommand(splitCmd)

	splitCmd.Flags().StringVar(&inputPath, "input", "", "Input CSV file path (required)")
	splitCmd.Flags().StringVar(&outputDir, "output-dir", "./output", "Directory to save split files")
	splitCmd.Flags().IntVar(&rowsPerFile, "rows", 1000, "Max rows per output file")
	splitCmd.Flags().BoolVar(&withHeader, "with-header", true, "Include header in each output file")
	splitCmd.Flags().StringVar(&delimiter, "delimiter", ",", "CSV delimiter character")
}
