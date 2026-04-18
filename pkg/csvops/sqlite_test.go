package csvops

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func queryCount(t *testing.T, dbPath, table string) int {
	t.Helper()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	row := db.QueryRow("SELECT COUNT(*) FROM " + QuoteIdent(table))
	var n int
	if err := row.Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestQuoteIdent(t *testing.T) {
	tests := map[string]string{
		"users":                    `"users"`,
		"my table":                 `"my table"`,
		`weird"; drop table x; --`: `"weird""; drop table x; --"`,
		`a"b"c`:                    `"a""b""c"`,
	}
	for in, want := range tests {
		if got := QuoteIdent(in); got != want {
			t.Errorf("QuoteIdent(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestToSQLite_BasicImport(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "data.csv")
	db := filepath.Join(dir, "out.db")
	writeCSV(t, in, "id,name\n1,a\n2,b\n3,c\n")

	res, err := ToSQLite(context.Background(), ToSQLiteOptions{
		Input:  in,
		DBPath: db,
		Table:  "users",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.RowsImported != 3 {
		t.Errorf("RowsImported = %d, want 3", res.RowsImported)
	}
	if queryCount(t, db, "users") != 3 {
		t.Error("table row count mismatch")
	}
}

func TestToSQLite_MaliciousTableNameQuoted(t *testing.T) {
	// The key regression test: an adversarial --table value must not inject SQL.
	dir := t.TempDir()
	in := filepath.Join(dir, "data.csv")
	db := filepath.Join(dir, "out.db")
	writeCSV(t, in, "id\n1\n")

	nasty := `x"; DROP TABLE secrets; --`
	res, err := ToSQLite(context.Background(), ToSQLiteOptions{
		Input:  in,
		DBPath: db,
		Table:  nasty,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Table != nasty {
		t.Errorf("Table = %q, want %q", res.Table, nasty)
	}
	// The table must exist verbatim — if SQL were injected, CREATE would fail or split.
	if queryCount(t, db, nasty) != 1 {
		t.Error("adversarial table name not stored verbatim")
	}
}

func TestToSQLite_IfExistsSkip(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "data.csv")
	db := filepath.Join(dir, "out.db")
	writeCSV(t, in, "id\n1\n2\n")

	if _, err := ToSQLite(context.Background(), ToSQLiteOptions{
		Input: in, DBPath: db, Table: "t",
	}); err != nil {
		t.Fatal(err)
	}
	res, err := ToSQLite(context.Background(), ToSQLiteOptions{
		Input: in, DBPath: db, Table: "t", IfExists: IfExistsSkip,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Skipped {
		t.Error("expected Skipped=true")
	}
	if queryCount(t, db, "t") != 2 {
		t.Error("skip mode should not touch existing rows")
	}
}

func TestToSQLite_IfExistsFail(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "data.csv")
	db := filepath.Join(dir, "out.db")
	writeCSV(t, in, "id\n1\n")

	if _, err := ToSQLite(context.Background(), ToSQLiteOptions{
		Input: in, DBPath: db, Table: "t",
	}); err != nil {
		t.Fatal(err)
	}
	_, err := ToSQLite(context.Background(), ToSQLiteOptions{
		Input: in, DBPath: db, Table: "t", IfExists: IfExistsFail,
	})
	if err == nil {
		t.Error("expected error with IfExistsFail")
	}
}

func TestToSQLite_IfExistsAppend(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "data.csv")
	db := filepath.Join(dir, "out.db")
	writeCSV(t, in, "id\n1\n2\n")

	if _, err := ToSQLite(context.Background(), ToSQLiteOptions{
		Input: in, DBPath: db, Table: "t",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := ToSQLite(context.Background(), ToSQLiteOptions{
		Input: in, DBPath: db, Table: "t", IfExists: IfExistsAppend,
	}); err != nil {
		t.Fatal(err)
	}
	if n := queryCount(t, db, "t"); n != 4 {
		t.Errorf("append: count = %d, want 4", n)
	}
}

func TestToSQLite_IfExistsReplace(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "data.csv")
	db := filepath.Join(dir, "out.db")
	writeCSV(t, in, "id\n1\n2\n3\n")

	if _, err := ToSQLite(context.Background(), ToSQLiteOptions{
		Input: in, DBPath: db, Table: "t",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := ToSQLite(context.Background(), ToSQLiteOptions{
		Input: in, DBPath: db, Table: "t", IfExists: IfExistsReplace,
	}); err != nil {
		t.Fatal(err)
	}
	if n := queryCount(t, db, "t"); n != 3 {
		t.Errorf("replace: count = %d, want 3", n)
	}
}

func TestSanitizeTableName(t *testing.T) {
	tests := map[string]string{
		"/a/b/users.csv":        "users",
		"/a/b/my file.csv":      "my_file",
		"/a/b/weird$$chars.csv": "weird__chars",
	}
	for in, want := range tests {
		if got := SanitizeTableName(in); got != want {
			t.Errorf("SanitizeTableName(%q) = %q, want %q", in, got, want)
		}
	}
}
