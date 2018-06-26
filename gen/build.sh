#!/bin/sh
# Builds a static stripped binary with version information.
#
# This script is used to build a binary for the current platform. Cgo is
# disabled to make sure the binary is statically linked. Appropriate flags are
# given to the go compiler to strip the binary. Current git tag is passed to
# the compiler by default to be used as the version in the binary.

[ -z $version ] && version=$(git describe --tags)

CGO_ENABLED=0 go build -ldflags="-s -w -X main.gVersion=$version" "$@"
