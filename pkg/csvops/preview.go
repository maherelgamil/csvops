package csvops

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
)

// PreviewOptions configures a Preview operation.
type PreviewOptions struct {
	Input     string
	Rows      int  // max rows to return
	NoHeader  bool // if true, the first row is treated as data
	Delimiter rune
}

// PreviewResult is returned from Preview.
type PreviewResult struct {
	Headers    []string   // empty when NoHeader is true
	Rows       [][]string // up to opts.Rows rows
	SkipErrors []error    // per-row parse errors that were skipped
}

// Preview reads the first Rows data rows of the CSV and returns them in memory.
// Unlike the other ops it does not stream — callers expect a bounded slice.
func Preview(ctx context.Context, opts PreviewOptions) (PreviewResult, error) {
	var res PreviewResult

	if opts.Input == "" {
		return res, fmt.Errorf("input is required")
	}
	if opts.Rows <= 0 {
		opts.Rows = 5
	}
	if opts.Delimiter == 0 {
		opts.Delimiter = ','
	}

	f, err := os.Open(opts.Input)
	if err != nil {
		return res, fmt.Errorf("open input: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = opts.Delimiter
	reader.FieldsPerRecord = -1

	if !opts.NoHeader {
		h, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return res, nil
			}
			return res, fmt.Errorf("read header: %w", err)
		}
		res.Headers = h
	}

	res.Rows = make([][]string, 0, opts.Rows)
	for len(res.Rows) < opts.Rows {
		if err := ctx.Err(); err != nil {
			return res, err
		}
		rec, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			res.SkipErrors = append(res.SkipErrors, err)
			continue
		}
		res.Rows = append(res.Rows, rec)
	}
	return res, nil
}
