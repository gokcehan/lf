//go:generate gen/docstring.sh
//go:generate gen/man.sh

/*
lf is a terminal file manager.

Source code can be found in the repository at https://github.com/gokcehan/lf.

This documentation can either be read from terminal using 'lf -doc' or online
at https://godoc.org/github.com/gokcehan/lf. You can also use 'doc' command
(default '<f-1>') inside lf to view the documentation in a pager.

You can run 'lf -help' to see descriptions of command line options.

Reference

The following commands are provided by lf with default keybindings:

    up                       (default 'k' and '<up>')
    half-up                  (default '<c-u>')
    page-up                  (default '<c-b>' and '<pgup>')
    down                     (default 'j' and '<down>')
    half-down                (default '<c-d>')
    page-down                (default '<c-f>' and '<pgdn>')
    updir                    (default 'h' and '<left>')
    open                     (default 'l' and '<right>')
    quit                     (default 'q')
    top                      (default 'gg' and '<home>')
    bottom                   (default 'G' and '<end>')
    invert                   (default 'v')
    unselect                 (default 'u')
    copy                     (default 'y')
    cut                      (default 'd')
    paste                    (default 'p')
    clear                    (default 'c')
    redraw                   (default '<c-l>')
    reload                   (default '<c-r>')
    read                     (default ':')
    rename                   (default 'r')
    shell                    (default '$')
    shell-pipe               (default '%')
    shell-wait               (default '!')
    shell-async              (default '&')
    find                     (default 'f')
    find-back                (default 'F')
    find-next                (default ';')
    find-prev                (default ',')
    search                   (default '/')
    search-back              (default '?')
    search-next              (default 'n')
    search-prev              (default 'N')
    mark-save                (default 'm')
    mark-load                (default "'")
    mark-remove              (default `"`)

The following commands are provided by lf without default keybindings:

    draw           draw the ui
    load           load modified files and directories
    sync           synchronize copied/cut files with server
    echo           print arguments to the message line
    echomsg        same as echo but logging
    echoerr        same as echomsg but red color
    cd             change working directory to the argument
    select         change current file selection to the argument
    toggle         toggle the selection of the current file
    glob-select    select files that match the given glob
    glob-unselect  unselect files that match the given glob
    source         read the configuration file in the argument
    push           simulate key pushes given in the argument
    delete         remove the current file or selected file(s)

The following command line commands are provided by lf with default
keybindings:

    cmd-escape               (default '<esc>')
    cmd-complete             (default '<tab>')
    cmd-enter                (default '<c-j>' and '<enter>')
    cmd-history-next         (default '<c-n>')
    cmd-history-prev         (default '<c-p>')
    cmd-delete               (default '<c-d>' and '<delete>')
    cmd-delete-back          (default '<bs>' and '<bs2>')
    cmd-left                 (default '<c-b>' and '<left>')
    cmd-right                (default '<c-f>' and '<right>')
    cmd-home                 (default '<c-a>' and '<home>')
    cmd-end                  (default '<c-e>' and '<end>')
    cmd-delete-home          (default '<c-u>')
    cmd-delete-end           (default '<c-k>')
    cmd-delete-unix-word     (default '<c-w>')
    cmd-yank                 (default '<c-y>')
    cmd-transpose            (default '<c-t>')
    cmd-interrupt            (default '<c-c>')
    cmd-word                 (default '<a-f>')
    cmd-word-back            (default '<a-b>')
    cmd-capitalize-word      (default '<a-c>')
    cmd-delete-word          (default '<a-d>')
    cmd-uppercase-word       (default '<a-u>')
    cmd-lowercase-word       (default '<a-l>')
    cmd-transpose-word       (default '<a-t>')

The following options can be used to customize the behavior of lf:

    anchorfind      boolean  (default on)
    color256        boolean  (default off)
    dircounts       boolean  (default off)
    dirfirst        boolean  (default on)
    drawbox         boolean  (default off)
    globsearch      boolean  (default off)
    hidden          boolean  (default off)
    icons           boolean  (default off)
    ignorecase      boolean  (default on)
    ignoredia       boolean  (default off)
    incsearch       boolean  (default off)
    number          boolean  (default off)
    preview         boolean  (default on)
    relativenumber  boolean  (default off)
    reverse         boolean  (default off)
    smartcase       boolean  (default on)
    smartdia        boolean  (default off)
    wrapscan        boolean  (default on)
    wrapscroll      boolean  (default off)
    findlen         integer  (default 1) (zero to prompt until single match)
    period          integer  (default 0) (zero to disable periodic loading)
    scrolloff       integer  (default 0)
    tabstop         integer  (default 8)
    errorfmt        string   (default "\033[7;31;47m%s\033[0m")
    filesep         string   (default "\n")
    hiddenfiles     string   (default '.*')
    ifs             string   (default '') (not exported if empty)
    info            string   (default '')
    previewer       string   (default '') (not filtered if empty)
    promptfmt       string   (default "\033[32;1m%u@%h\033[0m:\033[34;1m%w/\033[0m\033[1m%f\033[0m")
    ratios          string   (default '1:2:3')
    shell           string   (default 'sh')
    shellopts       string   (default '')
    sortby          string   (default 'natural')
    timefmt         string   (default 'Mon Jan _2 15:04:05 2006')

The following variables are exported for shell commands:

    $f   current file
    $fs  selected file(s) separated with 'filesep'
    $fx  current file or selected file(s) if any
    $id  id number of the client

The following variables are set to the corresponding values:

    $LF_LEVEL  current nesting level

The following default values are set to the environmental variables on unix
when they are not set or empty:

    $OPENER  open      # macos
    $OPENER  xdg-open  # others
    $EDITOR  vi
    $PAGER   less
    $SHELL   sh

The following default values are set to the environmental variables on windows
when they are not set or empty:

    %OPENER%  start
    %EDITOR%  notepad
    %PAGER%   more
    %SHELL%   cmd

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

The following keybindings to applications are provided by default:

    map e $$EDITOR $f
    map i $$PAGER $f
    map w $$SHELL

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

Prefixes

The following command prefixes are used by lf:

    :  read (default)  builtin/custom command
    $  shell           shell command
    %  shell-pipe      shell command running with the ui
    !  shell-wait      shell command waiting for key press
    &  shell-async     shell command running asynchronously

The same evaluator is used for the command line and the configuration file for
read and shell commands. The difference is that prefixes are not necessary in
the command line. Instead, different modes are provided to read corresponding
commands. These modes are mapped to the prefix keys above by default.

Syntax

Characters from '#' to newline are comments and ignored:

    # comments start with '#'

There are three special commands ('set', 'map', and 'cmd') and their variants
for configuration.

Command 'set' is used to set an option which can be boolean, integer, or
string:

    set hidden         # boolean on
    set nohidden       # boolean off
    set hidden!        # boolean toggle
    set scrolloff 10   # integer value
    set sortby time    # string value w/o quotes
    set sortby 'time'  # string value with single quotes (whitespaces)
    set sortby "time"  # string value with double quotes (backslash escapes)

Command 'map' is used to bind a key to a command which can be builtin command,
custom command, or shell command:

    map gh cd ~        # builtin command
    map D trash        # custom command
    map i $less $f     # shell command
    map U !du -sh      # waiting shell command

Command 'cmap' is used to bind a key to a command line command which can only
be one of the builtin commands:

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

An explicit ':' can be provided to group statements until a newline which is
especially useful for 'map' and 'cmd' commands:

    map st :set sortby time; set info time

If you need multiline you can wrap statements in '{{' and '}}' after the proper
prefix.

    map st :{{
        set sortby time
        set info time
    }}

Key Mappings

Regular keys are assigned to a command with the usual syntax:

    map a down

Keys combined with the shift key simply use the uppercase letter:

    map A down

Special keys are written in between '<' and '>' characters and always use
lowercase letters:

    map <enter> down

Angle brackets can be assigned with their special names:

    map <lt> down
    map <gt> down

Function keys are prefixed with 'f' character:

    map <f-1> down

Keys combined with the control key are prefixed with 'c' character:

    map <c-a> down

Keys combined with the alt key are assigned in two different ways depending on
the behavior of your terminal. Older terminals (e.g. xterm) may set the 8th bit
of a character when the alt key is pressed. On these terminals, you can use the
corresponding byte for the mapping:

    map á down

Newer terminals (e.g. gnome-terminal) may prefix the key with an escape key
when the alt key is pressed. lf uses the escape delaying mechanism to recognize
alt keys in these terminals (delay is 100ms). On these terminals, keys combined
with the alt key are prefixed with 'a' character:

    map <a-a> down

Please note that, some key combinations are not possible due to the way
terminals work (e.g. control and h combination sends a backspace key instead).
The easiest way to find the name of a key combination is to press the key while
lf is running and read the name of the key from the unknown mapping error.

Push Mappings

The usual way to map a key sequence is to assign it to a named or unnamed
command. While this provides a clean way to remap builtin keys as well as other
commands, it can be limiting at times. For this reason 'push' command is
provided by lf. This command is used to simulate key pushes given as its
arguments. You can 'map' a key to a 'push' command with an argument to create
various keybindings.

This is mainly useful for two purposes. First, it can be used to map a command
with a command count:

    map <c-j> push 10j

Second, it can be used to avoid typing the name when a command takes arguments:

    map r push :rename<space>

One thing to be careful is that since 'push' command works with keys instead of
commands it is possible to accidentally create recursive bindings:

    map j push 2j

These types of bindings create a deadlock when executed.

Shell Commands

Regular shell commands are the most basic command type that is useful for many
purposes. For example, we can write a shell command to move selected file(s) to
trash. A first attempt to write such a command may look like this:

    cmd trash ${{
        mkdir -p ~/.trash
        if [ -z "$fs" ]; then
            mv "$f" ~/.trash
        else
            IFS="`printf '\n\t'`"; mv $fs ~/.trash
        fi
    }}

We check '$fs' to see if there are any selected files. Otherwise we just delete
the current file. Since this is such a common pattern, a separate '$fx'
variable is provided. We can use this variable to get rid of the conditional:

    cmd trash ${{
        mkdir -p ~/.trash
        IFS="`printf '\n\t'`"; mv $fx ~/.trash
    }}

The trash directory is checked each time the command is executed. We can move
it outside of the command so it would only run once at startup:

    ${{ mkdir -p ~/.trash }}

    cmd trash ${{ IFS="`printf '\n\t'`"; mv $fx ~/.trash }}

Since these are one liners, we can drop '{{' and '}}':

    $mkdir -p ~/.trash

    cmd trash $IFS="`printf '\n\t'`"; mv $fx ~/.trash

Finally note that we set 'IFS' variable manually in these commands. Instead we
could use the 'ifs' option to set it for all shell commands (i.e. 'set ifs
"\n"'). This can be especially useful for interactive use (e.g. '$rm $f' or
'$rm $fs' would simply work). This option is not set by default as it can
behave unexpectedly for new users. However, use of this option is highly
recommended and it is assumed in the rest of the documentation.

Piping Shell Commands

Regular shell commands have some limitations in some cases. When an output or
error message is given and the command exits afterwards, the ui is immediately
resumed and there is no way to see the message without dropping to shell again.
Also, even when there is no output or error, the ui still needs to be paused
while the command is running. This can cause flickering on the screen for short
commands and similar distractions for longer commands.

Instead of pausing the ui, piping shell commands connects stdin, stdout, and
stderr of the command to the statline in the bottom of the ui. This can be
useful for programs following the unix philosophy to give no output in the
success case, and brief error messages or prompts in other cases.

For example, following rename command prompts for overwrite in the statline if
there is an existing file with the given name:

    cmd rename %mv -i $f $1

You can also output error messages in the command and it will show up in the
statline. For example, an alternative rename command may look like this:

    cmd rename %[ -e $1 ] && printf "file exists" || mv $f $1

One thing to be careful is that although input is still line buffered, output
and error are byte buffered and verbose commands will be very slow to display.

Waiting Shell Commands

Waiting shell commands are similar to regular shell commands except that they
wait for a key press when the command is finished. These can be useful to see
the output of a program before the ui is resumed. Waiting shell commands are
more appropriate than piping shell commands when the command is verbose and the
output is best displayed as multiline.

Asynchronous Shell Commands

Asynchronous shell commands are used to start a command in the background and
then resume operation without waiting for the command to finish. Stdin, stdout,
and stderr of the command is neither connected to the terminal nor to the ui.

Remote Commands

One of the more advanced features in lf is remote commands. All clients connect
to a server on startup. It is possible to send commands to all or any of the
connected clients over the common server. This is used internally to notify
file selection changes to other clients.

To use this feature, you need to use a client which supports communicating with
a UNIX-domain socket. OpenBSD implementation of netcat (nc) is one such
example. You can use it to send a command to the socket file:

    echo 'send echo hello world' | nc -U /tmp/lf.${USER}.sock

Since such a client may not be available everywhere, lf comes bundled with a
command line flag to be used as such. When using lf, you do not need to specify
the address of the socket file. This is the recommended way of using remote
commands since it is shorter and immune to socket file address changes:

    lf -remote 'send echo hello world'

In this command 'send' is used to send the rest of the string as a command to
all connected clients. You can optionally give it an id number to send a
command to a single client:

    lf -remote 'send 1000 echo hello world'

All clients have a unique id number but you may not be aware of the id number
when you are writing a command. For this purpose, an '$id' variable is exported
to the environment for shell commands. You can use it to send a remote command
from a client to the server which in return sends a command back to itself. So
now you can display a message in the current client by calling the following in
a shell command:

    lf -remote "send $id echo hello world"

Since lf does not have control flow syntax, remote commands are used for such
needs. For example, you can configure the number of columns in the ui with
respect to the terminal width as follows:

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

Besides 'send' command, there are also two commands to get or set the current
file selection. Two possible modes 'copy' and 'move' specify whether selected
files are to be copied or moved. File names are separated by newline character.
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

Lastly, there is a 'conn' command to connect the server as a client. This
should not be needed for users.

File Operations

lf uses its own builtin copy and move operations by default. These are
implemented as asynchronous operations and progress is shown in the bottom
ruler. These commands do not overwrite existing files or directories with the
same name. Instead, a suffix that is compatible with '--backup=numbered' option
in GNU cp is added to the new files or directories. Only file modes are
preserved and all other attributes are ignored including ownership, timestamps,
context, links, and xattr. Special files such as character and block devices,
named pipes, and sockets are skipped and links are followed. Moving is
performed using the rename operation of the underlying OS. This can fail to
move files between different partitions when it needs to copy files. For these
cases, users are expected to explicitly copy files and then delete the old ones
manually. Operation errors are shown in the message line as well as the log
file and they do not preemptively finish the corresponding file operation.

File operations can be performed on the current selected file or alternatively
on multiple files by selecting them first. When you 'copy' a file, lf doesn't
actually copy the file on the disk, but only records its name to memory. The
actual file copying takes place when you 'paste'. Similarly 'paste' after a
'cut' operation moves the file.

You can customize copy and move operations by defining a 'paste' command. This
is a special command that is called when it is defined instead of the builtin
implementation. You can use the following example as a starting point:

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

Some useful things to be considered are to use the backup ('--backup') and/or
preserve attributes ('-a') options with 'cp' and 'mv' commands if they support
it (i.e. GNU implementation), change the command type to asynchronous, or use
'rsync' command with progress bar option for copying and feed the progress to
the client periodically with remote 'echo' calls.

By default, lf does not assign 'delete' command to a key to protect new users.
You can customize file deletion by defining a 'delete' command. You can also
assign a key to this command if you like. An example command to move selected
files to a trash folder and remove files completely after a prompt are provided
in the example configuration file.

Searching Files

There are two mechanisms implemented in lf to search a file in the current
directory. Searching is the traditional method to move the selection to a file
matching a given pattern. Finding is an alternative way to search for a pattern
possibly using fewer keystrokes.

Searching mechanism is implemented with commands 'search' (default '/'),
'search-back' (default '?'), 'search-next' (default 'n'), and 'search-prev'
(default 'N'). You can enable 'globsearch' option to match with a glob pattern.
Globbing supports '*' to match any sequence, '?' to match any character, and
'[...]' or '[^...] to match character sets or ranges. You can enable
'incsearch' option to jump to the current match at each keystroke while typing.
In this mode, you can either use 'cmd-enter' to accept the search or use
'cmd-escape' to cancel the search. Alternatively, you can also map some other
commands with 'cmap' to accept the search and execute the command immediately
afterwards. Possible candidates are 'up', 'down' and their variants, 'updir',
and 'open' commands. For example, you can use arrow keys to finish the search
with the following mappings:

    cmap <up> up
    cmap <down> down
    cmap <left> updir
    cmap <right> open

Finding mechanism is implemented with commands 'find' (default 'f'),
'find-back' (default 'F'), 'find-next' (default ';'), 'find-prev' (default
','). You can disable 'anchorfind' option to match a pattern at an arbitrary
position in the filename instead of the beginning. You can set the number of
keys to match using 'findlen' option. If you set this value to zero, then the
the keys are read until there is only a single match. Default values of these
two options are set to jump to the first file with the given initial.

Some options effect both searching and finding. You can disable 'wrapscan'
option to prevent searches to wrap around at the end of the file list. You can
disable 'ignorecase' option to match cases in the pattern and the filename.
This option is already automatically overridden if the pattern contains upper
case characters. You can disable 'smartcase' option to disable this behavior.
Two similar options 'ignoredia' and 'smartdia' are provided to control matching
diacritics in latin letters.

Opening Files

You can define a an 'open' command (default 'l' and '<right>') to configure
file opening. This command is only called when the current file is not a
directory, otherwise the directory is entered instead. You can define it just
as you would define any other command:

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

You may want to use 'setsid' before your opener command to have persistent
processes that continue to run after lf quits.

Following command is provided by default:

    cmd open &$OPENER $f

You may also use any other existing file openers as you like. Possible options
are 'libfile-mimeinfo-perl' (executable name is 'mimeopen'), 'rifle' (ranger's
default file opener), or 'mimeo' to name a few.

Previewing Files

lf previews files on the preview pane by printing the file until the end or the
preview pane is filled. This output can be enhanced by providing a custom
preview script for filtering. This can be used to highlight source codes, list
contents of archive files or view pdf or image files as text to name few. For
coloring lf recognizes ansi escape codes.

In order to use this feature you need to set the value of 'previewer' option to
the path of an executable file. lf passes the current file name as the first
argument and the height of the preview pane as the second argument when running
this file. Output of the execution is printed in the preview pane. You may want
to use the same script in your pager mapping as well if any:

    set previewer ~/.config/lf/pv.sh
    map i $~/.config/lf/pv.sh $f | less -R

Since this script is called for each file selection change it needs to be as
efficient as possible and this responsibility is left to the user. You may use
file extensions to determine the type of file more efficiently compared to
obtaining mime types from 'file' command. Extensions can then be used to match
cleanly within a conditional:

    #!/bin/sh

    case "$1" in
        *.tar*) tar tf "$1";;
        *.zip) unzip -l "$1";;
        *.rar) unrar l "$1";;
        *.7z) 7z l "$1";;
        *.pdf) pdftotext "$1" -;;
        *) highlight -O ansi "$1" || cat "$1";;
    esac

Another important consideration for efficiency is the use of programs with
short startup times for preview. For this reason, 'highlight' is recommended
over 'pygmentize' for syntax highlighting. Besides, it is also important that
the application is processing the file on the fly rather than first reading it
to the memory and then do the processing afterwards. This is especially
relevant for big files. lf automatically closes the previewer script output
pipe with a SIGPIPE when enough lines are read. When everything else fails, you
can make use of the height argument to only feed the first portion of the file
to a program for preview.

Changing Directory

lf changes the working directory of the process to the current directory so
that shell commands always work in the displayed directory. After quitting, it
returns to the original directory where it is first launched like all shell
programs. If you want to stay in the current directory after quitting, you can
use one of the example wrapper shell scripts provided in the repository.

There is a special command 'on-cd' that runs a shell command when it is defined
and the directory is changed. You can define it just as you would define any
other command:

    cmd on-cd &{{
        # display git repository status in your prompt
        source /usr/share/git/completion/git-prompt.sh
        GIT_PS1_SHOWDIRTYSTATE=auto
        GIT_PS1_SHOWSTASHSTATE=auto
        GIT_PS1_SHOWUNTRACKEDFILES=auto
        GIT_PS1_SHOWUPSTREAM=auto
        git=$(__git_ps1 " (%s)") || true
        fmt="\033[32;1m%u@%h\033[0m:\033[34;1m%w/\033[0m\033[1m%f$git\033[0m"
        lf -remote "send $id set promptfmt \"$fmt\""
    }}

This command runs whenever you change directory but not on startup. You can add
an extra call to make it run on startup as well:

    cmd on-cd &{{ # ... }}
    on-cd

Note that all shell commands are possible but `%` and `&` are usually more
appropriate as `$` and `!` causes flickers and pauses respectively.

Colorschemes

lf tries to automatically adapt its colors to the environment. On startup,
first '$LS_COLORS' environment variable is checked. This variable is used by
GNU ls to configure its colors based on file types and extensions. The value of
this variable is often set by GNU dircolors in a shell configuration file.
dircolors program itself can be configured with a configuration file. dircolors
supports 256 colors along with common attributes such as bold and underline.

If '$LS_COLORS' variable is not set, '$LSCOLORS' variable is checked instead.
This variable is used by ls programs on unix systems such as Mac and BSDs. This
variable has a simple syntax and supports 8 colors and bold attribute.

If both of these environment variables are not set, then lf fallbacks to its
default colorscheme. Default lf colors are taken from GNU dircolors defaults.
These defaults use 8 basic colors and bold attribute.

You should also note that lf uses 8 color mode by default which uses sgr 3-bit
color escapes (e.g. '\033[34m'). If you want to use 256 colors, you need to
enable 'color256' option which then makes lf use sgr 8-bit color escapes (e.g.
'\033[38;5;4m'). This option is intended to eliminate differences between
default colors used by ls and lf since terminals may render 3-bit and 8-bit
escapes differently even for the same color.

Keeping this mechanism in mind, you can configure lf colors in two different
ways. First, you can configure 8 basic colors used by your terminal and lf
should pick up those colors automatically. Depending on your terminal, you
should be able to select your colors from a 24-bit palette. This is the
recommended approach as colors used by other programs will also match each
other.

Second, you can set the values of environmental variables mentioned above for
fine grained customization. This is useful to change colors used for different
file types and extensions. '$LS_COLORS' is more powerful than '$LSCOLORS' and
it can be used even when GNU programs are not installed on the system. You can
combine this second method with the first method for best results.

Lastly, you may also want to configure the colors of the prompt line to match
the rest of the colors. Colors of the prompt line can be configured using the
'promptfmt' option which can include hardcoded colors as ansi escapes. See the
default value of this option to have an idea about how to color this line.
*/
package main
