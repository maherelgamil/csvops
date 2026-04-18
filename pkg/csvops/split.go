package csvops

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// SplitOptions configures a Split operation.
type SplitOptions struct {
	Input       string
	OutputDir   string
	RowsPerFile int
	WithHeader  bool
	Delimiter   rune
	Progress    Progress
}

// SplitResult is returned from Split.
type SplitResult struct {
	RowsProcessed int64
	FilesCreated  int
}

// Split streams the CSV at opts.Input and writes chunks of RowsPerFile rows
// into opts.OutputDir as part_1.csv, part_2.csv, ...
func Split(ctx context.Context, opts SplitOptions) (SplitResult, error) {
	var res SplitResult

	if opts.Input == "" {
		return res, fmt.Errorf("input is required")
	}
	if opts.RowsPerFile <= 0 {
		return res, fmt.Errorf("RowsPerFile must be > 0")
	}
	if opts.Delimiter == 0 {
		opts.Delimiter = ','
	}
	if opts.OutputDir == "" {
		opts.OutputDir = "."
	}

	if err := os.MkdirAll(opts.OutputDir, 0o755); err != nil {
		return res, fmt.Errorf("create output dir: %w", err)
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

	r := csv.NewReader(f)
	r.Comma = opts.Delimiter
	r.FieldsPerRecord = -1

	var header []string
	if opts.WithHeader {
		h, err := r.Read()
		if err != nil {
			return res, fmt.Errorf("read header: %w", err)
		}
		header = h
	}

	buf := make([][]string, 0, opts.RowsPerFile)
	part := 1

	flush := func() error {
		if len(buf) == 0 {
			return nil
		}
		if err := writeSplitChunk(opts.OutputDir, part, header, buf, opts.Delimiter, opts.WithHeader); err != nil {
			return err
		}
		part++
		buf = buf[:0]
		return nil
	}

	for {
		if err := ctx.Err(); err != nil {
			return res, err
		}
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return res, fmt.Errorf("read row: %w", err)
		}
		buf = append(buf, row)
		res.RowsProcessed++
		safeProgress(opts.Progress, res.RowsProcessed, total)

		if len(buf) == opts.RowsPerFile {
			if err := flush(); err != nil {
				return res, err
			}
		}
	}
	if err := flush(); err != nil {
		return res, err
	}
	res.FilesCreated = part - 1
	return res, nil
}

func writeSplitChunk(dir string, part int, header []string, rows [][]string, delim rune, withHeader bool) error {
	path := filepath.Join(dir, fmt.Sprintf("part_%d.csv", part))
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.Comma = delim

	if withHeader && len(header) > 0 {
		if err := w.Write(header); err != nil {
			return err
		}
	}
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}
