#!/bin/bash
set -euo pipefail

test_container_name="${TEST_CONTAINER_NAME:-test-nginx}"
test_image="${TEST_IMAGE:-nginx:1.25.3}"
test_interval="${TEST_INTERVAL:-10}"
test_duration_seconds="${TEST_DURATION_SECONDS:-}"
cleanup_test_container="${CLEANUP_TEST_CONTAINER:-1}"

cleanup() {
	if [[ "$cleanup_test_container" == "1" ]]; then
		docker rm -f "$test_container_name" >/dev/null 2>&1 || true
	fi
}

trap cleanup EXIT INT TERM

echo "Starting test container..."
docker rm -f "$test_container_name" 2>/dev/null || true
docker run -d --name "$test_container_name" "$test_image"

echo "Building dockwatch..."
go build -o dockwatch .

chmod +x ./dockwatch

if [[ ! -x ./dockwatch ]]; then
	echo "Error: ./dockwatch is not executable"
	ls -l ./dockwatch || true
	exit 1
fi

docker_api_version="$(docker version --format '{{.Server.APIVersion}}')"
run_args=(--interval "$test_interval" "$test_container_name")

echo "Running dockwatch..."
if [[ -n "$test_duration_seconds" ]]; then
	if ! command -v timeout >/dev/null 2>&1; then
		echo "Error: timeout command not found; install coreutils or unset TEST_DURATION_SECONDS"
		exit 1
	fi

	echo "Running dockwatch for ${test_duration_seconds}s..."
	set +e
	timeout --signal=INT "${test_duration_seconds}s" env DOCKER_API_VERSION="$docker_api_version" ./dockwatch "${run_args[@]}"
	run_status=$?
	set -e

	if [[ "$run_status" -ne 0 && "$run_status" -ne 124 && "$run_status" -ne 130 ]]; then
		exit "$run_status"
	fi
else
	DOCKER_API_VERSION="$docker_api_version" ./dockwatch "${run_args[@]}"
fi