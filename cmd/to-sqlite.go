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

var identSanitizer = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// quoteIdent safely quotes a SQLite identifier by wrapping in double quotes
// and escaping embedded quotes.
func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

var toSqliteCmd = &cobra.Command{
	Use:   "to-sqlite",
	Short: "Convert a CSV file into a SQLite database",
	RunE: func(cmd *cobra.Command, args []string) error {
		delim, err := parseDelimiter(csvToSqliteDelimiter)
		if err != nil {
			return err
		}

		switch csvToSqliteIfExists {
		case "replace", "skip", "append", "fail":
		default:
			return fmt.Errorf("--if-exists must be one of: replace, skip, append, fail")
		}

		if csvToSqliteTable == "" {
			base := filepath.Base(csvToSqliteInput)
			name := strings.TrimSuffix(base, filepath.Ext(base))
			csvToSqliteTable = identSanitizer.ReplaceAllString(name, "_")
		}

		file, err := os.Open(csvToSqliteInput)
		if err != nil {
			return fmt.Errorf("failed to open input file: %w", err)
		}
		defer file.Close()

		reader := csv.NewReader(file)
		reader.Comma = delim

		headers, err := reader.Read()
		if err != nil {
			return fmt.Errorf("failed to read CSV header: %w", err)
		}

		totalRows, err := countDataRows(csvToSqliteInput, delim)
		if err != nil {
			return err
		}

		dbPath, _ := filepath.Abs(csvToSqliteOutput)
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			return fmt.Errorf("failed to open sqlite db: %w", err)
		}
		defer db.Close()

		tableIdent := quoteIdent(csvToSqliteTable)

		exists, err := tableExists(db, csvToSqliteTable)
		if err != nil {
			return fmt.Errorf("failed to check if table exists: %w", err)
		}

		if exists {
			switch csvToSqliteIfExists {
			case "fail":
				return fmt.Errorf("table %q already exists (use --if-exists replace|skip|append)", csvToSqliteTable)
			case "skip":
				fmt.Printf("⚠️  Table %q already exists, skipping.\n", csvToSqliteTable)
				return nil
			case "replace":
				if _, err := db.Exec("DROP TABLE " + tableIdent); err != nil {
					return fmt.Errorf("failed to drop existing table: %w", err)
				}
				exists = false
			case "append":
				// keep existing table; rows will be inserted below
			}
		}

		if !exists {
			cols := make([]string, len(headers))
			for i, h := range headers {
				cols[i] = quoteIdent(h) + " TEXT"
			}
			createStmt := fmt.Sprintf("CREATE TABLE %s (%s)", tableIdent, strings.Join(cols, ", "))
			if _, err := db.Exec(createStmt); err != nil {
				return fmt.Errorf("failed to create table: %w", err)
			}
		}

		placeholders := strings.TrimRight(strings.Repeat("?,", len(headers)), ",")
		insertStmt := fmt.Sprintf("INSERT INTO %s VALUES (%s)", tableIdent, placeholders)

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		stmt, err := tx.Prepare(insertStmt)
		if err != nil {
			_ = tx.Rollback()
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
				_ = tx.Rollback()
				return fmt.Errorf("failed to read row: %w", err)
			}
			vals := make([]interface{}, len(record))
			for i := range record {
				vals[i] = record[i]
			}
			if _, err := stmt.Exec(vals...); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("failed to insert row: %w", err)
			}
			_ = bar.Add(1)
			rowCount++
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		fmt.Printf("\n✅ Imported %d rows into %s\n", rowCount, dbPath)
		return nil
	},
}

func tableExists(db *sql.DB, name string) (bool, error) {
	row := db.QueryRow(
		"SELECT 1 FROM sqlite_master WHERE type='table' AND name=?",
		name,
	)
	var one int
	err := row.Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func init() {
	rootCmd.AddCommand(toSqliteCmd)

	toSqliteCmd.Flags().StringVar(&csvToSqliteInput, "input", "", "Input CSV file path (required)")
	toSqliteCmd.Flags().StringVar(&csvToSqliteOutput, "output", "", "Output SQLite DB file path (required)")
	toSqliteCmd.Flags().StringVar(&csvToSqliteTable, "table", "", "Table name to create in SQLite (defaults to filename)")
	toSqliteCmd.Flags().StringVar(&csvToSqliteDelimiter, "delimiter", ",", "CSV delimiter character")
	toSqliteCmd.Flags().StringVar(&csvToSqliteIfExists, "if-exists", "replace", "Action if table exists: replace | skip | append | fail")

	_ = toSqliteCmd.MarkFlagRequired("input")
	_ = toSqliteCmd.MarkFlagRequired("output")
}
