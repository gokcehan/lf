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

[ -z "${version:-}" ] && version=$(git describe --tags)
[ -z "${date:-}" ] && date=$(date +%F)

PANDOC_IMAGE=pandoc/minimal:3.7

generate_man_page() {
  docker run --rm -v "$PWD:/data" "$@" "$PANDOC_IMAGE" \
      --standalone \
      --from gfm --to man \
      --metadata=title:"LF" \
      --metadata=section:"1" \
      --metadata=date:"$date" \
      --metadata=footer:"$version" \
      --metadata=header:"DOCUMENTATION" \
      doc.md -o lf.1
  # Patch the TH man page command.
  sed -Ei '/^\.TH /{s/(([^"]*"){7})[^"]*(".*)/\1\3/}' lf.1
}

generate_plain_text() {
  docker run --rm -v "$PWD:/data" "$@" "$PANDOC_IMAGE" \
      --standalone \
      --from gfm --to plain \
      doc.md -o doc.txt
}

# If you get
# pandoc: lf.1: withFile: permission denied (Permission denied)
# try setting the "ROOTLESS" environment variable to a non-empty value.

if [ -z "${ROOTLESS:-}" ]; then
  generate_man_page --user "$(id -u):$(id -g)" # Not rootless.
else
  generate_man_page
fi

if [ -z "${ROOTLESS:-}" ]; then
  generate_plain_text --user "$(id -u):$(id -g)" # Not rootless.
else
  generate_plain_text
fi
