#!/bin/sh
# Compresses a binary into an archived form.
#
# This script is used to compress a binary built from `build.sh` into an
# archive. `.zip` is used for Windows and `.tar.gz` otherwise. The archive is
# placed inside a directory named `dist`.

set -o errexit -o nounset

mkdir -p dist

if [ "$GOOS" = "windows" ]; then
    zip "dist/lf-${GOOS}-${GOARCH}.zip" lf.exe
else
    tar czf "dist/lf-${GOOS}-${GOARCH}.tar.gz" lf
fi

# vim: tabstop=4 shiftwidth=4 textwidth=80 colorcolumn=80
