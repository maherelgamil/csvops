package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"os/exec"
	goruntime "runtime"
	"strings"

	"github.com/maherelgamil/csvops/pkg/csvops"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails-bound struct. Each public method becomes callable from the
// frontend via the generated TypeScript bindings.
type App struct {
	ctx context.Context
}

func NewApp() *App { return &App{} }
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// emitProgress builds a Progress callback that sends a "progress" event to
// the frontend tagged with the operation name. The frontend subscribes once
// and routes by op.
func (a *App) emitProgress(op string) csvops.Progress {
	return func(done, total int64) {
		runtime.EventsEmit(a.ctx, "progress", map[string]any{
			"op":    op,
			"done":  done,
			"total": total,
		})
	}
}

// ----- File / dir pickers --------------------------------------------------

func (a *App) OpenCSVFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open CSV file",
		Filters: []runtime.FileFilter{
			{DisplayName: "CSV files (*.csv)", Pattern: "*.csv"},
			{DisplayName: "All files", Pattern: "*.*"},
		},
	})
}

func (a *App) SaveCSVFile(suggestedName string) (string, error) {
	if suggestedName == "" {
		suggestedName = "output.csv"
	}
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save CSV as…",
		DefaultFilename: suggestedName,
		Filters: []runtime.FileFilter{
			{DisplayName: "CSV files (*.csv)", Pattern: "*.csv"},
		},
	})
}

func (a *App) SaveDBFile(suggestedName string) (string, error) {
	if suggestedName == "" {
		suggestedName = "output.db"
	}
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save SQLite database as…",
		DefaultFilename: suggestedName,
		Filters: []runtime.FileFilter{
			{DisplayName: "SQLite databases (*.db)", Pattern: "*.db"},
			{DisplayName: "All files", Pattern: "*.*"},
		},
	})
}

func (a *App) OpenDirectory(title string) (string, error) {
	if title == "" {
		title = "Select directory"
	}
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
	})
}

// ----- File metadata --------------------------------------------------------

type FileInfo struct {
	Path    string   `json:"path"`
	Size    int64    `json:"size"`
	Rows    int64    `json:"rows"`
	Headers []string `json:"headers"`
}

// FileInfoCSV returns size + row count + headers for a CSV. The frontend uses
// it to show file context and to populate column dropdowns.
func (a *App) FileInfoCSV(path string) (FileInfo, error) {
	st, err := os.Stat(path)
	if err != nil {
		return FileInfo{}, err
	}
	rows, _ := csvops.CountDataRows(path, ',')
	headers, _ := readHeaders(path)
	return FileInfo{
		Path:    path,
		Size:    st.Size(),
		Rows:    rows,
		Headers: headers,
	}, nil
}

func readHeaders(path string) ([]string, error) {
	res, err := csvops.Preview(context.Background(), csvops.PreviewOptions{
		Input: path,
		Rows:  0,
	})
	if err != nil {
		return nil, err
	}
	return res.Headers, nil
}

// RevealFile shows a file (or dir) in the OS native file manager.
func (a *App) RevealFile(path string) error {
	if path == "" {
		return fmt.Errorf("empty path")
	}
	switch goruntime.GOOS {
	case "darwin":
		return exec.Command("open", "-R", path).Start()
	case "windows":
		return exec.Command("explorer", "/select,", path).Start()
	default:
		// best effort on Linux: open the parent dir
		dir := path
		if st, err := os.Stat(path); err == nil && !st.IsDir() {
			dir = strings.TrimSuffix(path, "/"+st.Name())
		}
		return exec.Command("xdg-open", dir).Start()
	}
}

// ----- ReadPage (paginated browsing) ----------------------------------------

type PagePayload struct {
	Headers   []string   `json:"headers"`
	Rows      [][]string `json:"rows"`
	Offset    int64      `json:"offset"`
	TotalRows int64      `json:"totalRows"`
}

