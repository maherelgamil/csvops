package cmd

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

var (
	csvToSqliteInput     string
	csvToSqliteOutput    string
	csvToSqliteTable     string
	csvToSqliteDelimiter string
	csvToSqliteIfExists  string
)

var toSqliteCmd = &cobra.Command{
	Use:   "to-sqlite",
	Short: "Convert a CSV file into a SQLite database",
	RunE: func(cmd *cobra.Command, args []string) error {
		file, err := os.Open(csvToSqliteInput)
		if err != nil {
			return fmt.Errorf("failed to open input file: %w", err)
		}
		defer file.Close()

		if csvToSqliteTable == "" {
			base := filepath.Base(csvToSqliteInput)
			name := strings.TrimSuffix(base, filepath.Ext(base))
			cleanName := regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(name, "_")
			csvToSqliteTable = cleanName
		}

		reader := csv.NewReader(file)
		reader.Comma = ([]rune(csvToSqliteDelimiter))[0]

		headers, err := reader.Read()
		if err != nil {
			return fmt.Errorf("failed to read CSV header: %w", err)
		}

		// Count total rows for progress bar
		totalRows := int64(0)
		lineCounter, _ := os.Open(csvToSqliteInput)
		defer lineCounter.Close()
		lcReader := csv.NewReader(lineCounter)
		lcReader.Comma = ([]rune(csvToSqliteDelimiter))[0]
		for {
			_, err := lcReader.Read()
			if err == io.EOF {
				break
			}
			if err == nil {
				totalRows++
			}
		}
		totalRows-- // remove header row

		// reopen file to re-read rows
		file.Seek(0, 0)
		reader = csv.NewReader(file)
		reader.Comma = ([]rune(csvToSqliteDelimiter))[0]
		_, _ = reader.Read() // skip header

		dbPath, _ := filepath.Abs(csvToSqliteOutput)
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			return fmt.Errorf("failed to open sqlite db: %w", err)
		}
		defer db.Close()

		if csvToSqliteIfExists == "replace" {
			db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", csvToSqliteTable))
		}

		columns := make([]string, len(headers))
		for i, h := range headers {
			columns[i] = fmt.Sprintf("%s TEXT", h)
		}

		createStmt := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)",
			csvToSqliteTable,
			strings.Join(columns, ", "))

		_, err = db.Exec(createStmt)
		if err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}

		placeholders := strings.Repeat("?,", len(headers))
		placeholders = placeholders[:len(placeholders)-1]
		insertStmt := fmt.Sprintf("INSERT INTO %s VALUES (%s)", csvToSqliteTable, placeholders)

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		stmt, err := tx.Prepare(insertStmt)
		if err != nil {
			return fmt.Errorf("failed to prepare insert statement: %w", err)
		}
		defer stmt.Close()

		bar := progressbar.Default(totalRows, "Converting")
		rowCount := 0
		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to read row: %w", err)
			}
			vals := make([]interface{}, len(record))
			for i := range record {
				vals[i] = record[i]
			}
			_, err = stmt.Exec(vals...)
			if err != nil {
				return fmt.Errorf("failed to insert row: %w", err)
			}
			bar.Add(1)
			rowCount++
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		fmt.Printf("\n✅ Imported %d rows into %s\n", rowCount, dbPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(toSqliteCmd)

	toSqliteCmd.Flags().StringVar(&csvToSqliteInput, "input", "", "Input CSV file path (required)")
	toSqliteCmd.Flags().StringVar(&csvToSqliteOutput, "output", "", "Output SQLite DB file path (required)")
	toSqliteCmd.Flags().StringVar(&csvToSqliteTable, "table", "", "Table name to create in SQLite (defaults to filename)")
	toSqliteCmd.Flags().StringVar(&csvToSqliteDelimiter, "delimiter", ",", "CSV delimiter character")
	toSqliteCmd.Flags().StringVar(&csvToSqliteIfExists, "if-exists", "replace", "Action if table exists: skip | replace")

	toSqliteCmd.MarkFlagRequired("input")
	toSqliteCmd.MarkFlagRequired("output")
}
