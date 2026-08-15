#!/bin/bash

BINFILE=dockwatch
if [ -n "$MSYSTEM" ]; then
    BINFILE=dockwatch.exe
fi
# A clone with no tags (or no git at all) made this empty, and the binary then
# linked an empty version and sent "User-Agent: Dockwatch/" to every registry.
VERSION=$(git describe --tags 2>/dev/null || true)
if [ -z "$VERSION" ]; then
    VERSION="dev-$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
fi
echo "Building $VERSION..."
go build -o $BINFILE -ldflags "-X github.com/fugginold/dockwatch/internal/meta.Version=$VERSION"
