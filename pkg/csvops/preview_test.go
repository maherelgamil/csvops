package csvops

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"
)

func TestPreview_WithHeader(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.csv")
	writeCSV(t, in, "id,name\n1,a\n2,b\n3,c\n4,d\n")

	res, err := Preview(context.Background(), PreviewOptions{Input: in, Rows: 2})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(res.Headers, []string{"id", "name"}) {
		t.Errorf("headers = %v", res.Headers)
	}
	if len(res.Rows) != 2 {
		t.Errorf("rows = %d, want 2", len(res.Rows))
	}
	if !reflect.DeepEqual(res.Rows[0], []string{"1", "a"}) {
		t.Errorf("row[0] = %v", res.Rows[0])
	}
}

func TestPreview_NoHeader(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.csv")
	writeCSV(t, in, "1,a\n2,b\n")

	res, err := Preview(context.Background(), PreviewOptions{Input: in, Rows: 5, NoHeader: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Headers) != 0 {
		t.Errorf("expected no headers, got %v", res.Headers)
	}
	if len(res.Rows) != 2 {
		t.Errorf("rows = %d, want 2", len(res.Rows))
	}
}

func TestPreview_AsksForMoreThanAvailable(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.csv")
	writeCSV(t, in, "h\n1\n")
	res, err := Preview(context.Background(), PreviewOptions{Input: in, Rows: 100})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Rows) != 1 {
		t.Errorf("rows = %d, want 1", len(res.Rows))
	}
}

func TestPreview_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.csv")
	writeCSV(t, in, "")
	res, err := Preview(context.Background(), PreviewOptions{Input: in, Rows: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Headers) != 0 || len(res.Rows) != 0 {
		t.Errorf("empty file should yield empty result, got %+v", res)
	}
}
