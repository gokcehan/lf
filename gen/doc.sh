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

[ -z $version ] && version=$(git describe --tags)
[ -z $date ] && date=$(date +%F)

pandoc \
    --standalone \
    --from gfm --to man \
    --metadata=title:"LF" \
    --metadata=section:"1" \
    --metadata=date:"$date" \
    --metadata=footer:"$version" \
    --metadata=header:"DOCUMENTATION" \
    doc.md -o lf.1

pandoc \
    --standalone \
    --from gfm --to plain \
    doc.md -o doc.txt
