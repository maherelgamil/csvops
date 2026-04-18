package csvops

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// FilterOptions configures a Filter operation.
//
// Conditions (Eq / Contains / Gt / Lt) are pointers so callers can distinguish
// "not set" from a zero value — important so that --eq="" matches empty cells.
// By default a row matches if ANY set condition matches; All=true requires ALL.
type FilterOptions struct {
	Input      string
	Output     io.Writer
	Column     string
	Eq         *string
	Contains   *string
	Gt         *float64
	Lt         *float64
	All        bool
	WithHeader bool
	Delimiter  rune
	Progress   Progress
}

// FilterResult is returned from Filter.
type FilterResult struct {
	TotalRows int64
	Matched   int64
}

// Filter streams rows from the input CSV to opts.Output, keeping only rows
// whose value in opts.Column matches the configured conditions.
func Filter(ctx context.Context, opts FilterOptions) (FilterResult, error) {
	var res FilterResult

	if opts.Input == "" {
		return res, fmt.Errorf("input is required")
	}
	if opts.Output == nil {
		return res, fmt.Errorf("output writer is required")
	}
	if opts.Column == "" {
		return res, fmt.Errorf("column is required")
	}
	if opts.Eq == nil && opts.Contains == nil && opts.Gt == nil && opts.Lt == nil {
		return res, fmt.Errorf("at least one condition (Eq, Contains, Gt, Lt) must be set")
	}
	if opts.Delimiter == 0 {
		opts.Delimiter = ','
	}

	total, err := CountDataRows(opts.Input, opts.Delimiter)
	if err != nil {
		return res, err
	}
	res.TotalRows = total

	f, err := os.Open(opts.Input)
	if err != nil {
		return res, fmt.Errorf("open input: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = opts.Delimiter
	reader.FieldsPerRecord = -1

	headers, err := reader.Read()
	if err != nil {
		return res, fmt.Errorf("read header: %w", err)
	}

	colIndex := -1
	for i, c := range headers {
		if c == opts.Column {
			colIndex = i
			break
		}
	}
	if colIndex == -1 {
		return res, fmt.Errorf("column %q not found", opts.Column)
	}

	writer := csv.NewWriter(opts.Output)
	writer.Comma = opts.Delimiter
	if opts.WithHeader {
		if err := writer.Write(headers); err != nil {
			return res, fmt.Errorf("write header: %w", err)
		}
	}

	var processed int64
	for {
		if err := ctx.Err(); err != nil {
			return res, err
		}
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(row) <= colIndex {
			processed++
			safeProgress(opts.Progress, processed, total)
			continue
		}
		if matchesFilter(row[colIndex], opts) {
			if err := writer.Write(row); err != nil {
				return res, fmt.Errorf("write row: %w", err)
			}
			res.Matched++
		}
		processed++
		safeProgress(opts.Progress, processed, total)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return res, fmt.Errorf("writer: %w", err)
	}
	return res, nil
}

func matchesFilter(val string, opts FilterOptions) bool {
	checks := []bool{}
	if opts.Eq != nil {
		checks = append(checks, val == *opts.Eq)
	}
	if opts.Contains != nil {
		checks = append(checks, strings.Contains(strings.ToLower(val), strings.ToLower(*opts.Contains)))
	}
	if opts.Gt != nil {
		num, err := strconv.ParseFloat(val, 64)
		checks = append(checks, err == nil && num > *opts.Gt)
	}
	if opts.Lt != nil {
		num, err := strconv.ParseFloat(val, 64)
		checks = append(checks, err == nil && num < *opts.Lt)
	}
	if len(checks) == 0 {
		return false
	}
	if opts.All {
		for _, c := range checks {
			if !c {
				return false
			}
		}
		return true
	}
	for _, c := range checks {
		if c {
			return true
		}
	}
	return false
}
