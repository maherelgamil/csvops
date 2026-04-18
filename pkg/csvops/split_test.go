package csvops

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeCSV(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestSplit_BasicWithHeader(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.csv")
	out := filepath.Join(dir, "parts")
	writeCSV(t, in, "id,name\n1,a\n2,b\n3,c\n4,d\n5,e\n")

	res, err := Split(context.Background(), SplitOptions{
		Input:       in,
		OutputDir:   out,
		RowsPerFile: 2,
		WithHeader:  true,
		Delimiter:   ',',
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.RowsProcessed != 5 {
		t.Errorf("RowsProcessed = %d, want 5", res.RowsProcessed)
	}
	if res.FilesCreated != 3 {
		t.Errorf("FilesCreated = %d, want 3", res.FilesCreated)
	}

	got := readFile(t, filepath.Join(out, "part_1.csv"))
	if !strings.HasPrefix(got, "id,name\n") {
		t.Errorf("part_1.csv missing header: %q", got)
	}

	got3 := readFile(t, filepath.Join(out, "part_3.csv"))
	if !strings.Contains(got3, "5,e") {
		t.Errorf("part_3.csv missing trailing row: %q", got3)
	}
}

func TestSplit_ProgressCallbackInvoked(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.csv")
	writeCSV(t, in, "h\n1\n2\n3\n")

	var calls int
	var lastDone, lastTotal int64
	_, err := Split(context.Background(), SplitOptions{
		Input:       in,
		OutputDir:   filepath.Join(dir, "out"),
		RowsPerFile: 1,
		WithHeader:  true,
		Delimiter:   ',',
		Progress: func(done, total int64) {
			calls++
			lastDone, lastTotal = done, total
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 3 {
		t.Errorf("progress called %d times, want 3", calls)
	}
	if lastDone != 3 || lastTotal != 3 {
		t.Errorf("final progress = (%d,%d), want (3,3)", lastDone, lastTotal)
	}
}

func TestSplit_CancelledContext(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.csv")
	writeCSV(t, in, "h\n1\n2\n3\n")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Split(ctx, SplitOptions{
		Input:       in,
		OutputDir:   filepath.Join(dir, "out"),
		RowsPerFile: 1,
		WithHeader:  true,
		Delimiter:   ',',
	})
	if err != context.Canceled {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

func TestSplit_ValidatesInputs(t *testing.T) {
	_, err := Split(context.Background(), SplitOptions{RowsPerFile: 10})
	if err == nil {
		t.Error("expected error for missing input")
	}
	_, err = Split(context.Background(), SplitOptions{Input: "x", RowsPerFile: 0})
	if err == nil {
		t.Error("expected error for RowsPerFile=0")
	}
}