// ReadPage returns a paginated window of rows from a CSV. Offset is the index
// of the first data row to return (0-based, header excluded). Limit caps the
// returned rows. The frontend uses this to render the unified table view.
func (a *App) ReadPage(path string, offset, limit int) (PagePayload, error) {
	if limit <= 0 {
		limit = 100
	}
	total, err := csvops.CountDataRows(path, ',')
	if err != nil {
		return PagePayload{}, err
	}

	f, err := os.Open(path)
	if err != nil {
		return PagePayload{}, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1

	headers, err := r.Read()
	if err != nil {
		return PagePayload{}, err
	}

	// Skip `offset` rows.
	for i := 0; i < offset; i++ {
		if _, err := r.Read(); err != nil {
			return PagePayload{Headers: headers, Offset: int64(offset), TotalRows: total}, nil
		}
	}

	rows := make([][]string, 0, limit)
	for len(rows) < limit {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		rows = append(rows, rec)
	}

	return PagePayload{
		Headers:   headers,
		Rows:      rows,
		Offset:    int64(offset),
		TotalRows: total,
	}, nil
}

// ----- Preview --------------------------------------------------------------

type PreviewPayload struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}

func (a *App) PreviewCSV(path string, rows int, noHeader bool) (PreviewPayload, error) {
	res, err := csvops.Preview(a.ctx, csvops.PreviewOptions{
		Input: path, Rows: rows, NoHeader: noHeader,
	})
	if err != nil {
		return PreviewPayload{}, err
	}
	return PreviewPayload{Headers: res.Headers, Rows: res.Rows}, nil
}

// ----- Stats ----------------------------------------------------------------

type StatsValueCount struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

type StatsColumn struct {
	Name         string            `json:"name"`
	Unique       int               `json:"unique"`
	UniqueCapped bool              `json:"uniqueCapped"`
	Empty        int               `json:"empty"`
	Top          []StatsValueCount `json:"top"`
}

type StatsPayload struct {
	TotalRows int64         `json:"totalRows"`
	Columns   []StatsColumn `json:"columns"`
}

func (a *App) StatsCSV(path string, maxUnique int) (StatsPayload, error) {
	res, err := csvops.Stats(a.ctx, csvops.StatsOptions{
		Input:     path,
		MaxUnique: maxUnique,
		Progress:  a.emitProgress("stats"),
	})
	if err != nil {
		return StatsPayload{}, err
	}
	cols := make([]StatsColumn, len(res.Columns))
	for i, c := range res.Columns {
		top := make([]StatsValueCount, len(c.Top))
		for j, v := range c.Top {
			top[j] = StatsValueCount{Value: v.Value, Count: v.Count}
		}
		cols[i] = StatsColumn{
			Name:         c.Name,
			Unique:       c.Unique,
			UniqueCapped: c.UniqueCapped,
			Empty:        c.Empty,
			Top:          top,
		}
	}
	return StatsPayload{TotalRows: res.TotalRows, Columns: cols}, nil
}

// ----- Filter ---------------------------------------------------------------

type FilterRequest struct {
	Input       string  `json:"input"`
	Output      string  `json:"output"`
	Column      string  `json:"column"`
	Eq          string  `json:"eq"`
	EqSet       bool    `json:"eqSet"`
	Contains    string  `json:"contains"`
	ContainsSet bool    `json:"containsSet"`
	Gt          float64 `json:"gt"`
	GtSet       bool    `json:"gtSet"`
	Lt          float64 `json:"lt"`
	LtSet       bool    `json:"ltSet"`
	All         bool    `json:"all"`
	WithHeader  bool    `json:"withHeader"`
}

type FilterPayload struct {
	TotalRows int64 `json:"totalRows"`
	Matched   int64 `json:"matched"`
}

func (a *App) FilterCSV(req FilterRequest) (FilterPayload, error) {
	out, err := os.Create(req.Output)
	if err != nil {
		return FilterPayload{}, err
	}
	defer out.Close()

	opts := csvops.FilterOptions{
		Input:      req.Input,
		Output:     out,
		Column:     req.Column,
		All:        req.All,
		WithHeader: req.WithHeader,
		Progress:   a.emitProgress("filter"),
	}
	if req.EqSet {
		opts.Eq = &req.Eq
	}
	if req.ContainsSet {
		opts.Contains = &req.Contains
	}
	if req.GtSet {
		opts.Gt = &req.Gt
	}
	if req.LtSet {
		opts.Lt = &req.Lt
	}

	res, err := csvops.Filter(a.ctx, opts)
	if err != nil {
		return FilterPayload{}, err
	}
	return FilterPayload{TotalRows: res.TotalRows, Matched: res.Matched}, nil
}

// ----- Split ----------------------------------------------------------------

type SplitRequest struct {
	Input       string `json:"input"`
	OutputDir   string `json:"outputDir"`
	RowsPerFile int    `json:"rowsPerFile"`
	WithHeader  bool   `json:"withHeader"`
}

