package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/maherelgamil/csvops/pkg/csvops"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	csvToSqliteInput     string
	csvToSqliteOutput    string
	csvToSqliteTable     string
	csvToSqliteDelimiter string
	csvToSqliteIfExists  string
)

var toSqliteCmd = &cobra.Command{
	Use:   "to-sqlite",
	Short: "Convert a CSV file into a SQLite database",
	RunE: func(cmd *cobra.Command, args []string) error {
		delim, err := parseDelimiter(csvToSqliteDelimiter)
		if err != nil {
			return err
		}

		var bar *progressbar.ProgressBar
		res, err := csvops.ToSQLite(context.Background(), csvops.ToSQLiteOptions{
			Input:     csvToSqliteInput,
			DBPath:    csvToSqliteOutput,
			Table:     csvToSqliteTable,
			Delimiter: delim,
			IfExists:  csvops.IfExistsAction(csvToSqliteIfExists),
			Progress: func(done, total int64) {
				if bar == nil {
					bar = progressbar.Default(total, "Converting")
				}
				_ = bar.Set64(done)
			},
		})
		if err != nil {
			return err
		}

		if res.Skipped {
			fmt.Printf("⚠️  Table %q already exists, skipped.\n", res.Table)
			return nil
		}
		dbPath, _ := filepath.Abs(csvToSqliteOutput)
		fmt.Printf("\n✅ Imported %d rows into %s\n", res.RowsImported, dbPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(toSqliteCmd)

	toSqliteCmd.Flags().StringVar(&csvToSqliteInput, "input", "", "Input CSV file path (required)")
	toSqliteCmd.Flags().StringVar(&csvToSqliteOutput, "output", "", "Output SQLite DB file path (required)")
	toSqliteCmd.Flags().StringVar(&csvToSqliteTable, "table", "", "Table name to create in SQLite (defaults to filename)")
	toSqliteCmd.Flags().StringVar(&csvToSqliteDelimiter, "delimiter", ",", "CSV delimiter character")
	toSqliteCmd.Flags().StringVar(&csvToSqliteIfExists, "if-exists", "replace", "Action if table exists: replace | skip | append | fail")

	_ = toSqliteCmd.MarkFlagRequired("input")
	_ = toSqliteCmd.MarkFlagRequired("output")
}
