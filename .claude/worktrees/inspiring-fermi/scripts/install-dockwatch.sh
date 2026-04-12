#!/bin/bash

# Install dockwatch with an optional cron schedule prompt.

set -euo pipefail

readonly IMAGE="fugginold/dockwatch:latest"
readonly CONTAINER_NAME="dockwatch"

if ! command -v docker >/dev/null 2>&1; then
    echo "Error: docker is required but was not found in PATH." >&2
    exit 1
fi

default_schedule="@every 24h"
read -r -p "Enter dockwatch cron schedule [${default_schedule}]: " schedule
schedule="${schedule:-$default_schedule}"

api_version="$(docker version --format '{{.Server.APIVersion}}')"

echo "Installing ${CONTAINER_NAME} using schedule: ${schedule}"
docker rm -f "${CONTAINER_NAME}" >/dev/null 2>&1 || true

docker run -d \
    --name "${CONTAINER_NAME}" \
    --restart unless-stopped \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -e DOCKER_API_VERSION="${api_version}" \
    "${IMAGE}" \
    --schedule "${schedule}"

echo "dockwatch installed successfully."
echo "Verify with: docker logs --tail=100 ${CONTAINER_NAME}"