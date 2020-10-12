#!/bin/sh
# Generates cross builds for all supported platforms.
#
# This script is used to build binaries for all supported platforms. Cgo is
# disabled to make sure binaries are statically linked. Appropriate flags are
# given to the go compiler to strip binaries. Current git tag is passed to the
# compiler by default to be used as the version in binaries. These are then
# compressed in an archive form (`.zip` for windows and `.tar.gz` for the rest)
# within a folder named `dist`.

set -o verbose

[ -z $version ] && version=$(git describe --tags)

mkdir -p dist

# https://golang.org/doc/install/source#environment

CGO_ENABLED=0 GOOS=android   GOARCH=arm64    go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-android-arm64.tar.gz   lf --remove-files
CGO_ENABLED=0 GOOS=darwin    GOARCH=amd64    go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-darwin-amd64.tar.gz    lf --remove-files
CGO_ENABLED=0 GOOS=dragonfly GOARCH=amd64    go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-dragonfly-amd64.tar.gz lf --remove-files
CGO_ENABLED=0 GOOS=freebsd   GOARCH=386      go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-freebsd-386.tar.gz     lf --remove-files
CGO_ENABLED=0 GOOS=freebsd   GOARCH=amd64    go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-freebsd-amd64.tar.gz   lf --remove-files
CGO_ENABLED=0 GOOS=freebsd   GOARCH=arm      go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-freebsd-arm.tar.gz     lf --remove-files
CGO_ENABLED=0 GOOS=illumos   GOARCH=amd64    go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-illumos-amd64.tar.gz   lf --remove-files
CGO_ENABLED=0 GOOS=linux     GOARCH=386      go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-linux-386.tar.gz       lf --remove-files
CGO_ENABLED=0 GOOS=linux     GOARCH=amd64    go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-linux-amd64.tar.gz     lf --remove-files
CGO_ENABLED=0 GOOS=linux     GOARCH=arm      go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-linux-arm.tar.gz       lf --remove-files
CGO_ENABLED=0 GOOS=linux     GOARCH=arm64    go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-linux-arm64.tar.gz     lf --remove-files
CGO_ENABLED=0 GOOS=linux     GOARCH=ppc64    go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-linux-ppc64.tar.gz     lf --remove-files
CGO_ENABLED=0 GOOS=linux     GOARCH=ppc64le  go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-linux-ppc64le.tar.gz   lf --remove-files
CGO_ENABLED=0 GOOS=linux     GOARCH=mips     go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-linux-mips.tar.gz      lf --remove-files
CGO_ENABLED=0 GOOS=linux     GOARCH=mipsle   go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-linux-mipsle.tar.gz    lf --remove-files
CGO_ENABLED=0 GOOS=linux     GOARCH=mips64   go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-linux-mips64.tar.gz    lf --remove-files
CGO_ENABLED=0 GOOS=linux     GOARCH=mips64le go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-linux-mips64le.tar.gz  lf --remove-files
CGO_ENABLED=0 GOOS=linux     GOARCH=s390x    go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-linux-s390x.tar.gz     lf --remove-files
CGO_ENABLED=0 GOOS=netbsd    GOARCH=386      go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-netbsd-386.tar.gz      lf --remove-files
CGO_ENABLED=0 GOOS=netbsd    GOARCH=amd64    go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-netbsd-amd64.tar.gz    lf --remove-files
CGO_ENABLED=0 GOOS=netbsd    GOARCH=arm      go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-netbsd-arm.tar.gz      lf --remove-files
CGO_ENABLED=0 GOOS=openbsd   GOARCH=386      go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-openbsd-386.tar.gz     lf --remove-files
CGO_ENABLED=0 GOOS=openbsd   GOARCH=amd64    go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-openbsd-amd64.tar.gz   lf --remove-files
CGO_ENABLED=0 GOOS=openbsd   GOARCH=arm      go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-openbsd-arm.tar.gz     lf --remove-files
CGO_ENABLED=0 GOOS=openbsd   GOARCH=arm64    go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-openbsd-arm64.tar.gz   lf --remove-files
CGO_ENABLED=0 GOOS=solaris   GOARCH=amd64    go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-solaris-amd64.tar.gz   lf --remove-files

CGO_ENABLED=0 GOOS=windows   GOARCH=386      go build -ldflags="-s -w -X main.gVersion=$version" && zip dist/lf-windows-386.zip            lf.exe --move
CGO_ENABLED=0 GOOS=windows   GOARCH=amd64    go build -ldflags="-s -w -X main.gVersion=$version" && zip dist/lf-windows-amd64.zip          lf.exe --move

# not supported
# CGO_ENABLED=0 GOOS=aix       GOARCH=ppc64    go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-aix-ppc64.tar.gz       lf --remove-files
# CGO_ENABLED=0 GOOS=android   GOARCH=386      go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-android-386.tar.gz     lf --remove-files
# CGO_ENABLED=0 GOOS=android   GOARCH=amd64    go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-android-amd64.tar.gz   lf --remove-files
# CGO_ENABLED=0 GOOS=android   GOARCH=arm      go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-android-arm.tar.gz     lf --remove-files
# CGO_ENABLED=0 GOOS=darwin    GOARCH=arm64    go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-darwin-arm64.tar.gz    lf --remove-files
# CGO_ENABLED=0 GOOS=js        GOARCH=wasm     go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-js-wasm.tar.gz         lf --remove-files
# CGO_ENABLED=0 GOOS=plan9     GOARCH=386      go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-plan9-386.tar.gz       lf --remove-files
# CGO_ENABLED=0 GOOS=plan9     GOARCH=amd64    go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-plan9-amd64.tar.gz     lf --remove-files
# CGO_ENABLED=0 GOOS=plan9     GOARCH=arm      go build -ldflags="-s -w -X main.gVersion=$version" && tar czf dist/lf-plan9-arm.tar.gz       lf --remove-files
