# Changelog

All notable changes to this project are documented here.

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
