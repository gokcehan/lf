# LF

[Google Groups](https://groups.google.com/forum/#!forum/lf-fm)
| [Wiki](https://github.com/gokcehan/lf/wiki)
| [#lf](https://webchat.freenode.net/?channels=lf) (on Freenode)
| [#lf:matrix.org](https://matrix.to/#/#lf:matrix.org) (with IRC bridge)

[![Build Status](https://travis-ci.org/gokcehan/lf.svg?branch=master)](https://travis-ci.org/gokcehan/lf)
[![Go Report Card](https://goreportcard.com/badge/github.com/gokcehan/lf)](https://goreportcard.com/report/github.com/gokcehan/lf)
[![GoDoc](https://godoc.org/github.com/gokcehan/lf?status.svg)](https://godoc.org/github.com/gokcehan/lf)
[![Join the chat at https://gitter.im/lf-fm/Lobby](https://badges.gitter.im/lf-fm/Lobby.svg)](https://gitter.im/lf-fm/Lobby?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

> This is a work in progress. Use at your own risk.

`lf` (as in "list files") is a terminal file manager written in Go.
It is heavily inspired by ranger with some missing and extra features.
Some of the missing features are deliberately omitted since they are better handled by external tools.
See [faq](https://github.com/gokcehan/lf/wiki/FAQ) for more information.

![multicol-screenshot](http://i.imgur.com/DaTUenu.png)
![singlecol-screenshot](http://i.imgur.com/p95xzUj.png)

## Features

- Cross-platform (Linux, OSX, BSDs, Windows (partial))
- Single binary without any runtime dependencies (except for terminfo database)
- Fast startup and low memory footprint (due to native code and static binaries)
- Server/client architecture to share file selection between multiple instances
- Configuration with shell commands
- Customizable keybindings (vi and readline defaults)
- Preview filtering (for source highlight, archives, pdfs/images as text etc.)

## Non-Features

- Tabs or windows (handled by window manager or terminal multiplexer)
- Builtin pager/editor (handled by your pager/editor of choice)
- Image previews (cool but no standard available)

## May-Futures

- Bookmarks
- Mouse control

## Installation

See [releases](https://github.com/gokcehan/lf/releases) for pre-built binaries.

If you like to build from the source:

    go get -u github.com/gokcehan/lf

See [packages](https://github.com/gokcehan/lf/wiki/Packages) for community maintained packages.

## Usage

See [tutorial](https://github.com/gokcehan/lf/wiki/Tutorial) for a gentle introduction with screencasts.

After the installation `lf` command should start the application in the current directory.

Run `lf -help` to see command line options.

Run `lf -doc` to see the [documentation](https://godoc.org/github.com/gokcehan/lf).

See [etc](etc) directory to integrate `lf` to your shell or editor.
An example configuration file can also be found in this directory.

See [integrations](https://github.com/gokcehan/lf/wiki/Integrations) to integrate `lf` to other tools.

See [tips](https://github.com/gokcehan/lf/wiki/Tips) for more examples.

## Contributing

See [contributing](https://github.com/gokcehan/lf/wiki/Contributing) for guidelines.
