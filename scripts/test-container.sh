#!/bin/bash
set -e

IMAGE=$1

echo "Running container test for $IMAGE..."

docker run -d --name dockwatch-test -v /var/run/docker.sock:/var/run/docker.sock $IMAGE

sleep 5

# Check container exists (running or exited)
if ! docker ps -a | grep dockwatch-test; then
  echo "Container never started"
  exit 1
fi

EXIT_CODE=$(docker inspect --format='{{.State.ExitCode}}' dockwatch-test)
STATUS=$(docker inspect --format='{{.State.Status}}' dockwatch-test)

echo "Container status: $STATUS (exit code: $EXIT_CODE)"
docker logs dockwatch-test

# Fail only on non-zero exit — a clean exit (0) is acceptable for a scheduler
if [ "$EXIT_CODE" != "0" ] && [ "$STATUS" != "running" ]; then
  echo "Container failed with exit code $EXIT_CODE"
  exit 1
fi

echo "Container test passed"

docker stop dockwatch-test 2>/dev/null || true
docker rm dockwatch-test
