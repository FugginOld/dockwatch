# Dockwatch

A daemon that automates Docker container image updates: it watches running
containers, checks whether their image digests are stale in the registry,
pulls newer images, and recreates the containers using their existing runtime
configuration.

![Dockwatch logo](logo.png)

[![Go Report Card](https://goreportcard.com/badge/github.com/fugginold/dockwatch)](https://goreportcard.com/report/github.com/fugginold/dockwatch)
[![Apache-2.0 License](https://img.shields.io/github/license/fugginold/dockwatch.svg)](https://www.apache.org/licenses/LICENSE-2.0)

## What it does

- Automatic update detection and rollout
- One-shot update mode for immediate checks
- Optional HTTP API for update and schedule control
- Interactive shell environment for on-the-fly execution
- Prometheus metrics endpoint
- Dependency-aware container restart ordering

## Quick start

Run the published image with Docker Compose:

```yaml
# docker-compose.yml
services:
  dockwatch:
    image: ghcr.io/fugginold/dockwatch:latest
    container_name: dockwatch
    restart: unless-stopped
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    command: --interval 300
```

```bash
docker compose up -d
```

Or run it directly:

```bash
docker run -d \
  --name dockwatch \
  --restart unless-stopped \
  -v /var/run/docker.sock:/var/run/docker.sock \
  ghcr.io/fugginold/dockwatch:latest --interval 300
```

Because it runs a published image, dockwatch also keeps **itself** up to date on each run.

**Which tag to use.** `latest` is the highest released version and only ever moves
on a release tag. `main` tracks the tip of the default branch and is pre-release —
useful for trying an unreleased fix, not for anything you care about. Pin
`1.2.3` when you want a specific version. Full table in
[HOWTO.md](docs/HOWTO.md#published-tags).

### Build the image from source

To hack on dockwatch, build the image locally instead:

```bash
git clone https://github.com/fugginold/dockwatch.git
cd dockwatch
docker compose up -d --build
```

A locally-built image has no registry upstream, so dockwatch cannot self-update
it — you'll see harmless "pull access denied" warnings for its own container.
Rebuild with `docker compose up -d --build` to update.

Check it:

```bash
docker ps --filter name=dockwatch
docker logs --tail=100 dockwatch
```

## Common commands

These are the binary's flags, not host commands. For the daemon's own settings put
them in `command:`; to trigger a scan on a daemon that is already running, use the
shell (`docker attach dockwatch`, then `update`) or the HTTP API — not
`docker exec`, which starts a second process that ignores the daemon's flags. See
[HOWTO.md](docs/HOWTO.md#running).

```bash
# One immediate check, then exit
dockwatch --run-once

# Check on a schedule (cron) or a fixed interval (seconds)
dockwatch --schedule "@every 6h"
dockwatch --interval 300

# Detect updates without applying them
dockwatch --monitor-only
```

Run `dockwatch --help` for the full flag list.

## Documentation

- **[docs/HOWTO.md](docs/HOWTO.md)** — running, container selection, the HTTP
  API, metrics, and the full flag / environment-variable reference.
- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** — module layout, the update
  flow, and where to add new behavior.

## Build from source

```bash
go build -o dockwatch .
./dockwatch --help
```

## License

Apache-2.0 — see [LICENSE.md](LICENSE.md).
