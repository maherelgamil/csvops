package cmd

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
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

		records, err := reader.ReadAll()
		if err != nil {
			return fmt.Errorf("failed to read CSV file: %w", err)
		}
		if len(records) < 1 {
			return fmt.Errorf("CSV file is empty")
		}

		headers := records[0]
		rows := records[1:]

		dbPath, _ := filepath.Abs(csvToSqliteOutput)
		db, err := sql.Open("sqlite3", dbPath)
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
		stmt, err := db.Prepare(insertStmt)
		if err != nil {
			return fmt.Errorf("failed to prepare insert statement: %w", err)
		}
		defer stmt.Close()

		bar := progressbar.Default(int64(len(rows)), "Converting")

		for _, row := range rows {
			vals := make([]interface{}, len(row))
			for i := range row {
				vals[i] = row[i]
			}
			_, err := stmt.Exec(vals...)
			if err != nil {
				return fmt.Errorf("failed to insert row: %w", err)
			}
			bar.Add(1)
		}

		fmt.Printf("\n✅ CSV converted and saved to %s\n", dbPath)
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
