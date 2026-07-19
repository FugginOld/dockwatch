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

Build and run with Docker Compose:

```bash
git clone https://github.com/fugginold/dockwatch.git
cd dockwatch
docker compose up -d --build
```

Or build the image and run it directly:

```bash
docker build -t dockwatch:latest .

docker run -d \
  --name dockwatch \
  --restart unless-stopped \
  -v /var/run/docker.sock:/var/run/docker.sock \
  dockwatch:latest --interval 300
```

Check it:

```bash
docker ps --filter name=dockwatch
docker logs --tail=100 dockwatch
```

## Usage

Run one immediate check and exit:

```bash
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  dockwatch:latest --force-update
```

Interactive shell (`dockwatch>` prompt) — detach without stopping via `Ctrl-p` then `Ctrl-q`:

```bash
docker run --rm -it \
  -v /var/run/docker.sock:/var/run/docker.sock \
  dockwatch:latest
```

HTTP API plus periodic scheduler:

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

```bash
curl -X POST -H "Authorization: Bearer ${DW_TOKEN}" http://localhost:8080/v1/update
curl         -H "Authorization: Bearer ${DW_TOKEN}" http://localhost:8080/v1/schedule
```

See all options with `dockwatch --help`.

## Build from source

```bash
go build -o dockwatch .
./dockwatch --help
```

## License

Apache-2.0 — see [LICENSE.md](LICENSE.md).
