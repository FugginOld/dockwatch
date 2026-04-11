# Changelog

## 0.2.0 - 2026-04-11

### Added

- Added an interactive shell interface that automatically activates when Dockwatch operates within a TTY environment.
- Supported runtime `schedule` and `update` commands directly inside the interactive `dockwatch>` shell.

---

## 0.1.9 - 2026-04-07

### Changed

- Updated `scripts/push-dockerhub.sh` to support multi-architecture Buildx publishing by default with configurable platforms and builder name.
- Added `--platforms` and `--builder` options to release script for reproducible Docker publish workflow.

### Fixed

- Added local release and preview artifact patterns to `.gitignore` to prevent noisy/unintended git status changes during release operations.

---

## 0.1.8 - 2026-04-07

### Fixed

- Applied TLS skip-verify handling to registry auth HTTP clients, not only digest checks.
- Shared a single HTTP client via `sync.Once` and reduced duplicate TLS warning logs.
- Added tests for TLS-skip verification helper behavior.

### Removed

- Removed unused `gopher-watchtower.png` image asset.

---

## 0.1.7 - 2026-04-07

### Added

- Added `CODE_OF_CONDUCT.md` based on Contributor Covenant with actionable reporting contact.
- Added `docs/ARCHITECTURE.md`, `docs/CONFIGURATION.md`, `docs/API.md`, `docs/DEVELOPMENT.md`, and `docs/TROUBLESHOOTING.md`.

### Changed

- Updated `SECURITY.md` to use GitHub private vulnerability disclosure flow.
- Applied modernization refactors across core runtime and registry packages.
- Rewrote `README.md` to focused sections for quick start, install flow, and docs references.

### Fixed

- Added zero-time fallback in container created-time sorting and improved error return on start failures.
- Ran `go mod tidy` updates, including `golang.org/x/text` indirect dependency metadata.
- Improved docs validation and markdown lint handling in CI.

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
