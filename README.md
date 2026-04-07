<div align="center">

  <h3>This project is actively maintained</h3>
  <p>Issues and pull requests are welcome.</p>

  <img alt="Dockwatch logo" src="logo.png" width="450" />

  <h1>Dockwatch</h1>

  <p>A process for automating Docker container base image updates.</p>

  <p>
    <a href="https://circleci.com/gh/fugginold/dockwatch"><img alt="Circle CI" src="https://circleci.com/gh/fugginold/dockwatch.svg?style=shield"></a>
    <a href="https://codecov.io/gh/fugginold/dockwatch"><img alt="codecov" src="https://codecov.io/gh/fugginold/dockwatch/branch/main/graph/badge.svg"></a>
    <a href="https://godoc.org/github.com/fugginold/dockwatch"><img alt="GoDoc" src="https://godoc.org/github.com/fugginold/dockwatch?status.svg"></a>
    <a href="https://goreportcard.com/report/github.com/fugginold/dockwatch"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/fugginold/dockwatch"></a>
    <a href="https://github.com/fugginold/dockwatch/releases"><img alt="latest version" src="https://img.shields.io/github/tag/fugginold/dockwatch.svg"></a>
    <a href="https://www.apache.org/licenses/LICENSE-2.0"><img alt="Apache-2.0 License" src="https://img.shields.io/github/license/fugginold/dockwatch.svg"></a>
    <a href="https://www.codacy.com/gh/fugginold/dockwatch/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=fugginold/dockwatch&amp;utm_campaign=Badge_Grade"><img alt="Codacy Badge" src="https://app.codacy.com/project/badge/Grade/1c48cfb7646d4009aa8c6f71287670b8"></a>
    <a href="https://hub.docker.com/repository/docker/fugginold/dockwatch/"><img alt="Pulls from DockerHub" src="https://img.shields.io/docker/pulls/fugginold/dockwatch.svg"></a>
  </p>

</div>

## What It Does

**Dockwatch** is a Docker container auto-update daemon written in Go. It monitors running Docker containers, checks if their base images have been updated in a registry, and automatically pulls the new image and restarts the container with the same configuration. Think of it as a self-hosted alternative to tools like Watchtower.

## Technology Stack

| Technology | Purpose |
|---|---|
| **Go** (≥1.25) | Primary language |
| **Docker SDK** (`docker/docker`, `docker/cli`) | Container management via Unix socket |
| **Cobra + Pflag** | CLI framework and flag parsing |
| **Viper** | Config from env vars / files |
| **robfig/cron** | Cron-style scheduling (`--schedule "@every 24h"`) |
| **Logrus** | Structured logging |
| **Prometheus client** | Metrics exposure via HTTP |
| **Ginkgo + Gomega** | BDD-style test suite |
| **Testify** | Additional test assertions/mocks |
| **MkDocs** | Documentation site |

## Repository Structure

```
dockwatch/
├── main.go                    # Entrypoint — calls cmd.Execute()
├── cmd/
│   └── root.go                # Cobra root command: PreRun (config) + Run (main loop)
├── internal/                  # Private application code
│   ├── actions/               # Core business logic: update.go, check.go
│   ├── flags/                 # CLI flag registration, env-var binding, logging setup
│   ├── meta/                  # Version metadata
│   └── util/                  # Misc helpers (random names, SHA256)
├── pkg/                       # Reusable/public packages
│   ├── api/                   # HTTP API server + handlers:
│   │   ├── api.go             #   - base server with token auth
│   │   ├── update/            #   - POST endpoint to trigger updates
│   │   ├── schedule/          #   - GET/POST to inspect/modify schedule
│   │   └── metrics/           #   - Prometheus metrics endpoint
│   ├── container/             # Docker client wrapper (list, stop, start, pull, rename)
│   ├── filters/               # Container filtering by name, label, scope
│   ├── lifecycle/             # Pre/post update hook execution inside containers
│   ├── metrics/               # Internal metrics model (scanned/updated/failed counters)
│   ├── registry/              # Image registry interaction
│   │   ├── auth/              #   - Docker Hub + private registry authentication
│   │   ├── digest/            #   - Image digest comparison
│   │   ├── manifest/          #   - Manifest fetching
│   │   └── helpers/           #   - URL/ref parsing utilities
│   ├── session/               # Per-run progress tracking and final report
│   ├── sorter/                # Topological sort of containers by dependency links
│   └── types/                 # Shared interfaces and data structs (Container, Filter, Report…)
├── dockerfiles/               # Dockerfile variants (dev, self-contained, networking)
├── docker-compose.yml         # Full local dev stack (dockwatch + Prometheus + Grafana + demo containers)
├── prometheus/                # Prometheus config for scraping dockwatch metrics
├── grafana/                   # Grafana provisioning/dashboards
├── docs/                      # MkDocs documentation source
└── scripts/                   # install-dockwatch.sh helper script
```

## How It Works (Execution Flow)

