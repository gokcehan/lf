# LF

[Doc](doc.md)
| [Wiki](https://github.com/gokcehan/lf/wiki)
| [#lf:matrix.org](https://matrix.to/#/#lf:matrix.org) (with IRC bridge)

[![Go Build](https://github.com/gokcehan/lf/actions/workflows/go.yml/badge.svg)](https://github.com/gokcehan/lf/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gokcehan/lf)](https://goreportcard.com/report/github.com/gokcehan/lf)

`lf` (as in "list files") is a terminal file manager written in Go with a heavy inspiration from [`ranger`](https://github.com/ranger/ranger) file manager.
See [faq](https://github.com/gokcehan/lf/wiki/FAQ) for more information and [tutorial](https://github.com/gokcehan/lf/wiki/Tutorial) for a gentle introduction with screencasts.

![icons-and-border](https://github.com/user-attachments/assets/d5623462-05ab-4921-aeb3-d377d4732f9e)
![image-preview](https://github.com/user-attachments/assets/9ff42a21-19dd-42fa-8407-5d055aa9d561)
![new-features](https://github.com/user-attachments/assets/2a9a32aa-a764-438f-a810-e687e979dcee)

## Features

- Cross-platform (Linux, macOS, BSDs, Windows)
- Single binary without any runtime dependencies
- Fast startup and low memory footprint due to native code and static binaries
- Asynchronous IO operations to avoid UI locking
- Server/client architecture and remote commands to manage multiple instances
- Extendable and configurable with shell commands
- Customizable keybindings (vi and readline defaults)
- A reasonable set of other features (see the [documentation](doc.md))

## Non-Features

- Tabs or windows (better handled by window manager or terminal multiplexer)
- Builtin pager/editor (better handled by your pager/editor of choice)
- Builtin commands for file operations (better handled by the underlying shell tools including but not limited to `mkdir`, `touch`, `chmod`, `chown`, `chgrp`, and `ln`)

## Installation

See [packages](https://github.com/gokcehan/lf/wiki/Packages) for community maintained packages.

See [releases](https://github.com/gokcehan/lf/releases) for pre-built binaries.

Building from the source requires [Go](https://go.dev/).

On Unix:

```bash
env CGO_ENABLED=0 go install -ldflags="-s -w" github.com/gokcehan/lf@latest
```

On Windows `cmd`:

```cmd
set CGO_ENABLED=0
go install -ldflags="-s -w" github.com/gokcehan/lf@latest
```

On Windows `powershell`:

```powershell
$env:CGO_ENABLED = '0'
go install -ldflags="-s -w" github.com/gokcehan/lf@latest
```

## Usage

After the installation `lf` command should start the application in the current directory.

Run `lf -help` to see command line options.

Run `lf -doc` to see the [documentation](doc.md).

See [etc](etc) directory to integrate `lf` to your shell and/or editor.
Example configuration files along with example colors and icons files can also be found in this directory.

See [integrations](https://github.com/gokcehan/lf/wiki/Integrations) to integrate `lf` to other tools.

See [tips](https://github.com/gokcehan/lf/wiki/Tips) for more examples.

## Contributing

See [contributing](CONTRIBUTING.md) for guidelines.
