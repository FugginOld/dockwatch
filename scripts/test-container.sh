#!/bin/bash
set -e

IMAGE=$1

echo "Running container test for $IMAGE..."

docker run -d --name dockwatch-test -v /var/run/docker.sock:/var/run/docker.sock $IMAGE

sleep 5

if ! docker ps | grep dockwatch-test; then
  echo "Container failed to start"
  docker logs dockwatch-test
  exit 1
fi

HEALTH=$(docker inspect --format='{{.State.Health.Status}}' dockwatch-test 2>/dev/null || echo "none")

if [ "$HEALTH" = "unhealthy" ]; then
  echo "Container is unhealthy"
  docker logs dockwatch-test
  exit 1
fi

echo "Container test passed"

docker stop dockwatch-test
docker rm dockwatch-test
