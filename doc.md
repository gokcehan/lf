# NAME

lf - terminal file manager

# SYNOPSIS

**lf**
[**-command** *command*]
[**-config** *path*]
[**-cpuprofile** *path*]
[**-doc**]
[**-help**]
[**-last-dir-path** *path*]
[**-log** *path*]
[**-memprofile** *path*]
[**-print-last-dir**]
[**-print-selection**]
[**-remote** *command*]
[**-selection-path** *path*]
[**-server**]
[**-single**]
[**-version**]
[*cd-or-select-path*]

# DESCRIPTION

lf is a terminal file manager.

The source code can be found in the repository at https://github.com/gokcehan/lf

This documentation can either be read from the terminal using `lf -doc` or online at https://github.com/gokcehan/lf/blob/master/doc.md
You can also use the `help` command (default `<f-1>`) inside lf to view the documentation in a pager.
A man page with the same content is also available in the repository at https://github.com/gokcehan/lf/blob/master/lf.1

# OPTIONS

## POSITIONAL ARGUMENTS

**cd-or-select-path**

Set the starting location. If *path* is a directory, start in there. If it's a file, start in the file's parent directory and select the file. When no *path* is supplied, lf uses the current directory. Only accepts one argument.

## META OPTIONS

**-doc**

Show lf's documentation (same content as this file) and exit.

**-help**

Show command-line usage and exit.

**-version**

Show version information and exit.

## STARTUP & CONFIGURATION

**-command** *command*

Execute *command* during client initialization (i.e. after reading configuration, before `on-init`). To execute more than one command, you can either use the **-command** flag multiple times or pass multiple commands at once by chaining them with ";".

**-config** *path*

Use the config file at *path* instead of the normal search locations. This only affects which `lfrc` is read at startup.

## SHELL INTEGRATION

**-print-last-dir**

Print the last directory to stdout when lf exits. This can be used to let lf change your shells working directory. See `CHANGING DIRECTORY` for more details.

**-last-dir-path** *path*

Same as **-print-last-dir**, but write the directory to *path* instead of stdout.

**-print-selection**

Print selected files to stdout when opening a file in lf. This can be used to use lf as an "open file" dialog. First, select the files you want to pass to another program. Then, confirm the selection by opening a file. This causes lf to quit and print out the selection. Quitting lf prematurely discards the selection.

**-selection-path** *path*

Same as **-print-selection**, but write the newline-separated list to *path* instead of stdout.

## SERVER

**-remote** *command*

Send *command* to the running server (i.e. `send`, `query`, `list`, `quit`, or `quit!`). See `REMOTE COMMANDS` for more details.

**-server**

Start the (headless) server process explicitly. Runs in the foreground and writes server logs to stderr (or the file set with **-log**). Clients auto-start a server if none is running unless **-single** is used.

**-single**

Start a stand-alone client without a server. Disables remote control.

## DIAGNOSTICS

**-log** *path*

Append runtime log messages to *path*.

**-cpuprofile** *path*

Write a CPU profile to *path*. The profile can be used by `go tool pprof`.

**-memprofile** *path*

Write a memory profile to *path*. The profile can be used by `go tool pprof`.

## EXAMPLES

Use `lf` to select files (while hiding certain file types):

	lf -command 'set nohidden' -command 'set hiddenfiles "*mp4:*pdf:*txt"' -print-selection

Another sophisticated "open file" dialog focusing on design:

	lf -command 'set nopreview; set ratios 1; set drawbox; set promptfmt "Select files [%w] %S q: cancel, l: confirm"' -print-selection

Open Downloads and set `sortby` and `info` to creation date:

	lf -command 'set sortby btime; set info btime' ~/Downloads

Temporarily prevent `lf` from modifying the command history:

	lf -command 'set nohistory'

Use default settings and log current session:

	lf -config /dev/null -log /tmp/lf.log

Force-quit the server:

	lf -remote 'quit!'

Inherit lf's working directory in your shell:

	cd "$(lf -print-last-dir)"

# QUICK REFERENCE

The following commands are provided by lf:

	quit                     (default 'q')
	up                       (default 'k' and '<up>')
	half-up                  (default '<c-u>')
	page-up                  (default '<c-b>' and '<pgup>')
	scroll-up                (default '<c-y>')
	down                     (default 'j' and '<down>')
	half-down                (default '<c-d>')
	page-down                (default '<c-f>' and '<pgdn>')
	scroll-down              (default '<c-e>')
	updir                    (default 'h' and '<left>')
	open                     (default 'l' and '<right>')
	jump-next                (default ']')
	jump-prev                (default '[')
	top                      (default 'gg' and '<home>')
	bottom                   (default 'G' and '<end>')
	high                     (default 'H')
	middle                   (default 'M')
	low                      (default 'L')
	toggle
	invert                   (default 'v')
	unselect                 (default 'u')
	glob-select
	glob-unselect
	copy                     (default 'y')
	cut                      (default 'd')
	paste                    (default 'p')
	clear                    (default 'c')
	sync
	draw
	redraw                   (default '<c-l>')
	load
	reload                   (default '<c-r>')
	delete         (modal)
	rename         (modal)   (default 'r')
	read           (modal)   (default ':')
	shell          (modal)   (default '$')
	shell-pipe     (modal)   (default '%')
	shell-wait     (modal)   (default '!')
	shell-async    (modal)   (default '&')
	find           (modal)   (default 'f')
	find-back      (modal)   (default 'F')
	find-next                (default ';')
	find-prev                (default ',')
	search         (modal)   (default '/')
	search-back    (modal)   (default '?')
	search-next              (default 'n')
	search-prev              (default 'N')
	filter         (modal)
	setfilter
	mark-save      (modal)   (default 'm')
	mark-load      (modal)   (default "'")
	mark-remove    (modal)   (default '"')
	tag
	tag-toggle               (default 't')
	echo
	echomsg
	echoerr
	cd
	select
	source
	push
	addcustominfo
	calcdirsize
	clearmaps
	tty-write
	visual                   (default 'V')

The following Visual mode commands are provided by lf:

	visual-accept            (default 'V')
	visual-unselect
	visual-discard           (default '<esc>')
	visual-change            (default 'o')

The following Command-line mode commands are provided by lf:

	cmd-insert
	cmd-escape               (default '<esc>')
	cmd-complete             (default '<tab>')
	cmd-menu-complete
	cmd-menu-complete-back
	cmd-menu-accept
	cmd-menu-discard
	cmd-enter                (default '<c-j>' and '<enter>')
	cmd-interrupt            (default '<c-c>')
	cmd-history-next         (default '<c-n>' and '<down>')
	cmd-history-prev         (default '<c-p>' and '<up>')
	cmd-left                 (default '<c-b>' and '<left>')
	cmd-right                (default '<c-f>' and '<right>')
	cmd-home                 (default '<c-a>' and '<home>')
	cmd-end                  (default '<c-e>' and '<end>')
	cmd-delete               (default '<c-d>' and '<delete>')
	cmd-delete-back          (default '<backspace>' and '<backspace2>')
	cmd-delete-home          (default '<c-u>')
	cmd-delete-end           (default '<c-k>')
	cmd-delete-unix-word     (default '<c-w>')
	cmd-yank                 (default '<c-y>')
	cmd-transpose            (default '<c-t>')
	cmd-transpose-word       (default '<a-t>')
	cmd-word                 (default '<a-f>')
	cmd-word-back            (default '<a-b>')
	cmd-delete-word          (default '<a-d>')
	cmd-delete-word-back     (default '<a-backspace>' and '<a-backspace2>')
	cmd-capitalize-word      (default '<a-c>')
	cmd-uppercase-word       (default '<a-u>')
	cmd-lowercase-word       (default '<a-l>')

The following options can be used to customize the behavior of lf:

	anchorfind        bool      (default true)
	autoquit          bool      (default true)
	borderfmt         string    (default "\033[0m")
	cleaner           string    (default '')
	copyfmt           string    (default "\033[7;33m")
	cursoractivefmt   string    (default "\033[7m")
	cursorparentfmt   string    (default "\033[7m")
	cursorpreviewfmt  string    (default "\033[4m")
	cutfmt            string    (default "\033[7;31m")
	dircounts         bool      (default false)
	dirfirst          bool      (default true)
	dironly           bool      (default false)
	dirpreviews       bool      (default false)
	drawbox           bool      (default false)
	dupfilefmt        string    (default '%f.~%n~')
	errorfmt          string    (default "\033[7;31;47m")
	filesep           string    (default "\n")
	filtermethod      string    (default 'text')
	findlen           int       (default 1)
	hidden            bool      (default false)
	hiddenfiles       []string  (default '.*' for Unix and '' for Windows)
	history           bool      (default true)
	icons             bool      (default false)
	ifs               string    (default '')
	ignorecase        bool      (default true)
	ignoredia         bool      (default true)
	incfilter         bool      (default false)
	incsearch         bool      (default false)
	info              []string  (default '')
	infotimefmtnew    string    (default 'Jan _2 15:04')
	infotimefmtold    string    (default 'Jan _2  2006')
	menufmt           string    (default "\033[0m")
	menuheaderfmt     string    (default "\033[1m")
	menuselectfmt     string    (default "\033[7m")
	mergeindicators   bool      (default false)
	mouse             bool      (default false)
	number            bool      (default false)
	numberfmt         string    (default "\033[33m")
	period            int       (default 0)
	preload           bool      (default false)
	preserve          []string  (default "mode")
	preview           bool      (default true)
	previewer         string    (default '')
	promptfmt         string    (default "\033[32;1m%u@%h\033[0m:\033[34;1m%d\033[0m\033[1m%f\033[0m")
	ratios            []int     (default '1:2:3')
	relativenumber    bool      (default false)
	reverse           bool      (default false)
	roundbox          bool      (default false)
	rulerfile         string    (default "")
	rulerfmt          string    (default "")
	scrolloff         int       (default 0)
	searchmethod      string    (default 'text')
	selectfmt         string    (default "\033[7;35m")
	selmode           string    (default 'all')
	shell             string    (default 'sh' for Unix and 'cmd' for Windows)
	shellflag         string    (default '-c' for Unix and '/c' for Windows)
	shellopts         []string  (default '')
	showbinds         bool      (default true)
	sizeunits         string    (default 'binary')
	smartcase         bool      (default true)
	smartdia          bool      (default false)
	sortby            string    (default 'natural')
	statfmt           string    (default "\033[36m%p\033[0m| %c| %u| %g| %S| %t| -> %l")
	tabstop           int       (default 8)
	tagfmt            string    (default "\033[31m")
	tempmarks         string    (default '')
	timefmt           string    (default 'Mon Jan _2 15:04:05 2006')
	truncatechar      string    (default '~')
	truncatepct       int       (default 100)
	visualfmt         string    (default "\033[7;36m")
	waitmsg           string    (default 'Press any key to continue')
	watch             bool      (default false)
	wrapscan          bool      (default true)
	wrapscroll        bool      (default false)
	user_{option}     string    (default none)

