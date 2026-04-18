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

// StatsOptions configures a Stats operation.
type StatsOptions struct {
	Input string
	// MaxUnique caps the number of distinct values tracked per column.
	// 0 means unlimited. When a column hits the cap, new values are not
	// recorded but existing values' counts continue to increment and the
	// column is flagged as UniqueCapped.
	MaxUnique int
	Delimiter rune
	Progress  Progress
}

// ValueCount is a single (value, occurrence count) pair.
type ValueCount struct {
	Value string
	Count int
}

// ColumnStats is the per-column summary.
type ColumnStats struct {
	Name         string
	Unique       int  // number of distinct non-empty values seen (bounded by MaxUnique)
	UniqueCapped bool // true if MaxUnique was reached
	Empty        int
	Top          []ValueCount // top N by count (N=3)
}

// StatsResult is returned from Stats.
type StatsResult struct {
	TotalRows int64
	Columns   []ColumnStats
}

// Stats scans the CSV and returns row count plus per-column summary
// (unique value count, empty cell count, and top 3 most frequent values).
func Stats(ctx context.Context, opts StatsOptions) (StatsResult, error) {
	var res StatsResult

	if opts.Input == "" {
		return res, fmt.Errorf("input is required")
	}
	if opts.Delimiter == 0 {
		opts.Delimiter = ','
	}

	total, err := CountDataRows(opts.Input, opts.Delimiter)
	if err != nil {
		return res, err
	}

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
		return res, fmt.Errorf("read headers: %w", err)
	}

	type acc struct {
		empty   int
		uniques map[string]int
		capped  bool
	}
	cols := make([]acc, len(headers))
	for i := range cols {
		cols[i].uniques = make(map[string]int)
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
		if err != nil {
			continue
		}
		res.TotalRows++
		for i := range headers {
			cell := ""
			if i < len(row) {
				cell = strings.TrimSpace(row[i])
			}
			if cell == "" {
				cols[i].empty++
				continue
			}
			if opts.MaxUnique > 0 && len(cols[i].uniques) >= opts.MaxUnique {
				if _, exists := cols[i].uniques[cell]; exists {
					cols[i].uniques[cell]++
				} else {
					cols[i].capped = true
				}
			} else {
				cols[i].uniques[cell]++
			}
		}
		processed++
		safeProgress(opts.Progress, processed, total)
	}

	res.Columns = make([]ColumnStats, len(headers))
	for i, name := range headers {
		c := &cols[i]
		sorted := make([]ValueCount, 0, len(c.uniques))
		for k, v := range c.uniques {
			sorted = append(sorted, ValueCount{Value: k, Count: v})
		}
		sort.Slice(sorted, func(a, b int) bool {
			return sorted[a].Count > sorted[b].Count
		})
		top := sorted
		if len(top) > 3 {
			top = top[:3]
		}
		res.Columns[i] = ColumnStats{
			Name:         name,
			Unique:       len(c.uniques),
			UniqueCapped: c.capped,
			Empty:        c.empty,
			Top:          top,
		}
	}
	return res, nil
}
