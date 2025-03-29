package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
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

		lineCounter, err := os.Open(inputPath)
		if err != nil {
			fmt.Printf("❌ Failed to open file for counting: %v\n", err)
			return
		}
		lcReader := csv.NewReader(lineCounter)
		lcReader.Comma = ([]rune(delimiter))[0]
		total := int64(0)
		for {
			_, err := lcReader.Read()
			if err == io.EOF {
				break
			}
			if err == nil {
				total++
			}
		}
		lineCounter.Close()
		if withHeader && total > 0 {
			total--
		}

		file, err := os.Open(inputPath)
		if err != nil {
			fmt.Printf("❌ Failed to open input file: %v\n", err)
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		reader.Comma = ([]rune(delimiter))[0]

		header := []string{}
		rowBuffer := [][]string{}
		rowCount := 0
		part := 1

		bar := progressbar.Default(total, "Splitting")

		if withHeader {
			h, err := reader.Read()
			if err != nil {
				fmt.Printf("❌ Failed to read header: %v\n", err)
				return
			}
			header = h
		}

		for {
			row, err := reader.Read()
			if err != nil {
				break
			}
			rowBuffer = append(rowBuffer, row)
			rowCount++
			bar.Add(1)

			if len(rowBuffer) == rowsPerFile {
				writeChunk(rowBuffer, header, part)
				part++
				rowBuffer = [][]string{}
			}
		}

		if len(rowBuffer) > 0 {
			writeChunk(rowBuffer, header, part)
		}

		fmt.Printf("\n✅ Finished splitting %d rows into %d file(s).\n", rowCount, part)
	},
}

func writeChunk(rows [][]string, header []string, part int) {
	filename := filepath.Join(outputDir, fmt.Sprintf("part_%d.csv", part))
	outFile, err := os.Create(filename)
	if err != nil {
		fmt.Printf("❌ Failed to create file %s: %v\n", filename, err)
		return
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	writer.Comma = ([]rune(delimiter))[0]

	if withHeader {
		_ = writer.Write(header)
	}

	for _, row := range rows {
		_ = writer.Write(row)
	}

	writer.Flush()
}

func init() {
	rootCmd.AddCommand(splitCmd)

	splitCmd.Flags().StringVar(&inputPath, "input", "", "Input CSV file path (required)")
	splitCmd.Flags().StringVar(&outputDir, "output-dir", "./output", "Directory to save split files")
	splitCmd.Flags().IntVar(&rowsPerFile, "rows", 1000, "Max rows per output file")
	splitCmd.Flags().BoolVar(&withHeader, "with-header", true, "Include header in each output file")
	splitCmd.Flags().StringVar(&delimiter, "delimiter", ",", "CSV delimiter character")
}
