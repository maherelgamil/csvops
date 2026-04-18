package csvops

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "modernc.org/sqlite"
)

// IfExistsAction controls ToSQLite behavior when the target table already exists.
type IfExistsAction string

const (
	IfExistsReplace IfExistsAction = "replace"
	IfExistsSkip    IfExistsAction = "skip"
	IfExistsAppend  IfExistsAction = "append"
	IfExistsFail    IfExistsAction = "fail"
)

// ToSQLiteOptions configures a ToSQLite operation.
type ToSQLiteOptions struct {
	Input     string
	DBPath    string
	Table     string // defaults to sanitized input filename
	IfExists  IfExistsAction
	Delimiter rune
	Progress  Progress
}

// ToSQLiteResult is returned from ToSQLite.
type ToSQLiteResult struct {
	Table        string
	RowsImported int64
	Skipped      bool // true when IfExistsSkip was honored
}

var identSanitizer = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// QuoteIdent wraps a SQLite identifier in double quotes, escaping embedded quotes.
// Exported so the CLI and its tests can verify injection safety.
func QuoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// SanitizeTableName returns an ASCII-safe table name derived from a path.
func SanitizeTableName(path string) string {
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	return identSanitizer.ReplaceAllString(name, "_")
}

// ToSQLite imports a CSV file into a SQLite database. All columns are created
// as TEXT. Identifiers (table and column names) are quoted, so untrusted names
// cannot inject SQL.
func ToSQLite(ctx context.Context, opts ToSQLiteOptions) (ToSQLiteResult, error) {
	var res ToSQLiteResult

	if opts.Input == "" {
		return res, fmt.Errorf("input is required")
	}
	if opts.DBPath == "" {
		return res, fmt.Errorf("DBPath is required")
	}
	if opts.IfExists == "" {
		opts.IfExists = IfExistsReplace
	}
	switch opts.IfExists {
	case IfExistsReplace, IfExistsSkip, IfExistsAppend, IfExistsFail:
	default:
		return res, fmt.Errorf("IfExists must be one of: replace, skip, append, fail (got %q)", opts.IfExists)
	}
	if opts.Delimiter == 0 {
		opts.Delimiter = ','
	}
	if opts.Table == "" {
		opts.Table = SanitizeTableName(opts.Input)
	}
	res.Table = opts.Table

	total, err := CountDataRows(opts.Input, opts.Delimiter)
	if err != nil {
		return res, err
	}

	f, err := os.Open(opts.Input)
	if err != nil {
		return res, fmt.Errorf("open input: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = opts.Delimiter
	reader.FieldsPerRecord = -1

	headers, err := reader.Read()
	if err != nil {
		return res, fmt.Errorf("read header: %w", err)
	}

	dbPath, _ := filepath.Abs(opts.DBPath)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return res, fmt.Errorf("open sqlite: %w", err)
	}
	defer db.Close()

	tableIdent := QuoteIdent(opts.Table)

	exists, err := tableExists(db, opts.Table)
	if err != nil {
		return res, fmt.Errorf("check table existence: %w", err)
	}

	if exists {
		switch opts.IfExists {
		case IfExistsFail:
			return res, fmt.Errorf("table %q already exists", opts.Table)
		case IfExistsSkip:
			res.Skipped = true
			return res, nil
		case IfExistsReplace:
			if _, err := db.Exec("DROP TABLE " + tableIdent); err != nil {
				return res, fmt.Errorf("drop existing table: %w", err)
			}
			exists = false
		case IfExistsAppend:
			// keep existing table; rows inserted below
		}
	}

	if !exists {
		cols := make([]string, len(headers))
		for i, h := range headers {
			cols[i] = QuoteIdent(h) + " TEXT"
		}
		createStmt := fmt.Sprintf("CREATE TABLE %s (%s)", tableIdent, strings.Join(cols, ", "))
		if _, err := db.Exec(createStmt); err != nil {
			return res, fmt.Errorf("create table: %w", err)
		}
	}

	placeholders := strings.TrimRight(strings.Repeat("?,", len(headers)), ",")
	insertStmt := fmt.Sprintf("INSERT INTO %s VALUES (%s)", tableIdent, placeholders)

	tx, err := db.Begin()
	if err != nil {
		return res, fmt.Errorf("begin tx: %w", err)
	}
	stmt, err := tx.Prepare(insertStmt)
	if err != nil {
		_ = tx.Rollback()
		return res, fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	var processed int64
	for {
		if err := ctx.Err(); err != nil {
			_ = tx.Rollback()
			return res, err
		}
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			_ = tx.Rollback()
			return res, fmt.Errorf("read row: %w", err)
		}
		vals := make([]any, len(rec))
		for i := range rec {
			vals[i] = rec[i]
		}
		if _, err := stmt.Exec(vals...); err != nil {
			_ = tx.Rollback()
			return res, fmt.Errorf("insert row: %w", err)
		}
		processed++
		res.RowsImported++
		safeProgress(opts.Progress, processed, total)
	}

	if err := tx.Commit(); err != nil {
		return res, fmt.Errorf("commit: %w", err)
	}
	return res, nil
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
