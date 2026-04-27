# Changelog

<!-- BEGIN UNRELEASED -->
## [Unreleased]

_Last updated: 2026-04-15 (UTC)._

### Changed

- Changelog automation enabled for `main`.

---
<!-- END UNRELEASED -->

## [0.3.23] - 2026-04-14

### Fixed

- Avoid entering interactive shell mode in non-interactive containers (prevents restart loops when stdin is not a real TTY).

### Added

- Added regression tests for interactive input detection.

### Changed

- Updated interactive shell documentation to clarify how to start Dockwatch with a `dockwatch>` prompt.

---

## [0.3.22] - 2026-04-14

### Fixed

- Stamp release builds with the actual version tag instead of the stale fallback version.

---

## [0.3.21] - 2026-04-14

### Fixed

- Updated `.gitignore` to avoid committing local/assistant settings.

---

## [0.3.20] - 2026-04-14

### Fixed

- Updated `.gitignore` to exclude an additional generated/local file.

---

## [0.3.19] - 2026-04-14

### Changed

- Release workflow updates.

---

## [0.3.18] - 2026-04-14

### Fixed

- Bumped `goreleaser/goreleaser-action` to `v6` to support GoReleaser v2 config.

---

## [0.3.17] - 2026-04-13

### Fixed

- Auto-release pipeline and container test fixes.

---

## [0.3.16] - 2026-04-13

### Fixed

- Updated `scripts/test-container.sh` for the current image name.

---

## [0.3.15] - 2026-04-13

### Fixed

- Ensured `CGO_ENABLED=0` is set during release builds.

---

## [0.3.14] - 2026-04-13

### Fixed

- Fixed release workflow path issues.

---

## [0.3.13] - 2026-04-13

### Changed

- Updated the `docker-dev` workflow.

---

## [0.3.12] - 2026-04-13

### Changed

- Updated the `docker-dev` workflow.

---

## [0.3.11] - 2026-04-13

### Changed

- Removed the devcontainer and fixed workflow secrets handling.

---

## [0.3.10] - 2026-04-13

### Changed

- Default bymp.

---

## [0.3.9] - 2026-04-13

### Changed

- chore(deps): bump `github.com/docker/cli` in the go-dependencies group.

---

## [0.3.8] - 2026-04-13

### Changed

- chore(deps): bump `DavidAnson/markdownlint-cli2-action` from 17 to 23.

---

## [0.3.7] - 2026-04-13

### Changed

- chore(deps): bump `codecov/codecov-action` from 3 to 6.

---

## [0.3.6] - 2026-04-13

### Changed

- chore(deps): bump `actions/setup-python` from 5 to 6.

---

## [0.3.5] - 2026-04-13

### Changed

- chore(deps): bump `tj-actions/changed-files` from 46 to 47.

---

## [0.3.4] - 2026-04-13

### Changed

- chore(deps): bump `github/codeql-action` from 3 to 4.

---

## [0.3.3] - 2026-04-12

### Fixed

- Replaced deprecated Docker API types and added a missing vulnerability allowlist entry.

---

## [0.3.2] - 2026-04-12

### Fixed

- Resolved GoReleaser v2 deprecation warnings.

---

## [0.3.1] - 2026-04-12

### Fixed

- Restored GoReleaser builds/archives to resume asset generation.

---

## [0.3.0] - 2026-04-12

### Added

- Added release configuration to `.goreleaser.yml`.

---

## [0.2.3] - 2026-04-12

### Removed

- Removed accidentally committed `.claude/worktrees` directory.

---

## [0.2.2] - 2026-04-12

### Added

- Added Claude worktrees metadata.

---

## [0.2.1] - 2026-04-12

### Fixed

- Addressed code review issues across security, reliability, and correctness.

---

## [0.2.0] - 2026-04-11

### Added

- Added an interactive shell interface that automatically activates when Dockwatch operates within a TTY environment.
- Supported runtime `schedule` and `update` commands directly inside the interactive `dockwatch>` shell.

### Fixed

- Resolved duplicate GitHub Action pipeline triggers on version tags in `ci.yml`.
- Standardized `meta.Version` testing assertions to map accurately against dynamic module versions.
- Fixed GHCR Push `403 Forbidden` errors by swapping out an expired `GHCR_PAT` for the native secure `GITHUB_TOKEN`.
- Resolved `docker manifest create` failures during release by modifying `release.yml` to safely parse and merge `config.json` instead of erasing existing registry authentication tokens.
- Fixed a final `Refresh pkg.go.dev` documentation pipeline failure by bypassing a third-party GitHub Action that was incompatible with uppercase organization names.

