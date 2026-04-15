# Dockwatch

A container-based solution for automating Docker container base image updates.

![Dockwatch logo](./images/logo-450px.png)

[![GitHub CI](https://github.com/fugginold/dockwatch/actions/workflows/ci.yml/badge.svg)](https://github.com/fugginold/dockwatch/actions/workflows/ci.yml)
[![Codecov](https://codecov.io/gh/fugginold/dockwatch/branch/main/graph/badge.svg)](https://codecov.io/gh/fugginold/dockwatch)
[![GoDoc](https://godoc.org/github.com/fugginold/dockwatch?status.svg)](https://godoc.org/github.com/fugginold/dockwatch)
[![Go Report Card](https://goreportcard.com/badge/github.com/fugginold/dockwatch)](https://goreportcard.com/report/github.com/fugginold/dockwatch)
[![latest version](https://img.shields.io/github/tag/fugginold/dockwatch.svg)](https://github.com/fugginold/dockwatch/releases)
[![Apache-2.0 License](https://img.shields.io/github/license/fugginold/dockwatch.svg)](https://www.apache.org/licenses/LICENSE-2.0)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/1c48cfb7646d4009aa8c6f71287670b8)](https://www.codacy.com/gh/fugginold/dockwatch/dashboard?utm_source=github.com&utm_medium=referral&utm_content=fugginold/dockwatch&utm_campaign=Badge_Grade)
[![Pulls from DockerHub](https://img.shields.io/docker/pulls/fugginold/dockwatch.svg)](https://hub.docker.com/repository/docker/fugginold/dockwatch/)

## Quick Start

Dockwatch is actively maintained and continues to support automated image update workflows for Docker-based environments.

With dockwatch you can update the running version of your containerized app simply by pushing a new image to the Docker
Hub or your own image registry. Dockwatch will pull down your new image, gracefully shut down your existing container
and restart it with the same options that were used when it was deployed initially. Run the dockwatch container with
the following command:

=== "docker run"

    ```bash
    $ docker run -d \
    --name dockwatch \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v dockwatch-config:/config \
    fugginold/dockwatch
    ```

=== "docker-compose.yml"

    ```yaml
    services:
      dockwatch:
        image: fugginold/dockwatch
        volumes:
          - /var/run/docker.sock:/var/run/docker.sock
          - dockwatch-config:/config

    volumes:
      dockwatch-config:
    ```
