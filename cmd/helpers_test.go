package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDelimiter(t *testing.T) {
	tests := []struct {
		in      string
		want    rune
		wantErr bool
	}{
		{",", ',', false},
		{";", ';', false},
		{"|", '|', false},
		{"\t", '\t', false},
		{`\t`, '\t', false}, // literal escape shortcut
		{"", 0, true},
		{",,", 0, true},
		{"abc", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := parseDelimiter(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseDelimiter(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("parseDelimiter(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCountDataRows(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")
	if err := os.WriteFile(path, []byte("a,b\n1,2\n3,4\n5,6\n"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := countDataRows(path, ',')
	if err != nil {
		t.Fatal(err)
	}
	if got != 3 {
		t.Fatalf("countDataRows = %d, want 3", got)
	}
}

func TestCountDataRowsEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.csv")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := countDataRows(path, ',')
	if err != nil {
		t.Fatal(err)
	}
	if got != 0 {
		t.Fatalf("countDataRows on empty = %d, want 0", got)
	}
}

func TestCountDataRowsHeaderOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "h.csv")
	if err := os.WriteFile(path, []byte("a,b\n"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := countDataRows(path, ',')
	if err != nil {
		t.Fatal(err)
	}
	if got != 0 {
		t.Fatalf("countDataRows header-only = %d, want 0", got)
	}
}
