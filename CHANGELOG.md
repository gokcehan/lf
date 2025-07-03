# Changelog

All changes observable to end users should be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and should contain the following sections for each release:

- `Changed`
- `Added`
- `Fixed`

## [r35](https://github.com/gokcehan/lf/releases/tag/r35)

### Added

- Support is added for displaying underline styles (#1896).
- Support is added for displaying underline colors (#1933).
- A new subcommand `files` is added to the `query` server command to list the files in the current directory as displayed in `lf` (#1949).
- A new `tty-write` command is added for sending escape sequences to the terminal (#1961). **Writing directly to `/dev/tty` is not recommended as it not synchronized and can interfere with drawing the UI.**

### Fixed

- The `trash` command in `lfrc.example` now verifes if the trash directory exists before moving files there (#1918).
- `lf` should no longer crash if it fails to open the log file (#1937).
- Arrow keys are now handled properly when waiting for a key press after executing a shell-wait (`!`) command (#1956).
- The `previewer` script is now only invoked for the current directory (instead of all directories), when starting `lf` with `dirpreviews` enabled (#1958).

## [r34](https://github.com/gokcehan/lf/releases/tag/r34)

### Changed

- The `autoquit` option is now enabled by default (#1839).

### Added

- A new option `locale` is added to sort files based on the collation rules of the provided locale (#1818). **This feature is currently experimental.**
- A new `on-init` hook command is added to allow triggering custom actions when `lf` has finished initializing and connecting to the server (#1838).

### Fixed

- The background color now renders properly when displaying filenames (#1849).
- A bug where the `on-quit` hook command causes an infinite loop has been fixed (#1856).
- File sizes now display correctly after being copied when `watch` is enabled (#1881).
- Files are now automatically unselected when removed by an external process, when `watch` is enabled (#1901).

## [r33](https://github.com/gokcehan/lf/releases/tag/r33)

### Changed

- The `globsearch` option, which previously affected both searching and filtering, now affects only searching. A new `globfilter` option is introduced to enable globs when filtering, and acts independently from `globsearch` (#1650).
- The `hidecursorinactive` option is replaced by the `on-focus-gained` and `on-focus-lost` hook commands. These commands can be used to invoke custom behavior when the terminal gains or loses focus (#1763).
- The `ruler` option (deprecated in favor of `rulerfmt`) is now removed (#1766).
- Line numbers from the `number` and `relativenumber` options are displayed in the main window only, instead of all windows (#1789).

### Fixed

- Support for UNIX domain sockets (for communicating with the `lf` server) is added for Windows (#1637).
- Color and icon configurations now support the `target` keyword for symbolic links (#1644).
- A new option `roundbox` is added to use rounded corners when `drawbox` is enabled (#1653).
- A new option `watch` is added to allow using filesystem notifications to detect and display changes to files. This is an alternative to the `period` option, which polls the filesystem periodically for changes (#1667).
- Icons can now be colored independently of the filename (#1674).
- The `info` option now supports `perm`, `user` and `group` to display the permissions, user and group respectively for each file (#1799).
- A new option `showbinds` is added to toggle whether the keybinding hints are shown when a keybinding is partially typed (#1815).

### Added

- Sorting by extension is fixed for hidden files (#1670).
- The `on-quit` hook command is now triggered when the terminal is closed (#1681).
- Previews no longer flicker when deleting files (#1691).
- Previews no longer flicker when directories are being reloaded (#1697).
- `lfcd.nu` now runs properly without raising errors (#1728).
- Image previews (composed of ASCII art) containing long lines should now display properly (#1737).
- The performance is improved when copying files (#1749).
- `lfcd.cmd` now handles directories with special characters (#1772).
- Icon colors are no longer clipped when displaying in Windows Terminal (#1777).
- The file stat info is now cleared when changing to an empty directory (#1808).
- Error messages are cleared when opening files (#1809).
