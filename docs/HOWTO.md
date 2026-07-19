# How-to guide

Practical recipes for running and operating Dockwatch. For how the code is put
together see [ARCHITECTURE.md](ARCHITECTURE.md).

## Contents

- [Build & install](#build--install)
- [Running](#running)
- [Choosing which containers to watch](#choosing-which-containers-to-watch)
- [HTTP API](#http-api)
- [Prometheus metrics](#prometheus-metrics)
- [Lifecycle hooks & restart behavior](#lifecycle-hooks--restart-behavior)
- [Connecting to a remote Docker host](#connecting-to-a-remote-docker-host)
- [Configuration reference](#configuration-reference)
- [Development](#development)

## Build & install

Dockwatch builds from source into a self-contained image — nothing is published
to a registry.

Docker Compose (builds and runs a daily updater):

```bash
git clone https://github.com/fugginold/dockwatch.git
cd dockwatch
docker compose up -d --build
```

Docker directly:

```bash
docker build -t dockwatch:latest .

docker run -d \
  --name dockwatch \
  --restart unless-stopped \
  -v /var/run/docker.sock:/var/run/docker.sock \
  dockwatch:latest --interval 300
```

Local binary:

```bash
go build -o dockwatch .
./dockwatch --help
```

Dockwatch needs access to the Docker socket (`/var/run/docker.sock`) to inspect
and recreate containers.

## Running

By default Dockwatch runs as a daemon and checks every 24 hours. The cadence is
controlled two ways:

```bash
# Fixed interval, in seconds
dockwatch --interval 300

# Cron schedule (six-field cron; also available as --cron)
dockwatch --schedule "@every 6h"
dockwatch --schedule "0 0 4 * * *"     # every day at 04:00
```

One-shot — check once and exit (useful in a cron job on the host or CI):

```bash
dockwatch --run-once        # or --force-update
```

Monitor only — report what would update without pulling or recreating anything:

```bash
dockwatch --monitor-only
```

Interactive shell — when attached to a TTY, Dockwatch opens a `dockwatch>`
prompt:

```bash
docker run --rm -it \
  -v /var/run/docker.sock:/var/run/docker.sock \
  dockwatch:latest
```

Shell commands: `update` (trigger an immediate scan), `schedule <cron>`
(set/view the schedule), `help`, and `exit` / `quit`. Detach from a detached
`docker attach` session without stopping the container with `Ctrl-p` `Ctrl-q`.

## Choosing which containers to watch

By default every running container is watched. Narrow it down:

```bash
# Only containers with the enable label set to true
dockwatch --label-enable
# label: com.centurylinklabs.dockwatch.enable=true

# Watch only the containers named on the command line
dockwatch nginx redis

# Exclude specific containers
dockwatch --disable-containers db,cache

# Scope this instance (lets multiple Dockwatch instances coexist)
dockwatch --scope production

# Also consider stopped / restarting containers
dockwatch --include-stopped --include-restarting
```

## HTTP API

Enable the API to trigger updates or inspect/set the schedule over HTTP. The API
listens on `:8080` and requires a bearer token.

```bash
DW_TOKEN='replace-with-strong-token'

docker run -d \
  --name dockwatch \
  --restart unless-stopped \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e DOCKWATCH_HTTP_API_TOKEN="${DW_TOKEN}" \
  dockwatch:latest \
  --http-api-update \
  --http-api-periodic-polls \
  --schedule '@every 24h'
```

By default `--http-api-update` disables periodic scans (updates only happen when
requested). Add `--http-api-periodic-polls` to keep the schedule running
alongside the API.

Endpoints (all require `Authorization: Bearer <token>`):

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/v1/update` | Run a full update scan (or a targeted one, see below) |
| `GET` | `/v1/schedule` | Return the current cron spec and next run |
| `POST` | `/v1/schedule?schedule=<cron>` | Set the schedule |
| `GET` | `/v1/metrics` | Prometheus metrics (requires `--http-api-metrics`) |

```bash
# Full scan
curl -X POST -H "Authorization: Bearer ${DW_TOKEN}" http://localhost:8080/v1/update

# Targeted scan — only these images (comma-separated or repeated)
curl -X POST -H "Authorization: Bearer ${DW_TOKEN}" \
  "http://localhost:8080/v1/update?image=nginx,redis"

# Read / change the schedule
curl -H "Authorization: Bearer ${DW_TOKEN}" http://localhost:8080/v1/schedule
curl -X POST -H "Authorization: Bearer ${DW_TOKEN}" \
  "http://localhost:8080/v1/schedule?schedule=@every%2030m"
```

A full scan is skipped if one is already running; a targeted image update waits
for the in-progress scan to finish. The API serves plain HTTP — put it behind a
TLS-terminating proxy in production.

## Prometheus metrics

```bash
dockwatch --http-api-metrics --http-api-token "${DW_TOKEN}"
# scrape http://localhost:8080/v1/metrics
```

Exposed series: `dockwatch_containers_scanned`, `dockwatch_containers_updated`,
`dockwatch_containers_failed`, `dockwatch_scans_total`, and
`dockwatch_scans_skipped`.

## Lifecycle hooks & restart behavior

```bash
# Run per-container pre/post update commands (from container labels)
dockwatch --enable-lifecycle-hooks

# Restart updated containers one at a time instead of all at once
dockwatch --rolling-restart

# Remove the old image after a successful update
dockwatch --cleanup

# Check for updates but never restart containers
dockwatch --no-restart
```

Container start order respects dependencies declared with the
`com.centurylinklabs.dockwatch.depends-on` label. `--rolling-restart` cannot be
combined with `--monitor-only`.

## Connecting to a remote Docker host

```bash
dockwatch \
  --host tcp://docker-host:2376 \
  --tlsverify \
  --api-version 1.41
```

The equivalent standard Docker environment variables (`DOCKER_HOST`,
`DOCKER_TLS_VERIFY`, `DOCKER_API_VERSION`) are also honored. Minimum Docker API
version is `1.25`.

## Configuration reference

Every flag has an environment-variable equivalent, so anything below can be set
with `-e` on `docker run`. Boolean flags accept `true`/`false`.

| Flag | Short | Environment variable | Default | Description |
|---|---|---|---|---|
| `--interval` | `-i` | `DOCKWATCH_POLL_INTERVAL` | `86400` | Poll interval, in seconds |
| `--schedule` | `-s` | `DOCKWATCH_SCHEDULE` | — | Cron expression for updates |
| `--cron` | | — | — | Alias for `--schedule` |
| `--run-once` | `-R` | `DOCKWATCH_RUN_ONCE` | `false` | Run once now and exit |
| `--force-update` | | — | `false` | Alias for `--run-once` |
| `--monitor-only` | `-m` | `DOCKWATCH_MONITOR_ONLY` | `false` | Detect updates but don't apply them |
| `--no-pull` | | `DOCKWATCH_NO_PULL` | `false` | Do not pull new images |
| `--no-restart` | | `DOCKWATCH_NO_RESTART` | `false` | Do not restart containers |
| `--cleanup` | `-c` | `DOCKWATCH_CLEANUP` | `false` | Remove old images after updating |
| `--remove-volumes` | | `DOCKWATCH_REMOVE_VOLUMES` | `false` | Remove attached volumes before updating |
| `--rolling-restart` | | `DOCKWATCH_ROLLING_RESTART` | `false` | Restart containers one at a time |
| `--stop-timeout` | `-t` | `DOCKWATCH_TIMEOUT` | — | Timeout before force-stopping a container |
| `--label-enable` | `-e` | `DOCKWATCH_LABEL_ENABLE` | `false` | Only watch labelled containers |
| `--label-take-precedence` | | `DOCKWATCH_LABEL_TAKE_PRECEDENCE` | `false` | Labels override CLI arguments |
| `--disable-containers` | `-x` | `DOCKWATCH_DISABLE_CONTAINERS` | — | Comma-separated exclusion list |
| `--scope` | | `DOCKWATCH_SCOPE` | — | Monitoring scope for this instance |
| `--include-stopped` | `-S` | `DOCKWATCH_INCLUDE_STOPPED` | `false` | Also include created/exited containers |
| `--include-restarting` | | `DOCKWATCH_INCLUDE_RESTARTING` | `false` | Also include restarting containers |
| `--revive-stopped` | | `DOCKWATCH_REVIVE_STOPPED` | `false` | Start updated stopped containers |
| `--enable-lifecycle-hooks` | | `DOCKWATCH_LIFECYCLE_HOOKS` | `false` | Run pre/post update hook commands |
| `--http-api-update` | | `DOCKWATCH_HTTP_API_UPDATE` | `false` | Enable HTTP API; updates on request |
| `--http-api-periodic-polls` | | `DOCKWATCH_HTTP_API_PERIODIC_POLLS` | `false` | Keep periodic scans while API is on |
| `--http-api-metrics` | | `DOCKWATCH_HTTP_API_METRICS` | `false` | Enable the Prometheus metrics endpoint |
| `--http-api-token` | | `DOCKWATCH_HTTP_API_TOKEN` | — | Bearer token required by the API |
| `--host` | `-H` | `DOCKER_HOST` | — | Docker daemon socket |
| `--tlsverify` | `-v` | `DOCKER_TLS_VERIFY` | `false` | Use TLS and verify the remote |
| `--api-version` | `-a` | `DOCKER_API_VERSION` | — | Docker API version |
| `--registry-tls-skip-verify` | | `DOCKWATCH_REGISTRY_TLS_SKIP_VERIFY` | `false` | Skip registry TLS verify (testing only) |
| `--log-format` | `-l` | `DOCKWATCH_LOG_FORMAT` | `Auto` | `Auto`, `LogFmt`, `Pretty`, or `JSON` |
| `--log-level` | | `DOCKWATCH_LOG_LEVEL` | `info` | `panic`…`trace` |
| `--debug` | `-d` | `DOCKWATCH_DEBUG` | `false` | Verbose logging |
| `--trace` | | `DOCKWATCH_TRACE` | `false` | Very verbose logging (exposes credentials) |
| `--no-color` | | `NO_COLOR` | — | Disable ANSI colors |
| `--no-startup-message` | | `DOCKWATCH_NO_STARTUP_MESSAGE` | `false` | Suppress the startup message |
| `--porcelain` | `-P` | `DOCKWATCH_PORCELAIN` | — | Stable machine-readable output (`v1`) |
| `--health-check` | | — | `false` | Run the container health check and exit |

Run `dockwatch --help` for the authoritative list.

## Development

```bash
# Build
go build ./...

# Test (fakes/mocks; no Docker daemon required)
go test ./...

# Vet
go vet ./...
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for the module layout and where to add
new behavior. In short: extend a deep module behind its existing interface
rather than adding a thin pass-through package.