The following environment variables are exported for shell commands:

	f
	fs
	fv
	fx
	id
	PWD
	OLDPWD
	LF_LEVEL
	OPENER
	VISUAL
	EDITOR
	PAGER
	SHELL
	lf
	lf_{option}
	lf_user_{option}
	lf_flag_{flag}
	lf_width
	lf_height
	lf_count
	lf_mode

The following special shell commands are used to customize the behavior of lf when defined:

	open
	paste
	rename
	delete
	pre-cd
	on-cd
	on-load
	on-focus-gained
	on-focus-lost
	on-init
	on-select
	on-redraw
	on-quit

The following commands/keybindings are provided by default:

	Unix
	cmd open &$OPENER "$f"
	map e $$EDITOR "$f"
	map i $$PAGER "$f"
	map w $$SHELL
	cmd help $$lf -doc | $PAGER
	map <f-1> help
	cmd maps $lf -remote "query $id maps" | $PAGER
	cmd nmaps $lf -remote "query $id nmaps" | $PAGER
	cmd vmaps $lf -remote "query $id vmaps" | $PAGER
	cmd cmaps $lf -remote "query $id cmaps" | $PAGER
	cmd cmds $lf -remote "query $id cmds" | $PAGER

	Windows
	cmd open &%OPENER% %f%
	map e $%EDITOR% %f%
	map i !%PAGER% %f%
	map w $%SHELL%
	cmd help !%lf% -doc | %PAGER%
	map <f-1> help
	cmd maps !%lf% -remote "query %id% maps" | %PAGER%
	cmd nmaps !%lf% -remote "query %id% nmaps" | %PAGER%
	cmd vmaps !%lf% -remote "query %id% vmaps" | %PAGER%
	cmd cmaps !%lf% -remote "query %id% cmaps" | %PAGER%
	cmd cmds !%lf% -remote "query %id% cmds" | %PAGER%

The defaults for Windows are using `cmd` syntax.
A `PowerShell` compatible configuration file can be found at
https://github.com/gokcehan/lf/blob/master/etc/lfrc.ps1.example

The following additional keybindings are provided by default:

	map zh set hidden!
	map zr set reverse!
	map zn set info
	map zs set info size
	map zt set info time
	map za set info size:time
	map sn :set sortby natural; set info
	map ss :set sortby size; set info size
	map st :set sortby time; set info time
	map sa :set sortby atime; set info atime
	map sb :set sortby btime; set info btime
	map sc :set sortby ctime; set info ctime
	map se :set sortby ext; set info
	map gh cd ~
	nmap <space> :toggle; down

If the `mouse` option is enabled, mouse buttons have the following default effects:

	Left mouse button
	    Click on a file or directory to select it.

	Right mouse button
	    Enter a directory or open a file. Also works on the preview pane.

	Scroll wheel
	    Move up or down. If Ctrl is pressed, scroll up or down.

# CONFIGURATION

Configuration files should be located at:

	OS       system-wide               user-specific
	Unix     /etc/lf/lfrc              ~/.config/lf/lfrc
	Windows  C:\ProgramData\lf\lfrc    C:\Users\<user>\AppData\Roaming\lf\lfrc

The colors file should be located at:

	OS       system-wide               user-specific
	Unix     /etc/lf/colors            ~/.config/lf/colors
	Windows  C:\ProgramData\lf\colors  C:\Users\<user>\AppData\Roaming\lf\colors

The icons file should be located at:

	OS       system-wide               user-specific
	Unix     /etc/lf/icons             ~/.config/lf/icons
	Windows  C:\ProgramData\lf\icons   C:\Users\<user>\AppData\Roaming\lf\icons

The selection file should be located at:

	Unix     ~/.local/share/lf/files
	Windows  C:\Users\<user>\AppData\Local\lf\files

The marks file should be located at:

	Unix     ~/.local/share/lf/marks
	Windows  C:\Users\<user>\AppData\Local\lf\marks

The tags file should be located at:

	Unix     ~/.local/share/lf/tags
	Windows  C:\Users\<user>\AppData\Local\lf\tags

The history file should be located at:

	Unix     ~/.local/share/lf/history
	Windows  C:\Users\<user>\AppData\Local\lf\history

You can configure these locations with the following variables given with their order of precedences and their default values:

	Unix
	    $LF_CONFIG_HOME
	    $XDG_CONFIG_HOME
	    ~/.config

	    $LF_DATA_HOME
	    $XDG_DATA_HOME
	    ~/.local/share

	Windows
	    %LF_CONFIG_HOME%
	    %XDG_CONFIG_HOME%
	    %APPDATA%

	    %LF_DATA_HOME%
	    %XDG_DATA_HOME%
	    %LOCALAPPDATA%

A sample configuration file can be found at
https://github.com/gokcehan/lf/blob/master/etc/lfrc.example

# COMMANDS

This section shows information about built-in commands.
Modal commands do not take any arguments, but instead change the operation mode to read their input conveniently, and so they are meant to be assigned to keybindings.

## quit (default `q`)

Quit lf and return to the shell.

## up (default `k` and `<up>`), half-up (default `<c-u>`), page-up (default `<c-b>` and `<pgup>`), scroll-up (default `<c-y>`), down (default `j` and `<down>`), half-down (default `<c-d>`), page-down (default `<c-f>` and `<pgdn>`), scroll-down (default `<c-e>`)

Move/scroll the current file selection upwards/downwards by one/half a page/full page.

## updir (default `h` and `<left>`)

Change the current working directory to the parent directory.

## open (default `l` and `<right>`)

If the current file is a directory, then change the current directory to it, otherwise, execute the `open` command.
A default `open` command is provided to call the default system opener asynchronously with the current file as the argument.
A custom `open` command can be defined to override this default.

## jump-next (default `]`), jump-prev (default `[`)

Change the current working directory to the next/previous jumplist item.

## top (default `gg` and `<home>`), bottom (default `G` and `<end>`)

Move the current file selection to the top/bottom of the directory.
A count can be specified to move to a specific line, for example, use `3G` to move to the third line.

## high (default `H`), middle (default `M`), low (default `L`)

Move the current file selection to the high/middle/low of the screen.

## toggle

Toggle the selection of the current file or files given as arguments.

## invert (default `v`)

Reverse the selection of all files in the current directory (i.e. `toggle` all files).
Selections in other directories are not affected by this command.
You can define a new command to select all files in the directory by combining `invert` with `unselect` (i.e. `cmd select-all :unselect; invert`), though this will also remove selections in other directories.

## unselect (default `u`)

Remove the selection of all files in all directories.

## glob-select, glob-unselect

Select/unselect files that match the given glob.

## copy (default `y`)

Save the paths of selected files to the clipboard as files to be copied.
If there are no selected files, the path of the current file is used instead.

## cut (default `d`)

Save the paths of selected files to the clipboard as files to be moved.
If there are no selected files, the path of the current file is used instead.

## paste (default `p`)

Copy/Move files in the clipboard to the current working directory.
A custom `paste` command can be defined to override this default.

## clear (default `c`)

Clear file paths in the clipboard.

## sync

Synchronize copied/cut files with the server.
This command is automatically called when required.

## draw

Draw the screen.
This command is automatically called when required.

## redraw (default `<c-l>`)

Synchronize the terminal and redraw the screen.

## load

Load modified files and directories.
This command is automatically called when required.

## reload (default `<c-r>`)

Flush the cache and reload all files and directories.

## delete (modal)

Remove the current file or selected file(s).
A custom `delete` command can be defined to override this default.

## rename (modal) (default `r`)

Rename the current file using the built-in method.
A custom `rename` command can be defined to override this default.

## read (modal) (default `:`)

Read a command to evaluate.

## shell (modal) (default `$`)

Read a shell command to execute.

## shell-pipe (modal) (default `%`)

Read a shell command to execute piping its standard I/O to the bottom statline.

## shell-wait (modal) (default `!`)

Read a shell command to execute and wait for a key press at the end.

## shell-async (modal) (default `&`)

Read a shell command to execute asynchronously without standard I/O.

## find (modal) (default `f`), find-back (modal) (default `F`), find-next (default `;`), find-prev (default `,`)

Read key(s) to find the appropriate filename match in the forward/backward direction and jump to the next/previous match.

## search (default `/`), search-back (default `?`), search-next (default `n`), search-prev (default `N`)

Read a pattern to search for a filename match in the forward/backward direction and jump to the next/previous match.

## filter (modal), setfilter

Command `filter` reads a pattern to filter out and only view files matching the pattern.
Command `setfilter` does the same but uses an argument to set the filter immediately.
You can supply an argument to `filter` to use as the starting prompt.

