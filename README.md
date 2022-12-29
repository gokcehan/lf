# LF-sixel

This is a fork of [lf](https://github.com/gokcehan/lf) with support for [sixel](https://wikipedia.org/wiki/Sixel) graphics.

Sixel support requires a sixel-capable terminal and is only available on Unix systems (only tested on Linux but BSD distributions and MacOS should work).

![sixel-screenshot](https://i.imgur.com/5RGKrw2.png)

## For lf users moving to lf-sixel...

Sixel sequences can be passed directly to lf for displaying, cleaning, and caching. An example config based on cirala's lfimg can be found [here](https://github.com/horriblename/lfimg-sixel). For those managing their own previewer:

- use `chafa "$1" -f sixel -s "$(($2-2))x$3" | sed 's/#/\n#/g'` in your previewer script
- we need to pipe chafa's output into `sed 's/#/\n#/g'` as sixels can be too long for lf-sixel to handle internally
- remember to use `exit 0`, any other exit code indicates that the user's previewer and cleaner scripts will handle cleaning and caching.
- you **don't** need a cleaner script for sixel previewing

# LF

[Google Groups](https://groups.google.com/forum/#!forum/lf-fm)
| [Wiki](https://github.com/gokcehan/lf/wiki)
| [#lf](https://web.libera.chat/#lf) (on Libera.Chat)
| [#lf:matrix.org](https://matrix.to/#/#lf:matrix.org) (with IRC bridge)

[![Go Report Card](https://goreportcard.com/badge/github.com/gokcehan/lf)](https://goreportcard.com/report/github.com/gokcehan/lf)
[![Go Reference](https://pkg.go.dev/badge/github.com/gokcehan/lf.svg)](https://pkg.go.dev/github.com/gokcehan/lf)

> This is a work in progress. Use at your own risk.

`lf` (as in "list files") is a terminal file manager written in Go with a heavy inspiration from ranger file manager.
See [faq](https://github.com/gokcehan/lf/wiki/FAQ) for more information and [tutorial](https://github.com/gokcehan/lf/wiki/Tutorial) for a gentle introduction with screencasts.

![multicol-screenshot](http://i.imgur.com/DaTUenu.png)
![singlecol-screenshot](http://i.imgur.com/p95xzUj.png)

## Features

- Cross-platform (Linux, macOS, BSDs, Windows)
- Single binary without any runtime dependencies
- Fast startup and low memory footprint due to native code and static binaries
- Asynchronous IO operations to avoid UI locking
- Server/client architecture and remote commands to manage multiple instances
- Extendable and configurable with shell commands
- Customizable keybindings (vi and readline defaults)
- A reasonable set of other features (see the [documentation](https://pkg.go.dev/github.com/gokcehan/lf))

## Non-Features

- Tabs or windows (better handled by window manager or terminal multiplexer)
- Builtin pager/editor (better handled by your pager/editor of choice)

## Installation

<!-- See [packages](https://github.com/gokcehan/lf/wiki/Packages) for community maintained packages. -->

See [releases](https://github.com/horriblename/lf/releases) for pre-built binaries.

Building from the source requires [Go](https://go.dev/).

On Unix (Go version < 1.17):

```bash
env CGO_ENABLED=0 GO111MODULE=on go get -u -ldflags="-s -w" github.com/horriblename/lf
```

On Unix (Go version >= 1.17):

```bash
env CGO_ENABLED=0 go install -ldflags="-s -w" github.com/horriblename/lf@latest
```

On Windows `cmd` (Go version < 1.17):

```cmd
set CGO_ENABLED=0
set GO111MODULE=on
go get -u -ldflags="-s -w" github.com/horriblename/lf
```

On Windows `cmd` (Go version >= 1.17):

```cmd
set CGO_ENABLED=0
go install -ldflags="-s -w" github.com/horriblename/lf@latest
```

On Windows `powershell` (Go version < 1.17):

```powershell
$env:CGO_ENABLED = '0'
$env:GO111MODULE = 'on'
go get -u -ldflags="-s -w" github.com/horriblename/lf
```

On Windows `powershell` (Go version >= 1.17):

```powershell
$env:CGO_ENABLED = '0'
go install -ldflags="-s -w" github.com/horriblename/lf@latest
```

## Usage

After the installation `lf` command should start the application in the current directory.

Run `lf -help` to see command line options.

Run `lf -doc` to see the [documentation](https://pkg.go.dev/github.com/gokcehan/lf).

See [etc](etc) directory to integrate `lf` to your shell and/or editor.
Example configuration files along with example colors and icons files can also be found in this directory.

See [integrations](https://github.com/gokcehan/lf/wiki/Integrations) to integrate `lf` to other tools.

See [tips](https://github.com/gokcehan/lf/wiki/Tips) for more examples.

## Contributing

See [contributing](https://github.com/gokcehan/lf/wiki/Contributing) for guidelines.
