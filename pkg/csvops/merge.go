package csvops

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MergeOptions configures a Merge operation.
type MergeOptions struct {
	// InputDir is walked for *.csv files (non-recursive). Ignored if InputFiles is set.
	InputDir string
	// InputFiles is an explicit list of files to merge in order. Takes precedence over InputDir.
	InputFiles []string
	Output     io.Writer
	WithHeader bool
	Delimiter  rune
	// SkipErrors controls behavior when a file fails mid-read: true skips and records
	// a warning, false returns the error. Default false.
	SkipErrors bool
	// OnWarn is invoked for per-file skip events when SkipErrors is true.
	OnWarn   func(file string, err error)
	Progress Progress
}

// MergeResult is returned from Merge.
type MergeResult struct {
	FilesProcessed int
	RowsWritten    int64
}

// Merge streams one CSV at a time into opts.Output, writing the header from the
// first file only when WithHeader is set.
func Merge(ctx context.Context, opts MergeOptions) (MergeResult, error) {
	var res MergeResult

	if opts.Output == nil {
		return res, fmt.Errorf("output writer is required")
	}
	if opts.Delimiter == 0 {
		opts.Delimiter = ','
	}

	files := opts.InputFiles
	if len(files) == 0 {
		if opts.InputDir == "" {
			return res, fmt.Errorf("either InputDir or InputFiles is required")
		}
		entries, err := os.ReadDir(opts.InputDir)
		if err != nil {
			return res, fmt.Errorf("read input dir: %w", err)
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if strings.HasSuffix(strings.ToLower(e.Name()), ".csv") {
				files = append(files, filepath.Join(opts.InputDir, e.Name()))
			}
		}
		sort.Strings(files)
	}

	if len(files) == 0 {
		return res, nil
	}

	writer := csv.NewWriter(opts.Output)
	writer.Comma = opts.Delimiter
	defer writer.Flush()

	writtenHeader := false

	for i, path := range files {
		if err := ctx.Err(); err != nil {
			return res, err
		}
		n, err := mergeOneFile(path, writer, opts.WithHeader, &writtenHeader, opts.Delimiter)
		if err != nil {
			if opts.SkipErrors {
				if opts.OnWarn != nil {
					opts.OnWarn(path, err)
				}
				safeProgress(opts.Progress, int64(i+1), int64(len(files)))
				continue
			}
			return res, fmt.Errorf("%s: %w", filepath.Base(path), err)
		}
		res.RowsWritten += n
		res.FilesProcessed++
		safeProgress(opts.Progress, int64(i+1), int64(len(files)))
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return res, fmt.Errorf("writer: %w", err)
	}
	return res, nil
}

func mergeOneFile(path string, writer *csv.Writer, withHeader bool, writtenHeader *bool, delim rune) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = delim
	r.FieldsPerRecord = -1

	first := true
	var count int64
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, err
		}
		if first {
			first = false
			if withHeader {
				if !*writtenHeader {
					if err := writer.Write(row); err != nil {
						return count, err
					}
					*writtenHeader = true
				}
				continue
			}
		}
		if err := writer.Write(row); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
