#!/bin/sh
# Generates cross builds for all supported platforms.
#
# This script is used to build binaries for all supported platforms. Cgo is
# disabled to make sure binaries are statically linked. Appropriate flags are
# given to the go compiler to strip binaries. Current git tag is passed to the
# compiler by default to be used as the version in binaries. These are then
# compressed in an archive form (`.zip` for windows and `.tar.gz` for the rest)
# within a folder named `dist`.

[ -z "$version" ] && version=$(git describe --tags)

test -d dist && {
    echo "gen/xbuild.sh: WARNING: removing preexisting subdir 'dist'"
    rm -r dist || exit 5
}
mkdir dist || exit 5

ERRORS=

build() {
    # https://golang.org/doc/install/source#environment
    echo "=== Building for GOOS=$1 and GOARCH=$2."
    CGO_ENABLED=0 GOOS="$1" GOARCH="$2" go build -o dist/ \
        -ldflags="-s -w -X main.gVersion=$version"
    if test "$?" != "0"; then
        ERRORS=1
    else
        package "$1" "$2" || exit 5
    fi
}

package() (
    cd dist || return 1 
    # Since the function is surrounded by (), the cd only affects a subshell
    OUTFILE=
    case "$1" in
    windows)
        OUTFILE="lf-$1-$2.zip"
        zip "$OUTFILE" lf.exe --move || return 1
        ;;
    *)
        OUTFILE="lf-$1-$2.tar.gz"
        tar czf "$OUTFILE" lf --remove-files || return 1
        ;;
    esac
    echo "dist/$OUTFILE successfully created."
)

build android arm64
build darwin amd64
build darwin arm64
build dragonfly amd64
build freebsd 386
build freebsd amd64
build freebsd arm
build illumos amd64
build linux 386
build linux amd64
build linux arm
build linux arm64
build linux ppc64
build linux ppc64le
build linux mips
build linux mipsle
build linux mips64
build linux mips64le
build linux s390x
build netbsd 386
build netbsd amd64
build netbsd arm
build openbsd 386
build openbsd amd64
build openbsd arm
build openbsd arm64
build solaris amd64
build windows 386
build windows amd64
# Unsupported
# build aix ppc64
# build android 386
# build android amd64
# build android arm
# build js wasm
# build plan9 386
# build plan9 amd64
# build plan9 arm

if test -n "$ERRORS"; then
    printf "\ngen/xbuild.sh: some targets failed to compile.\n"
    exit 1
fi
