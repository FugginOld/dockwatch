# Architecture

## Technology Stack

- Go 1.25+ for core runtime and CLI
- Docker SDK and Docker CLI libraries for container interaction
- Cobra and Pflag for command and flag parsing
- Viper for environment variable binding and defaults
- robfig/cron for scheduled update loops
- Logrus for structured logging

## Repository Structure

- cmd: application entry and runtime flow
- internal/actions: update and sanity logic
- internal/flags: flag registration, alias processing, logging setup
- pkg/api: authenticated HTTP API and handlers
- pkg/container: Docker client abstraction and operations
- pkg/filters: container target selection and label filtering
- pkg/lifecycle: hook execution model
- pkg/registry: digest and manifest checks, auth helpers
- pkg/session: per-run report assembly
- pkg/sorter: dependency-aware container ordering

## Architecture-specific installs

Dockwatch ships multi-arch manifests, plus explicit per-arch tags for pinning.

- Multi-arch default: fugginold/dockwatch:latest
- AMD64: fugginold/dockwatch:amd64-latest
- i386: fugginold/dockwatch:i386-latest
- ARMv6/v7: fugginold/dockwatch:armhf-latest
- ARM64: fugginold/dockwatch:arm64v8-latest

Use explicit tags when deterministic architecture pinning is required by your deployment tooling.

## Execution Flow

1. main initializes logger level and executes Cobra root command.
2. PreRun resolves flags and env, configures logging, creates Docker client.
3. Run chooses one-shot update, interactive shell, or daemon mode.
4. Daemon mode runs an immediate update check before scheduled operation begins.
5. Daemon mode creates a schedule controller and optional HTTP API handlers.
6. Scheduler and API both call the same update pipeline with lock protection.
7. Update pipeline computes stale containers, orders dependencies, updates safely.

## Dependency Sorting

Dockwatch sorts dependent containers before update operations:

- Stale targets are sorted by dependency graph in sorter package.
- Stop operations occur in reverse dependency order.
- Start/recreate operations occur in forward dependency order.

This avoids breaking linked service chains during rolling updates.

## Internal Packages

- internal/actions: high-level update orchestration and runtime checks
- internal/flags: all user-facing flags and aliases, env mapping, defaults
- internal/meta: version injection and metadata
- internal/util: utility helpers used by multiple internal flows

## Design Patterns

- Interface-driven Docker client boundary for easier test mocking
- Controller-based schedule abstraction with explicit Set and Stop lifecycle
- Channel lock (per-handler, not global) to ensure one update run at a time across scheduler and API
- Local `http.ServeMux` per API instance — handlers are not registered on the global default mux, allowing multiple independent API instances and avoiding test cross-contamination
- `RunFlags` struct returned by `ReadFlags` replaces positional multi-value returns, preventing accidental argument swap
- `NewClient` returns `(Client, error)` so callers control error handling rather than taking a fatal exit inside library code
- Separation of concerns between filtering, updating, sorting, and reporting
- Configuration normalization through aliases like cron->schedule and force-update->run-once
- Container staleness and link-restarting state are unexported fields accessed only through typed getter/setter methods (`IsStale`, `SetStale`, `IsLinkedToRestarting`, `SetLinkedToRestarting`)

## Security Notes

- HTTP API requires bearer token and returns unauthorized when missing or invalid.
- Token must be set through environment variable before API startup.
- Trace logging can expose sensitive values and should be used cautiously.
