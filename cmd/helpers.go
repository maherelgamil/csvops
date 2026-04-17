package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

// parseDelimiter validates that the delimiter string is exactly one rune
// and returns it. Also accepts the literal string "\t" as a tab shortcut.
func parseDelimiter(s string) (rune, error) {
	if s == `\t` {
		return '\t', nil
	}
	if s == "" {
		return 0, fmt.Errorf("--delimiter must be a single character")
	}
	r, size := utf8.DecodeRuneInString(s)
	if size != len(s) {
		return 0, fmt.Errorf("--delimiter must be a single character, got %q", s)
	}
	return r, nil
}

// countDataRows counts the number of non-header data rows in a CSV file.
// Returns 0 if the file is empty.
func countDataRows(path string, delim rune) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("failed to open file for counting: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = delim
	r.FieldsPerRecord = -1

	var total int64
	for {
		_, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		total++
	}
	if total > 0 {
		total-- // exclude header
	}
	return total, nil
}
