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
    echo "=== Building for GOOS=$1 and GOARCH=$2."
    # https://golang.org/doc/install/source#environment
    CGO_ENABLED=0 GOOS="$1" GOARCH="$2" go build -o dist/ \
        -ldflags="-s -w -X main.gVersion=$version"
    if test "$?" != "0"; then
        ERRORS=1
    else
        case "$3" in
        *.tar.gz)
            (
                cd dist
                tar czf "$3" "$4" --remove-files
            ) || exit 5
            ;;
        *.zip)
            (
                cd dist
                zip "$3" "$4" --move
            ) || exit 5
            ;;
        esac
        echo "dist/$3 successfully created."
    fi
}

build android arm64 lf-android-arm64.tar.gz lf
build darwin amd64 lf-darwin-amd64.tar.gz lf
build dragonfly amd64 lf-dragonfly-amd64.tar.gz lf
build freebsd 386 lf-freebsd-386.tar.gz lf
build freebsd amd64 lf-freebsd-amd64.tar.gz lf
build freebsd arm lf-freebsd-arm.tar.gz lf
build illumos amd64 lf-illumos-amd64.tar.gz lf
build linux 386 lf-linux-386.tar.gz lf
build linux amd64 lf-linux-amd64.tar.gz lf
build linux arm lf-linux-arm.tar.gz lf
build linux arm64 lf-linux-arm64.tar.gz lf
build linux ppc64 lf-linux-ppc64.tar.gz lf
build linux ppc64le lf-linux-ppc64le.tar.gz lf
build linux mips lf-linux-mips.tar.gz lf
build linux mipsle lf-linux-mipsle.tar.gz lf
build linux mips64 lf-linux-mips64.tar.gz lf
build linux mips64le lf-linux-mips64le.tar.gz lf
build linux s390x lf-linux-s390x.tar.gz lf
build netbsd 386 lf-netbsd-386.tar.gz lf
build netbsd amd64 lf-netbsd-amd64.tar.gz lf
build netbsd arm lf-netbsd-arm.tar.gz lf
build openbsd 386 lf-openbsd-386.tar.gz lf
build openbsd amd64 lf-openbsd-amd64.tar.gz lf
build openbsd arm lf-openbsd-arm.tar.gz lf
build openbsd arm64 lf-openbsd-arm64.tar.gz lf
build solaris amd64 lf-solaris-amd64.tar.gz lf
build windows 386 lf-windows-386.zip lf.exe
build windows amd64 lf-windows-amd64.zip lf.exe
# Unsupported
# build aix ppc64 lf-aix-ppc64.tar.gz lf
# build android 386 lf-android-386.tar.gz lf
# build android amd64 lf-android-amd64.tar.gz lf
# build android arm lf-android-arm.tar.gz lf
# build darwin arm64 lf-darwin-arm64.tar.gz lf
# build js wasm lf-js-wasm.tar.gz lf
# build plan9 386 lf-plan9-386.tar.gz lf
# build plan9 amd64 lf-plan9-amd64.tar.gz lf
# build plan9 arm lf-plan9-arm.tar.gz lf

if test -n "$ERRORS"; then
    printf "\ngen/xbuild.sh: some targets failed to compile.\n"
    exit 1
fi
