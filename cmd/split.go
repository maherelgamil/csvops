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
	RunE: func(cmd *cobra.Command, args []string) error {
		if inputPath == "" {
			return fmt.Errorf("please provide an input CSV file using --input")
		}
		if rowsPerFile <= 0 {
			return fmt.Errorf("--rows must be > 0")
		}
		delim, err := parseDelimiter(delimiter)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		total, err := countDataRows(inputPath, delim)
		if err != nil {
			return err
		}

		file, err := os.Open(inputPath)
		if err != nil {
			return fmt.Errorf("failed to open input file: %w", err)
		}
		defer file.Close()

		reader := csv.NewReader(file)
		reader.Comma = delim
		reader.FieldsPerRecord = -1

		var header []string
		if withHeader {
			h, err := reader.Read()
			if err != nil {
				return fmt.Errorf("failed to read header: %w", err)
			}
			header = h
		}

		rowBuffer := [][]string{}
		rowCount := 0
		part := 1
		bar := progressbar.Default(total, "Splitting")

		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to read row: %w", err)
			}
			rowBuffer = append(rowBuffer, row)
			rowCount++
			_ = bar.Add(1)

			if len(rowBuffer) == rowsPerFile {
				if err := writeChunk(rowBuffer, header, part, delim); err != nil {
					return err
				}
				part++
				rowBuffer = rowBuffer[:0]
			}
		}

		if len(rowBuffer) > 0 {
			if err := writeChunk(rowBuffer, header, part, delim); err != nil {
				return err
			}
		} else {
			part-- // no partial final chunk
		}

		fmt.Printf("\n✅ Finished splitting %d rows into %d file(s).\n", rowCount, part)
		return nil
	},
}

func writeChunk(rows [][]string, header []string, part int, delim rune) error {
	filename := filepath.Join(outputDir, fmt.Sprintf("part_%d.csv", part))
	outFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	writer.Comma = delim

	if withHeader && len(header) > 0 {
		if err := writer.Write(header); err != nil {
			return err
		}
	}
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func init() {
	rootCmd.AddCommand(splitCmd)

	splitCmd.Flags().StringVar(&inputPath, "input", "", "Input CSV file path (required)")
	splitCmd.Flags().StringVar(&outputDir, "output-dir", "./output", "Directory to save split files")
	splitCmd.Flags().IntVar(&rowsPerFile, "rows", 1000, "Max rows per output file")
	splitCmd.Flags().BoolVar(&withHeader, "with-header", true, "Include header in each output file")
	splitCmd.Flags().StringVar(&delimiter, "delimiter", ",", "CSV delimiter character")
}
