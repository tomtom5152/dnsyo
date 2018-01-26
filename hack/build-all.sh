#!/usr/bin/env bash
# Copyright 2017 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.
#
# This script will build dep and calculate hash for each
# (DEP_BUILD_PLATFORMS, DEP_BUILD_ARCHS) pair.
# DEP_BUILD_PLATFORMS="linux" DEP_BUILD_ARCHS="amd64" ./hack/build-all.sh
# can be called to build only for linux-amd64

set -e

COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null)
DATE=$(date "+%Y-%m-%d")
VERSION=$(git describe --tags --dirty || echo "$COMMIT_HASH")

GO_BUILD_CMD="go build -a"
GO_BUILD_LDFLAGS="-s -w -X main.commitHash=$COMMIT_HASH -X main.buildDate=$DATE -X main.version=$VERSION"

if [ -z "$BUILD_PLATFORMS" ]; then
    BUILD_PLATFORMS="linux windows darwin freebsd"
fi

if [ -z "$BUILD_ARCHS" ]; then
    BUILD_ARCHS="amd64 386"
fi

mkdir -p release

for OS in ${BUILD_PLATFORMS[@]}; do
  for ARCH in ${BUILD_ARCHS[@]}; do
    NAME="dnsyo-$OS-$ARCH"
    if [ "$OS" == "windows" ]; then
      NAME="$NAME.exe"
    fi
    echo "Building for $OS/$ARCH"
    GOARCH=$ARCH GOOS=$OS $GO_BUILD_CMD -ldflags "$GO_BUILD_LDFLAGS" -o "release/$NAME" .
    shasum -a 256 "release/$NAME" > "release/$NAME".sha256
  done
done
