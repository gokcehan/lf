#!/bin/sh

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
    doc.md -o doc.1

pandoc \
    --standalone \
    --from gfm --to plain \
    doc.md -o doc.txt