## mark-save (modal) (default `m`)

Save the current directory as a bookmark assigned to the given key.

## mark-load (modal) (default `'`)

Change the current directory to the bookmark assigned to the given key.
A special bookmark `'` holds the previous directory after a `mark-load`, `cd`, or `select` command.

## mark-remove (modal) (default `"`)

Remove a bookmark assigned to the given key.

## tag

Tag a file with `*` or a single-width character given in the argument.
You can define a new tag-clearing command by combining `tag` with `tag-toggle` (i.e. `cmd tag-clear :tag; tag-toggle`).

## tag-toggle (default `t`)

Tag a file with `*` or a single-width character given in the argument if the file is untagged, otherwise remove the tag.

## echo

Print the given arguments to the message line at the bottom.

## echomsg

Print the given arguments to the message line at the bottom and also to the log file.

## echoerr

Print given arguments to the message line at the bottom as `errorfmt` and also to the log file.

## cd

Change the working directory to the given argument.

## select

Change the current file selection to the given argument.

## source

Read the configuration file given in the argument.

## push

Simulate key pushes given in the argument.

## addcustominfo

Update the `custom` info and `.Stat.CustomInfo` field of the given file with the given string.
The info string may contain ANSI escape codes to further customize its appearance.
If no info is provided, clear the file's info instead.

## calcdirsize

Calculate the total size for each of the selected directories.
Option `info` should include `size` and option `dircounts` should be disabled to show this size.
If the total size of a directory is not calculated, it will be shown as `-`.

## clearmaps

Remove all keybindings associated with the `map`, `nmap` and `vmap` command.
This command can be used in the config file to remove the default keybindings.
For safety purposes, `:` is left mapped to the `read` command, and `cmap` keybindings are retained so that it is still possible to exit `lf` using `:quit`.

## tty-write

Write the given string to the tty.
This is useful for sending escape sequences to the terminal to control its behavior (e.g. OSC 0 to set the window title).
Using `tty-write` is preferred over directly writing to `/dev/tty` because the latter is not synchronized and can interfere with drawing the UI.

## visual (default `V`)

Switch to Visual mode.
If already in Visual mode, discard the visual selection and stay in Visual mode.

# VISUAL MODE COMMANDS

## visual-accept (default `V`)

Add the visual selection to the selection list, quit Visual mode and return to Normal mode.

## visual-unselect

Remove the visual selection from the selection list, quit Visual mode and return to Normal mode.

## visual-discard (default `<esc>`)

Discard the visual selection, quit Visual mode and return to Normal mode.

## visual-change (default `o`)

Go to the other end of the current Visual mode selection.

# COMMAND-LINE MODE COMMANDS

The prompt character specifies which of the several Command-line modes you are in.
For example, the `read` command takes you to the `:` mode.

When the cursor is at the first character in `:` mode, pressing one of the keys `!`, `$`, `%`, or `&` takes you to the corresponding mode.
You can go back with `cmd-delete-back` (`<backspace>` by default).

The command line commands should be mostly compatible with readline keybindings.
A character refers to a Unicode code point, a word consists of letters and digits, and a Unix word consists of any non-blank characters.

## cmd-insert

Insert the character given in the argument.
This command is automatically called when required.

## cmd-escape (default `<esc>`)

Quit Command-line mode and return to Normal mode.

## cmd-complete (default `<tab>`)

Autocomplete the current word.

## cmd-menu-complete, cmd-menu-complete-back

Autocomplete the current word with the menu selection.
You need to assign keys to these commands (e.g. `cmap <tab> cmd-menu-complete; cmap <backtab> cmd-menu-complete-back`).
You can use the assigned keys to display the menu and then cycle through completion options.

## cmd-menu-accept

Accept the currently selected match in menu completion and close the menu.

## cmd-menu-discard

Discard the currently selected match in menu completion and close the menu.

## cmd-enter (default `<c-j>` and `<enter>`)

Execute the current line.

## cmd-interrupt (default `<c-c>`)

Interrupt the current shell-pipe command and return to the Normal mode.

## cmd-history-next (default `<c-n>` and `<down>`), cmd-history-prev (default `<c-p>` and `<up>`)

Go to the next/previous entry in the command history.
If part of the command is already typed, then only matching entries will be considered, and consecutive duplicate entries are skipped.

## cmd-left (default `<c-b>` and `<left>`), cmd-right (default `<c-f>` and `<right>`)

Move the cursor to the left/right.

## cmd-home (default `<c-a>` and `<home>`), cmd-end (default `<c-e>` and `<end>`)

Move the cursor to the beginning/end of the line.

## cmd-delete (default `<c-d>` and `<delete>`)

Delete the next character.

## cmd-delete-back (default `<backspace>` and `<backspace2>`)

Delete the previous character.
When at the beginning of a prompt, returns either to Normal mode or to `:` mode.

## cmd-delete-home (default `<c-u>`), cmd-delete-end (default `<c-k>`)

Delete everything up to the beginning/end of the line.

## cmd-delete-unix-word (default `<c-w>`)

Delete the previous Unix word.

## cmd-yank (default `<c-y>`)

Paste the buffer content containing the last deleted item.

## cmd-transpose (default `<c-t>`), cmd-transpose-word (default `<a-t>`)

Transpose the positions of the last two characters/words.

## cmd-word (default `<a-f>`), cmd-word-back (default `<a-b>`)

Move the cursor by one word in the forward/backward direction.

## cmd-delete-word (default `<a-d>`)

Delete the next word in the forward direction.

## cmd-delete-word-back (default `<a-backspace>` and `<a-backspace2>`)

Delete the previous word in the backward direction.

## cmd-capitalize-word (default `<a-c>`), cmd-uppercase-word (default `<a-u>`), cmd-lowercase-word (default `<a-l>`)

Capitalize/uppercase/lowercase the current word and jump to the next word.

# SETTINGS

This section shows information about options to customize the behavior.
Character `:` is used as the separator for list options `[]int` and `[]string`.

## anchorfind (bool) (default true)

When this option is enabled, the find command starts matching patterns from the beginning of filenames, otherwise, it can match at an arbitrary position.

## autoquit (bool) (default true)

Automatically quit the server when there are no clients left connected.

## borderfmt (string) (default `\033[0m`)

Format string of the box drawing characters enabled by the `drawbox` option.

## cleaner (string) (default ``) (not called if empty)

Set the path of a cleaner file.
The file should be executable.
This file is called if previewing is enabled, the previewer is set, and the previously selected file has its preview cache disabled.
The following arguments are passed to the file, (1) current filename, (2) width, (3) height, (4) horizontal position, (5) vertical position of preview pane and (6) next filename to be previewed respectively.
Preview cleaning is disabled when the value of this option is left empty.

## copyfmt (string) (default `\033[7;33m`)

Format string of the indicator for files to be copied.

## cursoractivefmt (string) (default `\033[7m`), cursorparentfmt (string) (default `\033[7m`), cursorpreviewfmt (string) (default `\033[4m`)

Format strings for highlighting the cursor.
`cursoractivefmt` applies in the current directory pane,
`cursorparentfmt` applies in panes that show parents of the current directory,
and `cursorpreviewfmt` applies in panes that preview directories.

The default is to make the active cursor and the parent directory cursor inverted. The preview cursor is underlined.

Some other possibilities to consider for the preview or parent cursors: an empty string for no cursor, `\033[7;2m` for dimmed inverted text (visibility varies by terminal), `\033[7;90m` for inverted text with grey (aka "brightblack") background.

If the format string contains the characters `%s`, it is interpreted as a format string for `fmt.Sprintf`. Such a string should end with the terminal reset sequence.
For example, `\033[4m%s\033[0m` has the same effect as `\033[4m`.

## cutfmt (string) (default `\033[7;31m`)

Format string of the indicator for files to be cut.

## dircounts (bool) (default false)

When this option is enabled, directory sizes show the number of items inside instead of the total size of the directory, which needs to be calculated for each directory using `calcdirsize`.
This information needs to be calculated by reading the directory and counting the items inside.
Therefore, this option is disabled by default for performance reasons.
This option only has an effect when `info` has a `size` field and the pane is wide enough to show the information.
999 items are counted per directory at most, and bigger directories are shown as `999+`.

## dirfirst (bool) (default true)

Show directories first above regular files.
With `dircounts` enabled, sorting by `size` always separates directories and files, regardless of `dirfirst`.

## dironly (bool) (default false)

Show only directories.

## dirpreviews (bool) (default false)

If enabled, directories will also be passed to the previewer script. This allows custom previews for directories.

## drawbox (bool) (default false)

Draw boxes around panes with box drawing characters.

## dupfilefmt (string) (default `%f.~%n~`)

Format string of filename when creating duplicate files. With the default format, copying a file `abc.txt` to the same directory will result in a duplicate file called `abc.txt.~1~`.
Special expansions are provided, `%f` as the file name, `%b` for the base name (file name without extension), `%e` as the extension (including the dot) and `%n` as the number of duplicates.

## errorfmt (string) (default `\033[7;31;47m`)

Format string of error messages shown in the bottom message line.

If the format string contains the characters `%s`, it is interpreted as a format string for `fmt.Sprintf`. Such a string should end with the terminal reset sequence.
For example, `\033[4m%s\033[0m` has the same effect as `\033[4m`.

## filesep (string) (default `\n`)

File separator used in environment variables `fs`, `fv` and `fx`.

## filtermethod (string) (default `text`)

How filter command patterns are treated.
Currently supported methods are `text` (i.e. string literals), `glob` (i.e. shell globs) and `regex` (i.e. regular expressions).
See `SEARCHING FILES` for more details.

## findlen (int) (default 1)

Number of characters prompted for the find command.
When this value is set to 0, find command prompts until there is only a single match left.

