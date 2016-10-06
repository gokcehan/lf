# LF

[Google Groups](https://groups.google.com/forum/#!forum/lf-fm)

[![Build Status](https://travis-ci.org/gokcehan/lf.svg?branch=master)](https://travis-ci.org/gokcehan/lf)
[![Join the chat at https://gitter.im/lf-fm/Lobby](https://badges.gitter.im/lf-fm/Lobby.svg)](https://gitter.im/lf-fm/Lobby?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![GoDoc](https://godoc.org/github.com/gokcehan/lf?status.svg)](https://godoc.org/github.com/gokcehan/lf)

> This is a work in progress. Use at your own risk.

`lf` (as in "list files") is a terminal file manager written in Go.
It is heavily inspired by ranger with some missing and extra features.
Some of the missing features are deliberately omitted
since it is better if they are handled by external tools.

![multicol-screenshot](http://i.imgur.com/DaTUenu.png)
![singlecol-screenshot](http://i.imgur.com/p95xzUj.png)

## Features

- Cross-platform (Linux, OSX, BSDs, Windows (partial))
- Single binary without any runtime dependencies (except for terminfo database)
- Fast startup and low memory footprint (due to native code and static binaries)
- Server/client architecture to share file selection between multiple instances
- Custom commands as shell commands
- Sync (waiting and non-waiting) and async commands
- Fully customizable keybindings
- Preview filtering (for source highlight, archives, pdfs/images as text etc.)

## Non-Features

- Tabs or windows (handled by the window manager or the terminal multiplexer)
- Built-in pager (handled by your pager of choice)
- Image previews (cool but no standard available)
- Periodic refresh (use explicit renew instead)

## May-Futures

- Bookmarks
- Colorschemes

## Installation

See [releases](https://github.com/gokcehan/lf/releases) for pre-built binaries.

If you like to build from the source:

    go get -u github.com/gokcehan/lf

## Usage

After the installation `lf` command should start the application in the current directory.

Run `lf -help` to see command line options.

Run `lf -doc` to see the [documentation](https://godoc.org/github.com/gokcehan/lf).

See [etc](etc) directory to integrate `lf` to your shell or editor.
An example configuration file can also be found in this directory.
