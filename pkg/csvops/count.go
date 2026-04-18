package csvops

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

// CountDataRows counts non-header data rows in a CSV file, treating the first
// line as a header. Returns 0 for an empty file or a header-only file.
func CountDataRows(path string, delim rune) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", path, err)
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
