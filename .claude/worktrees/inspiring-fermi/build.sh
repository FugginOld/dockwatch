#!/bin/bash

BINFILE=dockwatch
if [ -n "$MSYSTEM" ]; then
    BINFILE=dockwatch.exe
fi
VERSION=$(git describe --tags)
echo "Building $VERSION..."
go build -o $BINFILE -ldflags "-X github.com/fugginold/dockwatch/internal/meta.Version=$VERSION"
