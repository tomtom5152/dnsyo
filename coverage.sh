#!/usr/bin/env bash
# Copyright 2017 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.
#
# This script will generate coverage.txt
set -e

PKGS=$(go list ./... | grep -v /vendor/)
echo "mode: atomic" > cover.out
for pkg in $PKGS; do
  go test -v -race -coverprofile=profile.out -covermode=atomic -coverpkg=./... $pkg
  if [[ -f profile.out ]]; then
    grep -v "^mode:" profile.out >> cover.out
    rm profile.out
  fi
done

# Print out the coverage.txt file with a total value so gitlab can parse it
go tool cover -func=cover.out
