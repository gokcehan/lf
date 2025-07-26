# Changelog

All changes observable to end users should be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and should contain the following sections for each release:

- `Changed`
- `Added`
- `Fixed`

## r37 (Unreleased)

### Changed

- The default paths of files read by `lf` is changed on Windows, to separate configuration files from data files (#2051).
  - Configuration files (`lfrc`/`colors`/`icons`) are now stored in `%APPDATA%`, which can be overridden by `%LF_CONFIG_HOME%`.
  - Data files (`files`/`marks`/`tags`/`history`) are now stored in `%LOCALAPPDATA%`, which can be overridden by `%LF_DATA_HOME%`.
- The change for following symbolic links when tagging files from the previous release has been reverted (#2055). The previous change made it impossible to tag symbolic links separately from their targets, and also caused `lf` to run slowly in some cases.
- The existing `globfilter` and `globsearch` options are now deprecated in favor of the new `filtermethod` and `searchmethod` options, which support regex patterns (#2058).
  - `set globfilter true` should be replaced by `set filtermethod glob`.
  - `set globsearch true` should be replaced by `set searchmethod glob`.
- File sizes are now displayed using binary units (e.g. `1.0K` means 1024 bytes, not 1000 bytes) (#2062). The maximum width for displaying the file size has been increased from four to five characters.

### Added

- `dircounts` are now respected when sorting by size (#2025).
- The `info` and `sortby` options now support `btime` (file creation time) (#2042). This depends on support for file creation times from the underlying system.
- The selection in Visual mode now follows wrapping when `wrapscan`/`wrapscroll` is enabled (#2056).
- Input pasted from the terminal is now ignored while in Normal mode (#2059). This prevents pasted content from being treated as keybindings, which can result in dangerous unintended behavior.
- The Command-line mode completion now supports keywords for the `selmode` and `sortby` options (#2061).

### Fixed

- `dircounts` are now automatically populated after enabling it (#2049).
- A bug where directories are unsorted after reloading when `dircache` is disabled is now fixed (#2050).

## [r36](https://github.com/gokcehan/lf/releases/tag/r36)

### Changed

- Tagging symbolic links now affects the target instead of the symbolic link itself. This mimics the behavior in `ranger` (#1997).
- The experimental command `invert-below` has been removed in favor of the newly added support for Visual mode (#2021).

### Added

- A new placeholder `%P` representing the scroll percentage is added to the `rulerfmt` option (#1985).
- A new `on-load` hook command is added, which is triggered when files in a directory are loaded in `lf` (#2010).
- The `info` option now supports `custom`, allowing users to display custom information for each file (#2012). The custom information should be added by the user via the `addcustominfo` command. Sorting by the custom information is also supported (#2019).
- Support for `visual-mode` has now been added (#2021) (#2035). This includes the following changes:
  - A new command `visual` (default `V`) can be used to enter Visual mode.
  - A new command `visual-change` (default `o` in Visual mode) can be used to swap the positions of the cursor and anchor (start of the visual selection).
  - A new command `visual-accept` (default `V` in Visual mode) can be used to exit Visual mode, adding the visual selection to the selection list.
  - A new command `visual-discard` (default `<esc>` in Visual mode) can be used to exit Visual mode, without adding the visual selection to the selection list.
  - A new command `visual-unselect` can be used to exit Visual mode, removing the visual selection from the selection list.
  - The existing `map` command now adds keybindings for both Normal and Visual modes. Two new commands `nmap` and `vmap` are added which can be used to add keybindings for only Normal or Visual mode respectively.
  - Two new commands `nmaps` and `vmaps` are added to display the list of keybindings in Normal and Visual mode respectively. These, along with the existing `maps` and `cmaps` commands, now display an extra column indicating the mode for which the keybindings apply to.
  - A new option `visualfmt` is added to customize the appearance of the visual selection.
  - Two new placeholders `%m` and `%M` are added to `statfmt` to display the mode in the status line. Both will display `VISUAL` when in Visual mode, however in Normal mode `%m` will display as a blank string while `%M` will display `NORMAL`.
  - A new placeholder `%v` is added to `rulerfmt` which displays the number of files in the Visual selection. This is included in the default setting for `rulerfmt`.
  - The `lf_mode` environment variable will now be set to `visual` while in Visual mode.
  - The environment variable `$fv` is now exported to shell commands, which lists the files in the visual selection.
- A `CHANGELOG.md` file has been added to the repo (#2027). This will be updated to describe `Changed`, `Added` and `Fixed` functionality for each new release.

### Fixed

- Displaying sixel images now uses the screen locking API in Tcell, which reduces flickering in the UI (#1943).
- The `cmd-history` command is now ignored outside of Normal or Command-line mode, to prevent accidentally escaping out of other modes (#1971).
- A potential crash when using the `cmd-delete-word-back` command is fixed (#1976).
- The `preserve` option now applies to directories in addition to files when copying. This includes preserving `timestamps` (#1979) and `mode` (#1981).
- The `lfrc.ps1.example` example config file is updated to include PowerShell equivalents for the default commands and keybindings (#1989).
- Quoting for the `lf` environment variable is fixed for PowerShell users (#1990).
- `tempmarks` are no longer cleared after the `sync` command is called (#1996).
- The file stat information is no longer displayed during the execution of a `shell-pipe` command even if the file is updated (#2002).
- Directories are now reloaded properly if any component in the current path is renamed (#2005).
- Write updates for the log file are now ignored when `watch` is enabled. This helps to reduce notification spam and potential of infinite loops (#2015).
- Attempting to `cut`/`copy` files into a directory without execute permissions no longer causes `lf` to crash, and an error message will be displayed instead (#2024).

## [r35](https://github.com/gokcehan/lf/releases/tag/r35)

### Added

- Support is added for displaying underline styles (#1896).
- Support is added for displaying underline colors (#1933).
- A new subcommand `files` is added to the `query` server command to list the files in the current directory as displayed in `lf` (#1949).
- A new `tty-write` command is added for sending escape sequences to the terminal (#1961). **Writing directly to `/dev/tty` is not recommended as it not synchronized and can interfere with drawing the UI.**

### Fixed

- The `trash` command in `lfrc.example` now verifies if the trash directory exists before moving files there (#1918).
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

- Support for Unix domain sockets (for communicating with the `lf` server) is added for Windows (#1637).
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
- Filenames should only be quoted for `cmd` on Windows to avoid quoting issues for PowerShell (#1371).
- Inaccessible files should now be included in the directory list and display their `lstat` errors in the status line at the bottom (#1382).
- Command line command `cmd-delete-word` should now add the deleted text to the yank buffer (#1409).

## [r30](https://github.com/gokcehan/lf/releases/tag/r30)

### Added

- A new builtin command `jumps` is added to display the jump list (#1233).
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

## [r28](https://github.com/gokcehan/lf/releases/tag/r28)

### Changed

- Extension matching for colors and icons are now case insensitive (#908).

### Added

- Three new commands `high`, `middle`, and `low` are added to move the current selection relative to the screen (#824).
- Backspace on empty prompt now switches to Normal mode (#836).
- A new `history` option is now added to be able to disable history (#866).
- A new special expansion `%S` spacer is added for `promptfmt` to be able to right align parts (#867).
- A new command-line command `cmd-menu-accept` is now added to accept the currently selected match (#934).
- Command-line commands should now be shown in completion for `map` and `cmap` (#934).
- Italic escape codes should now be working in previews (#936).
- Position and size information are now also passed to the `cleaner` script as arguments (#945).
- A new option `dirpreviews` is now added to also pass directories to the `previewer` script (#842).
- A new option `selmode` is now added to be able to limit the selection to the current directory (#849).
- User defined options with `user_` prefix are now supported (#865).
- Adding or removing `$`/`%`/`!`/`&` characters in `:` mode should now change the mode accordingly (#960).
- A new special command `on-select` is now added to be able to run a command after the selection changes (#864).
- Mouse support is extended to be able to click filenames for selection and opening (#963).
- Two new environment variables `lf_width` and `lf_height` are now exported for shell commands.

### Fixed

- Option `tagfmt` can now be changed properly.
- User name, group name, and link count should now be displayed as before where available (#829).
- Tagging files with colons in their names should now work as expected (#857).
- Some multibyte characters should now be handled properly for completion (#934).
- Menu completion for a file in a subdirectory should now be working properly (#934).
- File completion should now be escaped properly in menu completion (#934).
- First use of `cmd-menu-complete-back` should now select the last completion as expected (#934).
- Broken symlinks should now be working properly in completion (#934).
- Files with stat errors should now be skipped properly in completion (#934).
- Empty search with `incsearch` option should now be handled properly (#944).
- History position is now also reset when leaving the command line (#953).
- Mouse drag events are now ignored properly to avoid command repetition (#962).
- Environment variables `HOME` and `USER` should now be used as fallback for locations on some systems (#972).
- File information is now displayed in the status line at first launch when there are no errors in the configuration file (#994).

## [r27](https://github.com/gokcehan/lf/releases/tag/r27)

### Changed

- Creation of log files are now disabled by default. Instead, a new command line option `-log` is provided.
- `copy` selections are now kept after `paste` (#745). You can use `map p :paste; clear` to get the old behavior.
- The socket file is now created in `XDG_RUNTIME_DIR` when set, with a fallback to the temporary directory otherwise.
- Directory counting with `dircounts` option is moved from UI drawing to directory reading to be run asynchronously without locking the UI. With this change, manual `reload` commands might be necessary when `dircounts` is changed at runtime. Indicators for errors are changed to `!` instead of `?` to distinguish them from missing values.
- The default icons are now replaced with ASCII characters to avoid font issues.

### Added

- Files and options are now exported for `previewer` and `cleaner` scripts. For `cleaner` scripts, this can be used to detect if the file selection is changed or not (e.g. `$1 == $f`) and act accordingly (e.g. skip cleaning).
- A new `tempmarks` option is added to set some marks as temporary (#744).
- The pattern `*filename` is added for colors and icons.
- A new `calcdirsize` command is added to calculate directory sizes (#750).
- Two new options `infotimefmtnew` and `infotimefmtold` are added to configure the time format used in `info` (#751).
- Two new commands `jump-next` (default `]`) and `jump-prev` (default `[`) are added to navigate the jumplist (#755).
- Colors and icons file support is now added to be able to configure without environment variables. Example colors and icons files are added to the repository under `etc` directory. See the documentation for more information.
- For Windows, an example `open` command is now provided in the PowerShell example configuration (#765).
- Two new commands `scroll-up` (default `<c-y>`) and `scroll-down` (default `<c-e>`) are added to be able to scroll the file list without moving (#764).
- A new special command `on-quit` is added to be able to run a command before quitting.
- Two new commands `tag` and `tag-toggle` (default `t`) are now added to be able to tag files (#791).

### Fixed

- `Chmod` calls in the codebase are now removed to avoid TOC/TOU exploits. Instead, file permissions are now set at file creation.
- Socket and log files are now created with only user permissions.
- On Windows, `PWD` variable is now quoted properly.
- Shell commands `%` and `&` are now run in a separate process group (#753).
- Navigation initialization is now delayed after the evaluation of configuration files to avoid startup races and redundant loadings (#759).
- The error message shown when the current working directory does not exist at startup is made more clear.
- Trailing slashes in `PWD` variable are now handled properly.
- Files with `stat` errors are now skipped while reading directories.

## [r26](https://github.com/gokcehan/lf/releases/tag/r26)

### Fixed

- On Windows, input handling is properly resumed after shell commands.

## [r25](https://github.com/gokcehan/lf/releases/tag/r25)

### Added

- A new `dironly` option is added to only show directories and hide regular files (#669).
- A new `dircache` option is added to disable caching of directories (#673).
- Two new commands `filter` and `setfilter` is added along with a new option `incfilter` and a `promptfmt` expansion `%F` to implement directory filtering feature (#667).
- A new special command `pre-cd` is added to run a command before a directory is changed (#685).
- `cmap` command now accepts all expressions similar to `map` (#686).

### Fixed

- Marking a symlink directory should now save the symlink path instead of the target path (#659).
- A number of crashes have been fixed when the `hidden` option is changed.

## [r24](https://github.com/gokcehan/lf/releases/tag/r24)

### Fixed

- Data directory is automatically created before the selection file is written.
- An error is returned for remote commands when the given ID is not connected to the server.
- Prompts longer than the width should not crash the program anymore.

## [r23](https://github.com/gokcehan/lf/releases/tag/r23)

### Changed

- There has been some changes in the server protocol. Make sure to kill the old server process when you update to avoid errors.
- Server `load` and `save` commands are now removed. Instead a local file is used to record file selections (e.g. `~/.local/share/lf/files`). See the documentation for more information.
- Clients are now disconnected from server on quit. The old server `quit` command is renamed to `quit!` to act as a force quit by closing connected client connections first. A new `quit` command is added to only quit when there are no connected clients left.

### Added

- A new `autoquit` option is added to automatically quit the server when there are no connected clients left. This option is disabled by default to keep the old behavior. This is added as an option to avoid respawning server repeatedly when there is often a single client involved but more clients are spawned from time to time.
- A new `-single` command line flag is added to avoid spawning and/or connecting to server on startup. Remote commands would not work in this case as the client does not connect to a server. Local versions of internal `load` and `sync` commands are implemented properly.
- Errors for remote commands are now also shown in the output in addition to the server log file.
- Bright ANSI color escape codes (i.e. 90-97 and 100-107) are now supported.

### Fixed

- Lookahead size for escape codes are increased to recognize longer escape codes used in some image previewers.
- The file preview cache is invalidated when the terminal height changes to fill the screen properly.
- The file preview cache is invalidated when the `drawbox` option changes and true image previews should be triggered to be drawn at updated positions.
- A crash scenario is fixed when `hidden` option is changed.
- Pane widths should now be calculated properly when big numbers are used in `ratios` (#622).
- The special bookmark `'` is now preserved properly after `sync` commands (#624).
- On some platforms, a bug has been fixed on the Tcell side to avoid an extra key press after terminal suspend/resume and the Tcell version used in `lf` is bumped accordingly to include the fix.
- The prompt line should now scroll accordingly when the text is wider than the screen.
- Text width in the prompt line should now be calculated properly when non-ASCII characters are involved.
- Erase line escape codes (i.e. `\033[K`) used in some command outputs should now be ignored properly.

## [r22](https://github.com/gokcehan/lf/releases/tag/r22)

### Added

- A new `-config` command line flag is added to use a custom config file path (#587).
- The current working directory is now exported as `PWD` environment variable (#591). Subshells in symlink directories should now start in their own paths properly.
- The initial working directory is now exported as `OLDPWD` environment variable.
- A new `shellflag` option is added to customize the shell flag used for passing commands (i.e. default `-c` for Unix and `/c` for Windows).
- Using the command `cmd-enter` during `find` and `find-back` now jumps to the first match (#605).
- A new `waitmsg` option is added to customize the prompt message after `shell-wait` commands (i.e. default `Press any key to continue`) (#604).

### Fixed

- A regression bug is fixed to print a newline in the prompt message properly after `shell-wait` commands.
- A regression bug is fixed to avoid CPU stuck at 100% when the terminal is closed unexpectedly.
- A regression bug is fixed to make shell commands use the alternate screen properly and keep the terminal history after quitting.
- Enter keypad terminfo sequence is now sent on startup so the `delete` key should be recognized properly in `st` terminal.

## [r21](https://github.com/gokcehan/lf/releases/tag/r21)

### Changed

- `cut` and `copy` do not follow symlinks anymore. Broken symlinks can now be selected for the `cut` and `copy` commands (#581).

### Added

- User name, group name, and hard link counts are now shown in the status line at the bottom when available.
- Number of selected, copied, and cut files are now shown in the ruler at the bottom when they are non-zero.
- Hard-coded shell commands with `stty` (Unix) and `pause` (Windows) to implement the `Press any key to continue` behavior are now implemented properly with a Go terminal handling library. With this change, the requirement for a POSIX compatible shell for `shell` option is now dropped and other shells can be used.

### Fixed

- A longstanding issue regarding UI suspend/resume for shell commands in MacOS is now fixed in Tcell.
- Renaming a symlink to its target or a symlink to another with the same target should now be handled properly (#581).
- Autocompletion in a directory containing a broken symlink should now work as intended (#581).
- Setting `shellopts` to empty in the configuration file should not pass an extra empty argument to shell commands anymore.
- Previously given tip to trap `SIGPIPE` in the preview script to enable caching is now updated in the documentation. Trapping the signal in the preview script avoids sending the signal to the program when enough lines are read. This may result in reading redundant lines especially for big files. The recommended method is now to add a trailing `|| true` to each command exiting with a non-zero return code after a `SIGPIPE`.

## [r20](https://github.com/gokcehan/lf/releases/tag/r20)

### Added

- A new `mouse` option is added to enable mouse events. This option is disabled by default to leave mouse events to the terminal. Also unbound mouse events when `mouse` is enabled should now show an `unknown mapping` error in the message line.

### Fixed

- Newline characters in the output of `%` commands should no longer shift the content up which was a bug introduced in the previous release due to a fix to handle combining characters in texts.
- Redundant preview loadings for the `search` and `find` commands are now avoided (#569).
- Scanner now only considers ASCII characters for spaces and digits which should avoid unexpected splits in some non-ASCII inputs.

## [r19](https://github.com/gokcehan/lf/releases/tag/r19)

### Changed

- Changes have been made to enable the use of true image previews. See the documentation and the previews wiki page for more information.
  - Non-zero exit codes should now make the preview volatile to avoid caching. Programs that may not behave well to `SIGPIPE` may trigger this behavior unintentionally. You may trap `SIGPIPE` in your preview script to get the old behavior.
  - Preview scripts should now get as arguments the current file path, width, height, horizontal position, and vertical position. Note that height is still passed as an argument but its order is changed.
  - A new `cleaner` option is added to set the path to a file to be executed when the preview is changed.
  - Redundant preview loadings for movement commands are now avoided.
- Expansion `%w` in `promptfmt` is changed back to its old behavior without a trailing separator. Instead, a new expansion `%d` is added with a trailing separator (#545). Expansion `%w` is meant to be used to display the current working directory, whereas `%d%f` is meant to be used to display the current file.
- A new `LF_COLORS` environment variable is now checked to be able to make `lf` specific configurations. Also, environment variables for colors are now read cumulatively starting from the default behavior (i.e. default, `LSCOLORS`, `LS_COLORS`, `LF_COLORS`).

### Added

- Full path, dir name, file name, and base name matching patterns are added to colors and icons. See the updated documentation for more information.
- PowerShell keybinding example has been added to `etc/lfcd.ps1` (#532).
- PowerShell autocompletion script has been added as `etc/lf.ps1` (#535).
- Multiple `-command` flags can now be given (#552).
- Basic mouse support has been added. Mouse buttons (e.g. `<m-1>` for primary button, `<m-2>` for secondary button, `<m-3>` for middle button etc.) and mouse wheels (e.g. `<m-up>` for wheel up, `<m-down>` for wheel down etc.) can be used in bindings.
- Commands `top` and `bottom` are now allowed in `cmap` mappings in addition to movement commands.

### Fixed

- Extension sorting should now handle extensions with different lengths properly (#539).
- Heuristic used to show `info` should now take into account the `number` and `icons` options properly.
- The environment variable `id` is now set to the process ID instead to avoid two clients getting the same ID when launched at the same time (#550).
- Unicode combining characters in texts should now be displayed properly.

## [r18](https://github.com/gokcehan/lf/releases/tag/r18)

### Changed

- The `ignorecase` and `ignoredia` options should now also apply to sorting in addition to searching.
- The `ignoredia` option is now enabled by default to be consistent with `ignorecase`.
- The terminal UI library Tcell has been updated to version 2. This version highlights adding 24-bit true colors on Windows and better support for colors on Unix. The environment variable `TCELL_TRUECOLOR` is not required anymore so that terminal themes and true colors can be used at the same time.
- The deprecated option `color256` is now removed.

### Added

- Two new command line commands `cmd-menu-complete` and `cmd-menu-complete-back` are added for completion menu cycling (#482).
- Simple configuration files for Windows `etc/lfrc.cmd.example` and `etc/lfrc.ps1.example` are now added to the repository.
- Bash completion script `etc/lf.bash` is now added to the repository.
- Time formats in `info` option should now show the year instead of `hh:mm` for times older than the current year.

### Fixed

- Signals `SIGHUP`, `SIGQUIT`, and `SIGTERM` should now quit the program properly.
- Setting `info` to an empty value should not print errors to the log file anymore.
- Natural sorting is optimized to work faster using less memory.
- Files and directories that incorrectly show modification times in the future (e.g. Linux builtin exFAT driver) should not cause CPU hogging anymore.
- The keybinding example in `etc/lfcd.fish` is now updated to avoid hanging in shell commands.
- Using the `bottom` command immediately after startup should not crash the program anymore.
- Changing sorting options during sorting operations should not crash the program anymore.
- Output in `shell-pipe` commands now uses lazy redrawing so that verbose commands should not block the program anymore.
- The server is now daemonized properly on Unix so that it is not killed anymore when the controlling terminal is killed (#517).

## [r17](https://github.com/gokcehan/lf/releases/tag/r17)

### Changed

- The terminal UI library has been changed from Termbox to Tcell as the former has been unmaintained for a while (#439). Some of the changes are listed below, though the list may not be complete as this is a relatively big change.
  - Some special key names are changed to be consistent with the Tcell documentation (e.g. `<bs>` renamed to `<backspace>`). On the other hand, there are also additional keybindings that were not available before (e.g. `<backtab>` for <kbd>Shift+Tab</kbd>). You can either check the Tcell documentation for the list of keys or hit the key combination in `lf` to read the name of the key from the `unknown mapping` error message.
  - 24-bit true colors are now supported on Unix systems. See the updated documentation for more information. There is an ongoing version 2.0 of Tcell in development that we plan to switch to once it becomes stable and it is expected to add support for true colors in Windows consoles as well.
  - Additional platforms are now supported and the list of pre-built binaries provided are updated accordingly.
  - Wide characters are now displayed properly in Windows consoles.

### Added

- Descriptions of commands and options are now added to the documentation. Undocumented behaviors should now be considered documentation bugs and they can be reported.
- Keys are now evaluated with a lazy drawing approach so `push` commands to set the prompt and pasting something to the command line should feel instantaneous.

### Fixed

- Corrupted history files should no longer crash the program.
- The server now only listens connections from `localhost` on Windows so firewall permissions are not required anymore.
- `push` commands that change the operation mode should now work consistently as expected.
- Loading directories should now display the previous file list if any, which was a regression due to a bug fix in a previous release.
- `shell-pipe` commands should now automatically update previews when necessary.
- Errors from failed shell commands should not be overwritten by file information anymore.
- The server can now also be started automatically when the program is called with a relative path, which was a regression due to a bug fix in a previous release (#463).
- Environment variables are now exported automatically for preview scripts without having to call a shell command first (#468).
- The `<esc>` key can now be bound to be used on its own, instead of escaping a keybinding combination, which was a regression due to a bug fix in a previous release (#475).
- Changing the `hiddenfiles` option should now automatically trigger directory updates when necessary.

## [r16](https://github.com/gokcehan/lf/releases/tag/r16)

### Added

- Option values are now available in shell commands as environment variables with a prefix of `lf_` (e.g. `$lf_hidden`, `$lf_ratios`) (#448).

### Fixed

- Directories containing internal Windows links that show permission denied errors should now display properly.

## [r15](https://github.com/gokcehan/lf/releases/tag/r15)

### Changed

- The `toggle` command does not move the selection down anymore. The default binding for `<space>` is now assigned to `:toggle; down` instead to keep the default behavior same as before.
- The expansion `%w` in option `promptfmt` should now have a trailing slash. The default value of `promptfmt` is now changed accordingly, and should not display double slashes in the root directory anymore.
- The key `<esc>` is now used as the escape key. It should not display an error message when used to cancel a keybinding menu as before. However, it is not possible to bind `<esc>` key to another command anymore.

### Added

- Symbolic link destinations are now shown in the bottom status line (#374).
- A new `hiddenfiles` option which takes a list of globs is implemented to customize which files should be `hidden` (#372).
- Expressions consisting of multiple commands can now use counts (#394).
- Moving operations now fall back to copy and then delete strategy automatically for cross-device linking.
- The `hidden` option now works in Windows.
- The `toggle` command can now take optional arguments to toggle given names instead of the current file (#409).
- A new option `truncatechar` is implemented to customize the truncate character used in long filenames (#417).
- Copy and move operations now display a success message when they are finished (#427).

### Fixed

- `SIGHUP` and `SIGTERM` signals are now properly handled. Log files should not remain when terminals are directly closed (#305).
- The `info` option should now align properly when used with the `number` and `relativenumber` options (#373).
- Tilde (`~`) is now only expanded at the beginning of the path for the `cd` and `select` commands (#373).
- The `rename` command should now work properly with names differing only cases on case-insensitive filesystems.
- Tab characters are now expanded to spaces in Windows.
- The `incsearch` option now respects the search direction accordingly.
- The server is now started in the home folder and will not hold mounted filesystems busy.
- Trailing spaces in configuration files do not confuse the parser anymore.
- Termbox version is updated to fix a keyboard problem in FreeBSD (#404).
- Async commands do not leave zombie processes anymore (#407).
- The `hidden` option now works consistently as expected when set at the initial launch.
- The `rename` command should now select the new file after the operation.
- The `rename` command should now handle absolute paths properly.
- The `select` command should now work properly on loading directories. Custom commands that select a file after an operation should now work properly without an explicit `load` operation beforehand.
- Previous errors in the bottom message line should not persist through the prompt usage anymore.
- The `push` command should not fail with non-ASCII characters anymore.
- The `select` command should not fail with broken links anymore.
- The `load` command should not clear toggled broken links anymore.
- Copy and move operations do not overwrite broken links anymore.

## [r14](https://github.com/gokcehan/lf/releases/tag/r14)

### Added

- The `delete` command now shows a prompt with the current filename or the number of selected files (#206).
- Backslash can now be escaped with a backslash even without quotes.
- A new desktop entry file `lf.desktop` is added (#222).
- Three new `sortby` types are added, access time (i.e. `atime`), change time (i.e. `ctime`) (#226), and extension (i.e. `ext`) (#230). New default keybindings are added for these sorts correspondingly (i.e. `sa`, `sc`, and `se`). The `info` option can now also contain `atime` and `ctime` values accordingly.
- A new shell completion for `zsh` is added to `etc/lf.zsh` (#239).
- The `delete` command now works asynchronously and shows the progress (#238).
- Completion and directory change scripts are added for `csh` and `tcsh` as `etc/lf.csh` and `etc/lfcd.csh` respectively (#264).
- A new special command `on-cd` is added to run a shell command when the directory is changed. See the documentation for more information (#291).

### Fixed

- Some directories with special permissions that previously show a file icon now shows a directory icon properly.
- The `etc/lfcd.cmd` script can now also change to a different volume drive (#221).
- The proper use of `setsid` for opening files is now added to the example configuration and the documentation.
- The home directory abbreviation `~` is now only applied properly to paths starting with the home directory (#241).
- The `rename` command now cancels the operation if the old and new paths are the same (#266).
- Autocompletion and word movements should now work properly with all Unicode characters.
- The `shell-pipe` command which was broken some time ago should now work as expected.
- The `$TERM` environment variable can now work with values containing `tmux` with custom `$TERMINFO` values. @doronbehar now maintains a Termbox fork for `lf` (https://github.com/doronbehar/termbox-go).

## [r13](https://github.com/gokcehan/lf/releases/tag/r13)

### Added

- A new `wrapscroll` option is added to wrap top and bottom while scrolling (#166).
- The `up`, `down` movement commands and their variants, `updir`, and `open` are now allowed in `cmap` mappings.
- Two new `glob-select` and `glob-unselect` commands are added to use globbing for toggling files (#184).
- A new `mark-remove` (default `"`) command is added to allow removing marks (#190).
- Icon support is added with the `icon` option. See the wiki page for more details.
- A new builtin `rename` command is added (#197).

### Fixed

- The `cmd-history-next` command now remains in Command-line mode after the last item (#168).
- The `select` command does not change directories anymore when used on a directory.
- The working directory is now changed to the first argument when it is a directory.
- The `ratios` option is now checked before `preview` to avoid crashes (#174).
- Previous error messages are now cleared after successful commands (#192).
- Symlink to directories are now colored as symlinks (#195).
- Permission errors for directories are now displayed properly instead of showing as empty (#203).

## [r12](https://github.com/gokcehan/lf/releases/tag/r12)

### Added

- Go modules replaced `godep` for dependency management. Package maintainers may need to update accordingly.
- A new `errorfmt` option is added to customize the colors and attributes of error messages.

### Fixed

- Autocompletion for searches now complete filenames instead of commands.
- Permanent environment variables (e.g. `$id`, `$EDITOR`, `$LF_LEVEL`) are now exported on startup so they can be used in preview scripts without running a shell command first.
- On Windows, quotes are added to the exported values `$f`, `$fs`, and `$fx` to handle filenames with spaces properly.
- On Windows, filenames starting with `.` characters are now shown to avoid crashes when filenames show up as empty.

## [r11](https://github.com/gokcehan/lf/releases/tag/r11)

### Changed

- Copy and move operations are now implemented as builtins instead of using the underlying shell primitives (i.e. `cp` and `mv`). Users who want the old behavior can define a custom `paste` command. See the updated documentation for more information. Please report bugs regarding this change.
- Preview messages (i.e. `empty`, `binary`, and `loading...`) are now shown with the reverse attribute.

### Added

- Copy and move operations now run asynchronously and the progress is shown in the bottom ruler.
- Two new commands `echomsg` and `echoerr` are added to print a message to the message line and to the log file at the same time.

### Fixed

- Terminal initialization errors are now shown in the terminal instead of the log file.

## [r10](https://github.com/gokcehan/lf/releases/tag/r10)

### Changed

- The ability to map Normal mode commands in `cmap` is removed. This has caused a number of bugs in the previous release. A different mechanism for a similar functionality is planned.

### Added

- A new command line flag `-command` has been added to execute a command on client initialization (#135).
- A `select` command is now executed after initialization if the first command line argument is a file.
- A prompting mechanism has been added to the builtin `delete` command.

### Fixed

- Input and output in `shell-pipe` commands were broken with the `cmap` patch. This should now work as before.
- Some `push` commands were broken with the `cmap` patch and sometimes ignored Command-line mode for some keys to execute as in Normal mode. This should now work as before.
- `read` and shell commands should now also work when typed manually (e.g. typing `:shell` should switch the prefix to `$`).
- Configuration files are now read after initialization.
- Background colors are removed from defaults to avoid confusion with selection highlighting.

## [r9](https://github.com/gokcehan/lf/releases/tag/r9)

### Changed

- The default number of colors is set to 8 to have better defaults in some terminals. A new option `color256` is added to use 256 colors instead. Users who want the old behavior should enable this option in their configuration files.

### Added

- A new `incsearch` option is added to enable incremental matching while searching.
- Two new options `ignoredia` and `smartdia` are added to ignore diacritics in Latin letters for `search` and `find` (#118).
- A new builtin `delete` command is added for file deletion (#121). This command is not assigned to a key by default to prevent accidental deletions. In the future, a prompting mechanism may be added to this command for more safety.
- Normal mode commands can now be used in `cmap` which can be used to immediately finish Command-line mode and execute a Normal mode command afterwards.
- A new `fish` completion script is added to the `etc` folder (#131).
- Two new options `number` and `relativenumber` are added to enable line numbers in directories (#133).

### Fixed

- Autocompletion should now show only a single match for redefined builtin commands.

## [r8](https://github.com/gokcehan/lf/releases/tag/r8)

### Added

- Four new commands `find`, `find-back`, `find-next`, and `find-prev` are added to implement file finding. Two options `anchorfind` and `findlen` are added to customize the behavior of these commands.
- A new `quit` command is added to the server protocol to quit the server.
- A new `$LF_LEVEL` environment variable is added to show the nesting level.

### Fixed

- The `load` and `reload` commands now work properly when the current directory is deleted. Also `lf` does not start in deleted directories anymore.
- The server is now started as a detached process in Windows so its lifetime is not tied to the command line window anymore.
- Clients now try to reconnect to the server at startup with exponentially increasing intervals when they fail. This is to avoid connection failures due to the server not being ready for the first client that automatically starts the server.
- The old index is now kept when the current selection is deleted.
- The `shell-pipe` command now triggers `load` instead of `reload`.
- Error messages are now more informative when `lf` fails to start due to either `$HOME` or `$USER` variables being empty or not set.
- Searching for the next/previous item is now based on the direction of the initial search.

## [r7](https://github.com/gokcehan/lf/releases/tag/r7)

### Changed

- The system-wide configuration path on Unix is changed from `/etc/lfrc` to `/etc/lf/lfrc`.

### Added

- A man page is now automatically generated from the documentation which can be installed to make the documentation available with the `man` command. On a related note, there is now a packaging guide section in packages wiki page.
- A new `doc` command (default `<f-1>`) is added to view the documentation in a pager.
- Commands `mark-save` (default `m`) and `mark-load` (default `'`) are added to implement builtin bookmarks. Marks are saved in a file in the data folder which can be found in the documentation.
- The history is now saved in a file in the data folder which can be found in the documentation.

## [r6](https://github.com/gokcehan/lf/releases/tag/r6)

### Changed

- The `yank`, `delete`, and `put` commands are renamed to `copy`, `cut`, and `paste` respectively. In the example configuration, the `remove` command is renamed to `delete`.
- The special command `open-file` to configure file opening is renamed to `open`.

### Added

- A new option `shellopts` is added to be able to pass command line arguments to the shell interpreter (i.e. `<shell> <shellopts> -c <cmd> -- <args>`) which is useful to set safety options for all shell commands (i.e. `sh -eu ...`). See the example configuration file for more information.
- The special keys `<home>`, `<end>`, `<pgup>`, and `<pgdn>` are mapped to the `top`, `bottom`, `page-up`, and `page-down` commands respectively by default.
- A new command `source` is added to read a configuration file.
- Support is added to read a system-wide configuration file on startup located in `/etc/lfrc` on Unix and `C:\ProgramData\lf\lfrc` on Windows. The documentation is updated to show the locations of all configuration files.
- Environment variables used for configuration (i.e. `$EDITOR`, `$PAGER`, `$SHELL`) are set to their default values when they are not set or empty and they are exported to shell commands.
- A new environment variable `$OPENER` is added to configure the default file opener using the previous default values and it is exported to shell commands.

### Fixed

- Executable completion now works on Windows as well.

## [r5](https://github.com/gokcehan/lf/releases/tag/r5)

### Added

- The server is automatically restarted on startup if it does not work anymore.
- A new option `period` is added to set time duration in seconds for periodic refreshes. Setting the value of this option to zero disables periodic refreshes which is the default behavior.
- A new command `load` is added to refresh only modified files and directories which is more efficient than `reload` command.

### Fixed

- `cmd-word-back` does not change the command line anymore.
- Modified files and directories are automatically detected and refreshed when they are loaded from cache.
- All clients are now refreshed when the `put` command is used.
- The correct hidden parent is selected when the `hidden` option is changed.
- The preview is properly updated when the `hidden` option is changed.

## [r4](https://github.com/gokcehan/lf/releases/tag/r4)

### Changed

- The following commands are renamed for clarity and consistency:
  - `bot` is renamed to `bottom`
  - `cmd-delete-word` is renamed to `cmd-delete-unix-word`
  - `cmd-beg` is renamed to `cmd-home`
  - `cmd-delete-beg` is renamed to `cmd-delete-home`
  - `cmd-comp` is renamed to `cmd-complete`
  - `cmd-hist-next` is renamed to `cmd-history-next`
  - `cmd-hist-prev` is renamed to `cmd-history-prev`
  - `cmd-put` is renamed to `cmd-yank`

### Added

- Support for alt key bindings have been added using the commonly used escape delaying mechanism. The delay value is set to 100ms which is also used for other escape codes in Termbox. Keys are named with an `a` prefix, as in `<a-f>` for the `alt` and `f` keys. Also note that the old mechanism for alt keybindings on 8-bit terminals still works as before.
- The following command line commands and their default alt keybindings have been added:
  - `cmd-word` with `<a-f>`
  - `cmd-word-back` with `<a-b>`
  - `cmd-capitalize-word` with `<a-c>`
  - `cmd-delete-word` with `<a-d>`
  - `cmd-uppercase-word` with `<a-u>`
  - `cmd-lowercase-word` with `<a-l>`
  - `cmd-transpose-word` with `<a-t>`

### Fixed

- The default editor, pager, and opener commands should now work in Windows. Opener still only works with paths without spaces though.
- 8-bit color codes and attributes are not confused anymore.
- History selection is disabled when a `shell-pipe` command is running.
- Searches are now excluded from the history.

## [r3](https://github.com/gokcehan/lf/releases/tag/r3)

### Changed

- Command counts are now only applied for the `up`/`down` (and variants), `updir`, `toggle`, `search-next`, and `search-prev` commands. These commands are now handled more efficiently when used with counts.

### Added

- Pressed keys are now shown in the ruler when they are not matched yet.
- A new builtin `draw` command has been added which is more efficient than the `redraw` command. The latter is replaced with the former in many places to prevent flickers on the screen.
- Support for the `$LS_COLORS` and `$LSCOLORS` environment variables are added for color customization (#96). See the updated documentation for more information.
- A new option `drawbox` is added to draw a box around panes.

### Fixed

- Resize events that change the height are now handled properly.
- Changes in sorting methods and options are checked for cached directories and these directories are sorted again if necessary while loading.
- A `~` character is added as a suffix to file names when they do not fit in the window.

## [r2](https://github.com/gokcehan/lf/releases/tag/r2)

### Changed

- Shell command names are shortened (e.g. `read-shell-wait` is renamed to `shell-wait`).

### Added

- A new shell command type named `shell-pipe` is introduced that runs with the UI. See the updated documentation for the motivation and some example use cases.
- A new command named `cmd-interrupt` (default `<c-c>`) is introduced to interrupt the current `shell-pipe` command.
- A new command named `select` is introduced that changes the current file selection to its argument.

### Fixed

- Running `cmd-hist-prev` in Normal mode now always starts with the last item to avoid confusion. Running `cmd-hist-next` in Normal mode now has no effect for consistency.

## [r1](https://github.com/gokcehan/lf/releases/tag/r1)

### Added

- Initial release
