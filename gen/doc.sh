#!/bin/sh
# Generates `lf.1` and `doc.txt` from the `doc.md` file.
#
# This script is used to generate a man page and a plain text conversion of the
# markdown documentation using pandoc (https://pandoc.org/). GitHub Flavored
# Markdown (GFM) (https://github.github.com/gfm/) is used for the markdown
# input format. The markdown file is automatically rendered in the GitHub
# repository (https://github.com/gokcehan/lf/blob/master/doc.md). The man page
# file `lf.1` is meant to be used for installations on Unix systems. The plain
# text file `doc.txt` is embedded in the binary to be displayed on request with
# the `-doc` command line flag. Thus the same documentation is used for online
# and terminal display.

set -o errexit -o nounset

get_version() {
    printf "r%s" $(($(git describe --tags --abbrev=0 | tr -d r) + 1))
}

[ -z "${version:-}" ] && version=$(get_version)
[ -z "${date:-}" ] && date=$(date +%F)

PANDOC_IMAGE=pandoc/minimal:3.7

generate_man_page() {
    "${OCI_PROGRAM?}" run \
        --rm \
        --volume "$PWD:/data" \
        "$@" "$PANDOC_IMAGE" \
        --standalone \
        --from gfm --to man \
        --metadata=title:"LF" \
        --metadata=section:"1" \
        --metadata=date:"$date" \
        --metadata=footer:"$version" \
        --metadata=header:"DOCUMENTATION" \
        doc.md -o lf.1
}

generate_plain_text() {
    "${OCI_PROGRAM?}" run \
        --rm \
        --volume "$PWD:/data" \
        "$@" "$PANDOC_IMAGE" \
        --standalone \
        --from gfm --to plain \
        doc.md -o doc.txt
}

is_rootless() {
    case "$OCI_PROGRAM" in
        podman) podman info -f '{{.Host.Security.Rootless}}' | grep -q true ;;
        docker) docker info -f '{{.SecurityOptions}}' | grep -q rootless ;;
        *) echo >&2 \
            "Unknown OCI program \"$OCI_PROGRAM\", assuming rootless mode" ;;
    esac
}

# You can set your own OCI_PROGRAM, which assumes it runs in rootless mode.
if [ -z "${OCI_PROGRAM:-}" ]; then
    if command -v podman > /dev/null; then
        OCI_PROGRAM=podman
    elif command -v docker > /dev/null; then
        OCI_PROGRAM=docker
    fi
fi

if is_rootless; then
    generate_man_page
else
    generate_man_page --user "$(id -u):$(id -g)"
fi

if is_rootless; then
    generate_plain_text
else
    generate_plain_text --user "$(id -u):$(id -g)"
fi

# vim: tabstop=4 shiftwidth=4 textwidth=80 colorcolumn=80
