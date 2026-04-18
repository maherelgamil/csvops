package cmd

import (
	"fmt"
	"unicode/utf8"

	"github.com/maherelgamil/csvops/pkg/csvops"
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

// countDataRows delegates to the library implementation.
func countDataRows(path string, delim rune) (int64, error) {
	return csvops.CountDataRows(path, delim)
}
