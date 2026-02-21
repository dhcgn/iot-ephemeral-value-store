#!/bin/bash
set -euo pipefail

if ! command -v git >/dev/null 2>&1; then
    echo "Error: git is not installed or not available in PATH." >&2
    exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
    echo "Error: docker is not installed or not available in PATH." >&2
    exit 1
fi

# Get version info (format: v1.0.0-5-g1234567 or v1.0.0)
GIT_DESC=$(git describe --tags 2>/dev/null || echo "v0.0.0")

# Parse version components
BASE_TAG=$(echo "$GIT_DESC" | cut -d- -f1)
COMMITS_SINCE=$(echo "$GIT_DESC" | grep -o '[0-9]*-g[0-9a-f]*$' | cut -d- -f1 || true)

# Build final version string
if [ -n "$COMMITS_SINCE" ]; then
    VERSION="${BASE_TAG}_beta${COMMITS_SINCE}"
else
    VERSION="${BASE_TAG}"
fi

COMMIT=$(git rev-parse --short HEAD)
BUILDTIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

# Build the Docker image
docker build \
    --build-arg VERSION="${VERSION}" \
    --build-arg COMMIT="${COMMIT}" \
    --build-arg BUILDTIME="${BUILDTIME}" \
    -t iot-ephemeral-value-store-server:${VERSION} \
    .

# Also tag as latest
docker tag iot-ephemeral-value-store-server:${VERSION} iot-ephemeral-value-store-server:latest

echo "Built iot-ephemeral-value-store-server:${VERSION} and tagged as latest"