type SplitPayload struct {
	RowsProcessed int64 `json:"rowsProcessed"`
	FilesCreated  int   `json:"filesCreated"`
}

func (a *App) SplitCSV(req SplitRequest) (SplitPayload, error) {
	res, err := csvops.Split(a.ctx, csvops.SplitOptions{
		Input:       req.Input,
		OutputDir:   req.OutputDir,
		RowsPerFile: req.RowsPerFile,
		WithHeader:  req.WithHeader,
		Progress:    a.emitProgress("split"),
	})
	if err != nil {
		return SplitPayload{}, err
	}
	return SplitPayload{
		RowsProcessed: res.RowsProcessed,
		FilesCreated:  res.FilesCreated,
	}, nil
}

// ----- Dedupe ---------------------------------------------------------------

type DedupeRequest struct {
	Input         string `json:"input"`
	Output        string `json:"output"`
	KeyColumns    string `json:"keyColumns"` // comma-separated
	KeepLast      bool   `json:"keepLast"`
	CaseSensitive bool   `json:"caseSensitive"`
}

type DedupePayload struct {
	TotalRows  int64 `json:"totalRows"`
	UniqueRows int   `json:"uniqueRows"`
	Duplicates int   `json:"duplicates"`
}

func (a *App) DedupeCSV(req DedupeRequest) (DedupePayload, error) {
	keys := []string{}
	for _, k := range strings.Split(req.KeyColumns, ",") {
		k = strings.TrimSpace(k)
		if k != "" {
			keys = append(keys, k)
		}
	}
	res, err := csvops.Dedupe(a.ctx, csvops.DedupeOptions{
		Input:         req.Input,
		Output:        req.Output,
		KeyColumns:    keys,
		KeepLast:      req.KeepLast,
		CaseSensitive: req.CaseSensitive,
		Progress:      a.emitProgress("dedupe"),
	})
	if err != nil {
		return DedupePayload{}, err
	}
	return DedupePayload{
		TotalRows:  res.TotalRows,
		UniqueRows: res.UniqueRows,
		Duplicates: res.Duplicates,
	}, nil
}

// ----- Merge ----------------------------------------------------------------

type MergeRequest struct {
	InputDir   string `json:"inputDir"`
	Output     string `json:"output"`
	WithHeader bool   `json:"withHeader"`
}

type MergePayload struct {
	FilesProcessed int   `json:"filesProcessed"`
	RowsWritten    int64 `json:"rowsWritten"`
}

func (a *App) MergeCSV(req MergeRequest) (MergePayload, error) {
	out, err := os.Create(req.Output)
	if err != nil {
		return MergePayload{}, err
	}
	defer out.Close()

	res, err := csvops.Merge(a.ctx, csvops.MergeOptions{
		InputDir:   req.InputDir,
		Output:     out,
		WithHeader: req.WithHeader,
		SkipErrors: true,
		Progress:   a.emitProgress("merge"),
	})
	if err != nil {
		return MergePayload{}, err
	}
	return MergePayload{
		FilesProcessed: res.FilesProcessed,
		RowsWritten:    res.RowsWritten,
	}, nil
}

// ----- ToSQLite -------------------------------------------------------------

type ToSQLiteRequest struct {
	Input    string `json:"input"`
	DBPath   string `json:"dbPath"`
	Table    string `json:"table"`
	IfExists string `json:"ifExists"`
}

type ToSQLitePayload struct {
	Table        string `json:"table"`
	RowsImported int64  `json:"rowsImported"`
	Skipped      bool   `json:"skipped"`
}

func (a *App) ToSQLiteCSV(req ToSQLiteRequest) (ToSQLitePayload, error) {
	if req.IfExists == "" {
		req.IfExists = "replace"
	}
	res, err := csvops.ToSQLite(a.ctx, csvops.ToSQLiteOptions{
		Input:    req.Input,
		DBPath:   req.DBPath,
		Table:    req.Table,
		IfExists: csvops.IfExistsAction(req.IfExists),
		Progress: a.emitProgress("to-sqlite"),
	})
	if err != nil {
		return ToSQLitePayload{}, fmt.Errorf("%w", err)
	}
	return ToSQLitePayload{
		Table:        res.Table,
		RowsImported: res.RowsImported,
		Skipped:      res.Skipped,
	}, nil
}
