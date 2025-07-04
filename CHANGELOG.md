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
- Arrow keys are now handled properly when waiting for a key press after executing a `shell-wait` (`!`) command (#1956).
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

### Added

- Support for UNIX domain sockets (for communicating with the `lf` server) is added for Windows (#1637).
- Color and icon configurations now support the `target` keyword for symbolic links (#1644).
- A new option `roundbox` is added to use rounded corners when `drawbox` is enabled (#1653).
- A new option `watch` is added to allow using filesystem notifications to detect and display changes to files. This is an alternative to the `period` option, which polls the filesystem periodically for changes (#1667).
- Icons can now be colored independently of the filename (#1674).
- The `info` option now supports `perm`, `user` and `group` to display the permissions, user and group respectively for each file (#1799).
- A new option `showbinds` is added to toggle whether the keybinding hints are shown when a keybinding is partially typed (#1815).

### Fixed

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

## [r32](https://github.com/gokcehan/lf/releases/tag/r32)

### Changed

- The example script `etc/lfcd.cmd` is updated to use the `-print-last-dir` option instead of `-last-dir-path` (#1444). Similar changes have been made for `etc/lfcd.ps1` (#1491), `etc/lfcd.fish` (#1503), and `etc/lfcd.nu` (#1575).
- The documentation from `lf -doc` and the `doc` command is now generated from Markdown using `pandoc` (#1474).

### Added

- A new option `hidecursorinactive` is added to hide the cursor when the terminal is not focused (#965).
- A new special command `on-redraw` is added to be able to run a command when the screen is redrawn or when the terminal is resized (#1479).
- Options `cutfmt`, `copyfmt` and `selectfmt` are added to configure the indicator color for cut/copied/selected files respectively (#1540).
- `zsh` completion is added for the `lfcd` command (#1564).
- The file stat information now falls back to displaying user/group ID if looking up the user/group name fails (#1590).
- A new environment variable `lf_mode` is now exported to indicate which mode `lf` is currently running in (#1594).
- Default icons are added for Docker Compose files (#1626).

### Fixed

- Default value of `rulerfmt` option is now left-padded with spaces to visually separate it from the file stat information (#1437).
- Previews should now work for files containing lines with 65536 characters or more (#1447).
- Sixel previews should now work when using `lfcd` scripts (#1451).
- Colors and icons should now display properly for character device files (#1469).
- The selection file is now immediately synced to physical storage after writing to it (#1480).
- Timestamps are preserved when moving files across devices (#1482).
- Fix crash for `high` and `low` commands when `scrolloff` is set to a large value (#1504).
- Documentation is updated with various spelling and grammar fixes (#1518).
- Files beginning with a dot (e.g. `.gitignore`) are named correctly after `paste` if another file with the same name already exists (#1525).
- Prevent potential race condition when sorting directory contents (#1526).
- Signals are now handled properly even after receiving and ignoring `SIGINT` (#1549).
- The file stat information should now update properly after using the `cd` command to change to a directory for the first time (#1536).
- Previous error messages should now be cleared after a `mark-save`/`mark-remove` operation (#1544).
- Fix high CPU usage issue when viewing CryFS filesystems (#1607).
- Invalid entries in the `marks` and `tags` files now raise an error message instead of crashing (#1614).
- Startup time is improved on Windows (#1617).
- Sixel previews are now resized properly when the horizontal size of the preview window changes (#1629).
- The cut buffer is only cleared if the `paste` operation succeeds (#1652).
- The extension after `.` is ignored to set the cursor position when renaming a directory (#1664).
- The option `period` should not cause flickers in sixel previews anymore (#1666).

## [r31](https://github.com/gokcehan/lf/releases/tag/r31)

### Changed

- There has been some changes in the server protocol. Make sure to kill the old server process when you update to avoid errors (i.e. `lf -remote 'quit!'`).
- A new server command `query` is added to expose internal state to users (#1384). A new builtin command `cmds` is added to display the commands. The old builtin command `jumps` is now removed. The builtin commands `maps` and `cmaps` now use the new server command.
- Environment variable exporting for files and options are not performed anymore while previewing and cleaning to avoid various issues with race conditions (#1354). Cleaning program should now instead receive an additional sixth argument for the next file path to be previewed to allow comparisons with the previous file path. User options (i.e. `user_{option}`) are now exported whenever they are changed (#1418).
- Command outputs are now exclusively attached to `stderr` to allow printing the last directory or selection to `stdout` (#1399) (#1402). Two new command line options `-print-last-dir` and `-print-selection` are added to print the last directory and selection to `stdout`. The example script `etc/lfcd.sh` is updated to use `-print-last-dir` instead. Other `lfcd` scripts are also likely to be updated in the future to use the new method (patches are welcome).
- The option `ruler` is now deprecated in favor of its replacement `rulerfmt` (#1386). The new `rulerfmt` option is more capable (i.e. displays option values, supports colors and attributes, and supports optional fields) and more consistent with the rest of our options. See the documentation for more information.

### Added

- Modifier keys (i.e. control, shift, alt) with special keys (e.g. arrows, enter) are now supported for most combinations (#1248).
- A new option `borderfmt` is added to configure colors for pane borders (#1251).
- New `lf` specific environment variables, `LF_CONFIG_HOME` on Windows and `LF_CONFIG/DATA_HOME` on Unix, are now supported to set the configuration directory (#1253).
- Tilde (i.e. `~`) expansion is performed during completion to be able to use expanded tilde paths as command arguments (#1246).
- A new option `preserve` is added to preserve attributes (i.e. mode and timestamps) while copying (#1026).
- The file `etc/icons.example` is updated for nerd-fonts v3.0.0 (#1271).
- A new builtin command `clearmaps` is added to clear all default keybindings except for `read` (i.e. `:`) and `cmap` keybindings to be able to `:quit` (#1286).
- A new option `statfmt` is added to configure the status line at the bottom (#1288).
- A new option `truncatepct` is added to determine the location of truncation from the beginning in terms of percentage (#1029).
- A new option `dupfilefmt` is added to configure the names of duplicate files while copying (#1315).
- Shell scripts `etc/lf.nu` and `etc/lfcd.nu` are added to the repository to allow completion and directory change with Nushell (#1341).
- Sixels are now supported for previewing (#1211). A new option `sixel` is added to enable this behavior.
- A new configuration keyword `setlocal` is added to configure directory specific options (#1381).
- A new command line command `cmd-delete-word-back` (default `a-backspace` and `a-backspace2`) is added to use word boundaries when deleting a word backwards (#1409).

### Fixed

- Cursor positions in the directory should now be preserved after file operations that changes the directory (e.g. create or delete) (#1247).
- Option `reverse` should now respect to sort stability requirements (#1261).
- Backspace should not exit `filter` mode anymore (#1269).
- Truncated double width characters should not cause misalignment for the file information (#1272).
- Piping shell commands should not refresh the preview anymore (#1281).
- Cursor position should now update properly after a terminal resize (#1290).
- Directories should now be reloaded properly after a `delete` operation (#1292).
- Executable file completion should not add entries to the log file anymore (#1307).
- Blank input lines are now allowed in piping shell commands (#1308).
- Shell commands arguments on Windows should now be quoted properly to fix various issues (#1309).
- Reloading in a symlink directory should not follow the symlink anymore (#1327).
- Command `load` should not flicker image previews anymore (#1335).
- Asynchronous shell commands should now trigger `load` automatically when they are finished (#1345).
- Changing the value of `preview` option should now clear image previews (#1350).
- Cursor position in the status line at the bottom should now consider double width characters properly (#1348).
- Filenames should only be quoted for `cmd` on Windows to avoid quoting issues for `powershell` (#1371).
- Inaccessible files should now be included in the directory list and display their `lstat` errors in the status line at the bottom (#1382).
- Command line command `cmd-delete-word` should now add the deleted text to the yank buffer (#1409).

## [r30](https://github.com/gokcehan/lf/releases/tag/r30)

### Added

- A new builtin command `jumps` is addded to display the jump list (#1233).
- A new possible field `filter` is added to `ruler` option to display the filter indicator (#1223).

### Fixed

- Broken mappings for `bottom` command due to recent changes are fixed (#1240).
- Selecting a file does not scroll to bottom anymore (#1222).
- Broken builds on some platforms due to recent changes are fixed (#1168).

## [r29](https://github.com/gokcehan/lf/releases/tag/r29)

### Changed

- Three new options `cursoractivefmt`, `cursorparentfmt` and `cursorpreviewfmt` have been added (#1086) (#1106). The default style for the preview cursor is changed to underline. You can revert back to the old style with `set cursorpreviewfmt "\033[7m"`.
- An alternative boolean option syntax `set option true/false` is added in addition to the previous syntax `set option/nooption` (#758). If you have `set option true` in your configuration, then there is no need for any changes as it was already working as expected accidentally. If you have `set option false` in your configuration, then previously it was enabling the option instead accidentally but now it is disabling the option as intended. Any other syntax including `set option on/off` are now considered errors and result in error messages. Boolean option toggling `set option!` remains unchanged with no new alternative syntax added.
- Cursor is now placed at the file extension by default in rename prompts (#1162).
- The environment variable `VISUAL` is checked before `EDITOR` for the default editor choice (#1197).

### Added

- Mouse wheel events with the Control modifier have been bound to scrolling by default (#1051).
- Option values for `tagfmt` and `errorfmt` have been simplified to be able to avoid the reset sequence (#1086).
- Two default command line bindings for `<down>` and `<up>` have been added for `cmd-history-next` and `cmd-history-prev` respectively (#1112).
- A new command `invert-below` is added to invert all selections below the cursor (#1101). **This feature is currently experimental.**
- Two new commands `maps` and `cmaps` have been added to display the current list of bindings (#1146) (#1201).
- A new option `numberfmt` is added to customize line numbers (#1177).
- A new environment variable `lf_count` is now exported to use the count in shell commands (#1187).
- A new environment variable `lf` is now exported to be used as the executable path (#1176).
- An example `mkdir` binding is added to the example configuration (#1188).
- An example binding to show execution results is added to the example configuration (#1188).
- Commands `top` and `bottom` now accepts counts to move to a specific line (#1196).
- A new option `ruler` is added to customize the ruler information with a new addition for free disk space (#1168) (#1205).

### Fixed

- Example `lfcd` files have been made safer to be able to alias the commands as `lf` (#1049).
- Backspace should not exit from `rename:` mode anymore (#1060).
- Preview is now refreshed even if the selection does not change (#1074).
- Stale directory cache entry is now deleted during rename (#1138).
- File information is now updated properly after reloading (#1149).
- Window widths are now calculated properly when `drawbox` is enabled (#1150).
- Line number widths are now calculated properly when there are exactly 10 entries (#1151).
- Preview is not redrawn in async shell commands (#1164).
- A small delay is added before showing loading text in preview pane to avoid flickering (#1154).
- Hard-coded box drawing characters are replaced with Tcell constants to enable the fallback mechanism (#1170).
- Option `relativenumber` now shows zero in the current line (#1171).
- Completion is not stuck in an infinite loop anymore when a match is longer than the window width (#1183).
- Completion now inserts the longest match even if there is no word before the cursor (#1184).
- Command `doc` should now work even if `lf` is not in the `PATH` variable (#1176).
- Directory option changes should not crash the program anymore (#1204).
- Option `selmode` is now validated for the accepted values (#1206).
