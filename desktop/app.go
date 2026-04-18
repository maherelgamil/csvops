package main

import (
	"context"
	"os"

	"github.com/maherelgamil/csvops/pkg/csvops"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails-bound struct. Each public method becomes callable from the
// frontend via the generated TypeScript bindings.
type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ----- File pickers --------------------------------------------------------

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

// FilterRequest is intentionally JSON-friendly: each *Set boolean tells the
// backend which condition to apply. This keeps the wire format simple instead
// of wrestling with optional/null pointers across the JS bridge.
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
