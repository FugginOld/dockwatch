#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

IMAGE="${IMAGE:-fugginold/dockwatch}"
VERSION_TAG=""
PUSH_LATEST=1
NO_CACHE=0
USE_LEGACY_BUILDER=0
PLATFORMS="${PLATFORMS:-linux/amd64,linux/arm64,linux/arm/v7,linux/386}"
BUILDER_NAME="${BUILDER_NAME:-dockwatch-builder}"

usage() {
  cat <<'EOF'
Build and push Dockwatch Docker images to Docker Hub.

Usage:
  ./scripts/push-dockerhub.sh [--tag <version>] [--image <repo/name>] [--no-latest] [--no-cache] [--legacy-builder] [--platforms <csv>] [--builder <name>]

Options:
  --tag <version>     Also push an explicit version tag (example: 0.1.8).
  --image <repo/name> Override image name (default: fugginold/dockwatch).
  --no-latest         Do not push the latest tag.
  --no-cache          Build without using Docker layer cache.
  --legacy-builder    Use single-arch local docker build/push flow.
  --platforms <csv>   Target platforms for buildx (default: linux/amd64,linux/arm64,linux/arm/v7,linux/386).
  --builder <name>    Buildx builder name (default: dockwatch-builder).
  -h, --help          Show this help message.

Examples:
  ./scripts/push-dockerhub.sh --tag 0.1.8
  ./scripts/push-dockerhub.sh --image myuser/dockwatch --tag 0.1.8
  ./scripts/push-dockerhub.sh --platforms linux/amd64,linux/arm64 --tag 0.1.8
  ./scripts/push-dockerhub.sh --legacy-builder --no-cache --tag 0.1.8
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --tag)
      [[ $# -ge 2 ]] || { echo "Error: --tag requires a value"; exit 1; }
      VERSION_TAG="$2"
      shift 2
      ;;
    --image)
      [[ $# -ge 2 ]] || { echo "Error: --image requires a value"; exit 1; }
      IMAGE="$2"
      shift 2
      ;;
    --no-latest)
      PUSH_LATEST=0
      shift
      ;;
    --no-cache)
      NO_CACHE=1
      shift
      ;;
    --legacy-builder)
      USE_LEGACY_BUILDER=1
      shift
      ;;
    --platforms)
      [[ $# -ge 2 ]] || { echo "Error: --platforms requires a value"; exit 1; }
      PLATFORMS="$2"
      shift 2
      ;;
    --builder)
      [[ $# -ge 2 ]] || { echo "Error: --builder requires a value"; exit 1; }
      BUILDER_NAME="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Error: Unknown option '$1'"
      usage
      exit 1
      ;;
  esac
done

if [[ "$PUSH_LATEST" -eq 0 && -z "$VERSION_TAG" ]]; then
  echo "Error: Nothing to push. Provide --tag or omit --no-latest."
  exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "Error: docker is not installed or not on PATH"
  exit 1
fi

if ! docker info >/dev/null 2>&1; then
  echo "Error: Docker daemon is not reachable"
  exit 1
fi

cd "$REPO_ROOT"

if [[ "$USE_LEGACY_BUILDER" -eq 1 ]]; then
  build_args=(-f dockerfiles/Dockerfile.dev-self-contained -t "${IMAGE}:latest")
  if [[ "$NO_CACHE" -eq 1 ]]; then
    build_args+=(--no-cache)
  fi
  build_args+=(.)

  echo "Building image ${IMAGE}:latest using legacy single-arch flow..."
  DOCKER_BUILDKIT=0 docker build "${build_args[@]}"

  if [[ -n "$VERSION_TAG" ]]; then
    echo "Tagging image ${IMAGE}:${VERSION_TAG}..."
    docker tag "${IMAGE}:latest" "${IMAGE}:${VERSION_TAG}"
  fi

  if [[ "$PUSH_LATEST" -eq 1 ]]; then
    echo "Pushing ${IMAGE}:latest..."
    docker push "${IMAGE}:latest"
  fi

  if [[ -n "$VERSION_TAG" ]]; then
    echo "Pushing ${IMAGE}:${VERSION_TAG}..."
    docker push "${IMAGE}:${VERSION_TAG}"
  fi
else
  if ! docker buildx inspect "$BUILDER_NAME" >/dev/null 2>&1; then
    docker buildx create --name "$BUILDER_NAME" --use >/dev/null
  else
    docker buildx use "$BUILDER_NAME"
  fi

  docker buildx inspect --bootstrap >/dev/null

  tags=()
  if [[ "$PUSH_LATEST" -eq 1 ]]; then
    tags+=(--tag "${IMAGE}:latest")
  fi
  if [[ -n "$VERSION_TAG" ]]; then
    tags+=(--tag "${IMAGE}:${VERSION_TAG}")
  fi

  buildx_args=(
    --platform "$PLATFORMS"
    --file dockerfiles/Dockerfile.dev-self-contained
    --provenance=false
    --push
  )
  if [[ "$NO_CACHE" -eq 1 ]]; then
    buildx_args+=(--no-cache)
  fi

  echo "Building and pushing multi-arch images for ${IMAGE} on ${PLATFORMS}..."
  docker buildx build "${buildx_args[@]}" "${tags[@]}" .
fi

echo "Done."
