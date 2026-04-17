# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build / Run / Lint

```bash
go build -o csvops          # local build
./csvops --help

go fmt ./...
go vet ./...
```

Cross-platform release builds: `scripts/build.sh <version>` (darwin/linux/windows, amd64/arm64). Production releases go through `.goreleaser.yaml`, which also publishes a Homebrew formula to `maherelgamil/homebrew-tap`.

There is no test suite yet. When adding tests, use `go test ./...` and prefer golden-file tests in `cmd/` (sample CSV → expected output).

## Architecture

Single-binary Cobra CLI. `main.go` calls `cmd.SetVersion(...)` then `cmd.Execute()`; everything else lives in `cmd/`.

**One file per subcommand.** Each file in `cmd/` (`split.go`, `dedupe.go`, `filter.go`, `stats.go`, `preview.go`, `merge.go`, `to-sqlite.go`) defines:
- package-level flag vars (prefixed with the command name to avoid collisions across the shared `cmd` package — e.g. `dedupeInput`, `filterInput`),
- a `*cobra.Command`,
- an `init()` that registers flags and calls `rootCmd.AddCommand(...)`.

`rootCmd` lives in `cmd/root.go`. To add a command, create a new file in `cmd/` following the same shape — no central registration table to update.

**Common processing pattern** used by most commands:
1. Pre-scan the input file once with `csv.NewReader` just to count rows for a `progressbar.Default(...)`.
2. Re-open and stream rows via `csv.NewReader.Read()` in a loop.
3. Write through `csv.NewWriter` (or `database/sql` for `to-sqlite`).

This double-open is intentional for the progress bar; preserve it when modifying existing commands unless you're changing the UX deliberately.

**SQLite driver**: `to-sqlite.go` uses `modernc.org/sqlite` (pure-Go, no CGO) — keep cross-compilation working, don't switch to `mattn/go-sqlite3`.

**Per-command docs** live in `docs/commands/*.md` and are referenced from the README. When you add or change a command's flags, update the matching doc file.

## Versioning gotcha

`scripts/build.sh` and `.goreleaser.yaml` both inject the version via `-ldflags "-X main.version=..."`, but `main.go` doesn't declare a `version` var — it calls `cmd.SetVersion("v0.0.1")` with a hardcoded string. The ldflag is currently a no-op. If you touch the version wiring, fix both sides together (either expose `main.version` and pass it into `cmd.SetVersion`, or change the ldflag target to `github.com/maherelgamil/csvops/cmd.version`).

## Conventions

- All commands use Cobra; keep that consistent.
- UX libs already in use: `schollz/progressbar/v3` for long operations, `olekukonko/tablewriter` for terminal tables (`stats`, `preview`).
- User-facing errors are printed with a leading `❌`, info with `✅`/`📊`/`⚠️`. Match the existing style if you add output.
- Most commands today use `Run` (not `RunE`) and print errors instead of returning them — so failures exit 0. New commands should prefer `RunE` so the process exits non-zero on error.
