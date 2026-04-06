<div align="center">

  ### This project is actively maintained
  Issues and pull requests are welcome.

  ---
  
  <img src="./logo.png" width="450" />
  
  # Dockwatch
  
  A process for automating Docker container base image updates.
  <br/><br/>
  
  [![Circle CI](https://circleci.com/gh/fugginold/dockwatch.svg?style=shield)](https://circleci.com/gh/fugginold/dockwatch)
  [![codecov](https://codecov.io/gh/fugginold/dockwatch/branch/main/graph/badge.svg)](https://codecov.io/gh/fugginold/dockwatch)
  [![GoDoc](https://godoc.org/github.com/fugginold/dockwatch?status.svg)](https://godoc.org/github.com/fugginold/dockwatch)
  [![Go Report Card](https://goreportcard.com/badge/github.com/fugginold/dockwatch)](https://goreportcard.com/report/github.com/fugginold/dockwatch)
  [![latest version](https://img.shields.io/github/tag/fugginold/dockwatch.svg)](https://github.com/fugginold/dockwatch/releases)
  [![Apache-2.0 License](https://img.shields.io/github/license/fugginold/dockwatch.svg)](https://www.apache.org/licenses/LICENSE-2.0)
  [![Codacy Badge](https://app.codacy.com/project/badge/Grade/1c48cfb7646d4009aa8c6f71287670b8)](https://www.codacy.com/gh/fugginold/dockwatch/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=fugginold/dockwatch&amp;utm_campaign=Badge_Grade)
  [![Pulls from DockerHub](https://img.shields.io/docker/pulls/fugginold/dockwatch.svg)](https://hub.docker.com/r/fugginold/dockwatch)

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
