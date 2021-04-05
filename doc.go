//go:generate gen/docstring.sh
//go:generate gen/man.sh

/*
lf is a terminal file manager.

Source code can be found in the repository at https://github.com/gokcehan/lf.

This documentation can either be read from terminal using 'lf -doc' or online
at https://godoc.org/github.com/gokcehan/lf. You can also use 'doc' command
(default '<f-1>') inside lf to view the documentation in a pager.

You can run 'lf -help' to see descriptions of command line options.

Quick Reference

The following commands are provided by lf:

    quit                     (default 'q')
    up                       (default 'k' and '<up>')
    half-up                  (default '<c-u>')
    page-up                  (default '<c-b>' and '<pgup>')
    down                     (default 'j' and '<down>')
    half-down                (default '<c-d>')
    page-down                (default '<c-f>' and '<pgdn>')
    updir                    (default 'h' and '<left>')
    open                     (default 'l' and '<right>')
    top                      (default 'gg' and '<home>')
    bottom                   (default 'G' and '<end>')
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
    echo
    echomsg
    echoerr
    cd
    select
    delete         (modal)
    rename         (modal)   (default 'r')
    source
    push
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
    mark-save      (modal)   (default 'm')
    mark-load      (modal)   (default "'")
    mark-remove    (modal)   (default `"`)

The following command line commands are provided by lf:

    cmd-escape               (default '<esc>')
    cmd-complete             (default '<tab>')
    cmd-menu-complete
    cmd-menu-complete-back
    cmd-enter                (default '<c-j>' and '<enter>')
    cmd-interrupt            (default '<c-c>')
    cmd-history-next         (default '<c-n>')
    cmd-history-prev         (default '<c-p>')
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
    cmd-capitalize-word      (default '<a-c>')
    cmd-uppercase-word       (default '<a-u>')
    cmd-lowercase-word       (default '<a-l>')

The following options can be used to customize the behavior of lf:

    anchorfind     bool      (default on)
    dircounts      bool      (default off)
    dirfirst       bool      (default on)
    drawbox        bool      (default off)
    errorfmt       string    (default "\033[7;31;47m%s\033[0m")
    filesep        string    (default "\n")
    findlen        int       (default 1)
    globsearch     bool      (default off)
    hidden         bool      (default off)
    hiddenfiles    []string  (default '.*')
    icons          bool      (default off)
    ifs            string    (default '')
    ignorecase     bool      (default on)
    ignoredia      bool      (default on)
    incsearch      bool      (default off)
    info           []string  (default '')
    mouse          bool      (default off)
    number         bool      (default off)
    period         int       (default 0)
    preview        bool      (default on)
    previewer      string    (default '')
    cleaner        string    (default '')
    promptfmt      string    (default "\033[32;1m%u@%h\033[0m:\033[34;1m%d\033[0m\033[1m%f\033[0m")
    ratios         []int     (default '1:2:3')
    relativenumber bool      (default off)
    reverse        bool      (default off)
    scrolloff      int       (default 0)
    shell          string    (default 'sh' for unix and 'cmd' for windows)
    shellflag      string    (default '-c' for unix and '/c' for windows)
    shellopts      []string  (default '')
    smartcase      bool      (default on)
    smartdia       bool      (default off)
    sortby         string    (default 'natural')
    tabstop        int       (default 8)
    timefmt        string    (default 'Mon Jan _2 15:04:05 2006')
    truncatechar   string    (default '~')
    waitmsg        string    (default 'Press any key to continue')
    wrapscan       bool      (default on)
    wrapscroll     bool      (default off)

The following environment variables are exported for shell commands:

    f
    fs
    fx
    id
    PWD
    LF_LEVEL
    OPENER
    EDITOR
    PAGER
    SHELL

The following commands/keybindings are provided by default:

    unix                     windows
    cmd open &$OPENER "$f"   cmd open &%OPENER% %f%
    map e $$EDITOR "$f"      map e $%EDITOR% %f%
    map i $$PAGER "$f"       map i !%PAGER% %f%
    map w $$SHELL            map w $%SHELL%

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
    map sc :set sortby ctime; set info ctime
    map se :set sortby ext; set info
    map gh cd ~
    map <space> :toggle; down

Configuration

Configuration files should be located at:

    os       system-wide             user-specific
    unix     /etc/lf/lfrc            ~/.config/lf/lfrc
    windows  C:\ProgramData\lf\lfrc  C:\Users\<user>\AppData\Local\lf\lfrc

Marks file should be located at:

    unix     ~/.local/share/lf/marks
    windows  C:\Users\<user>\AppData\Local\lf\marks

History file should be located at:

    unix     ~/.local/share/lf/history
    windows  C:\Users\<user>\AppData\Local\lf\history

You can configure the default values of following variables to change these
locations:

    $XDG_CONFIG_HOME  ~/.config
    $XDG_DATA_HOME    ~/.local/share
    %ProgramData%     C:\ProgramData
    %LOCALAPPDATA%    C:\Users\<user>\AppData\Local

A sample configuration file can be found at
https://github.com/gokcehan/lf/blob/master/etc/lfrc.example.

Commands

This section shows information about builtin commands.
Modal commands do not take any arguments, but instead change the operation mode to read their input conveniently, and so they are meant to be assigned to keybindings.

    quit                     (default 'q')

Quit lf and return to the shell.

    up                       (default 'k' and '<up>')
    half-up                  (default '<c-u>')
    page-up                  (default '<c-b>' and '<pgup>')
    down                     (default 'j' and '<down>')
    half-down                (default '<c-d>')
    page-down                (default '<c-f>' and '<pgdn>')

Move the current file selection upwards/downwards by one/half a page/full page.

    updir                    (default 'h' and '<left>')

Change the current working directory to the parent directory.

    open                     (default 'l' and '<right>')

If the current file is a directory, then change the current directory to it, otherwise, execute the 'open' command.
A default 'open' command is provided to call the default system opener asynchronously with the current file as the argument.
A custom 'open' command can be defined to override this default.

(See also 'OPENER' variable and 'Opening Files' section)

    top                      (default 'gg' and '<home>')
    bottom                   (default 'G' and '<end>')

Move the current file selection to the top/bottom of the directory.

    toggle

Toggle the selection of the current file or files given as arguments.

    invert                   (default 'v')

Reverse the selection of all files in the current directory (i.e. 'toggle' all files).
Selections in other directories are not effected by this command.
You can define a new command to select all files in the directory by combining 'invert' with 'unselect' (i.e. `cmd select-all :unselect; invert`), though this will also remove selections in other directories.

    unselect                 (default 'u')

Remove the selection of all files in all directories.

    glob-select

Select files that match the given glob.

    glob-unselect

Unselect files that match the given glob.

    copy                     (default 'y')

If there are no selections, save the path of the current file to the copy buffer, otherwise, copy the paths of selected files.

    cut                      (default 'd')

If there are no selections, save the path of the current file to the cut buffer, otherwise, copy the paths of selected files.

    paste                    (default 'p')

Copy/Move files in copy/cut buffer to the current working directory.

    clear                    (default 'c')

Clear file paths in copy/cut buffer.

    sync

Synchronize copied/cut files with server.
This command is automatically called when required.

    draw

Draw the screen.
This command is automatically called when required.

    redraw                   (default '<c-l>')

Synchronize the terminal and redraw the screen.

    load

Load modified files and directories.
This command is automatically called when required.

    reload                   (default '<c-r>')

Flush the cache and reload all files and directories.

    echo

Print given arguments to the message line at the bottom.

    echomsg

Print given arguments to the message line at the bottom and also to the log file.

    echoerr

Print given arguments to the message line at the bottom in red color and also to the log file.

    cd

Change the working directory to the given argument.

    select

Change the current file selection to the given argument.

    delete         (modal)

Remove the current file or selected file(s).

    rename         (modal)   (default 'r')

Rename the current file using the builtin method.
A custom 'rename' command can be defined to override this default.

    source

Read the configuration file given in the argument.

    push

Simulate key pushes given in the argument.

    read           (modal)   (default ':')

Read a command to evaluate.

    shell          (modal)   (default '$')

Read a shell command to execute.

(See also 'Prefixes' and 'Shell Commands' sections)

    shell-pipe     (modal)   (default '%')

Read a shell command to execute piping its standard I/O to the bottom statline.

(See also 'Prefixes' and 'Piping Shell Commands' sections)

    shell-wait     (modal)   (default '!')

Read a shell command to execute and wait for a key press in the end.

(See also 'Prefixes' and 'Waiting Shell Commands' sections)

    shell-async    (modal)   (default '&')

Read a shell command to execute synchronously without standard I/O.

    find           (modal)   (default 'f')
    find-back      (modal)   (default 'F')
    find-next                (default ';')
    find-prev                (default ',')

Read key(s) to find the appropriate file name match in the forward/backward direction and jump to the next/previous match.

(See also 'anchorfind', 'findlen', 'wrapscan', 'ignorecase', 'smartcase', 'ignoredia', and 'smartdia' options and 'Searching Files' section)

    search                   (default '/')
    search-back              (default '?')
    search-next              (default 'n')
    search-prev              (default 'N')

Read a pattern to search for a file name match in the forward/backward direction and jump to the next/previous match.

(See also 'globsearch', 'incsearch', 'wrapscan', 'ignorecase', 'smartcase', 'ignoredia', and 'smartdia' options and 'Searching Files' section)

    mark-save      (modal)   (default 'm')

Save the current directory as a bookmark assigned to the given key.

    mark-load      (modal)   (default "'")

Change the current directory to the bookmark assigned to the given key.
A special bookmark "'" holds the previous directory after a 'mark-load', 'cd', or 'select' command.

    mark-remove    (modal)   (default `"`)

Remove a bookmark assigned to the given key.

Command Line Commands

This section shows information about command line commands.
These should be mostly compatible with readline keybindings.
A character refers to a unicode code point, a word consists of letters and digits, and a unix word consists of any non-blank characters.

    cmd-escape               (default '<esc>')

Quit command line mode and return to normal mode.

    cmd-complete             (default '<tab>')

Autocomplete the current word.

    cmd-menu-complete

Autocomplete the current word, then you can press the binded key/s
again to cycle completition options.

    cmd-menu-complete-back

Autocomplete the current word, then you can press the binded key/s
again to cycle completition options backwards.

    cmd-enter                (default '<c-j>' and '<enter>')

Execute the current line.

    cmd-interrupt            (default '<c-c>')

Interrupt the current shell-pipe command and return to the normal mode.

    cmd-history-next         (default '<c-n>')
    cmd-history-prev         (default '<c-p>')

Go to next/previous item in the history.

    cmd-left                 (default '<c-b>' and '<left>')
    cmd-right                (default '<c-f>' and '<right>')

Move the cursor to the left/right.

    cmd-home                 (default '<c-a>' and '<home>')
    cmd-end                  (default '<c-e>' and '<end>')

Move the cursor to the beginning/end of line.

    cmd-delete               (default '<c-d>' and '<delete>')
    cmd-delete-back          (default '<backspace>' and '<backspace2>')

Delete the next character in forward/backward direction.

    cmd-delete-home          (default '<c-u>')
    cmd-delete-end           (default '<c-k>')

Delete everything up to the beginning/end of line.

    cmd-delete-unix-word     (default '<c-w>')

Delete the previous unix word.

    cmd-yank                 (default '<c-y>')

Paste the buffer content containing the last deleted item.

    cmd-transpose            (default '<c-t>')
    cmd-transpose-word       (default '<a-t>')

Transpose the positions of last two characters/words.

    cmd-word                 (default '<a-f>')
    cmd-word-back            (default '<a-b>')

Move the cursor by one word in forward/backward direction.

    cmd-delete-word          (default '<a-d>')

Delete the next word in forward direction.

    cmd-capitalize-word      (default '<a-c>')
    cmd-uppercase-word       (default '<a-u>')
    cmd-lowercase-word       (default '<a-l>')

Capitalize/uppercase/lowercase the current word and jump to the next word.

Options

This section shows information about options to customize the behavior.
Character ':' is used as the separator for list options '[]int' and '[]string'.

    anchorfind     bool      (default on)

When this option is enabled, find command starts matching patterns from the beginning of file names, otherwise, it can match at an arbitrary position.

    dircounts      bool      (default off)

When this option is enabled, directory sizes show the number of items inside instead of the size of directory file.
The former needs to be calculated by reading the directory and counting the items inside.
The latter is directly provided by the operating system and it does not require any calculation, though it is non-intuitive and it can often be misleading.
This option is disabled by default for performance reasons.
This option only has an effect when 'info' has a 'size' field and the pane is wide enough to show the information.
A thousand items are counted per directory at most, and bigger directories are shown as '999+'.

    dirfirst       bool      (default on)

Show directories first above regular files.

    drawbox        bool      (default off)

Draw boxes around panes with box drawing characters.

    errorfmt       string    (default "\033[7;31;47m%s\033[0m")

Format string of error messages shown in the bottom message line.

    filesep        string    (default "\n")

File separator used in environment variables 'fs' and 'fx'.

    findlen        int       (default 1)

Number of characters prompted for the find command.
When this value is set to 0, find command prompts until there is only a single match left.

    globsearch     bool      (default off)

When this option is enabled, search command patterns are considered as globs, otherwise they are literals.
With globbing, '*' matches any sequence, '?' matches any character, and '[...]' or '[^...] matches character sets or ranges.
Otherwise, these characters are interpreted as they are.

    hidden         bool      (default off)

Show hidden files.
On unix systems, hidden files are determined by the value of 'hiddenfiles'.
On windows, only files with hidden attributes are considered hidden files.

    hiddenfiles    []string  (default '.*')

List of hidden file glob patterns.
Patterns can be given as relative or absolute paths.
Globbing supports the usual special characters, '*' to match any sequence, '?' to match any character, and '[...]' or '[^...] to match character sets or ranges.
In addition, if a pattern starts with '!', then its matches are excluded from hidden files.

    icons          bool      (default off)

Show icons before each item in the list.
By default, only two icons, 🗀 (U+1F5C0) and 🗎 (U+1F5CE), are used for directories and files respectively, as they are supported in the unicode standard.
Icons can be configured with an environment variable named 'LF_ICONS'.
The syntax of this variable is similar to 'LS_COLORS'.
See the wiki page for an example icon configuration.

    ifs            string    (default '')

Sets 'IFS' variable in shell commands.
It works by adding the assignment to the beginning of the command string as 'IFS='...'; ...'.
The reason is that 'IFS' variable is not inherited by the shell for security reasons.
This method assumes a POSIX shell syntax and so it can fail for non-POSIX shells.
This option has no effect when the value is left empty.
This option does not have any effect on windows.

    ignorecase     bool      (default on)

Ignore case in sorting and search patterns.

    ignoredia      bool      (default on)

Ignore diacritics in sorting and search patterns.

    incsearch      bool      (default off)

Jump to the first match after each keystroke during searching.

    info           []string  (default '')

List of information shown for directory items at the right side of pane.
Currently supported information types are 'size', 'time', 'atime', and 'ctime'.
Information is only shown when the pane width is more than twice the width of information.

    mouse          bool      (default off)

Send mouse events as input.

    number         bool      (default off)

Show the position number for directory items at the left side of pane.
When 'relativenumber' is enabled, only the current line shows the absolute position and relative positions are shown for the rest.

    period         int       (default 0)

Set the interval in seconds for periodic checks of directory updates.
This works by periodically calling the 'load' command.
Note that directories are already updated automatically in many cases.
This option can be useful when there is an external process changing the displayed directory and you are not doing anything in lf.
Periodic checks are disabled when the value of this option is set to zero.

    preview        bool      (default on)

Show previews of files and directories at the right most pane.
If the file has more lines than the preview pane, rest of the lines are not read.
Files containing the null character (U+0000) in the read portion are considered binary files and displayed as 'binary'.

    previewer      string    (default '') (not filtered if empty)

Set the path of a previewer file to filter the content of regular files for previewing.
The file should be executable.
Five arguments are passed to the file, first is the current file name; the second, third, fourth, and fifth are width, height, horizontal position, and vertical position of preview pane respectively.
SIGPIPE signal is sent when enough lines are read.
If the previewer returns a non-zero exit code, then the preview cache for the given file is disabled. This means that if the file is selected in the future, the previewer is called once again.
Preview filtering is disabled and files are displayed as they are when the value of this option is left empty.

    cleaner        string    (default '') (not called if empty)

Set the path of a cleaner file. This file will be called if previewing is enabled, the previewer is set, and the previously selected file had its preview cache disabled.
The file should be executable.
One argument is passed to the file; the path to the file whose preview should be cleaned.
Preview clearing is disabled when the value of this option is left empty.

    promptfmt      string    (default "\033[32;1m%u@%h\033[0m:\033[34;1m%d\033[0m\033[1m%f\033[0m")

Format string of the prompt shown in the top line.
Special expansions are provided, '%u' as the user name, '%h' as the host name, '%w' as the working directory, '%d' as the working directory with a trailing path separator, and '%f' as the file name.
Home folder is shown as '~' in the working directory expansion.
Directory names are automatically shortened to a single character starting from the left most parent when the prompt does not fit to the screen.

    ratios         []int     (default '1:2:3')

List of ratios of pane widths.
Number of items in the list determines the number of panes in the ui.
When 'preview' option is enabled, the right most number is used for the width of preview pane.

    relativenumber bool      (default off)

Show the position number relative to the current line.
When 'number' is enabled, current line shows the absolute position, otherwise nothing is shown.

    reverse        bool      (default off)

Reverse the direction of sort.

    scrolloff      int       (default 0)

Minimum number of offset lines shown at all times in the top and the bottom of the screen when scrolling.
The current line is kept in the middle when this option is set to a large value that is bigger than the half of number of lines.
A smaller offset can be used when the current file is close to the beginning or end of the list to show the maximum number of items.

    shell          string    (default 'sh' for unix and 'cmd' for windows)

Shell executable to use for shell commands.
Shell commands are executed as 'shell shellopts shellflag command -- arguments'.

    shellflag      string    (default '-c' for unix and '/c' for windows)

Command line flag used to pass shell commands.

    shellopts      []string  (default '')

List of shell options to pass to the shell executable.

    smartcase      bool      (default on)

Override 'ignorecase' option when the pattern contains an uppercase character.
This option has no effect when 'ignorecase' is disabled.

    smartdia       bool      (default off)

Override 'ignoredia' option when the pattern contains a character with diacritic.
This option has no effect when 'ignoredia' is disabled.

    sortby         string    (default 'natural')

Sort type for directories.
Currently supported sort types are 'natural', 'name', 'size', 'time', 'ctime', 'atime', and 'ext'.

    tabstop        int       (default 8)

Number of space characters to show for horizontal tabulation (U+0009) character.

    timefmt        string    (default 'Mon Jan _2 15:04:05 2006')

Format string of the file modification time shown in the bottom line.

    truncatechar   string    (default '~')

Truncate character shown at the end when the file name does not fit to the pane.

    waitmsg        string    (default 'Press any key to continue')

String shown after commands of shell-wait type.

    wrapscan       bool      (default on)

Searching can wrap around the file list.

    wrapscroll     bool      (default off)

Scrolling can wrap around the file list.

Environment Variables

The following variables are exported for shell commands:
These are referred with a '$' prefix on POSIX shells (e.g. '$f'), between '%' characters on Windows cmd (e.g. '%f%'), and with a '$env:' prefix on Windows powershell (e.g. '$env:f').

    f

Current file selection as a full path.

    fs

Selected file(s) separated with the value of 'filesep' option as full path(s).

    fx

Selected file(s) (i.e. 'fs') if there are any selected files, otherwise current file selection (i.e. 'f').

    id

Id of the running client.

    PWD

Present working directory.

    LF_LEVEL

The value of this variable is set to the current nesting level when you run lf from a shell spawned inside lf.
You can add the value of this variable to your shell prompt to make it clear that your shell runs inside lf.
For example, with POSIX shells, you can use '[ -n "$LF_LEVEL" ] && PS1="$PS1""(lf level: $LF_LEVEL) "' in your shell configuration file (e.g. '~/.bashrc').

    OPENER

If this variable is set in the environment, use the same value, otherwise set the value to 'start' in Windows, 'open' in MacOS, 'xdg-open' in others.

    EDITOR

If this variable is set in the environment, use the same value, otherwise set the value to 'vi' on unix, 'notepad' in Windows.

    PAGER

If this variable is set in the environment, use the same value, otherwise set the value to 'less' on unix, 'more' in Windows.

    SHELL

If this variable is set in the environment, use the same value, otherwise set the value to 'sh' on unix, 'cmd' in Windows.

Prefixes

The following command prefixes are used by lf:

    :  read (default)  builtin/custom command
    $  shell           shell command
    %  shell-pipe      shell command running with the ui
    !  shell-wait      shell command waiting for key press
    &  shell-async     shell command running asynchronously

The same evaluator is used for the command line and the configuration file for read and shell commands.
The difference is that prefixes are not necessary in the command line.
Instead, different modes are provided to read corresponding commands.
These modes are mapped to the prefix keys above by default.

Syntax

Characters from '#' to newline are comments and ignored:

    # comments start with '#'

There are three special commands ('set', 'map', and 'cmd') and their variants for configuration.

Command 'set' is used to set an option which can be boolean, integer, or string:

    set hidden         # boolean on
    set nohidden       # boolean off
    set hidden!        # boolean toggle
    set scrolloff 10   # integer value
    set sortby time    # string value w/o quotes
    set sortby 'time'  # string value with single quotes (whitespaces)
    set sortby "time"  # string value with double quotes (backslash escapes)

Command 'map' is used to bind a key to a command which can be builtin command, custom command, or shell command:

    map gh cd ~        # builtin command
    map D trash        # custom command
    map i $less $f     # shell command
    map U !du -sh      # waiting shell command

Command 'cmap' is used to bind a key to a command line command which can only be one of the builtin commands:

    cmap <c-g> cmd-escape

You can delete an existing binding by leaving the expression empty:

    map gh             # deletes 'gh' mapping
    cmap <c-g>         # deletes '<c-g>' mapping

Command 'cmd' is used to define a custom command:

    cmd usage $du -h -d1 | less

You can delete an existing command by leaving the expression empty:

    cmd trash          # deletes 'trash' command

If there is no prefix then ':' is assumed:

    map zt set info time

An explicit ':' can be provided to group statements until a newline which is especially useful for 'map' and 'cmd' commands:

    map st :set sortby time; set info time

If you need multiline you can wrap statements in '{{' and '}}' after the proper prefix.

    map st :{{
        set sortby time
        set info time
    }}

Key Mappings

Regular keys are assigned to a command with the usual syntax:

    map a down

Keys combined with the shift key simply use the uppercase letter:

    map A down

Special keys are written in between '<' and '>' characters and always use lowercase letters:

    map <enter> down

Angle brackets can be assigned with their special names:

    map <lt> down
    map <gt> down

Function keys are prefixed with 'f' character:

    map <f-1> down

Keys combined with the control key are prefixed with 'c' character:

    map <c-a> down

Keys combined with the alt key are assigned in two different ways depending on the behavior of your terminal.
Older terminals (e.g. xterm) may set the 8th bit of a character when the alt key is pressed.
On these terminals, you can use the corresponding byte for the mapping:

    map á down

Newer terminals (e.g. gnome-terminal) may prefix the key with an escape key when the alt key is pressed.
lf uses the escape delaying mechanism to recognize alt keys in these terminals (delay is 100ms).
On these terminals, keys combined with the alt key are prefixed with 'a' character:

    map <a-a> down

Please note that, some key combinations are not possible due to the way terminals work (e.g. control and h combination sends a backspace key instead).
The easiest way to find the name of a key combination is to press the key while lf is running and read the name of the key from the unknown mapping error.

Mouse buttons are prefixed with 'm' character:

    map <m-1> down  # primary
    map <m-2> down  # secondary
    map <m-3> down  # middle
    map <m-4> down
    map <m-5> down
    map <m-6> down
    map <m-7> down
    map <m-8> down

Mouse wheel events are also prefixed with 'm' character:

    map <m-up>    down
    map <m-down>  down
    map <m-left>  down
    map <m-right> down

Push Mappings

The usual way to map a key sequence is to assign it to a named or unnamed command.
While this provides a clean way to remap builtin keys as well as other commands, it can be limiting at times.
For this reason 'push' command is provided by lf.
This command is used to simulate key pushes given as its arguments.
You can 'map' a key to a 'push' command with an argument to create various keybindings.

This is mainly useful for two purposes.
First, it can be used to map a command with a command count:

    map <c-j> push 10j

Second, it can be used to avoid typing the name when a command takes arguments:

    map r push :rename<space>

One thing to be careful is that since 'push' command works with keys instead of commands it is possible to accidentally create recursive bindings:

    map j push 2j

These types of bindings create a deadlock when executed.

Shell Commands

Regular shell commands are the most basic command type that is useful for many purposes.
For example, we can write a shell command to move selected file(s) to trash.
A first attempt to write such a command may look like this:

    cmd trash ${{
        mkdir -p ~/.trash
        if [ -z "$fs" ]; then
            mv "$f" ~/.trash
        else
            IFS="`printf '\n\t'`"; mv $fs ~/.trash
        fi
    }}

We check '$fs' to see if there are any selected files.
Otherwise we just delete the current file.
Since this is such a common pattern, a separate '$fx' variable is provided.
We can use this variable to get rid of the conditional:

    cmd trash ${{
        mkdir -p ~/.trash
        IFS="`printf '\n\t'`"; mv $fx ~/.trash
    }}

The trash directory is checked each time the command is executed.
We can move it outside of the command so it would only run once at startup:

    ${{ mkdir -p ~/.trash }}

    cmd trash ${{ IFS="`printf '\n\t'`"; mv $fx ~/.trash }}

Since these are one liners, we can drop '{{' and '}}':

    $mkdir -p ~/.trash

    cmd trash $IFS="`printf '\n\t'`"; mv $fx ~/.trash

Finally note that we set 'IFS' variable manually in these commands.
Instead we could use the 'ifs' option to set it for all shell commands (i.e. 'set ifs "\n"').
This can be especially useful for interactive use (e.g. '$rm $f' or '$rm $fs' would simply work).
This option is not set by default as it can behave unexpectedly for new users.
However, use of this option is highly recommended and it is assumed in the rest of the documentation.

Piping Shell Commands

Regular shell commands have some limitations in some cases.
When an output or error message is given and the command exits afterwards, the ui is immediately resumed and there is no way to see the message without dropping to shell again.
Also, even when there is no output or error, the ui still needs to be paused while the command is running.
This can cause flickering on the screen for short commands and similar distractions for longer commands.

Instead of pausing the ui, piping shell commands connects stdin, stdout, and stderr of the command to the statline in the bottom of the ui.
This can be useful for programs following the unix philosophy to give no output in the success case, and brief error messages or prompts in other cases.

For example, following rename command prompts for overwrite in the statline if there is an existing file with the given name:

    cmd rename %mv -i $f $1

You can also output error messages in the command and it will show up in the statline.
For example, an alternative rename command may look like this:

    cmd rename %[ -e $1 ] && printf "file exists" || mv $f $1

Note that input is line buffered and output and error are byte buffered.

Waiting Shell Commands

Waiting shell commands are similar to regular shell commands except that they wait for a key press when the command is finished.
These can be useful to see the output of a program before the ui is resumed.
Waiting shell commands are more appropriate than piping shell commands when the command is verbose and the output is best displayed as multiline.

Asynchronous Shell Commands

Asynchronous shell commands are used to start a command in the background and then resume operation without waiting for the command to finish.
Stdin, stdout, and stderr of the command is neither connected to the terminal nor to the ui.

Remote Commands

One of the more advanced features in lf is remote commands.
All clients connect to a server on startup.
It is possible to send commands to all or any of the connected clients over the common server.
This is used internally to notify file selection changes to other clients.

To use this feature, you need to use a client which supports communicating with a UNIX-domain socket.
OpenBSD implementation of netcat (nc) is one such example.
You can use it to send a command to the socket file:

    echo 'send echo hello world' | nc -U /tmp/lf.${USER}.sock

Since such a client may not be available everywhere, lf comes bundled with a command line flag to be used as such.
When using lf, you do not need to specify the address of the socket file.
This is the recommended way of using remote commands since it is shorter and immune to socket file address changes:

    lf -remote 'send echo hello world'

In this command 'send' is used to send the rest of the string as a command to all connected clients.
You can optionally give it an id number to send a command to a single client:

    lf -remote 'send 1000 echo hello world'

All clients have a unique id number but you may not be aware of the id number when you are writing a command.
For this purpose, an '$id' variable is exported to the environment for shell commands.
You can use it to send a remote command from a client to the server which in return sends a command back to itself.
So now you can display a message in the current client by calling the following in a shell command:

    lf -remote "send $id echo hello world"

Since lf does not have control flow syntax, remote commands are used for such needs.
For example, you can configure the number of columns in the ui with respect to the terminal width as follows:

    cmd recol %{{
        w=$(tput cols)
        if [ $w -le 80 ]; then
            lf -remote "send $id set ratios 1:2"
        elif [ $w -le 160 ]; then
            lf -remote "send $id set ratios 1:2:3"
        else
            lf -remote "send $id set ratios 1:2:3:5"
        fi
    }}

Besides 'send' command, there are also two commands to get or set the current file selection.
Two possible modes 'copy' and 'move' specify whether selected files are to be copied or moved.
File names are separated by newline character.
Setting the file selection is done with 'save' command:

    lf -remote "$(printf 'save\ncopy\nfoo.txt\nbar.txt\nbaz.txt\n')"

Getting the file selection is similarly done with 'load' command:

    load=$(lf -remote 'load')
    mode=$(echo "$load" | sed -n '1p')
    list=$(echo "$load" | sed '1d')
    if [ $mode = 'copy' ]; then
        # do something with $list
    elif [ $mode = 'move' ]; then
        # do something else with $list
    fi

There is a 'quit' command to close client connections and quit the server:

    lf -remote 'quit'

Lastly, there is a 'conn' command to connect the server as a client.
This should not be needed for users.

File Operations

lf uses its own builtin copy and move operations by default.
These are implemented as asynchronous operations and progress is shown in the bottom ruler.
These commands do not overwrite existing files or directories with the same name.
Instead, a suffix that is compatible with '--backup=numbered' option in GNU cp is added to the new files or directories.
Only file modes are preserved and all other attributes are ignored including ownership, timestamps, context, and xattr.
Special files such as character and block devices, named pipes, and sockets are skipped and links are not followed.
Moving is performed using the rename operation of the underlying OS.
For cross-device moving, lf falls back to copying and then deletes the original files if there are no errors.
Operation errors are shown in the message line as well as the log file and they do not preemptively finish the corresponding file operation.

File operations can be performed on the current selected file or alternatively on multiple files by selecting them first.
When you 'copy' a file, lf doesn't actually copy the file on the disk, but only records its name to memory.
The actual file copying takes place when you 'paste'.
Similarly 'paste' after a 'cut' operation moves the file.

You can customize copy and move operations by defining a 'paste' command.
This is a special command that is called when it is defined instead of the builtin implementation.
You can use the following example as a starting point:

    cmd paste %{{
        load=$(lf -remote 'load')
        mode=$(echo "$load" | sed -n '1p')
        list=$(echo "$load" | sed '1d')
        if [ $mode = 'copy' ]; then
            cp -R $list .
        elif [ $mode = 'move' ]; then
            mv $list .
        fi
        lf -remote 'send load'
        lf -remote 'send clear'
    }}

Some useful things to be considered are to use the backup ('--backup') and/or preserve attributes ('-a') options with 'cp' and 'mv' commands if they support it (i.e. GNU implementation), change the command type to asynchronous, or use 'rsync' command with progress bar option for copying and feed the progress to the client periodically with remote 'echo' calls.

By default, lf does not assign 'delete' command to a key to protect new users.
You can customize file deletion by defining a 'delete' command.
You can also assign a key to this command if you like.
An example command to move selected files to a trash folder and remove files completely after a prompt are provided in the example configuration file.

Searching Files

There are two mechanisms implemented in lf to search a file in the current directory.
Searching is the traditional method to move the selection to a file matching a given pattern.
Finding is an alternative way to search for a pattern possibly using fewer keystrokes.

Searching mechanism is implemented with commands 'search' (default '/'), 'search-back' (default '?'), 'search-next' (default 'n'), and 'search-prev' (default 'N').
You can enable 'globsearch' option to match with a glob pattern.
Globbing supports '*' to match any sequence, '?' to match any character, and '[...]' or '[^...] to match character sets or ranges.
You can enable 'incsearch' option to jump to the current match at each keystroke while typing.
In this mode, you can either use 'cmd-enter' to accept the search or use 'cmd-escape' to cancel the search.
Alternatively, you can also map some other commands with 'cmap' to accept the search and execute the command immediately afterwards.
Possible candidates are 'up', 'down' and their variants, 'top', 'bottom', 'updir', and 'open' commands.
For example, you can use arrow keys to finish the search with the following mappings:

    cmap <up>    up
    cmap <down>  down
    cmap <left>  updir
    cmap <right> open

Finding mechanism is implemented with commands 'find' (default 'f'), 'find-back' (default 'F'), 'find-next' (default ';'), 'find-prev' (default ',').
You can disable 'anchorfind' option to match a pattern at an arbitrary position in the filename instead of the beginning.
You can set the number of keys to match using 'findlen' option.
If you set this value to zero, then the the keys are read until there is only a single match.
Default values of these two options are set to jump to the first file with the given initial.

Some options effect both searching and finding.
You can disable 'wrapscan' option to prevent searches to wrap around at the end of the file list.
You can disable 'ignorecase' option to match cases in the pattern and the filename.
This option is already automatically overridden if the pattern contains upper case characters.
You can disable 'smartcase' option to disable this behavior.
Two similar options 'ignoredia' and 'smartdia' are provided to control matching diacritics in latin letters.

Opening Files

You can define a an 'open' command (default 'l' and '<right>') to configure file opening.
This command is only called when the current file is not a directory, otherwise the directory is entered instead.
You can define it just as you would define any other command:

    cmd open $vi $fx

It is possible to use different command types:

    cmd open &xdg-open $f

You may want to use either file extensions or mime types from 'file' command:

    cmd open ${{
        case $(file --mime-type $f -b) in
            text/*) vi $fx;;
            *) for f in $fx; do xdg-open $f > /dev/null 2> /dev/null & done;;
        esac
    }}

You may want to use 'setsid' before your opener command to have persistent processes that continue to run after lf quits.

Following command is provided by default:

    cmd open &$OPENER $f

You may also use any other existing file openers as you like.
Possible options are 'libfile-mimeinfo-perl' (executable name is 'mimeopen'), 'rifle' (ranger's default file opener), or 'mimeo' to name a few.

Previewing Files

lf previews files on the preview pane by printing the file until the end or the preview pane is filled.
This output can be enhanced by providing a custom preview script for filtering.
This can be used to highlight source codes, list contents of archive files or view pdf or image files as text to name few.
For coloring lf recognizes ansi escape codes.

In order to use this feature you need to set the value of 'previewer' option to the path of an executable file.
lf passes the current file name as the first argument and the height of the preview pane as the second argument when running this file.
Output of the execution is printed in the preview pane.
You may want to use the same script in your pager mapping as well if any:

    set previewer ~/.config/lf/pv.sh
    map i $~/.config/lf/pv.sh $f | less -R

For 'less' pager, you may instead utilize 'LESSOPEN' mechanism so that useful information about the file such as the full path of the file can be displayed in the statusline below:

    set previewer ~/.config/lf/pv.sh
    map i $LESSOPEN='| ~/.config/lf/pv.sh %s' less -R $f

Since this script is called for each file selection change it needs to be as efficient as possible and this responsibility is left to the user.
You may use file extensions to determine the type of file more efficiently compared to obtaining mime types from 'file' command.
Extensions can then be used to match cleanly within a conditional:

    #!/bin/sh

    case "$1" in
        *.tar*) tar tf "$1";;
        *.zip) unzip -l "$1";;
        *.rar) unrar l "$1";;
        *.7z) 7z l "$1";;
        *.pdf) pdftotext "$1" -;;
        *) highlight -O ansi "$1";;
    esac

Another important consideration for efficiency is the use of programs with short startup times for preview.
For this reason, 'highlight' is recommended over 'pygmentize' for syntax highlighting.
Besides, it is also important that the application is processing the file on the fly rather than first reading it to the memory and then do the processing afterwards.
This is especially relevant for big files.
lf automatically closes the previewer script output pipe with a SIGPIPE when enough lines are read.
When everything else fails, you can make use of the height argument to only feed the first portion of the file to a program for preview.
Note that some programs may not respond well to SIGPIPE to exit with a non-zero return code and avoid caching.
You may add a trailing '|| true' command to avoid such errors:

    highlight -O ansi "$1" || true

You may also use an existing preview filter as you like.
Your system may already come with a preview filter named 'lesspipe'.
These filters may have a mechanism to add user customizations as well.
See the related documentations for more information.

Changing Directory

lf changes the working directory of the process to the current directory so that shell commands always work in the displayed directory.
After quitting, it returns to the original directory where it is first launched like all shell programs.
If you want to stay in the current directory after quitting, you can use one of the example wrapper shell scripts provided in the repository.

There is a special command 'on-cd' that runs a shell command when it is defined and the directory is changed.
You can define it just as you would define any other command:

    cmd on-cd &{{
        # display git repository status in your prompt
        source /usr/share/git/completion/git-prompt.sh
        GIT_PS1_SHOWDIRTYSTATE=auto
        GIT_PS1_SHOWSTASHSTATE=auto
        GIT_PS1_SHOWUNTRACKEDFILES=auto
        GIT_PS1_SHOWUPSTREAM=auto
        git=$(__git_ps1 " (%s)") || true
        fmt="\033[32;1m%u@%h\033[0m:\033[34;1m%d\033[0m\033[1m%f$git\033[0m"
        lf -remote "send $id set promptfmt \"$fmt\""
    }}

If you want to print escape sequences, you may redirect 'printf' output to '/dev/tty'.
The following xterm specific escape sequence sets the terminal title to the working directory:

    cmd on-cd &{{
        printf "\033]0; $PWD\007" > /dev/tty
    }}

This command runs whenever you change directory but not on startup.
You can add an extra call to make it run on startup as well:

    cmd on-cd &{{ # ... }}
    on-cd

Note that all shell commands are possible but `%` and `&` are usually more appropriate as `$` and `!` causes flickers and pauses respectively.

Colors

lf tries to automatically adapt its colors to the environment.
It starts with a default colorscheme and updates colors using values of existing environment variables possibly by overwriting its previous values.
Colors are set in the following order:

    1. default
    2. LSCOLORS (Mac/BSD ls)
    3. LS_COLORS (GNU ls)
    4. LF_COLORS (lf specific)

Please refer to the corresponding man pages for more information about 'LSCOLORS' and 'LS_COLORS'.
'LF_COLORS' is provided with the same syntax as 'LS_COLORS' in case you want to configure colors only for lf but not ls.
This can be useful since there are some differences between ls and lf, though one should expect the same behavior for common cases.

You can configure lf colors in two different ways.
First, you can only configure 8 basic colors used by your terminal and lf should pick up those colors automatically.
Depending on your terminal, you should be able to select your colors from a 24-bit palette.
This is the recommended approach as colors used by other programs will also match each other.

Second, you can set the values of environmental variables mentioned above for fine grained customization.
Note that 'LS_COLORS/LF_COLORS' are more powerful than 'LSCOLORS' and they can be used even when GNU programs are not installed on the system.
You can combine this second method with the first method for best results.

Lastly, you may also want to configure the colors of the prompt line to match the rest of the colors.
Colors of the prompt line can be configured using the 'promptfmt' option which can include hardcoded colors as ansi escapes.
See the default value of this option to have an idea about how to color this line.

It is worth noting that lf uses as many colors are advertised by your terminal's entry in your systems terminfo or infocmp database, if this is not present lf will default to an internal database.
For terminals supporting 24-bit (or "true") color that do not have a database entry (or one that does not advertise all capabilities), support can be enabled by either setting the '$COLORTERM' variable to "truecolor" or ensuring '$TERM' is set to a value that ends with "-truecolor".

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

Note that, lf first tries matching file names and then falls back to file types.
The full order of matchings from most specific to least are as follows:

    1. Full Path (e.g. '~/.config/lf/lfrc')
    2. Dir Name  (e.g. '.git/') (only matches dirs with a trailing slash at the end)
    3. File Type (e.g. 'ln') (except 'fi')
    4. File Name (e.g. '.git*') (only matches files with a trailing star at the end)
    5. Base Name (e.g. 'README.*')
    6. Extension (e.g. '*.txt')
    7. Default   (i.e. 'fi')

For example, given a regular text file '/path/to/README.txt', the following entries are checked in the configuration and the first one to match is used:

    1. '/path/to/README.txt'
    2. (skipped since the file is not a directory)
    3. (skipped since the file is of type 'fi')
    4. 'README.txt*'
    5. 'README.*'
    6. '*.txt'
    7. 'fi'

Given a regular directory '/path/to/example.d', the following entries are checked in the configuration and the first one to match is used:

    1. '/path/to/example.d'
    2. 'example.d/'
    3. 'di'
    4. 'example.d*'
    5. 'example.*'
    6. '*.d'
    7. 'fi'

Note that glob-like patterns do not actually perform glob matching due to performance reasons.

For example, you can set a variable as follows:

    export LF_COLORS="~/Documents=01;31:~/Downloads=01;31:~/.local/share=01;31:~/.config/lf/lfrc=31:.git/=01;32:.git=32:.gitignore=32:Makefile=32:README.*=33:*.txt=34:*.md=34:ln=01;36:di=01;34:ex=01;32:"

Having all entries on a single line can make it hard to read.
You may instead divide it to multiple lines in between double quotes by escaping newlines with backslashes as follows:

    export LF_COLORS="\
    ~/Documents=01;31:\
    ~/Downloads=01;31:\
    ~/.local/share=01;31:\
    ~/.config/lf/lfrc=31:\
    .git/=01;32:\
    .git=32:\
    .gitignore=32:\
    Makefile=32:\
    README.*=33:\
    *.txt=34:\
    *.md=34:\
    ln=01;36:\
    di=01;34:\
    ex=01;32:\
    "

Having such a long variable definition in a shell configuration file might be undesirable.
You may instead put this definition in a separate file and source it in your shell configuration file as follows:

    [ -f "/path/to/colors" ] && source "/path/to/colors"

See the wiki page for ansi escape codes
https://en.wikipedia.org/wiki/ANSI_escape_code.

Icons

Icons are configured using 'LF_ICONS' environment variable.
This variable uses the same syntax as 'LS_COLORS/LF_COLORS'.
Instead of colors, you should put a single characters as values of entries.
Do not forget to enable 'icons' option to see the icons.
Default values are as follows given with their matching order in lf:

    ln  🗎
    or  🗎
    tw  🗀
    ow  🗀
    st  🗀
    di  🗀
    pi  🗎
    so  🗎
    bd  🗎
    cd  🗎
    su  🗎
    sg  🗎
    ex  🗎
    fi  🗎

See the wiki page for an example icons configuration
https://github.com/gokcehan/lf/wiki/Icons.
*/
package main
