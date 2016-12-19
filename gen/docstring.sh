#!/bin/sh
# Generates `docstring.go` having `genDocString` variable with `go doc` output.
#
# This script is called in `doc.go` using `go generate` to embed the
# documentation inside the binary in order to show it on request with `-doc`
# command line flag. Thus the same documentation is used for online and
# terminal display.

tmp=gen/docstring.go

echo "// DO NOT EDIT! (AUTO-GENERATED)" >> $tmp
echo                                    >> $tmp
echo "package main"                     >> $tmp
echo                                    >> $tmp
echo "var genDocString = \`"            >> $tmp
go doc | tr "\`" "'"                    >> $tmp
echo "\`"                               >> $tmp

mv $tmp docstring.go
