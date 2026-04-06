#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

IMAGE="${IMAGE:-fugginold/dockwatch}"
VERSION_TAG=""
PUSH_LATEST=1

usage() {
  cat <<'EOF'
Build and push Dockwatch Docker images to Docker Hub.

Usage:
  ./scripts/push-dockerhub.sh [--tag <version>] [--image <repo/name>] [--no-latest]

Options:
  --tag <version>     Also push an explicit version tag (example: 0.1.1).
  --image <repo/name> Override image name (default: fugginold/dockwatch).
  --no-latest         Do not push the latest tag.
  -h, --help          Show this help message.

Examples:
  ./scripts/push-dockerhub.sh
  ./scripts/push-dockerhub.sh --tag 0.1.1
  ./scripts/push-dockerhub.sh --image myuser/dockwatch --tag 0.1.1
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

echo "Building image ${IMAGE}:latest using dockerfiles/Dockerfile.dev-self-contained..."
docker build -f dockerfiles/Dockerfile.dev-self-contained -t "${IMAGE}:latest" .

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

echo "Done."