---

## [0.1.9] - 2026-04-07

### Changed

- Updated `scripts/push-dockerhub.sh` to support multi-architecture Buildx publishing by default with configurable platforms and builder name.
- Added `--platforms` and `--builder` options to release script for reproducible Docker publish workflow.

### Fixed

- Added local release and preview artifact patterns to `.gitignore` to prevent noisy/unintended git status changes during release operations.

---

## [0.1.8] - 2026-04-07

### Fixed

- Applied TLS skip-verify handling to registry auth HTTP clients, not only digest checks.
- Shared a single HTTP client via `sync.Once` and reduced duplicate TLS warning logs.
- Added tests for TLS-skip verification helper behavior.

### Removed

- Removed unused `gopher-watchtower.png` image asset.

---

## [0.1.7] - 2026-04-07

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

## [0.1.5] - 2026-04-06

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

## [0.1.4] - 2026-04-06

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

## [0.1.3] - 2026-04-06

- Published refreshed Docker Hub images for `fugginold/dockwatch`.
- Added multi-architecture image tags for `amd64`, `386`, `arm/v6`, and `arm64`.
- Updated the shared `0.1.3` and `latest` tags to point at the current manifest set.

<!-- Footer links -->
[Unreleased]: https://github.com/FugginOld/dockwatch/compare/v0.3.23...HEAD
[0.3.23]: https://github.com/FugginOld/dockwatch/compare/v0.3.22...v0.3.23
[0.3.22]: https://github.com/FugginOld/dockwatch/compare/v0.3.21...v0.3.22
[0.3.21]: https://github.com/FugginOld/dockwatch/compare/v0.3.20...v0.3.21
[0.3.20]: https://github.com/FugginOld/dockwatch/compare/v0.3.19...v0.3.20
[0.3.19]: https://github.com/FugginOld/dockwatch/compare/v0.3.18...v0.3.19
[0.3.18]: https://github.com/FugginOld/dockwatch/compare/v0.3.17...v0.3.18
[0.3.17]: https://github.com/FugginOld/dockwatch/compare/v0.3.16...v0.3.17
[0.3.16]: https://github.com/FugginOld/dockwatch/compare/v0.3.15...v0.3.16
[0.3.15]: https://github.com/FugginOld/dockwatch/compare/v0.3.14...v0.3.15
[0.3.14]: https://github.com/FugginOld/dockwatch/compare/v0.3.13...v0.3.14
[0.3.13]: https://github.com/FugginOld/dockwatch/compare/v0.3.12...v0.3.13
[0.3.12]: https://github.com/FugginOld/dockwatch/compare/v0.3.11...v0.3.12
[0.3.11]: https://github.com/FugginOld/dockwatch/compare/v0.3.10...v0.3.11
[0.3.10]: https://github.com/FugginOld/dockwatch/compare/v0.3.9...v0.3.10
[0.3.9]: https://github.com/FugginOld/dockwatch/compare/v0.3.8...v0.3.9
[0.3.8]: https://github.com/FugginOld/dockwatch/compare/v0.3.7...v0.3.8
[0.3.7]: https://github.com/FugginOld/dockwatch/compare/v0.3.6...v0.3.7
[0.3.6]: https://github.com/FugginOld/dockwatch/compare/v0.3.5...v0.3.6
[0.3.5]: https://github.com/FugginOld/dockwatch/compare/v0.3.4...v0.3.5
[0.3.4]: https://github.com/FugginOld/dockwatch/compare/v0.3.3...v0.3.4
[0.3.3]: https://github.com/FugginOld/dockwatch/compare/v0.3.2...v0.3.3
[0.3.2]: https://github.com/FugginOld/dockwatch/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/FugginOld/dockwatch/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/FugginOld/dockwatch/compare/v0.2.3...v0.3.0
[0.2.3]: https://github.com/FugginOld/dockwatch/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/FugginOld/dockwatch/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/FugginOld/dockwatch/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/FugginOld/dockwatch/compare/v0.1.9...v0.2.0
[0.1.9]: https://github.com/FugginOld/dockwatch/compare/v0.1.8...v0.1.9
[0.1.8]: https://github.com/FugginOld/dockwatch/compare/v0.1.7...v0.1.8
[0.1.7]: https://github.com/FugginOld/dockwatch/compare/v0.1.5...v0.1.7
[0.1.5]: https://github.com/FugginOld/dockwatch/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/FugginOld/dockwatch/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/FugginOld/dockwatch/releases/tag/v0.1.3
