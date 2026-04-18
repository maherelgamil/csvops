# csvops

[![Release](https://img.shields.io/github/v/release/maherelgamil/csvops)](https://github.com/maherelgamil/csvops/releases)
[![License](https://img.shields.io/github/license/maherelgamil/csvops)](LICENSE)
[![Stars](https://img.shields.io/github/stars/maherelgamil/csvops?style=social)](https://github.com/maherelgamil/csvops/stargazers)

A fast CSV toolkit shipped three ways:

- 🖥️ **CLI** — single static Go binary for scripts and pipelines.
- 📦 **Library** — `github.com/maherelgamil/csvops/pkg/csvops` for embedding in Go programs.
- 🪟 **Desktop app** — paginated table view + one-click operations, built with Wails.

All three share the same engine: streaming CSV ops with progress callbacks and `context.Context` cancellation.

```bash
csvops split    --input big.csv     --rows 10000 --output-dir ./parts
csvops dedupe   --input users.csv   --output clean.csv --key email
csvops filter   --input users.csv   --column country --eq Egypt
csvops stats    --input data.csv
csvops preview  --input data.csv    --rows 10
csvops merge    --input-dir ./parts --output all.csv
csvops to-sqlite --input data.csv   --output data.db
```

## Install (CLI)

### Homebrew

```bash
brew install maherelgamil/tap/csvops
```

### From a release binary

Grab the archive for your OS/arch from [Releases](https://github.com/maherelgamil/csvops/releases) and put `csvops` on your `PATH`.

### From source

```bash
git clone https://github.com/maherelgamil/csvops.git
cd csvops
go build -o csvops .
```

## Desktop app

Download the latest from [Releases](https://github.com/maherelgamil/csvops/releases) (look for tags starting with `desktop-v`):

- `csvops-desktop-macos-universal.zip` — macOS arm64 + amd64 fat binary
- `csvops-desktop-windows-amd64.zip` — Windows x64

> **macOS first launch**: the app is currently ad-hoc signed, so Gatekeeper will block it. Open **System Settings → Privacy & Security**, scroll down, and click **"Open Anyway"** next to the warning. Apple Developer ID signing is on the roadmap.

What it does:
- Open or **drag a CSV** onto the window.
- Browse the whole file in a paginated table (50–1000 rows/page).
- Click any column header for instant **per-column stats** (unique, empty, top values).
- Run **Filter / Dedupe / Split / Export to SQLite / Merge** from the Actions menu — output is auto-suggested next to the input. Success banners include a Reveal-in-Finder button.

Build it yourself: see [`desktop/README.md`](./desktop/README.md).

## Library

Use the same CSV engine in your own Go program:

```go
import "github.com/maherelgamil/csvops/pkg/csvops"

ctx := context.Background()
res, err := csvops.Filter(ctx, csvops.FilterOptions{
    Input:      "users.csv",
    Output:     os.Stdout,
    Column:     "country",
    Eq:         strPtr("Egypt"),
    WithHeader: true,
    Progress:   func(done, total int64) { /* update UI */ },
})
```

Each operation (`Split`, `Dedupe`, `Filter`, `Merge`, `Stats`, `Preview`, `ToSQLite`) takes a typed `Options` struct and returns a typed `Result`. See [`pkg/csvops/`](./pkg/csvops/) and the test files for full examples.

## Commands

| Command     | Purpose                                            |
| ----------- | -------------------------------------------------- |
| `split`     | Split a large CSV into smaller chunks              |
| `merge`     | Combine all CSV files in a directory into one      |
| `dedupe`    | Remove duplicate rows by one or more key columns   |
| `filter`    | Keep rows matching `eq` / `contains` / `gt` / `lt` |
| `stats`     | Row counts, unique values, empty cells, top values |
| `preview`   | Pretty-print the first N rows as a table           |
| `to-sqlite` | Import a CSV into a SQLite database                |

Run `csvops <command> --help` for the full flag list, or see [`docs/commands/`](./docs/commands).

### `split`

Streams the input file and writes chunks of `--rows` lines to `--output-dir`.

```bash
csvops split --input big.csv --rows 10000 --output-dir ./parts --with-header --delimiter ","
```

`--delimiter` accepts a single character or `\t` for tab.

### `merge`

Reads every `*.csv` file in `--input-dir` (sorted by name) and streams them into one output file. The header from the first file is written once when `--with-header` is set.

```bash
csvops merge --input-dir ./parts --output merged.csv
```

### `dedupe`

```bash
csvops dedupe --input users.csv --output clean.csv --key email
csvops dedupe --input users.csv --output clean.csv --key first_name,last_name --case-sensitive
csvops dedupe --input users.csv --output clean.csv --key email --keep-last
```

- Output preserves the original file row order.
- Case-insensitive by default; pass `--case-sensitive` to compare exactly.
- `--keep-last` retains the last occurrence (default keeps the first).

### `filter`

```bash
csvops filter --input users.csv --column country --eq Egypt
csvops filter --input users.csv --column name    --contains "ali"
csvops filter --input scores.csv --column score  --gt 80 --lt 100 --all
```

- Conditions combine with **OR** by default; pass `--all` for **AND**.
- A flag is only applied when explicitly set, so `--eq=""` matches empty values.
- Writes to `--output` if provided, otherwise to stdout.

### `stats`

```bash
csvops stats --input data.csv
csvops stats --input data.csv --max-unique 5000
```

Prints row/column counts and a per-column table with unique value count, empty cell count, and top 3 values. `--max-unique` (default `100000`) bounds memory on high-cardinality columns; columns that hit the cap are reported as `>=N (capped)`.

### `preview`

```bash
csvops preview --input data.csv --rows 20
csvops preview --input data.csv --no-header
```

### `to-sqlite`

```bash
csvops to-sqlite --input data.csv --output data.db
csvops to-sqlite --input data.csv --output data.db --table users --if-exists append
```

- Pure-Go SQLite (`modernc.org/sqlite`), no CGO required.
- All columns created as `TEXT`. SQL identifiers are quoted, so column names and table names with spaces or special characters are safe.
- Default table name is derived from the input filename.
- `--if-exists` modes: `replace` (default, drops then re-creates), `append` (insert into existing), `skip` (no-op if table exists), `fail` (error if table exists).

## Repo layout

```
cmd/                CLI commands (Cobra) — thin wrappers over pkg/csvops
pkg/csvops/         The CSV engine: Split, Dedupe, Filter, Merge, Stats, Preview, ToSQLite
desktop/            Wails React+TS desktop app, imports pkg/csvops
docs/commands/      Per-command CLI documentation
```

## Contributing

PRs welcome — see [CONTRIBUTING.md](CONTRIBUTING.md).

```bash
go build -o csvops .
go fmt ./...
go vet ./...
go test ./...
```

## License

MIT © [Maher El Gamil](LICENSE)
