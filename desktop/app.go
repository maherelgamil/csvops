package main

import (
	"context"

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

// OpenCSVFile shows a native file picker filtered to .csv and returns the
// chosen absolute path. Empty string means the user cancelled.
func (a *App) OpenCSVFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open CSV file",
		Filters: []runtime.FileFilter{
			{DisplayName: "CSV files (*.csv)", Pattern: "*.csv"},
			{DisplayName: "All files", Pattern: "*.*"},
		},
	})
}

// PreviewCSV reads the first `rows` data rows of `path` and returns a
// flattened result the frontend can render directly.
func (a *App) PreviewCSV(path string, rows int, noHeader bool) (PreviewPayload, error) {
	res, err := csvops.Preview(a.ctx, csvops.PreviewOptions{
		Input:    path,
		Rows:     rows,
		NoHeader: noHeader,
	})
	if err != nil {
		return PreviewPayload{}, err
	}
	return PreviewPayload{
		Headers: res.Headers,
		Rows:    res.Rows,
	}, nil
}

// PreviewPayload is what the frontend receives. Wails generates a matching
// TypeScript interface from this Go struct.
type PreviewPayload struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}
