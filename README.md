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

```bash
docker rm -f dockwatch 2>/dev/null || true
DW_API_VERSION=$(docker version --format '{{.Server.APIVersion}}')

docker run -d \
  --name dockwatch \
  --restart unless-stopped \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e DOCKER_API_VERSION="${DW_API_VERSION}" \
  fugginold/dockwatch:latest
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
