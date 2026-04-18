# csvops-desktop

A desktop UI for [csvops](https://github.com/maherelgamil/csvops), built with [Wails v2](https://wails.io) (Go backend + React/TypeScript frontend). Reuses 100% of the CSV logic from `pkg/csvops/` in the parent repo.

## Develop

Requirements:
- Go 1.24+ (older versions produce broken binaries on macOS Sequoia/Tahoe — missing `LC_UUID`)
- Node.js 20+
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

```bash
cd desktop
wails dev          # hot-reload dev mode (opens a window)
wails build        # produces a packaged .app / .exe under build/bin/
```

The repo uses a Go workspace (`go.work` at the root) so the `desktop` module imports the local `pkg/csvops` library. No `replace` editing needed for day-to-day work.

## Release

Push a tag matching `desktop-v*` to trigger `.github/workflows/desktop-release.yml`. It builds:
- `csvops-desktop-macos-universal.zip` (arm64 + amd64 fat binary)
- `csvops-desktop-windows-amd64.zip`

The workflow uploads them to a new GitHub release. Tag CLI releases as `v*` (separate cadence — UI-only changes shouldn't force a CLI version bump).

## Architecture

- `app.go` — Wails-bound `App` struct. Each public method becomes a TypeScript function in `frontend/wailsjs/go/main/App.ts`.
- `main.go` — Wails app boot.
- `frontend/src/App.tsx` — single-file React UI with one component per operation tab.
- Progress events: backend ops emit `runtime.EventsEmit(ctx, "progress", ...)`; the frontend subscribes via `EventsOn` and routes by op tag.
