# Dockwatch

> This project is actively maintained. Issues and pull requests are welcome.

A process for automating Docker container base image updates.

![Dockwatch logo](logo.png)

[![GoDoc](https://godoc.org/github.com/fugginold/dockwatch?status.svg)](https://godoc.org/github.com/fugginold/dockwatch)
[![Go Report Card](https://goreportcard.com/badge/github.com/fugginold/dockwatch)](https://goreportcard.com/report/github.com/fugginold/dockwatch)
[![latest version](https://img.shields.io/github/tag/fugginold/dockwatch.svg)](https://github.com/fugginold/dockwatch/releases)
[![Apache-2.0 License](https://img.shields.io/github/license/fugginold/dockwatch.svg)](https://www.apache.org/licenses/LICENSE-2.0)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/1c48cfb7646d4009aa8c6f71287670b8)](https://www.codacy.com/gh/fugginold/dockwatch/dashboard?utm_source=github.com&utm_medium=referral&utm_content=fugginold/dockwatch&utm_campaign=Badge_Grade)
[![Pulls from DockerHub](https://img.shields.io/docker/pulls/fugginold/dockwatch.svg)](https://hub.docker.com/repository/docker/fugginold/dockwatch/)

## What It Does

Dockwatch is a Go-based Docker update daemon. It monitors containers, checks whether image digests are stale in a registry, pulls newer images, and restarts containers using their existing runtime configuration.

Core capabilities:

- Automatic update detection and rollout
- One-shot update mode for immediate checks
- Optional HTTP API for update and schedule control
- Interactive shell environment for on-the-fly execution
- Prometheus metrics endpoint
- Dependency-aware container restart ordering

## Quick Start Instructions

Start Dockwatch with defaults:

```bash
docker run -d \
  --name dockwatch \
  --restart unless-stopped \
  -v /var/run/docker.sock:/var/run/docker.sock \
  fugginold/dockwatch:latest
```

Verify:

```bash
docker ps --filter name=dockwatch
docker logs --tail=100 dockwatch
```

Run one immediate check and exit:

```bash
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  fugginold/dockwatch:latest \
  --force-update
```

## Install & Usage Instructions

### Debian install

Install Docker Engine:

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

Optional no-sudo Docker usage:

```bash
sudo usermod -aG docker "$USER"
newgrp docker
```

Install Dockwatch with an explicit schedule:

```bash
docker rm -f dockwatch 2>/dev/null || true
DW_API_VERSION=$(docker version --format '{{.Server.APIVersion}}')
DW_SCHEDULE='@every 24h'

docker run -d \
  --name dockwatch \
  --restart unless-stopped \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e DOCKER_API_VERSION="${DW_API_VERSION}" \
  fugginold/dockwatch:latest \
  --schedule "${DW_SCHEDULE}"
```

Architecture-specific tags when pinning platform:

- AMD64: fugginold/dockwatch:amd64-latest
- i386: fugginold/dockwatch:i386-latest
- ARMv6/v7: fugginold/dockwatch:armhf-latest
- ARM64: fugginold/dockwatch:arm64v8-latest

### Runtime API usage

Enable update API plus periodic scheduler controls:

```bash
DW_TOKEN='replace-with-strong-token'
DW_API_VERSION=$(docker version --format '{{.Server.APIVersion}}')

docker rm -f dockwatch 2>/dev/null || true
docker run -d \
  --name dockwatch \
  --restart unless-stopped \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e DOCKER_API_VERSION="${DW_API_VERSION}" \
  -e DOCKWATCH_HTTP_API_TOKEN="${DW_TOKEN}" \
  fugginold/dockwatch:latest \
  --http-api-update \
  --http-api-periodic-polls \
  --schedule '@every 24h'
```

Examples:

```bash
curl -X POST -H "Authorization: Bearer ${DW_TOKEN}" http://localhost:8080/v1/update
curl -H "Authorization: Bearer ${DW_TOKEN}" http://localhost:8080/v1/schedule
curl -X POST -H "Authorization: Bearer ${DW_TOKEN}" "http://localhost:8080/v1/schedule?schedule=@every%2030m"
```

## Reference the following docs below in /docs

- /docs/ARCHITECTURE.md
- /docs/CONFIGURATION.md
- /docs/API.md
- /docs/DEVELOPMENT.md
- /docs/TROUBLESHOOTING.md
- /docs/arguments.md
- /docs/installation.md
- /docs/http-api-mode.md
- /docs/metrics.md
- /docs/lifecycle-hooks.md
- /docs/container-selection.md
- /docs/running-multiple-instances.md
- /docs/interactive-shell.md
