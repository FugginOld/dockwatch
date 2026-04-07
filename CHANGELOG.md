# Changelog

## Unreleased - 2026-04-06

### Docs

- Added `docs/ARCHITECTURE.md` covering technology stack, repository structure, execution flow, design patterns, and security notes.
- Added `docs/CONFIGURATION.md` documenting all environment variables, CLI flags, and scheduling examples.
- Added `docs/API.md` documenting all HTTP API endpoints, auth requirements, and enable-API-mode examples.
- Added `docs/DEVELOPMENT.md` covering local setup, build, test scripts, and debug tips.
- Added `docs/TROUBLESHOOTING.md` covering Docker socket issues, API version mismatch, permission errors, and diagnostics.
- Rewrote `README.md` to focused sections: What It Does, Quick Start, Install & Usage (Debian + Runtime API), and docs index.

---

## 0.1.5 - 2026-04-06

### Added

- HTTP API endpoint `GET /v1/schedule` returns current schedule and next run time.
- HTTP API endpoint `POST /v1/schedule` and `PUT /v1/schedule` allow updating the active schedule at runtime without restarting Dockwatch.
- Schedule controller abstraction in `cmd/root.go` with explicit `Set`, `Current`, `NextRun`, and `Stop` lifecycle methods.
- `--cron` flag as alias for `--schedule`.
- `--force-update` flag as alias for `--run-once` for immediate one-shot update.
- `scripts/install-dockwatch.sh` automated install script.

### Changed

- Schedule endpoint is only registered when periodic polling mode is active.
- Update channel lock prevents concurrent update sessions between scheduler and API triggers.
- CI release workflow: added `contents:write` permission to build job.

### Fixed

- `go mod tidy` to mark `golang.org/x/text` as indirect dependency.

---

## 0.1.4 - 2026-04-06

### Removed

- Entire `pkg/notifications` package and all notification-related code.
- `pkg/types/notifier.go` notifier interface.
- `tplprev` tooling (`main.go`, `main_wasm.go`).
- `docs/notifications.md` and `docs/template-preview.md`.

### Changed

- `cmd/root.go`: removed notifier wiring; simplified `runUpdatesWithNotifications` to `runUpdates`.
- `internal/flags/flags.go`: removed no-op notification flag registration.
- `mkdocs.yml`: removed notification nav entries.

---

## 0.1.3 - 2026-04-06

- Published refreshed Docker Hub images for `fugginold/dockwatch`.
- Added multi-architecture image tags for `amd64`, `386`, `arm/v6`, and `arm64`.
- Updated the shared `0.1.3` and `latest` tags to point at the current manifest set.
