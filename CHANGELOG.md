# Changelog

All notable changes to this project are documented here.

## [v0.4.0] - 2026-04-18

### ✨ Features
- **New library**: all CSV operations are now exposed as Go functions under `github.com/maherelgamil/csvops/pkg/csvops` — importable by other Go programs (including the upcoming desktop app). Each operation takes a typed `Options` struct and returns a typed `Result`.
- Every operation supports `context.Context` cancellation and an optional `Progress` callback.
- **`merge`** (library only): new `InputFiles` option for explicit-order merging, plus `SkipErrors` + `OnWarn` for per-file non-fatal errors.
- **`to-sqlite`** (library only): typed `IfExistsAction` enum (`Replace`, `Skip`, `Append`, `Fail`); `ToSQLiteResult.Skipped` reflects honored `Skip` mode.
- Test suite extended to ~35 tests in `pkg/csvops/`, including regression coverage for deterministic dedupe output order and the adversarial-table-name SQL-injection fix.

### 🛠 Internal
- CLI commands in `cmd/` are now thin wrappers over the library.
- `io.Writer`-based output in `Filter` and `Merge` so callers can stream results to memory.

### 🧩 Compatibility
- CLI flags and behavior unchanged. No breaking changes for end users.

## [v0.3.0] - 2026-04-18

### 🔒 Security
- **`to-sqlite`**: SQL identifiers (`--table` and CSV header column names) are now properly quoted, preventing SQL injection via crafted table names or headers.

### ✨ Features
- **`filter`**: new `--all` flag to require ALL conditions to match (AND) instead of the default ANY (OR).
- **`stats`**: new `--max-unique` flag (default `100000`) to cap per-column unique tracking and bound memory on high-cardinality columns.
- **`split`**, **`to-sqlite`**: `--delimiter` now accepts `\t` as a tab shortcut.
- **`to-sqlite`**: `--if-exists` now actually implements `skip`, `append`, and `fail` modes (previously only `replace` worked despite the docs).

### 🐛 Fixes
- **`dedupe`**: output rows are now emitted in original file order (previous map iteration produced non-deterministic output between runs).
- **`dedupe`**: in-place overwrite (`--input` == `--output`) now closes the output before renaming, fixing Windows behavior.
- **`dedupe`**: fixed nil-pointer panic when the row-count pre-scan failed to open the file.
- **`split`**, **`to-sqlite`**: validate `--delimiter` is exactly one character (previously panicked on empty string).
- **`filter`**: `--eq=""` now correctly matches empty string values (uses `Flags().Changed()` instead of zero-value comparison).
- **`to-sqlite`**: transaction is rolled back on row-insert errors instead of leaking.
- **All commands**: converted to `RunE` so failures exit with non-zero status (previously printed errors but exited 0).
- **Build**: `-ldflags "-X main.version=..."` injection now actually takes effect (previously a no-op because `main.version` didn't exist).

### 🚜 Performance
- **`merge`**: streams rows row-by-row instead of `ReadAll`'ing each input file into memory.
- **`dedupe`**: `keep-first` mode now streams output instead of buffering all kept rows.

### 💥 Breaking
- **`filter`**: removed `--enable-gt` and `--enable-lt` flags. The `--gt` and `--lt` flags now activate automatically when set.

## [v0.2.1] - earlier

### ✨ New Features
- Added `to-sqlite` command to convert CSV files to SQLite databases
- Auto-generate valid table names from filenames
- Added progress bar while inserting rows

### 🛠 Improvements
- Cleaned up table naming logic
- Docs added for all commands
