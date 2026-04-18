package csvops

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// DedupeOptions configures a Dedupe operation.
type DedupeOptions struct {
	Input         string
	Output        string
	KeyColumns    []string
	KeepLast      bool
	CaseSensitive bool
	Delimiter     rune
	Progress      Progress
}

// DedupeResult is returned from Dedupe.
type DedupeResult struct {
	TotalRows  int64
	UniqueRows int
	Duplicates int
}

// Dedupe removes duplicate rows from a CSV file based on one or more key columns.
// Output preserves the original file row order. KeepLast controls whether the
// first or last occurrence of a duplicate key is retained. When Output equals
// Input the file is overwritten in place.
func Dedupe(ctx context.Context, opts DedupeOptions) (DedupeResult, error) {
	var res DedupeResult

	if opts.Input == "" {
		return res, fmt.Errorf("input is required")
	}
	if opts.Output == "" {
		return res, fmt.Errorf("output is required")
	}
	if len(opts.KeyColumns) == 0 {
		return res, fmt.Errorf("at least one key column is required")
	}
	if opts.Delimiter == 0 {
		opts.Delimiter = ','
	}

	totalLines, err := CountDataRows(opts.Input, opts.Delimiter)
	if err != nil {
		return res, err
	}
	res.TotalRows = totalLines

	inFile, err := os.Open(opts.Input)
	if err != nil {
		return res, fmt.Errorf("open input: %w", err)
	}
	defer inFile.Close()

	reader := csv.NewReader(inFile)
	reader.Comma = opts.Delimiter
	reader.FieldsPerRecord = -1

	headers, err := reader.Read()
	if err != nil {
		return res, fmt.Errorf("read headers: %w", err)
	}

	keyIdx, err := resolveKeyIndexes(headers, opts.KeyColumns, opts.CaseSensitive)
	if err != nil {
		return res, err
	}

	tempPath := opts.Output + ".tmp"
	outFile, err := os.Create(tempPath)
	if err != nil {
		return res, fmt.Errorf("create temp output: %w", err)
	}

	writer := csv.NewWriter(outFile)
	writer.Comma = opts.Delimiter
	if err := writer.Write(headers); err != nil {
		outFile.Close()
		return res, fmt.Errorf("write header: %w", err)
	}

	var processed int64

	if opts.KeepLast {
		// Need to see all rows before knowing which is "last".
		rows := [][]string{}
		seen := map[string]int{}
		for {
			if err := ctx.Err(); err != nil {
				outFile.Close()
				return res, err
			}
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				processed++
				safeProgress(opts.Progress, processed, totalLines)
				continue
			}
			if len(row) < len(headers) {
				processed++
				safeProgress(opts.Progress, processed, totalLines)
				continue
			}
			key := BuildDedupeKey(row, keyIdx, opts.CaseSensitive)
			if _, exists := seen[key]; exists {
				res.Duplicates++
			}
			seen[key] = len(rows)
			rows = append(rows, row)
			processed++
			safeProgress(opts.Progress, processed, totalLines)
		}
		kept := make([]int, 0, len(seen))
		for _, idx := range seen {
			kept = append(kept, idx)
		}
		sort.Ints(kept)
		for _, idx := range kept {
			if err := writer.Write(rows[idx]); err != nil {
				outFile.Close()
				return res, fmt.Errorf("write row: %w", err)
			}
		}
		res.UniqueRows = len(kept)
	} else {
		// keep-first: stream as we go.
		seen := map[string]struct{}{}
		for {
			if err := ctx.Err(); err != nil {
				outFile.Close()
				return res, err
			}
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				processed++
				safeProgress(opts.Progress, processed, totalLines)
				continue
			}
			if len(row) < len(headers) {
				processed++
				safeProgress(opts.Progress, processed, totalLines)
				continue
			}
			key := BuildDedupeKey(row, keyIdx, opts.CaseSensitive)
			if _, exists := seen[key]; exists {
				res.Duplicates++
			} else {
				seen[key] = struct{}{}
				if err := writer.Write(row); err != nil {
					outFile.Close()
					return res, fmt.Errorf("write row: %w", err)
				}
				res.UniqueRows++
			}
			processed++
			safeProgress(opts.Progress, processed, totalLines)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		outFile.Close()
		return res, fmt.Errorf("writer: %w", err)
	}
	if err := outFile.Close(); err != nil {
		return res, fmt.Errorf("close output: %w", err)
	}

	// Close input before rename so in-place overwrite works on Windows.
	if opts.Output == opts.Input {
		inFile.Close()
	}
	if err := os.Rename(tempPath, opts.Output); err != nil {
		return res, fmt.Errorf("rename temp file: %w", err)
	}
	return res, nil
}

// BuildDedupeKey joins the values at the given column indexes into a single key.
// Exported so the CLI's tests can keep covering it.
func BuildDedupeKey(row []string, indexes []int, caseSensitive bool) string {
	parts := make([]string, len(indexes))
	for i, idx := range indexes {
		v := row[idx]
		if !caseSensitive {
			v = strings.ToLower(v)
		}
		parts[i] = v
	}
	return strings.Join(parts, "||")
}

func resolveKeyIndexes(headers, keys []string, caseSensitive bool) ([]int, error) {
	out := make([]int, 0, len(keys))
	for _, key := range keys {
		lookup := key
		if !caseSensitive {
			lookup = strings.ToLower(lookup)
		}
		found := false
		for i, h := range headers {
			candidate := h
			if !caseSensitive {
				candidate = strings.ToLower(h)
			}
			if candidate == lookup {
				out = append(out, i)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("column %q not found in headers", key)
		}
	}
	return out, nil
}
