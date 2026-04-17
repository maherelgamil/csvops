# csvops

[![Release](https://img.shields.io/github/v/release/maherelgamil/csvops)](https://github.com/maherelgamil/csvops/releases)
[![License](https://img.shields.io/github/license/maherelgamil/csvops)](LICENSE)
[![Stars](https://img.shields.io/github/stars/maherelgamil/csvops?style=social)](https://github.com/maherelgamil/csvops/stargazers)

A fast, modular command-line toolkit for working with CSV files. Built in Go — single static binary, no runtime dependencies.

```bash
csvops split    --input big.csv     --rows 10000 --output-dir ./parts
csvops dedupe   --input users.csv   --output clean.csv --key email
csvops filter   --input users.csv   --column country --eq Egypt
csvops stats    --input data.csv
csvops preview  --input data.csv    --rows 10
csvops merge    --input-dir ./parts --output all.csv
csvops to-sqlite --input data.csv   --output data.db
```

## Install

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

## Commands

| Command     | Purpose                                           |
| ----------- | ------------------------------------------------- |
| `split`     | Split a large CSV into smaller chunks             |
| `merge`     | Combine all CSV files in a directory into one     |
| `dedupe`    | Remove duplicate rows by one or more key columns  |
| `filter`    | Keep rows matching `eq` / `contains` / `gt` / `lt` |
| `stats`     | Row counts, unique values, empty cells, top values |
| `preview`   | Pretty-print the first N rows as a table          |
| `to-sqlite` | Import a CSV into a SQLite database               |

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

## Contributing

PRs welcome — see [CONTRIBUTING.md](CONTRIBUTING.md).

```bash
go build -o csvops .
go fmt ./...
go vet ./...
```

## License

MIT © [Maher El Gamil](LICENSE)
