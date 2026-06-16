#!/bin/bash
set -e

echo "Building sl-cli..."

# Build for current platform
go build -o sl-cli ./cmd/sl-cli

echo "Build complete: ./sl-cli"