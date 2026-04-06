<p style="text-align: center; margin-left: 1.6rem;">
  <img alt="Logotype depicting a lighthouse" src="./images/logo-450px.png" width="450" />
</p>
<h1 align="center">
  Dockwatch
</h1>

<p align="center">
  A container-based solution for automating Docker container base image updates.
  <br/><br/>
  <a href="https://circleci.com/gh/fugginold/dockwatch">
    <img alt="Circle CI" src="https://circleci.com/gh/fugginold/dockwatch.svg?style=shield" />
  </a>
  <a href="https://codecov.io/gh/fugginold/dockwatch">
    <img alt="Codecov" src="https://codecov.io/gh/fugginold/dockwatch/branch/main/graph/badge.svg">
  </a>
  <a href="https://godoc.org/github.com/fugginold/dockwatch">
    <img alt="GoDoc" src="https://godoc.org/github.com/fugginold/dockwatch?status.svg" />
  </a>
  <a href="https://goreportcard.com/report/github.com/fugginold/dockwatch">
    <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/fugginold/dockwatch" />
  </a>
  <a href="https://github.com/fugginold/dockwatch/releases">
    <img alt="latest version" src="https://img.shields.io/github/tag/fugginold/dockwatch.svg" />
  </a>
  <a href="https://www.apache.org/licenses/LICENSE-2.0">
    <img alt="Apache-2.0 License" src="https://img.shields.io/github/license/fugginold/dockwatch.svg" />
  </a>
  <a href="https://www.codacy.com/gh/fugginold/dockwatch/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=fugginold/dockwatch&amp;utm_campaign=Badge_Grade">
    <img alt="Codacy Badge" src="https://app.codacy.com/project/badge/Grade/1c48cfb7646d4009aa8c6f71287670b8"/>
  </a>
  <a href="https://hub.docker.com/r/fugginold/dockwatch">
    <img alt="Pulls from DockerHub" src="https://img.shields.io/docker/pulls/fugginold/dockwatch.svg" />
  </a>
</p>

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
    fugginold/dockwatch
    ```

=== "docker-compose.yml"

    ```yaml
    version: "3"
    services:
      dockwatch:
        image: fugginold/dockwatch
        volumes:
          - /var/run/docker.sock:/var/run/docker.sock
    ```
