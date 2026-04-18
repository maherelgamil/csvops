package csvops

import (
	"context"
	"path/filepath"
	"testing"
)

func TestStats_Basic(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.csv")
	writeCSV(t, in, "id,country\n1,Egypt\n2,USA\n3,Egypt\n4,\n5,Egypt\n")

	res, err := Stats(context.Background(), StatsOptions{Input: in})
	if err != nil {
		t.Fatal(err)
	}
	if res.TotalRows != 5 {
		t.Errorf("TotalRows = %d, want 5", res.TotalRows)
	}
	if len(res.Columns) != 2 {
		t.Fatalf("Columns len = %d, want 2", len(res.Columns))
	}

	country := res.Columns[1]
	if country.Name != "country" {
		t.Errorf("name = %q", country.Name)
	}
	if country.Empty != 1 {
		t.Errorf("empty = %d, want 1", country.Empty)
	}
	if country.Unique != 2 {
		t.Errorf("unique = %d, want 2 (Egypt, USA)", country.Unique)
	}
	if len(country.Top) == 0 || country.Top[0].Value != "Egypt" || country.Top[0].Count != 3 {
		t.Errorf("top[0] = %+v, want Egypt/3", country.Top)
	}
}

func TestStats_MaxUniqueCap(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.csv")
	writeCSV(t, in, "v\na\nb\nc\nd\ne\na\na\n")

	res, err := Stats(context.Background(), StatsOptions{
		Input:     in,
		MaxUnique: 2, // cap early
	})
	if err != nil {
		t.Fatal(err)
	}
	col := res.Columns[0]
	if !col.UniqueCapped {
		t.Error("expected UniqueCapped=true")
	}
	if col.Unique != 2 {
		t.Errorf("Unique = %d, want 2 (capped)", col.Unique)
	}
	// 'a' was encountered before the cap hit, so its count should keep growing.
	if col.Top[0].Value != "a" || col.Top[0].Count != 3 {
		t.Errorf("top[0] = %+v, want a/3", col.Top[0])
	}
}

func TestStats_TopThree(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.csv")
	writeCSV(t, in, "k\nx\nx\nx\ny\ny\nz\nw\n")

	res, err := Stats(context.Background(), StatsOptions{Input: in})
	if err != nil {
		t.Fatal(err)
	}
	col := res.Columns[0]
	if len(col.Top) != 3 {
		t.Fatalf("top len = %d, want 3", len(col.Top))
	}
	if col.Top[0].Value != "x" || col.Top[0].Count != 3 {
		t.Errorf("top[0] = %+v", col.Top[0])
	}
	if col.Top[1].Value != "y" || col.Top[1].Count != 2 {
		t.Errorf("top[1] = %+v", col.Top[1])
	}
}

func TestStats_EmptyFileHeaderOnly(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.csv")
	writeCSV(t, in, "a,b\n")
	res, err := Stats(context.Background(), StatsOptions{Input: in})
	if err != nil {
		t.Fatal(err)
	}
	if res.TotalRows != 0 {
		t.Errorf("TotalRows = %d, want 0", res.TotalRows)
	}
	if len(res.Columns) != 2 {
		t.Errorf("Columns = %d, want 2", len(res.Columns))
	}
}
