# LF

> Please note that this is an experimental file manager.
> One of the most dangerous pieces of software you can play with.
> You may accidentally lose your files or worse so use at your own risk.

> Likewise it is a work in progress.
> You will most likely come across some shameful bugs.
> Also some essentials may not have been implemented yet.

`lf` (as in "list files") is a terminal file manager written in Go.
It is heavily inspired by ranger with some missing and extra features.
Some of the missing features are deliberately ommited
since it is better if they are handled by external tools.

![multicol-screenshot](http://i.imgur.com/DaTUenu.png)
![singlecol-screenshot](http://i.imgur.com/p95xzUj.png)

## Features

- no external runtime dependencies (except for terminfo database)
- fast startup and low memory footprint (due to native code and static binaries)
- server/client architecture to share selection between multiple instances
- custom commands as shell scripts (hence any other language as well)
- sync (waiting and skipping) and async commands
- fully customizable keybindings

## Non-Features

- tabs or windows (handled by the window manager or the terminal multiplexer)
- built-in pager (handled by your pager of choice)

## May-Futures

- enchanced previews (image, pdf etc.)
- bookmarks
- colorschemes
- periodic refresh
- progress bar for file yank/delete paste

## Installation

You can build from the source using:

    go get -u github.com/gokcehan/lf

Currently there are no prebuilt binaries provided.

## Usage

After the installation `lf` command should start the application in the current directory.

See [tutorial](doc/tutorial.md) for an introduction to the configuration.

See [reference](doc/reference.md) for the list of keys, options and variables with their default values.

See [etc](etc) directory to integrate `lf` to your shell or editor.
