#!/usr/bin/env bash
set -euo pipefail

# CI-friendly wrapper around scripts/test-dockwatch.sh.
# Exits non-zero if Dockwatch fails during the bounded run.
DURATION_SECONDS="${TEST_DURATION_SECONDS:-30}"

if ! [[ "$DURATION_SECONDS" =~ ^[0-9]+$ ]] || [[ "$DURATION_SECONDS" -eq 0 ]]; then
  echo "Error: TEST_DURATION_SECONDS must be a positive integer"
  exit 1
fi

echo "Running bounded Dockwatch smoke test for ${DURATION_SECONDS}s..."
TEST_DURATION_SECONDS="$DURATION_SECONDS" ./scripts/test-dockwatch.sh

echo "Dockwatch smoke test completed successfully."
