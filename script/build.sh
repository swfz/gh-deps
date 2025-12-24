#!/usr/bin/env bash
set -e

VERSION=$1
BINARY_NAME="gh-deps"
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

# Create dist directory
mkdir -p dist

# Define platforms to build for
PLATFORMS=(
  "darwin-amd64"
  "darwin-arm64"
  "linux-amd64"
  "linux-arm64"
  "linux-386"
  "windows-amd64"
  "windows-arm64"
  "windows-386"
  "freebsd-amd64"
  "freebsd-arm64"
)

# Build for each platform
for platform in "${PLATFORMS[@]}"; do
  IFS='-' read -r GOOS GOARCH <<< "$platform"

  output_name="${BINARY_NAME}_${VERSION}_${platform}"

  # Add .exe extension for Windows
  if [ "$GOOS" = "windows" ]; then
    output_name="${output_name}.exe"
  fi

  echo "Building for $GOOS/$GOARCH..."

  GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build \
    -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" \
    -o "dist/${output_name}" \
    cmd/gh-deps/main.go
done

echo "Build complete! Binaries are in the dist directory."
