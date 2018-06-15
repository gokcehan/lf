// DO NOT EDIT! (AUTO-GENERATED)

package main

var genDocString = `
lf is a terminal file manager.

Source code can be found in the repository at
https://github.com/gokcehan/lf.

This documentation can either be read from terminal using 'lf -doc' or
online at https://godoc.org/github.com/gokcehan/lf.


Reference

The following commands are provided by lf with default keybindings:

    up           (default 'k' and '<up>')
    half-up      (default '<c-u>')
    page-up      (default '<c-b>')
    down         (default 'j' and '<down>')
    half-down    (default '<c-d>')
    page-down    (default '<c-f>')
    updir        (default 'h' and '<left>')
    open         (default 'l' and '<right>')
    quit         (default 'q')
    top          (default 'gg')
    bottom       (default 'G')
    toggle       (default '<space>')
    invert       (default 'v')
    unmark       (default 'u')
    yank         (default 'y')
    delete       (default 'd')
    put          (default 'p')
    clear        (default 'c')
    redraw       (default '<c-l>')
    reload       (default '<c-r>')
    read         (default ':')
    shell        (default '$')
    shell-pipe   (default '%')
    shell-wait   (default '!')
    shell-async  (default '&')
    search       (default '/')
    search-back  (default '?')
    search-next  (default 'n')
    search-prev  (default 'N')

The following commands are provided by lf without default keybindings:

    draw    draw the ui
    load    load modified files and directories
    sync    synchronizes yanked/deleted files with server
    echo    prints its arguments to the message line
    cd      changes working directory to its argument
    select  changes current file selection to its argument
    push    simulate key pushes given in its argument

The following command line commands are provided by lf with default
keybindings:

    cmd-escape            (default '<esc>')
    cmd-complete          (default '<tab>')
    cmd-enter             (default '<c-j>' and '<enter>')
    cmd-history-next      (default '<c-n>')
    cmd-history-prev      (default '<c-p>')
    cmd-delete            (default '<c-d>' and '<delete>')
    cmd-delete-back       (default '<bs>' and '<bs2>')
    cmd-left              (default '<c-b>' and '<left>')
    cmd-right             (default '<c-f>' and '<right>')
    cmd-home              (default '<c-a>' and '<home>')
    cmd-end               (default '<c-e>' and '<end>')
    cmd-delete-home       (default '<c-u>')
    cmd-delete-end        (default '<c-k>')
    cmd-delete-unix-word  (default '<c-w>')
    cmd-yank              (default '<c-y>')
    cmd-transpose         (default '<c-t>')
    cmd-interrupt         (default '<c-c>')
    cmd-word              (default '<a-f>')
    cmd-word-back         (default '<a-b>')
    cmd-capitalize-word   (default '<a-c>')
    cmd-delete-word       (default '<a-d>')
    cmd-uppercase-word    (default '<a-u>')
    cmd-lowercase-word    (default '<a-l>')
    cmd-transpose-word    (default '<a-t>')

The following options can be used to customize the behavior of lf:

    dircounts   boolean  (default off)
    dirfirst    boolean  (default on)
    drawbox     boolean  (default off)
    globsearch  boolean  (default off)
    hidden      boolean  (default off)
    ignorecase  boolean  (default on)
    preview     boolean  (default on)
    reverse     boolean  (default off)
    smartcase   boolean  (default on)
    wrapscan    boolean  (default on)
    period      integer  (default 0)
    scrolloff   integer  (default 0)
    tabstop     integer  (default 8)
    filesep     string   (default "\n")
    ifs         string   (default '') (not exported if empty)
    previewer   string   (default '') (not filtered if empty)
    promptfmt   string   (default "\033[32;1m%u@%h\033[0m:\033[34;1m%w/\033[0m\033[1m%f\033[0m")
    shell       string   (default 'sh')
    sortby      string   (default 'natural')
    timefmt     string   (default 'Mon Jan _2 15:04:05 2006')
    ratios      string   (default '1:2:3')
    info        string   (default '')

The following variables are exported for shell commands:

    $f   current file
    $fs  marked file(s) separated with 'filesep'
    $fx  current file or marked file(s) if any
    $id  id number of the client

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
    map gh cd ~

The following keybindings to applications are provided by default on unix:

    map e $$EDITOR $f ('vi' if empty)
    map i $$PAGER $f  ('less' if empty)
    map w $$SHELL     ('sh' if empty)

The following keybindings to applications are provided by default on
windows:

    map e $notepad %f%
    map i $more %f%
    map w $cmd


Configuration

The configuration file should be located at:

    $XDG_CONFIG_HOME/lf/lfrc

If '$XDG_CONFIG_HOME' is not set, it defaults to '$HOME/.config' so the
location should be:

    ~/.config/lf/lfrc

A sample configuration file can be found at
https://github.com/gokcehan/lf/blob/master/etc/lfrc.example.


Prefixes

The following command prefixes are used by lf:

    :  read (default)  builtin/custom command
    $  shell           shell command
    %  shell-pipe      shell command running with the ui
    !  shell-wait      shell command waiting for key press
    &  shell-async     shell command running asynchronously
    /  search          search file in current directory
    ?  search-back     search file in the reverse order

The same evaluator is used for the command line and the configuration file
for read and shell commands. The difference is that prefixes are not
necessary in the command line. Instead, different modes are provided to read
corresponding commands. These modes are mapped to the prefix keys above by
default. Searching commands are only used from the command line.


Syntax

Characters from '#' to newline are comments and ignored:

    # comments start with '#'

There are three special commands ('set', 'map', and 'cmd') and their
variants for configuration.

'set' is used to set an option which can be boolean, integer, or string:

    set hidden         # boolean on
    set nohidden       # boolean off
    set hidden!        # boolean toggle
    set scrolloff 10   # integer value
    set sortby time    # string value w/o quotes
    set sortby 'time'  # string value with single quotes (whitespaces)
    set sortby "time"  # string value with double quotes (backslash escapes)

'map' is used to bind a key to a command which can be builtin command,
custom command, or shell command:

    map gh cd ~        # builtin command
    map D trash        # custom command
    map i $less $f     # shell command
    map U !du -sh      # waiting shell command

'cmap' is used to bind a key to a command line command which can only be one
of the builtin commands:

    cmap <c-g> cmd-escape

You can delete an existing binding by leaving the expression empty:

    map gh             # deletes 'gh' mapping
    cmap <c-g>         # deletes '<c-g>' mapping

'cmd' is used to define a custom command

    cmd usage $du -h -d1 | less

You can delete an existing command by leaving the expression empty:

    cmd trash          # deletes 'trash' command

If there is no prefix then ':' is assumed:

    map zt set info time

An explicit ':' can be provided to group statements until a newline which is
especially useful for 'map' and 'cmd' commands:

    map st :set sortby time; set info time

If you need multiline you can wrap statements in '{{' and '}}' after the
proper prefix.

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

Keys combined with the alt key are assigned in two different ways depending
on the behavior of your terminal. Older terminals (e.g. xterm) may set the
8th bit of a character when the alt key is pressed. On these terminals, you
can use the corresponding byte for the mapping:

    map á down

Newer terminals (e.g. gnome-terminal) may prefix the key with an escape key
when the alt key is pressed. lf uses the escape delaying mechanism to
recognize alt keys in these terminals (delay is 100ms). On these terminals,
keys combined with the alt key are prefixed with 'a' character:

    map <a-a> down

Please note that, some key combinations are not possible due to the way
terminals work (e.g. control and h combination sends a backspace key
instead). The easiest way to find the name of a key combination is to press
the key while lf is running and read the name of the key from the unknown
mapping error.


Push Mappings

The usual way to map a key sequence is to assign it to a named or unnamed
command. While this provides a clean way to remap builtin keys as well as
other commands, it can be limiting at times. For this reason 'push' command
is provided by lf. This command is used to simulate key pushes given as its
arguments. You can 'map' a key to a 'push' command with an argument to
create various keybindings.

This is mainly useful for two purposes. First, it can be used to map a
command with a command count:

    map <c-j> push 10j

Second, it can be used to avoid typing the name when a command takes
arguments:

    map r push :rename<space>

One thing to be careful is that since 'push' command works with keys instead
of commands it is possible to accidentally create recursive bindings:

    map j push 2j

These types of bindings create a deadlock when executed.


Shell Commands

Regular shell commands are the most basic command type that is useful for
many purposes. For example, we can write a shell command to move selected
file(s) to trash. A first attempt to write such a command may look like
this:

    cmd trash ${{
        mkdir -p ~/.trash
        if [ -z "$fs" ]; then
            mv "$f" ~/.trash
        else
            IFS="'printf '\n\t''"; mv $fs ~/.trash
        fi
    }}

We check '$fs' to see if there are any marked files. Otherwise we just
delete the current file. Since this is such a common pattern, a separate
'$fx' variable is provided. We can use this variable to get rid of the
conditional:

    cmd trash ${{
        mkdir -p ~/.trash
        IFS="'printf '\n\t''"; mv $fx ~/.trash
    }}

The trash directory is checked each time the command is executed. We can
move it outside of the command so it would only run once at startup:

    ${{ mkdir -p ~/.trash }}

    cmd trash ${{ IFS="'printf '\n\t''"; mv $fx ~/.trash }}

Since these are one liners, we can drop '{{' and '}}':

    $mkdir -p ~/.trash

    cmd trash $IFS="'printf '\n\t''"; mv $fx ~/.trash

Finally note that we set 'IFS' variable manually in these commands. Instead
we could use the 'ifs' option to set it for all shell commands (i.e. 'set
ifs "\n"'). This can be especially useful for interactive use (e.g. '$rm $f'
or '$rm $fs' would simply work). This option is not set by default as it can
behave unexpectedly for new users. However, use of this option is highly
recommended and it is assumed in the rest of the documentation.


Piping Shell Commands

Regular shell commands have some limitations in some cases. When an output
or error message is given and the command exits afterwards, the ui is
immediately resumed and there is no way to see the message without dropping
to shell again. Also, even when there is no output or error, the ui still
needs to be paused while the command is running. This can cause flickering
on the screen for short commands and similar distractions for longer
commands.

Instead of pausing the ui, piping shell commands connects stdin, stdout, and
stderr of the command to the statline in the bottom of the ui. This can be
useful for programs following the unix philosophy to give no output in the
success case, and brief error messages or prompts in other cases.

For example, following rename command prompts for overwrite in the statline
if there is an existing file with the given name:

    cmd rename %mv -i $f $1

You can also output error messages in the command and it will show up in the
statline. For example, an alternative rename command may look like this:

    cmd rename %[ -e $1 ] && printf "file exists" || mv $f $1

One thing to be careful is that although input is still line buffered,
output and error are byte buffered and verbose commands will be very slow to
display.


Waiting Shell Commands

Waiting shell commands are similar to regular shell commands except that
they wait for a key press when the command is finished. These can be useful
to see the output of a program before the ui is resumed. Waiting shell
commands are more appropriate than piping shell commands when the command is
verbose and the output is best displayed as multiline.


Asynchronous Shell Commands

Asynchronous shell commands are used to start a command in the background
and then resume operation without waiting for the command to finish. Stdin,
stdout, and stderr of the command is neither connected to the terminal nor
to the ui.


Remote Commands

One of the more advanced features in lf is remote commands. All clients
connect to a server on startup. It is possible to send commands to all or
any of the connected clients over the common server. This is used internally
to notify file selection changes to other clients.

To use this feature, you need to use a client which supports communicating
with a UNIX-domain socket. OpenBSD implementation of netcat (nc) is one such
example. You can use it to send a command to the socket file:

    echo 'send echo hello world' | nc -U /tmp/lf.${USER}.sock

Since such a client may not be available everywhere, lf comes bundled with a
command line flag to be used as such. When using lf, you do not need to
specify the address of the socket file. This is the recommended way of using
remote commands since it is shorter and immune to socket file address
changes:

    lf -remote 'send echo hello world'

In this command 'send' is used to send the rest of the string as a command
to all connected clients. You can optionally give it an id number to send a
command to a single client:

    lf -remote 'send 1000 echo hello world'

All clients have a unique id number but you may not be aware of the id
number when you are writing a command. For this purpose, an '$id' variable
is exported to the environment for shell commands. You can use it to send a
remote command from a client to the server which in return sends a command
back to itself. So now you can display a message in the current client by
calling the following in a shell command:

    lf -remote "send $id echo hello world"

Since lf does not have control flow syntax, remote commands are used for
such needs. For example, you can configure the number of columns in the ui
with respect to the terminal width as follows:

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

Besides 'send' command, there are also two commands to get or set the
current file selection. Two possible modes 'copy' and 'move' specify whether
selected files are to be copied or moved. File names are separated by
newline character. Setting the file selection is done with 'save' command:

    lf -remote "$(printf 'save\ncopy\nfoo.txt\nbar.txt\nbaz.txt\n')"

Getting the file selection is similarly done with 'load' command:

    resp=$(lf -remote 'load')
    mode=$(echo "$resp" | sed -n '1p')
    list=$(echo "$resp" | sed '1d')
    if [ $mode = 'copy' ]; then
        # do something with $list
    elif [ $mode = 'move' ]; then
        # do something else with $list
    fi

Lastly, there is a 'conn' command to connect the server as a client. This
should not be needed for users.


File Operations

lf uses the underlying 'cp' and 'mv' shell commands for file operations. For
this purpose, when you 'yank' (i.e. copy) a file, it doesn't actually copy
the file on the disk, but only records its name to memory. The actual file
operation takes place when you do the 'put' in which case the 'cp' command
is used. Similarly the 'mv' command is used for 'delete' (i.e. cut or kill)
followed by 'put'. These traditional names (e.g. 'yank', 'delete', and
'put') are picked instead of the other common convention (e.g. copy and cut)
to resemble the default keybinds for these operations.

You can customize these operations by defining a 'put' command. This is a
special command that is called when it is defined instead of the builtin
implementation. The default behavior is similar to the following command:

    cmd put ${{
        load=$(lf -remote 'load')
        mode=$(echo "$load" | sed -n '1p')
        list=$(echo "$load" | sed '1d')
        if [ $mode = 'copy' ]; then
            cp -R -n $list .
        elif [ $mode = 'move' ]; then
            mv -n $list .
        fi
        lf -remote "$(printf 'save\nmove\n\n')"
        lf -remote 'send load'
        lf -remote 'send sync'
    }}

Some useful things are to use the backup option ('--backup') with 'cp' and
'mv' commands if they support it (i.e. GNU implementation), change the
command type to asynchronous, or use 'rsync' command with progress bar
option for copying and feed the progress to the client periodically with
remote 'echo' calls.

By default, lf does not provide an actual file deletion command to protect
new users. You can define such a command and optionally assign a key if you
like. An example command to move selected files to a trash folder and remove
files completely are provided in the example configuration file.


Opening Files

You can use 'open-file' command to open a file. This is a special command
called by 'open' when the current file is not a directory. Normally a user
maps the 'open' command to a key (default 'l') and customize 'open-file'
command as desired. You can define it just as you would define any other
command:

    cmd open-file $vi $fx

It is possible to use different command types:

    cmd open-file &xdg-open $f

You may want to use either file extensions or mime types from 'file'
command:

    cmd open-file ${{
        case $(file --mime-type $f -b) in
            text/*) vi $fx;;
            *) for f in $fx; do xdg-open $f > /dev/null 2> /dev/null & done;;
        esac
    }}

Following commands are provided by default:

    cmd open-file &start %f%      # windows
    cmd open-file &open "$f"      # mac
    cmd open-file &xdg-open "$f"  # others

You may also use any other existing file openers as you like. Possible
options are 'libfile-mimeinfo-perl' (executable name is 'mimeopen'), 'rifle'
(ranger's default file opener), or 'mimeo' to name a few.


Previewing Files

lf previews files on the preview pane by printing the file until the end or
the preview pane is filled. This output can be enhanced by providing a
custom preview script for filtering. This can be used to highlight source
codes, list contents of archive files or view pdf or image files as text to
name few. For coloring lf recognizes ansi escape codes.

In order to use this feature you need to set the value of 'previewer' option
to the path of an executable file. lf passes the current file name as the
first argument and the height of the preview pane as the second argument
when running this file. Output of the execution is printed in the preview
pane. You may want to use the same script in your pager mapping as well if
any:

    set previewer ~/.config/lf/pv.sh
    map i $~/.config/lf/pv.sh $f | less -R

Since this script is called for each file selection change it needs to be as
efficient as possible and this responsibility is left to the user. You may
use file extensions to determine the type of file more efficiently compared
to obtaining mime types from 'file' command. Extensions can then be used to
match cleanly within a conditional:

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
over 'pygmentize' for syntax highlighting. Besides, it is also important
that the application is processing the file on the fly rather than first
reading it to the memory and then do the processing afterwards. This is
especially relevant for big files. lf automatically closes the previewer
script output pipe with a SIGPIPE when enough lines are read. When
everything else fails, you can make use of the height argument to only feed
the first portion of the file to a program for preview.


Colorschemes

lf tries to automatically adapt its colors to the environment. On startup,
first '$LS_COLORS' environment variable is checked. This variable is used by
GNU ls to configure its colors based on file types and extensions. The value
of this variable is often set by GNU dircolors in a shell configuration
file. dircolors program itself can be configured with a configuration file.
dircolors supports 256 colors along with common attributes such as bold and
underline.

If '$LS_COLORS' variable is not set, '$LSCOLORS' variable is checked
instead. This variable is used by ls programs on unix systems such as Mac
and BSDs. This variable has a simple syntax and supports 8 colors and bold
attribute.

If both of these environment variables are not set, then lf fallbacks to its
default colorscheme. Default lf colors are taken from GNU dircolors
defaults. These defaults use 8 basic colors and bold attribute.

Keeping this mechanism in mind, you can configure lf colors in two different
ways. First, you can configure 8 basic colors used by your terminal and lf
should pick up those colors automatically. Depending on your terminal, you
should be able to select your colors from a 24-bit palette. This is the
recommended approach as colors used by other programs will also match each
other.

Second, you can set the values of environmental variables mentioned above
for fine grained customization. This is useful to change colors used for
different file types and extensions. '$LS_COLORS' is more powerful than
'$LSCOLORS' and it can be used even when GNU programs are not installed on
the system. You can combine this second method with the first method for
best results.
`