## hidden (bool) (default false)

Show hidden files.
On Unix systems, hidden files are determined by the value of `hiddenfiles`.
On Windows, files with hidden attributes are also considered hidden files.

## hiddenfiles ([]string) (default `.*` for Unix and `` for Windows)

List of hidden file glob patterns.
Patterns can be given as relative or absolute paths.
Globbing supports the usual special characters, `*` to match any sequence, `?` to match any character, and `[...]` or `[^...]` to match character sets or ranges.
In addition, if a pattern starts with `!`, then its matches are excluded from hidden files. To add multiple patterns, use `:` as a separator. Example: `.*:lost+found:*.bak`

## history (bool) (default true)

Save command history.

## icons (bool) (default false)

Show icons before each item in the list.

## ifs (string) (default ``)

Sets `IFS` variable in shell commands.
It works by adding the assignment to the beginning of the command string as `IFS=...; ...`.
The reason is that `IFS` variable is not inherited by the shell for security reasons.
This method assumes a POSIX shell syntax so it can fail for non-POSIX shells.
This option has no effect when the value is left empty.
This option does not have any effect on Windows.

## ignorecase (bool) (default true)

Ignore case in sorting and search patterns.

## ignoredia (bool) (default true)

Ignore diacritics in sorting and search patterns.

## incfilter (bool) (default false)

Apply filter pattern after each keystroke during filtering.

## incsearch (bool) (default false)

Jump to the first match after each keystroke during searching.

## info ([]string)  (default ``)

A list of information that is shown for directory items at the right side of the pane.

The following information types are supported:

	perm      file permission
	user      user name
	group     group name
	size      file size
	time      time of last data modification
	atime     time of last access
	btime     time of file birth
	ctime     time of last status (inode) change
	custom    property defined via `addcustominfo` (empty by default)

Information is only shown when the pane width is more than twice the width of information.

## infotimefmtnew (string) (default `Jan _2 15:04`)

Format string of the file time shown in the info column when it matches this year.

## infotimefmtold (string) (default `Jan _2  2006`)

Format string of the file time shown in the info column when it doesn't match this year.

## menufmt (string) (default `\033[0m`)

Format string of the menu.

## menuheaderfmt (string) (default `\033[1m`)

Format string of the header row in the menu.

## menuselectfmt (string) (default `\033[7m`)

Format string of the currently selected item in the menu.

## mergeindicators (bool) (default false)

When `mergeindicators` is enabled, tag and selection indicators are drawn in a single column to reduce the gap before filenames.
If a file is both tagged and selected, the tag uses the selection format (e.g. `copyfmt`) instead of `tagfmt`.

## mouse (bool) (default false)

Send mouse events as input.

## number (bool) (default false)

Show the position number for directory items on the left side of the pane.
When the `relativenumber` option is enabled, only the current line shows the absolute position and relative positions are shown for the rest.

## numberfmt (string) (default `\033[33m`)

Format string of the position number for each line.

## period (int) (default 0)

Set the interval in seconds for periodic checks of directory updates.
This works by periodically calling the `load` command.
Note that directories are already updated automatically in many cases.
This option can be useful when there is an external process changing the displayed directory and you are not doing anything in lf.
Periodic checks are disabled when the value of this option is set to zero.

## preload (bool) (default false)

Allow previews to be generated in advance using the `previewer` script as the user navigates through the filesystem.

## preserve ([]string) (default `mode`)

List of attributes that are preserved when copying files.
Currently supported attributes are `mode` (i.e. access mode) and `timestamps` (i.e. modification time and access time).
Note that preserving other attributes like ownership of change/birth timestamp is desirable, but not portably supported in Go.

## preview (bool) (default true)

Show previews of files and directories at the rightmost pane.
If the file has more lines than the preview pane, the rest of the lines are not read.
Files containing the null character (U+0000) in the read portion are considered binary files and displayed as `binary`.

## previewer (string) (default ``) (not filtered if empty)