1. **`main.go`** initializes logging and calls `cmd.Execute()`.
2. **`cmd/root.go` PreRun** parses flags/env-vars, creates the Docker `container.Client`.
3. **`cmd/root.go` Run** does one of two things:
   - **Run-once mode** (`--run-once`): calls `runUpdates()` immediately and exits.
   - **Daemon mode**: starts a `scheduleController` (cron job) and optionally an HTTP API.
4. **`runUpdates()`** calls `actions.Update()` which:
   - Lists containers matching the configured filter
   - For each container, calls `client.IsContainerStale()` — compares the local image digest vs. the registry
   - Sorts stale containers by dependency order (`sorter.SortByDependencies`)
   - Stops containers in **reverse** dependency order
   - Pulls new images and restarts containers in **forward** dependency order
   - Optionally runs **lifecycle hooks** (pre/post update commands inside containers)
   - Optionally cleans up old images
5. Results are collected into a **session report** and fed into **Prometheus metrics**.

## Key Architectural Patterns

- **Interface-driven design**: `container.Client` is an interface, making the whole system mockable for tests.
- **Dependency-aware updates**: The `sorter` package topologically sorts containers so that parents restart before children.
- **Label system**: Containers can opt in/out of updates and configure behavior via Docker labels (e.g., `com.centurylinklabs.dockwatch.depends-on`).
- **HTTP API** (optional): Exposes endpoints to trigger updates on demand and inspect/change the schedule — protected by a bearer token.
- **Prometheus metrics**: Exposes scan counts (scanned/updated/failed) at `:8080` for observability.
- **Concurrency safety**: A channel-based mutex (`updateLock`) ensures only one update run happens at a time, even if the HTTP API and scheduler fire simultaneously.

## Quick Start

Dockwatch is actively maintained as a container update automation tool for Docker environments.

With dockwatch you can update the running version of your containerized app simply by pushing a new image to the Docker Hub or your own image registry. 

Dockwatch will pull down your new image, gracefully shut down your existing container and restart it with the same options that were used when it was deployed initially. Run the dockwatch container with the following command:

```
$ docker run --detach \
    --name dockwatch \
    --volume /var/run/docker.sock:/var/run/docker.sock \
    fugginold/dockwatch
```

## Installation (Debian)

Install Docker Engine from the official Docker repository:

```bash
sudo apt-get update
sudo apt-get install -y ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian $(. /etc/os-release; echo $VERSION_CODENAME) stable" | sudo tee /etc/apt/sources.list.d/docker.list >/dev/null

sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo systemctl enable --now docker
```

Run Dockwatch with API-version auto-detection to avoid client/server mismatch errors:

Optional interactive install with cron prompt:

```bash
./scripts/install-dockwatch.sh
```

```bash
docker rm -f dockwatch 2>/dev/null || true
DW_API_VERSION=$(docker version --format '{{.Server.APIVersion}}')

docker run -d \
  --name dockwatch \
  --restart unless-stopped \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e DOCKER_API_VERSION="${DW_API_VERSION}" \
  fugginold/dockwatch:latest \
  --schedule "@every 24h"
```

Verify the installation:

```bash
docker ps --filter name=dockwatch
docker logs --tail=100 dockwatch
```

Docker Compose alternative:

```yaml
services:
  dockwatch:
    image: fugginold/dockwatch:latest
    container_name: dockwatch
    restart: unless-stopped
    command: --schedule "@every 24h"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      DOCKER_API_VERSION: "${DW_API_VERSION}"
```

```bash
DW_API_VERSION=$(docker version --format '{{.Server.APIVersion}}')
export DW_API_VERSION
docker compose up -d
```

For full setup details, see `docs/installation.md` and the docs site navigation.

Dockwatch is intended to be used in homelabs, media centers, local dev environments, and similar. We do **not** recommend using Dockwatch in a commercial or production environment. If that is you, you should be looking into using Kubernetes. If that feels like too big a step for you, please look into solutions like [MicroK8s](https://microk8s.io/) and [k3s](https://k3s.io/) that take away a lot of the toil of running a Kubernetes cluster. 

## Local Test Script

Use the smoke-test helper to build Dockwatch and run it against a temporary nginx container:

```bash
./scripts/test-dockwatch.sh
```

For CI or local bounded runs, set a duration (seconds):

```bash
TEST_DURATION_SECONDS=25 ./scripts/test-dockwatch.sh
```

Supported environment variables:

- `TEST_DURATION_SECONDS`: Optional run duration in seconds. If set, the script exits successfully after the bounded run unless Dockwatch returns an actual error.
- `TEST_CONTAINER_NAME`: Name for the temporary test container (default: `test-nginx`).
- `TEST_IMAGE`: Test container image (default: `nginx:1.25.3`).
- `TEST_INTERVAL`: Dockwatch interval in seconds (default: `10`).
- `CLEANUP_TEST_CONTAINER`: `1` (default) removes the test container on exit, `0` keeps it for debugging.

CI-friendly wrapper:

```bash
./scripts/test-dockwatch-ci.sh
```
