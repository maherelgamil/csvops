package csvops

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestMerge_FromDirectorySortedAndHeaderOnce(t *testing.T) {
	dir := t.TempDir()
	writeCSV(t, filepath.Join(dir, "a.csv"), "id,name\n1,a\n2,b\n")
	writeCSV(t, filepath.Join(dir, "b.csv"), "id,name\n3,c\n4,d\n")
	writeCSV(t, filepath.Join(dir, "c.csv"), "id,name\n5,e\n")

	var buf bytes.Buffer
	res, err := Merge(context.Background(), MergeOptions{
		InputDir:   dir,
		Output:     &buf,
		WithHeader: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.FilesProcessed != 3 {
		t.Errorf("FilesProcessed = %d, want 3", res.FilesProcessed)
	}
	if res.RowsWritten != 5 {
		t.Errorf("RowsWritten = %d, want 5", res.RowsWritten)
	}
	got := buf.String()
	want := "id,name\n1,a\n2,b\n3,c\n4,d\n5,e\n"
	if got != want {
		t.Errorf("output:\n%s\nwant:\n%s", got, want)
	}
	if strings.Count(got, "id,name") != 1 {
		t.Errorf("header should appear once, got %d", strings.Count(got, "id,name"))
	}
}

func TestMerge_ExplicitInputFiles(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "z.csv")
	b := filepath.Join(dir, "a.csv")
	writeCSV(t, a, "id\n1\n")
	writeCSV(t, b, "id\n2\n")

	var buf bytes.Buffer
	res, err := Merge(context.Background(), MergeOptions{
		InputFiles: []string{a, b}, // explicit order, NOT alphabetical
		Output:     &buf,
		WithHeader: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.RowsWritten != 2 {
		t.Errorf("RowsWritten = %d, want 2", res.RowsWritten)
	}
	got := buf.String()
	if got != "id\n1\n2\n" {
		t.Errorf("output order not respected:\n%s", got)
	}
}

func TestMerge_NoHeaderMode(t *testing.T) {
	dir := t.TempDir()
	writeCSV(t, filepath.Join(dir, "a.csv"), "1,a\n2,b\n")
	writeCSV(t, filepath.Join(dir, "b.csv"), "3,c\n")

	var buf bytes.Buffer
	_, err := Merge(context.Background(), MergeOptions{
		InputDir: dir,
		Output:   &buf,
	})
	if err != nil {
		t.Fatal(err)
	}
	// WithHeader=false: every row from every file is emitted verbatim.
	if buf.String() != "1,a\n2,b\n3,c\n" {
		t.Errorf("no-header output:\n%s", buf.String())
	}
}

func TestMerge_EmptyDir(t *testing.T) {
	var buf bytes.Buffer
	res, err := Merge(context.Background(), MergeOptions{
		InputDir: t.TempDir(),
		Output:   &buf,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.FilesProcessed != 0 || buf.Len() != 0 {
		t.Errorf("expected empty result, got %+v output=%q", res, buf.String())
	}
}

func TestMerge_SkipErrorsOnMissingFile(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "a.csv")
	writeCSV(t, good, "id\n1\n")
	missing := filepath.Join(dir, "nope.csv")

	var warned []string
	var buf bytes.Buffer
	res, err := Merge(context.Background(), MergeOptions{
		InputFiles: []string{missing, good},
		Output:     &buf,
		WithHeader: true,
		SkipErrors: true,
		OnWarn: func(file string, e error) {
			warned = append(warned, filepath.Base(file))
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.FilesProcessed != 1 {
		t.Errorf("FilesProcessed = %d, want 1", res.FilesProcessed)
	}
	if len(warned) != 1 || warned[0] != "nope.csv" {
		t.Errorf("warnings = %v", warned)
	}
}