Set the path of a previewer file to filter the content of regular files for previewing.
The file should be executable.
The following arguments are passed to the file, (1) current filename, (2) width, (3) height, (4) horizontal position, (5) vertical position, and (6) mode ("preview" or "preload").
SIGPIPE signal is sent when enough lines are read.
If the previewer returns a non-zero exit code, then the preview cache for the given file is disabled.
This means that if the file is selected in the future, the previewer is called once again.
Preview filtering is disabled and files are displayed as they are when the value of this option is left empty.
If the `preload` option is enabled, then this will be called with `preload` as the mode when preloading file previews.
Refer to the [PREVIEWING FILES section](https://github.com/gokcehan/lf/blob/master/doc.md#previewing-files) for more information about how to configure custom previews.

## promptfmt (string) (default `\033[32;1m%u@%h\033[0m:\033[34;1m%d\033[0m\033[1m%f\033[0m`)

Format string of the prompt shown in the top line.

The following special expansions are supported:

	%f        file name
	%h        host name
	%u        user name
	%w        working directory
	%d        working directory (with trailing path separator)
	%F        current filter
	%S        spacer to right-align the following parts (can be used once)

The home folder is shown as `~` in the working directory expansion.
Directory names are automatically shortened to a single character starting from the leftmost parent when the prompt does not fit the screen.

## ratios ([]int) (default `1:2:3`)

List of ratios of pane widths.
Number of items in the list determines the number of panes in the UI.
When the `preview` option is enabled, the rightmost number is used for the width of the preview pane.

## relativenumber (bool) (default false)

Show the position number relative to the current line.
When `number` is enabled, the current line shows the absolute position, otherwise nothing is shown.

## reverse (bool) (default false)

Reverse the direction of sort.

## roundbox (bool) (default false)

Draw rounded outer corners when the `drawbox` option is enabled.

## rulerfile (string) (default ``)

Set the path of the ruler file.
If not set, then a default template will be used for the ruler.
Refer to the [RULER section](https://github.com/gokcehan/lf/blob/master/doc.md#ruler) for more information about how the ruler file works.

## rulerfmt (string) (default ``)

Format string of the ruler shown in the bottom right corner.
When set, it will be used along with `statfmt` to draw the ruler, and `rulerfile` will be ignored.
However, using `rulerfile` is preferred and this option is provided for backwards compatibility.

The following special expansions are supported:

	%a        pressed keys
	%p        progress of file operations
	%m        number of files to be cut (moved)
	%c        number of files to be copied
	%s        number of selected files
	%v        number of visually selected files
	%t        number of shown files in the current directory
	%h        number of hidden files in the current directory
	%f        current filter
	%i        cursor position
	%P        scroll percentage
	%d        amount of free disk space

Additional expansions are provided for environment variables exported by lf, in the form `%{lf_<name>}` (e.g. `%{lf_selmode}`). This is useful for displaying the current settings.
Expansions are also provided for user-defined options, in the form `%{lf_user_<name>}` (e.g. `%{lf_user_foo}`).
The `|` character splits the format string into sections. Any section containing a failed expansion (result is a blank string) is discarded and not shown.

## scrolloff (int) (default 0)

Minimum number of offset lines shown at all times at the top and bottom of the screen when scrolling.
The current line is kept in the middle when this option is set to a large value that is bigger than half the number of lines.
A smaller offset can be used when the current file is close to the beginning or end of the list to show the maximum number of items.

## searchmethod (string) (default `text`)

How search command patterns are treated.
Currently supported methods are `text` (i.e. string literals), `glob` (i.e. shell globs) and `regex` (i.e. regular expressions).
See `SEARCHING FILES` for more details.

## selectfmt (string) (default `\033[7;35m`)

Format string of the indicator for files that are selected.

## selmode (string) (default `all`)

Selection mode for commands.
When set to `all` it will use the selected files from all directories.
When set to `dir` it will only use the selected files in the current directory.

## shell (string) (default `sh` for Unix and `cmd` for Windows)

Shell executable to use for shell commands.
Shell commands are executed as `shell shellopts shellflag command -- arguments`.

## shellflag (string) (default `-c` for Unix and `/c` for Windows)

Command line flag used to pass shell commands.

## shellopts ([]string)  (default ``)

List of shell options to pass to the shell executable.

## showbinds (bool) (default true)

Show bindings associated with pressed keys.

## sizeunits (string) (default `binary`)

Determines whether file sizes are displayed using binary units (`1K` is 1024 bytes) or decimal units (`1K` is 1000 bytes).

## smartcase (bool) (default true)

Override `ignorecase` option when the pattern contains an uppercase character.
This option has no effect when `ignorecase` is disabled.

## smartdia (bool) (default false)

Override `ignoredia` option when the pattern contains a character with diacritic.
This option has no effect when `ignoredia` is disabled.

## sortby (string) (default `natural`)

Sort type for directories.

The following sort types are supported:

	natural   file name (track_2.flac comes before track_10.flac)
	name      file name (track_10.flac comes before track_2.flac)
	ext       file extension
	size      file size
	time      time of last data modification
	atime     time of last access
	btime     time of file birth
	ctime     time of last status (inode) change
	custom    property defined via `addcustominfo` (empty by default)

## statfmt (string) (default `\033[36m%p\033[0m| %c| %u| %g| %S| %t| -> %l`)

Format string of the file info shown in the bottom left corner.
This option has no effect unless `rulerfmt` is also set.
Using `rulerfile` is preferred and this option is provided for backwards compatibility.

The following special expansions are supported:

	%p        file permission
	%c        link count
	%u        user name
	%g        group name
	%s        file size
	%S        file size (left-padded with spaces to a fixed width of 5 characters)
	%t        time of last data modification
	%l        link target
	%m        current mode
	%M        current mode (displaying `NORMAL` instead of a blank string in Normal mode)

The `|` character splits the format string into sections. Any section containing a failed expansion (result is a blank string) is discarded and not shown.

## tabstop (int) (default 8)

Number of space characters to show for horizontal tabulation (U+0009) character.

## tagfmt (string) (default `\033[31m`)

Format string of the tags.

If the format string contains the characters `%s`, it is interpreted as a format string for `fmt.Sprintf`. Such a string should end with the terminal reset sequence.
For example, `\033[4m%s\033[0m` has the same effect as `\033[4m`.

## tempmarks (string) (default ``)

Marks to be considered temporary (e.g. `abc` refers to marks `a`, `b`, and `c`).
These marks are not synced to other clients and they are not saved in the bookmarks file.
Note that the special bookmark `` ` `` is always treated as temporary and it does not need to be specified.

## timefmt (string) (default `Mon Jan _2 15:04:05 2006`)

Format string of the file modification time shown in the bottom line.

## truncatechar (string) (default `~`)

The truncate character that is shown at the end when the filename does not fit into the pane.

## truncatepct (int) (default 100)

When a filename is too long to be shown completely, the available space will be partitioned into two parts.
`truncatepct` is a percentage value between 0 and 100 that determines the size of the first part, which will be shown at the beginning of the filename.
The second part uses the rest of the available space, and will be shown at the end of the filename.
Both parts are separated by the truncation character (`truncatechar`).
Truncation is not applied to the file extension.

For example, with the filename `very_long_filename.txt`:

- `set truncatepct 100` -> `very_long_filena~.txt` (default)

- `set truncatepct 50`  -> `very_lon~filename.txt`

- `set truncatepct 0`   -> `~ry_long_filename.txt`

## visualfmt (string) (default `\033[7;36m`)

Format string of the indicator for files that are visually selected.

## waitmsg (string) (default `Press any key to continue`)

String shown after commands of shell-wait type.

## watch (bool) (default false)

Watch the filesystem for changes using `fsnotify` to automatically refresh file information.
FUSE is currently not supported due to limitations in `fsnotify`.

## wrapscan (bool) (default true)

Searching can wrap around the file list.

## wrapscroll (bool) (default false)

Scrolling can wrap around the file list.

## user_{option} (string) (default none)

Any option that is prefixed with `user_` is a user-defined option and can be set to any string.
Inside a user-defined command, the value will be provided in the `lf_user_{option}` environment variable.
These options are not used by lf and are not persisted.

# ENVIRONMENT VARIABLES

The following variables are exported for shell commands:
These are referred to with a `$` prefix on POSIX shells (e.g. `$f`), between `%` characters on Windows cmd (e.g. `%f%`), and with a `$env:` prefix on Windows PowerShell (e.g. `$env:f`).

## f

Current file selection as a full path.

## fs

Selected file(s) separated with the value of `filesep` option as full path(s).

## fv

Visually selected file(s) separated with the value of `filesep` option as full path(s).

## fx

Selected file(s) (i.e. `fs`, never `fv`) if there are any selected files, otherwise current file selection (i.e. `f`).

## id

Id of the running client.

## PWD

Present working directory.

## OLDPWD

Initial working directory.

## LF_LEVEL

The value of this variable is set to the current nesting level when you run lf from a shell spawned inside lf.
You can add the value of this variable to your shell prompt to make it clear that your shell runs inside lf.
For example, with POSIX shells, you can use `[ -n "$LF_LEVEL" ] && PS1="$PS1""(lf level: $LF_LEVEL) "` in your shell configuration file (e.g. `~/.bashrc`).

## OPENER

If this variable is set in the environment, use the same value. Otherwise, this is set to `start` in Windows, `open` in macOS, `xdg-open` in others.

## EDITOR

If VISUAL is set in the environment, use its value. Otherwise, use the value of the environment variable EDITOR. If neither variable is set, this is set to `vi` on Unix, `notepad` in Windows.

## PAGER

If this variable is set in the environment, use the same value. Otherwise, this is set to `less` on Unix, `more` in Windows.

## SHELL

If this variable is set in the environment, use the same value. Otherwise, this is set to `sh` on Unix, `cmd` in Windows.

## lf

Absolute path to the currently running lf binary, if it can be found. Otherwise, this is set to the string `lf`.

## lf_{option}

Value of the {option}.

## lf_user_{option}

Value of the user_{option}.

## lf_flag_{flag}

Value of the command line {flag}.

## lf_width, lf_height

Width/Height of the terminal.

## lf_count

Value of the count associated with the current command.

## lf_mode

Current mode that `lf` is operating in.
This is useful for customizing keybindings depending on what the current mode is.
Possible values are `compmenu`, `delete`, `rename`, `filter`, `find`, `mark`, `search`, `command`, `shell`, `pipe` (when running a shell-pipe command), `normal`, `visual` and `unknown`.

# SPECIAL COMMANDS

This section shows information about special shell commands.

## open

This shell command can be defined to override the default `open` command when the current file is not a directory.

## paste

This shell command can be defined to override the default `paste` command.

## rename

This shell command can be defined to override the default `rename` command.

## delete

This shell command can be defined to override the default `delete` command.

## pre-cd

This shell command can be defined to be executed before changing a directory.

## on-cd

This shell command can be defined to be executed after changing a directory.

## on-load

This shell command can be defined to be executed after loading a directory.
It provides the files inside the directory as arguments.

## on-focus-gained

This shell command can be defined to be executed when the terminal gains focus.

## on-focus-lost

This shell command can be defined to be executed when the terminal loses focus.

## on-init

This shell command can be defined to be executed after initializing and connecting to the server.

## on-select

This shell command can be defined to be executed after the selection changes.

## on-redraw

This shell command can be defined to be executed after the screen is redrawn or if the terminal is resized.

## on-quit

This shell command can be defined to be executed before quitting.

# PREFIXES

The following command prefixes are used by lf:

	:  read (default)  built-in/custom command
	$  shell           shell command
	%  shell-pipe      shell command running with the UI
	!  shell-wait      shell command waiting for a key press
	&  shell-async     shell command running asynchronously

The same evaluator is used for the command line and the configuration file for reading shell commands.
The difference is that prefixes are not necessary in the command line.
Instead, different modes are provided to read corresponding commands.
These modes are mapped to the prefix keys above by default.
Visual mode mappings are defined the same way Normal mode mappings are defined.

# SYNTAX

Characters from `#` to newline are comments and ignored:

	# comments start with `#`

The following commands (`set`, `setlocal`, `map`, `nmap`, `vmap`, `cmap`, and `cmd`) are used for configuration.

Command `set` is used to set an option which can be a boolean, integer, or string:

	set hidden         # boolean enable
	set hidden true    # boolean enable
	set nohidden       # boolean disable
	set hidden false   # boolean disable
	set hidden!        # boolean toggle
	set scrolloff 10   # integer value
	set sortby time    # string value without quotes
	set sortby 'time'  # string value with single quotes (whitespace)
	set sortby "time"  # string value with double quotes (backslash escapes)

Command `setlocal` is used to set a local option for a directory which can be a boolean or string.
Currently supported local options are `dircounts`, `dirfirst`, `dironly`, `hidden`, `info`, `reverse` and `sortby`.
Adding a trailing path separator (i.e. `/` for Unix and `\` for Windows) sets the option for the given directory along with its subdirectories:

	setlocal /foo/bar hidden         # boolean enable
	setlocal /foo/bar hidden true    # boolean enable
	setlocal /foo/bar nohidden       # boolean disable
	setlocal /foo/bar hidden false   # boolean disable
	setlocal /foo/bar hidden!        # boolean toggle
	setlocal /foo/bar sortby time    # string value without quotes
	setlocal /foo/bar sortby 'time'  # string value with single quotes (whitespace)
	setlocal /foo/bar sortby "time"  # string value with double quotes (backslash escapes)
	setlocal /foo/bar  hidden        # for only '/foo/bar' directory
	setlocal /foo/bar/ hidden        # for '/foo/bar' and its subdirectories (e.g. '/foo/bar/baz')

Command `map` is used to bind a key in Normal and Visual mode to a command which can be a built-in command, custom command, or shell command:

	map gh cd ~        # built-in command
	map D trash        # custom command
	map i $less $f     # shell command
	map U !du -csh *   # waiting shell command

Command `nmap` does the same but for Normal mode only.

Command `vmap` does the same but for Visual mode only.

Overview of which map command works in which mode:

	map                Normal, Visual
	nmap               Normal
	vmap               Visual
	cmap               Command-line

Command `cmap` is used to bind a key on the command line to a command line command or any other command:

	cmap <c-g> cmd-escape
	cmap <a-i> set incsearch!

You can delete an existing binding by leaving the expression empty:

	map gh             # deletes 'gh' mapping in Normal and Visual mode
	nmap v             # deletes 'v' mapping in Normal mode
	vmap o             # deletes 'o' mapping in Visual mode
	cmap <c-g>         # deletes '<c-g>' mapping

Command `cmd` is used to define a custom command:

	cmd usage $du -h -d1 | less

You can delete an existing command by leaving the expression empty:

	cmd trash          # deletes 'trash' command

If there is no prefix then `:` is assumed:

	map zt set info time

An explicit `:` can be provided to group statements until a newline which is especially useful for `map` and `cmd` commands:

	map st :set sortby time; set info time

If you need multiline you can wrap statements in `{{` and `}}` after the proper prefix.

	map st :{{
	    set sortby time
	    set info time
	}}

# KEY MAPPINGS

Regular keys are assigned to a command with the usual syntax:

	map a down

Keys combined with the Shift key simply use the uppercase letter:

	map A down

Special keys are written in between `<` and `>` characters and always use lowercase letters:

	map <enter> down

Angle brackets can be assigned with their special names:

	map <lt> down
	map <gt> down

Function keys are prefixed with `f` character:

	map <f-1> down

Keys combined with the Ctrl key are prefixed with a `c` character:

	map <c-a> down

Keys combined with the Alt key are assigned in two different ways depending on the behavior of your terminal.
Older terminals (e.g. xterm) may set the 8th bit of a character when the Alt key is pressed.
On these terminals, you can use the corresponding byte for the mapping:

	map á down

Newer terminals (e.g. gnome-terminal) may prefix the key with an escape character when the Alt key is pressed.
lf uses the escape delaying mechanism to recognize Alt keys in these terminals (delay is 100ms).
On these terminals, keys combined with the Alt key are prefixed with an `a` character:

	map <a-a> down

It is possible to combine special keys with modifiers:

	map <a-enter> down

WARNING: Some key combinations will likely be intercepted by your OS, window manager, or terminal.
Other key combinations cannot be recognized by lf due to the way terminals work (e.g. `Ctrl+h` combination sends a backspace key instead).
The easiest way to find out the name of a key combination and whether it will work on your system is to press the key while lf is running and read the name from the `unknown mapping` error.

Mouse buttons are prefixed with an `m` character:

	map <m-1> down  # primary
	map <m-2> down  # secondary
	map <m-3> down  # middle
	map <m-4> down
	map <m-5> down
	map <m-6> down
	map <m-7> down
	map <m-8> down

Mouse wheel events are also prefixed with an `m` character:

	map <m-up>    down
	map <m-down>  down
	map <m-left>  down
	map <m-right> down

# PUSH MAPPINGS

The usual way to map a key sequence is to assign it to a named or unnamed command.
While this provides a clean way to remap built-in keys as well as other commands, it can be limiting at times.
For this reason, the `push` command is provided by lf.
This command is used to simulate key pushes given as its arguments.
You can `map` a key to a `push` command with an argument to create various keybindings.

This is mainly useful for two purposes.
First, it can be used to map a command with a command count:

	map <c-j> push 10j

Second, it can be used to avoid typing the name when a command takes arguments:

	map r push :rename<space>

One thing to be careful of is that since the `push` command works with keys instead of commands it is possible to accidentally create recursive bindings:

	map j push 2j

These types of bindings create a deadlock when executed.

# SHELL COMMANDS

Regular shell commands are the most basic command type that is useful for many purposes.
For example, we can write a shell command to move the selected file(s) to trash.
A first attempt to write such a command may look like this:

	cmd trash ${{
	    mkdir -p ~/.trash
	    if [ -z "$fs" ]; then
	        mv "$f" ~/.trash
	    else
	        IFS="$(printf '\n\t')"; mv $fs ~/.trash
	    fi
	}}

We check `$fs` to see if there are any selected files.
Otherwise, we just delete the current file.
Since this is such a common pattern, a separate `$fx` variable is provided.
We can use this variable to get rid of the conditional:

	cmd trash ${{
	    mkdir -p ~/.trash
	    IFS="$(printf '\n\t')"; mv $fx ~/.trash
	}}

The trash directory is checked each time the command is executed.
We can move it outside of the command so it would only run once at startup:

	${{ mkdir -p ~/.trash }}

	cmd trash ${{ IFS="$(printf '\n\t')"; mv $fx ~/.trash }}

Since these are one-liners, we can drop `{{` and `}}`:

	$mkdir -p ~/.trash

	cmd trash $IFS="$(printf '\n\t')"; mv $fx ~/.trash

Finally, note that we set the `IFS` variable manually in these commands.
Instead, we could use the `ifs` option to set it for all shell commands (i.e. `set ifs "\n"`).
This can be especially useful for interactive use (e.g. `$rm $f` or `$rm $fs` would simply work).
This option is not set by default as it can behave unexpectedly for new users.
However, use of this option is highly recommended and it is assumed in the rest of the documentation.

# PIPING SHELL COMMANDS

Regular shell commands have some limitations in some cases.
When an output or error message is given and the command exits afterwards, the UI is immediately resumed and there is no way to see the message without dropping to shell again.
Also, even when there is no output or error, the UI still needs to be paused while the command is running.
This can cause flickering on the screen for short commands and similar distractions for longer commands.

Instead of pausing the UI, piping shell commands connect stdin, stdout, and stderr of the command to the statline at the bottom of the UI.
This can be useful for programs following the Unix philosophy to give no output in the success case, and brief error messages or prompts in other cases.

For example, the following rename command prompts for overwrite in the statline if there is an existing file with the given name:

	cmd rename %mv -i $f $1

You can also output error messages in the command and they will show up in the statline.
For example, an alternative rename command may look like this:

	cmd rename %[ -e $1 ] && printf "file exists" || mv $f $1

Note that input is line buffered and output and error are byte buffered.

# WAITING SHELL COMMANDS

Waiting shell commands are similar to regular shell commands except that they wait for a key press when the command is finished.
These can be useful to see the output of a program before the UI is resumed.
Waiting shell commands are more appropriate than piping shell commands when the command is verbose and the output is best displayed as multiline.

# ASYNCHRONOUS SHELL COMMANDS

Asynchronous shell commands are used to start a command in the background and then resume operation without waiting for the command to finish.
Stdin, stdout, and stderr of the command are neither connected to the terminal nor the UI.

# REMOTE COMMANDS

One of the more advanced features in lf is remote commands.
All clients connect to a server on startup.
It is possible to send commands to all or any of the connected clients over the common server.
This is used internally to notify file selection changes to other clients.

To use this feature, you need to use a client which supports communicating with a Unix domain socket.
OpenBSD implementation of netcat (nc) is one such example.
You can use it to send a command to the socket file:

	echo 'send echo hello world' | nc -U ${XDG_RUNTIME_DIR:-/tmp}/lf.${USER}.sock

Since such a client may not be available everywhere, lf comes bundled with a command line flag to be used as such.
When using lf, you do not need to specify the address of the socket file.
This is the recommended way of using remote commands since it is shorter and immune to socket file address changes:

	lf -remote 'send echo hello world'

In this command `send` is used to send the rest of the string as a command to all connected clients.
You can optionally give it an ID number to send a command to a single client:

	lf -remote 'send 1234 echo hello world'

All clients have a unique ID number but you may not be aware of the ID number when you are writing a command.
For this purpose, an `$id` variable is exported to the environment for shell commands.
The value of this variable is set to the process ID of the client.
You can use it to send a remote command from a client to the server which in return sends a command back to itself.
So now you can display a message in the current client by calling the following in a shell command:

	lf -remote "send $id echo hello world"

Since lf does not have control flow syntax, remote commands are used for such needs.
For example, you can configure the number of columns in the UI with respect to the terminal width as follows:

	cmd recol %{{
	    if [ $lf_width -le 80 ]; then
	        lf -remote "send $id set ratios 1:2"
	    elif [ $lf_width -le 160 ]; then
	        lf -remote "send $id set ratios 1:2:3"
	    else
	        lf -remote "send $id set ratios 1:2:3:5"
	    fi
	}}

In addition, the `query` command can be used to obtain information about a specific lf instance by providing its ID:

	lf -remote "query $id maps"

The following types of information are supported:

	maps     list of mappings created by the 'map', 'nmap' and 'vmap' command
	nmaps    list of mappings created by the 'nmap' and 'map' command
	vmaps    list of mappings created by the 'vmap' and 'map' command
	cmaps    list of mappings created by the 'cmap' command
	cmds     list of commands created by the 'cmd' command
	jumps    contents of the jump list, showing previously visited locations
	history  list of previously executed commands on the command line
	files    list of files in the currently open directory as displayed by lf, empty if dir is still loading

When listing mappings the characters in the first column are:

	n  Normal
	v  Visual
	c  Command-line

This is useful for scripting actions based on the internal state of lf.
For example, to select a previous command using fzf and execute it:

	map <a-h> ${{
	    clear
	    cmd=$(
	        lf -remote "query $id history" |
	        awk -F'\t' 'NR > 1 { print $NF}' |
	        sort -u |
	        fzf --reverse --prompt='Execute command: '
	    )
	    lf -remote "send $id $cmd"
	}}

The `list` command prints the IDs of all currently connected clients:

	lf -remote 'list'

There is also a `quit` command to quit the server when there are no connected clients left, and a `quit!` command to force quit the server by closing client connections first:

	lf -remote 'quit'
	lf -remote 'quit!'

Lastly, the commands `conn` and `drop` connect or disconnect ID to/from the server:

	lf -remote 'conn $id'
	lf -remote 'drop $id'

These are internal and generally not needed by users.

# FILE OPERATIONS

lf uses its own built-in copy and move operations by default.
These are implemented as asynchronous operations and progress is shown in the bottom ruler.
These commands do not overwrite existing files or directories with the same name.
Instead, a suffix that is compatible with the `--backup=numbered` option in GNU cp is added to the new files or directories.
Only file modes and (some) timestamps can be preserved (see `preserve` option), all other attributes are ignored including ownership, context, and xattr.
Special files such as character and block devices, named pipes, and sockets are skipped and links are not followed.
Moving is performed using the rename operation of the underlying OS.
For cross-device moving, lf falls back to copying and then deletes the original files if there are no errors.
Operation errors are shown in the message line as well as the log file and they do not prematurely terminate the corresponding file operation.

File operations can be performed on the currently selected file or on multiple files by selecting them first.
When you `copy` a file, lf doesn't actually copy the file on the disk, but only records its name to a file.
The actual file copying takes place when you `paste`.
Similarly `paste` after a `cut` operation moves the file.

You can customize copy and move operations by defining a `paste` command.
This is a special command that is called when it is defined instead of the built-in implementation.
You can use the following example as a starting point:

	cmd paste %{{
	    load=$(cat ~/.local/share/lf/files)
	    mode=$(echo "$load" | sed -n '1p')
	    list=$(echo "$load" | sed '1d')
	    if [ $mode = 'copy' ]; then
	        cp -R $list .
	    elif [ $mode = 'move' ]; then
	        mv $list .
	        rm ~/.local/share/lf/files
	        lf -remote 'send clear'
	    fi
	}}

Some useful things to be considered are to use the backup (`--backup`) and/or preserve attributes (`-a`) options with `cp` and `mv` commands if they support it (i.e. GNU implementation), change the command type to asynchronous, or use `rsync` command with progress bar option for copying and feed the progress to the client periodically with remote `echo` calls.

By default, lf does not assign `delete` command to a key to protect new users.
You can customize file deletion by defining a `delete` command.
You can also assign a key to this command if you like.
An example command to move selected files to a trash folder and remove files completely after a prompt is provided in the example configuration file.

# SEARCHING FILES

There are two mechanisms implemented in lf to search a file in the current directory.
Searching is the traditional method to move the selection to a file matching a given pattern.
Finding is an alternative way to search for a pattern possibly using fewer keystrokes.

The searching mechanism is implemented with commands `search` (default `/`), `search-back` (default `?`), `search-next` (default `n`), and `search-prev` (default `N`).
You can set `searchmethod` to `glob` to match using a glob pattern.
Globbing supports `*` to match any sequence, `?` to match any character, and `[...]` or `[^...]` to match character sets or ranges.
You can set `searchmethod` to `regex` to match using a regex pattern.
For a full overview of Go's RE2 syntax, see https://pkg.go.dev/regexp/syntax.
You can enable `incsearch` option to jump to the current match at each keystroke while typing.
In this mode, you can either use `cmd-enter` to accept the search or use `cmd-escape` to cancel the search.
You can also map some other commands with `cmap` to accept the search and execute the command immediately afterwards.
For example, you can use the right arrow key to finish the search and open the selected file with the following mapping:

	cmap <right> :cmd-enter; open

The finding mechanism is implemented with commands `find` (default `f`), `find-back` (default `F`), `find-next` (default `;`), `find-prev` (default `,`).
You can disable `anchorfind` option to match a pattern at an arbitrary position in the filename instead of the beginning.
You can set the number of keys to match using `findlen` option.
If you set this value to zero, then the keys are read until there is only a single match.
The default values of these two options are set to jump to the first file with the given initial.

Some options affect both searching and finding.
You can disable `wrapscan` option to prevent searches from being wrapped around at the end of the file list.
You can disable `ignorecase` option to match cases in the pattern and the filename.
This option is already automatically overridden if the pattern contains uppercase characters.
You can disable `smartcase` option to disable this behavior.
Two similar options `ignoredia` and `smartdia` are provided to control matching diacritics in Latin letters.

# OPENING FILES

You can define an `open` command (default `l` and `<right>`) to configure file opening.
This command is only called when the current file is not a directory, otherwise, the directory is entered instead.
You can define it just as you would define any other command:

	cmd open $vi $fx

It is possible to use different command types:

	cmd open &xdg-open $f

You may want to use either file extensions or MIME types from `file` command:

	cmd open ${{
	    case $(file --mime-type -Lb $f) in
	        text/*) vi $fx;;
	        *) for f in $fx; do xdg-open $f > /dev/null 2> /dev/null & done;;
	    esac
	}}

You may want to use `setsid` before your opener command to have persistent processes that continue to run after lf quits.

Regular shell commands (i.e. `$`) drop to the terminal which results in a flicker for commands that finish immediately (e.g. `xdg-open` in the above example).
If you want to use asynchronous shell commands (i.e. `&`) but also want to use the terminal when necessary (e.g. `vi` in the above example), you can use a remote command:

	cmd open &{{
	    case $(file --mime-type -Lb $f) in
	        text/*) lf -remote "send $id \$vi \$fx";;
	        *) for f in $fx; do xdg-open $f > /dev/null 2> /dev/null & done;;
	    esac
	}}

Note that asynchronous shell commands run in their own process group by default so they do not require the manual use of `setsid`.

The following command is provided by default:

	cmd open &$OPENER $f

You may also use any other existing file openers as you like.
Possible options are `libfile-mimeinfo-perl` (executable name is `mimeopen`), `rifle` (ranger's default file opener), or `mimeo` to name a few.

# PREVIEWING FILES

lf previews files on the preview pane by printing the file until the end or until the preview pane is filled.
This output can be enhanced by providing a custom preview script for filtering.
This can be used to highlight source code, list contents of archive files or view PDF or image files to name a few.
For coloring lf recognizes ANSI escape codes.

To use this feature, you need to set the value of `previewer` option to the path of an executable file.
The following arguments are passed to the file, (1) current filename, (2) width, (3) height, (4) horizontal position, (5) vertical position, and (6) mode ("preview" or "preload").
The output of the execution is printed in the preview pane.

Different types of files can be handled by matching by extension (or MIME type from the `file` command):

	#!/bin/sh

	case "$1" in
	    *.tar*) tar tf "$1";;
	    *.zip) unzip -l "$1";;
	    *.rar) unrar l "$1";;
	    *.7z) 7z l "$1";;
	    *.pdf) pdftotext "$1" -;;
	    *) highlight -O ansi "$1";;
	esac

Because files can be large, lf automatically closes the previewer script output pipe with a SIGPIPE when enough lines are read.
Note that some programs may not respond well to SIGPIPE and will exit with a non-zero return code, which avoids caching.
You may add a trailing `|| true` command to avoid such errors:

	highlight -O ansi "$1" || true

You may also want to use the same script in your pager mapping as well:

	set previewer ~/.config/lf/pv.sh
	map i $~/.config/lf/pv.sh $f | less -R

For `less` pager, you may instead utilize `LESSOPEN` mechanism so that useful information about the file such as the full path of the file can still be displayed in the statusline below:

	set previewer ~/.config/lf/pv.sh
	map i $LESSOPEN='| ~/.config/lf/pv.sh %s' less -R $f

Since the preview script is called for each file selection change, it may not generate previews fast enough if the user scrolls through files quickly.
To deal with this, the `preload` option can be set to enable file previews to be preloaded in advance.
If enabled, the preview script will be run on files in advance as the user navigates through them.
In this case, if the exit code of the preview script is zero, then the output will be cached in memory and displayed by lf (useful for text or sixel previews).
Otherwise, it will fallback to calling the preview script again when the file is actually selected (useful for previews managed by an external program).

# CHANGING DIRECTORY

lf changes the working directory of the process to the current directory so that shell commands always work in the displayed directory.
After quitting, it returns to the original directory where it is first launched like all shell programs.
If you want to stay in the current directory after quitting, you can use one of the example lfcd wrapper shell scripts provided in the repository at
https://github.com/gokcehan/lf/tree/master/etc

There is a special command `on-cd` that runs a shell command when it is defined and the directory is changed.
You can define it just as you would define any other command:

	cmd on-cd &{{
	    bash -c '
	    # display git repository status in your prompt
	    source /usr/share/git/completion/git-prompt.sh
	    GIT_PS1_SHOWDIRTYSTATE=auto
	    GIT_PS1_SHOWSTASHSTATE=auto
	    GIT_PS1_SHOWUNTRACKEDFILES=auto
	    GIT_PS1_SHOWUPSTREAM=auto
	    git=$(__git_ps1 " (%s)")
	    fmt="\033[32;1m%u@%h\033[0m:\033[34;1m%d\033[0m\033[1m%f$git\033[0m"
	    lf -remote "send $id set promptfmt \"$fmt\""
	    '
	}}

If you want to send escape sequences to the terminal, you can use the `tty-write` command to do so.
The following xterm-specific escape sequence sets the terminal title to the working directory:

	cmd on-cd &{{
	    lf -remote "send $id tty-write \"\033]0;$PWD\007\""
	}}

This command runs whenever you change the directory but not on startup.
You can add an extra call to make it run on startup as well:

	cmd on-cd &{{ ... }}
	on-cd

Note that all shell commands are possible but `%` and `&` are usually more appropriate as `$` and `!` causes flickers and pauses respectively.

There is also a `pre-cd` command, that works like `on-cd`, but is run before the directory is actually changed.
Another related command is `on-load` which gets executed when loading a directory.

# LOADING DIRECTORY

Similar to `on-cd` there also is `on-load` that when defined runs a shell command after loading a directory.
It works well when combined with `addcustominfo`.

The following example can be used to display git indicators in the info column:

	cmd on-load &{{
	    cd "$(dirname "$1")" || exit 1
	    [ "$(git rev-parse --is-inside-git-dir 2>/dev/null)" = false ] || exit 0

	    cmds=""

	    for file in "$@"; do
	        case "$file" in
	            */.git|*/.git/*) continue;;
	        esac

	        status=$(git status --porcelain --ignored -- "$file" | cut -c1-2 | head -n1)

	        if [ -n "$status" ]; then
	            cmds="${cmds}addcustominfo \"${file}\" \"$status\"; "
	        else
	            cmds="${cmds}addcustominfo \"${file}\" ''; "
	        fi
	    done

	    if [ -n "$cmds" ]; then
	        lf -remote "send $id :$cmds"
	    fi
	}}

Another use case could be showing the dimensions of images and videos:

	cmd on-load &{{
	    cmds=""

	    for file in "$@"; do
	        mime=$(file --mime-type -Lb -- "$file")
	        case "$mime" in
	            # vector images cause problems
	            image/svg+xml)
	                ;;
	            image/*|video/*)
	                dimensions=$(exiftool -s3 -imagesize -- "$file")
	                cmds="${cmds}addcustominfo \"${file}\" \"$dimensions\"; "
	                ;;
	        esac
	    done

	    if [ -n "$cmds" ]; then
	        lf -remote "send $id :$cmds"
	    fi
	}}

# COLORS

lf tries to automatically adapt its colors to the environment.
It starts with a default color scheme and updates colors using values of existing environment variables possibly by overwriting its previous values.
Colors are set in the following order:

 1. default
 2. LSCOLORS (macOS/BSD ls)
 3. LS_COLORS (GNU ls)
 4. LF_COLORS (lf specific)
 5. colors file (lf specific)

Please refer to the corresponding man pages for more information about `LSCOLORS` and `LS_COLORS`.
`LF_COLORS` is provided with the same syntax as `LS_COLORS` in case you want to configure colors only for lf but not ls.
This can be useful since there are some differences between ls and lf, though one should expect the same behavior for common cases.
The colors file (refer to the [CONFIGURATION section](https://github.com/gokcehan/lf/blob/master/doc.md#configuration)) is provided for easier configuration without environment variables.
This file should consist of whitespace-separated pairs with a `#` character to start comments until the end of the line.

You can configure lf colors in two different ways.
First, you can only configure 8 basic colors used by your terminal and lf should pick up those colors automatically.
Depending on your terminal, you should be able to select your colors from a 24-bit palette.
This is the recommended approach as colors used by other programs will also match each other.

Second, you can set the values of environment variables or colors file mentioned above for fine-grained customization.
Note that `LS_COLORS/LF_COLORS` are more powerful than `LSCOLORS` and they can be used even when GNU programs are not installed on the system.
You can combine this second method with the first method for the best results.

Lastly, you may also want to configure the colors of the prompt line to match the rest of the colors.
Colors of the prompt line can be configured using the `promptfmt` option which can include hardcoded colors as ANSI escapes.
See the default value of this option to have an idea about how to color this line.

It is worth noting that lf uses as many colors advertised by your terminal's entry in terminfo or infocmp databases on your system.
If an entry is not present, it falls back to an internal database.
If your terminal supports 24-bit colors but either does not have a database entry or does not advertise all capabilities, you can enable support by setting the `$COLORTERM` variable to `truecolor` or ensuring `$TERM` is set to a value that ends with `-truecolor`.

Default lf colors are mostly taken from GNU dircolors defaults.
These defaults use 8 basic colors and bold attribute.
Default dircolors entries with background colors are simplified to avoid confusion with current file selection in lf.
Similarly, there are only file type matchings and extension matchings are left out for simplicity.
Default values are as follows given with their matching order in lf:

	ln  01;36
	or  31;01
	tw  01;34
	ow  01;34
	st  01;34
	di  01;34
	pi  33
	so  01;35
	bd  33;01
	cd  33;01
	su  01;32
	sg  01;32
	ex  01;32
	fi  00

Note that lf first tries matching file names and then falls back to file types.
The full order of matchings from most specific to least are as follows:

 1. Full Path (e.g. `~/.config/lf/lfrc`)
 2. Dir Name  (e.g. `.git/`) (only matches dirs with a trailing slash at the end)
 3. File Type (e.g. `ln`) (except `fi`)
 4. File Name (e.g. `README*`)
 5. File Name (e.g. `*README`)
 6. Base Name (e.g. `README.*`)
 7. Extension (e.g. `*.txt`)
 8. Default   (i.e. `fi`)

For example, given a regular text file `/path/to/README.txt`, the following entries are checked in the configuration and the first one to match is used:

 1. `/path/to/README.txt`
 2. (skipped since the file is not a directory)
 3. (skipped since the file is of type `fi`)
 4. `README.txt*`
 5. `*README.txt`
 6. `README.*`
 7. `*.txt`
 8. `fi`

Given a regular directory `/path/to/example.d`, the following entries are checked in the configuration and the first one to match is used:

 1. `/path/to/example.d`
 2. `example.d/`
 3. `di`
 4. `example.d*`
 5. `*example.d`
 6. `example.*`
 7. `*.d`
 8. `fi`

Note that glob-like patterns do not perform glob matching for performance reasons.

For example, you can set a variable as follows:

	export LF_COLORS="~/Documents=01;31:~/Downloads=01;31:~/.local/share=01;31:~/.config/lf/lfrc=31:.git/=01;32:.git*=32:*.gitignore=32:*Makefile=32:README.*=33:*.txt=34:*.md=34:ln=01;36:di=01;34:ex=01;32:"

Having all entries on a single line can make it hard to read.
You may instead divide it into multiple lines in between double quotes by escaping newlines with backslashes as follows:

	export LF_COLORS="\
	~/Documents=01;31:\
	~/Downloads=01;31:\
	~/.local/share=01;31:\
	~/.config/lf/lfrc=31:\
	.git/=01;32:\
	.git*=32:\
	*.gitignore=32:\
	*Makefile=32:\
	README.*=33:\
	*.txt=34:\
	*.md=34:\
	ln=01;36:\
	di=01;34:\
	ex=01;32:\
	"

The `ln` entry supports the special value `target`, which will use the link target to select a style. Filename rules will still apply based on the link's name -- this mirrors GNU's `ls` and `dircolors` behavior.
Having such a long variable definition in a shell configuration file might be undesirable.
You may instead use the colors file (refer to the [CONFIGURATION section](https://github.com/gokcehan/lf/blob/master/doc.md#configuration)) for configuration.
A sample colors file can be found at
https://github.com/gokcehan/lf/blob/master/etc/colors.example
You may also see the wiki page for ANSI escape codes
https://en.wikipedia.org/wiki/ANSI_escape_code

# ICONS

Icons are configured using `LF_ICONS` environment variable or an icons file (refer to the [CONFIGURATION section](https://github.com/gokcehan/lf/blob/master/doc.md#configuration)).
The variable uses the same syntax as `LS_COLORS/LF_COLORS`.
Instead of colors, you should use single characters or symbols as values.
The `ln` entry supports the special value `target`, which will use the link target to select a icon. Filename rules will still apply based on the link's name -- this mirrors GNU's `ls` and `dircolors` behavior.
The icons file (refer to the [CONFIGURATION section](https://github.com/gokcehan/lf/blob/master/doc.md#configuration)) should consist of whitespace-separated arrays with a `#` character to start comments until the end of the line.
Each line should contain 1-3 columns: a file type or file name pattern, the icon, and an optional icon color. Using only one column disables all rules for that type or name.
Do not forget to add `set icons true` to your `lfrc` to see the icons.
Default values are listed below in the order lf matches them:

	ln  l
	or  l
	tw  t
	ow  d
	st  t
	di  d
	pi  p
	so  s
	bd  b
	cd  c
	su  u
	sg  g
	ex  x
	fi  -

A sample icons file can be found at
https://github.com/gokcehan/lf/blob/master/etc/icons.example

A sample colored icons file can be found at
https://github.com/gokcehan/lf/blob/master/etc/icons_colored.example

# RULER

The ruler can be configured using the `rulerfile` option (refer to the [CONFIGURATION section](https://github.com/gokcehan/lf/blob/master/doc.md#configuration)).
The contents of the ruler file should be a Go template which is then rendered to create the actual output (refer to https://pkg.go.dev/text/template for more details on the syntax).

The following data fields are exported:

	.Message          string              Includes internal messages, errors, and messages generated by the `echo`/`echomsg`/`echoerr` commands
	.Keys             string              Keys pressed by the user
	.Progress         []string            Progress indicators for copied, moved and deleted files
	.Copy             []string            List of files in the clipboard to be copied
	.Cut              []string            List of files in the clipboard to be moved
	.Select           []string            Selection list
	.Visual           []string            Visual selection
	.Index            int                 Index of the cursor
	.Total            int                 Number of visible files in the current directory
	.Hidden           int                 Number of hidden files in the current directory
	.LinePercentage   string              Line percentage (analogous to `%p` for the `statusline` option in Vim)
	.ScrollPercentage string              Scroll percentage (analogous to `%P` for the `statusline` option in Vim)
	.Filter           []string            Filter currently being applied
	.Mode             string              Current mode ("NORMAL" for Normal mode, and "VISUAL" for Visual mode)
	.Options          map[string]string   The value of options (e.g. `{{.Options.hidden}}`)
	.UserOptions      map[string]string   The value of user-defined options (e.g. `{{.UserOptions.foo}}`)
	.Stat.Path        string              Path of the current file
	.Stat.Name        string              Name of the current file
	.Stat.Extension   string              Extension of the current file
	.Stat.Size        uint64              Size of the current file
	.Stat.DirSize     *uint64             Total size of the current directory if calculated via `calcdirsize`
	.Stat.DirCount    *uint64             Number of items in the current directory if the `dircounts` option is enabled
	.Stat.Permissions string              Permissions of the current file
	.Stat.ModTime     string              Last modified time of the current file (formatted based on the `timefmt` option)
	.Stat.AccessTime  string              Last access time of the current file (formatted based on the `timefmt` option)
	.Stat.BirthTime   string              Birth time of the current file (formatted based on the `timefmt` option)
	.Stat.ChangeTime  string              Last status (inode) change time of the current file (formatted based on the `timefmt` option)
	.Stat.LinkCount   string              Number of hard links for the current file
	.Stat.User        string              User of the current file
	.Stat.Group       string              Group of the current file
	.Stat.Target      string              Target if the current file is a symbolic link, otherwise a blank string
	.Stat.CustomInfo  string              Custom property if defined via `addcustominfo`, otherwise a blank string

The following functions are exported:

	df       func() string                   Get an indicator representing the amount of free disk space available
	env      func(string) string             Get the value of an environment variable
	humanize func(uint64) string             Express a file size in a human-readable format
	join     func([]string, string) string   Join a string array by a separator
	lower    func(string) string             Convert a string to lowercase
	substr   func(string, int, int) string   Get a substring based on starting index and length
	upper    func(string) string             Convert a string to uppercase

The special identifier `{{.SPACER}}` can be used to divide the ruler into sections that are spaced evenly from each other.

The default ruler file can be found at
https://github.com/gokcehan/lf/blob/master/etc/ruler.default
