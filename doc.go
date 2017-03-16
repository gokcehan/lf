//go:generate gen/docstring.sh

/*
lf is a terminal file manager.

Source code can be found in the repository at https://github.com/gokcehan/lf.

This documentation can either be read from terminal using "lf -doc" or online
at https://godoc.org/github.com/gokcehan/lf.

Reference

The following commands are provided by lf with default keybindings:

    up                (default "k" and "<up>")
    half-up           (default "<c-u>")
    page-up           (default "<c-b>")
    down              (default "j" and "<down>")
    half-down         (default "<c-d>")
    page-down         (default "<c-f>")
    updir             (default "h" and "<left>")
    open              (default "l" and "<right>")
    quit              (default "q")
    bot               (default "G")
    top               (default "gg")
    read              (default ":")
    read-shell        (default "$")
    read-shell-wait   (default "!")
    read-shell-async  (default "&")
    search            (default "/")
    search-back       (default "?")
    search-next       (default "n")
    search-prev       (default "N")
    toggle            (default "<space>")
    invert            (default "v")
    yank              (default "y")
    clear             (default "c")
    delete            (default "d")
    put               (default "p")
    renew             (default "<c-l>")

The following commands are provided by lf without default keybindings:

    sync  synchronizes yanked/deleted files with server
    echo  prints its arguments to the message line
    cd    changes working directory to its argument
    push  simulate key pushes given in its argument

The following command line commands are provided by lf with default
keybindings:

    cmd-escape        (default "<esc>")
    cmd-comp          (default "<tab>")
    cmd-enter         (default "<c-j>" and "<enter>")
    cmd-delete        (default "<c-d>" and "<delete>")
    cmd-delete-back   (default "<bs>" and "<bs2>")
    cmd-left          (default "<c-b>" and "<left>")
    cmd-right         (default "<c-f>" and "<right>")
    cmd-beg           (default "<c-a>" and "<home>")
    cmd-end           (default "<c-e>" and "<end>")
    cmd-delete-end    (default "<c-k>")
    cmd-delete-beg    (default "<c-u>")
    cmd-delete-word   (default "<c-w>")
    cmd-put           (default "<c-y>")
    cmd-transpose     (default "<c-t>")

The following options can be used to customize the behavior of lf:

    dirfirst   boolean  (default on)
    hidden     boolean  (default off)
    preview    boolean  (default on)
    reverse    boolean  (default off)
    wrapscan   boolean  (default on)
    scrolloff  integer  (default 0)
    tabstop    integer  (default 8)
    filesep    string   (default ":")
    ifs        string   (default "") (not exported if empty)
    previewer  string   (default "") (not filtered if empty)
    shell      string   (default "/bin/sh")
    sortby     string   (default "natural")
    timefmt    string   (default "Mon Jan _2 15:04:05 2006")
    ratios     string   (default "1:2:3")
    info       string   (default "")

The following variables are exported for shell commands:

    $f   current file
    $fs  marked file(s) separated with ':'
    $fx  current file or marked file(s) if any
    $id  id number of the client

Configuration

The configuration file should be located at:

    $XDG_CONFIG_HOME/lf/lfrc"

If "$XDG_CONFIG_HOME" is not set, it defaults to "$HOME/.config" so the
location should be:

    ~/.config/lf/lfrc

A sample configuration file can be found at
https://github.com/gokcehan/lf/blob/master/etc/lfrc.example.

Prefixes

The following command prefixes are used by lf:

    :  read (default)    built-in command
    $  read-shell        shell command
    !  read-shell-wait   shell command waiting for key press
    &  read-shell-async  asynchronous shell command
    /  search            search file in current directory
    ?  search-back       search file in the reverse order

The same evaluator is used for the command line and the configuration file. The
difference is that prefixes are not necessary in the command line. Instead
different modes are provided to read corresponding commands. Note that by
default these modes are mapped to the prefix keys above.

Syntax

Characters from "#" to "\n" are comments and ignored:

    # comments start with '#'

There are three special commands ("set", "map", and "cmd") and their variants
for configuration.

"set" is used to set an option which can be boolean, integer, or string:

    set hidden         # boolean on
    set nohidden       # boolean off
    set hidden!        # boolean toggle
    set scrolloff 10   # integer value
    set sortby time    # string value w/o quotes
    set sortby 'time'  # string value with single quotes (whitespaces)
    set sortby "time"  # string value with double quotes (backslash escapes)

"map" is used to bind a key to a command which can be built-in command, custom
command, or shell command:

    map gh cd ~        # built-in command
    map D trash        # custom command
    map i $less "$f"   # shell command
    map u !du -h .     # waiting shell command

"cmap" is used to bind a key to a command line command which can only be one of
the built-in commands:

    cmap <c-g> cmd-escape

You can delete an existing binding by leaving the expression empty:

    map gh             # deletes 'gh' mapping
    cmap <c-g>         # deletes '<c-g>' mapping

"cmd" is used to define a custom command

    cmd usage $du -h . | less

You can delete an existing command by leaving the expression empty:

    cmd trash          # deletes trash command

If there is no prefix then ":" is assumed:

    map zt set info time

An explicit ":" can be provided to group statements until a "\n" occurs which
is especially useful for "map" and "cmd" commands:

    map st :set sortby time; set info time

If you need multiline you can wrap statements in "{{" and "}}" after the proper
prefix.

    map st :{{
        set sortby time
        set info time
    }}

Mappings

The usual way to map a key sequence is to assign it to a named or unnamed
command. While this provides a clean way to remap builtin keys as well as other
commands, it can be limiting at times. For this reason "push" command is
provided by lf. This command is used to simulate key pushes given as its
arguments. You can "map" a key to a "push" command with an argument to create
various keybindings.

This is mainly useful for two purposes. First, it can be used to map a command
with a command count:

    map <c-j> push 10j

Second, it can be used to avoid typing the name when a command takes arguments:

    map r push :rename<space>

One thing to be careful is that since "push" command works with keys instead of
commands it is possible to accidentally create recursive bindings:

    map j push 2j

These types of bindings create a deadlock when executed.

Commands

For demonstration let us write a shell command to move selected file(s) to
trash.

A first attempt to write such a command may look like this:

    cmd trash ${{
        mkdir -p ~/.trash
        if [ -z $fs ]; then
            mv --backup=numbered "$f" $HOME/.trash
        else
            IFS=':'; mv --backup=numbered $fs $HOME/.trash
        fi
    }}

We check "$fs" to see if there are any marked files. Otherwise we just delete
the current file. Since this is such a common pattern, a separate "$fx"
variable is provided. We can use this variable to get rid of the conditional:

    cmd trash ${{
        mkdir -p ~/.trash
        IFS=':'; mv --backup=numbered $fx $HOME/.trash
    }}

The trash directory is checked each time the command is executed. We can move
it outside of the command so it would only run once at startup:

    ${{ mkdir -p ~/.trash }}

    cmd trash ${{ IFS=':'; mv --backup=numbered $fx $HOME/.trash }}

Since these are one liners, we can drop "{{" and "}}":

    $mkdir -p ~/.trash

    cmd trash $IFS=':'; mv --backup=numbered $fx $HOME/.trash

Finally note that we set "IFS" variable accordingly in the command. Instead we
could use the "ifs" option to set it for all commands (e.g. "set ifs ':'").
This can be especially useful for interactive use (e.g. "rm $fs" would simply
work). This option is not set by default as things may behave unexpectedly at
other places.

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

In this command "send" is used to send the rest of the string as a command to
all connected clients. You can optionally give it an id number to send a
command to a single client:

    lf -remote 'send 1000 echo hello world'

All clients have a unique id number but you may not be aware of the id number
when you are writing a command. For this purpose, an "$id" variable is exported
to the environment for shell commands. You can use it to send a remote command
from a client to the server which in return sends a command back to itself. So
now you can display a message in the current client by calling the following in
a shell command:

    lf -remote "send $id echo hello world"

A common use for this feature is to display an error message back in the
client. You can implement a safe rename command which does not overwrite an
existing file or directory as such:

    cmd rename ${{
        if [ -e "$1" ]; then
            lf -remote "send $id echo file exists"
        else
            mv "$f" "$1"
        fi
    }}

Since lf does not have control flow syntax, remote commands are used for such
needs. Following example can be used to dynamically set the number of columns
on startup based on terminal width:

    ${{
        w=$(tput cols)
        if [ $w -le 80 ]; then
            lf -remote "send $id set ratios 1:2"
        elif [ $w -le 160 ]; then
            lf -remote "send $id set ratios 1:2:3"
        else
            lf -remote "send $id set ratios 1:2:3:5"
        fi
    }}

Besides "send" command, there are also two commands to get or set the current
file selection. Two possible modes "copy" and "move" specify whether selected
files are to be copied or moved. File names are separated ":" character.
Setting the file selection is done with "save" command:

    lf -remote 'save copy foo.txt:bar.txt:baz.txt'

Getting the file selection is similarly done with "load" command. You may need
to parse the response as such to achieve what you need:

    resp=$(echo 'load' | nc -U /tmp/lf.${USER}.sock)
    mode=$(echo $resp | cut -d' ' -f1)
    list=$(echo $resp | cut -d' ' -f2-)
    if [ $mode = 'copy' ]; then
        # do something with the $list
    elif [ $mode = 'move' ]; then
        # do something else with the $list
    fi

Lastly, there is a "conn" command to connect the server as a client. This
should not be needed for users.

File Operations

lf uses the underlying "cp" and "mv" shell commands for file operations. For
this purpose, when you "yank" (i.e. copy) a file, it doesn't actually copy the
file on the disk, but only records its name to memory. The actual file
operation takes place when you do the "put" in which case the "cp" command is
used. Similarly the "mv" command is used for "delete" (i.e. cut or kill)
followed by "put". These traditional names (e.g. "yank", "delete", and "put")
are picked instead of the other common convention (e.g. copy and cut) to
resemble the default keybinds for these operations.

By default, lf does not provide an actual file deletion command to protect new
users. You can define such a command and optionally assign a key if you like.
An example command to move selected files to a trash folder and remove files
completely are provided in the example configuration file.

Opening Files

You can use "open-file" command to open a file. This is a special command
called by "open" when the current file is not a directory. Normally a user maps
the "open" command to a key (default "l") and customize "open-file" command as
desired. You can define it just as you would define any other command:

    cmd open-file $IFS=':'; vim $fx

It is possible to use different command types:

    cmd open-file &xdg-open "$f"

You may want to use either file extensions or mime types from "file" command:

    cmd open-file ${{
        case $(file --mime-type "$f" -b) in
            text/*)
                IFS=':'; vim $fx;;
            *)
                IFS=':'; for f in $fx; do
                    xdg-open "$f" > /dev/null 2> /dev/null &
                done;;
        esac
    }}

lf does not come bundled with a file opener. You can use any of the existing
file openers as you like. Possible options are "open" (for Mac OS X only),
"xdg-utils" (executable name is "xdg-open"), "libfile-mimeinfo-perl"
(executable name is "mimeopen"), "rifle" (ranger's default file opener), or
"mimeo" to name a few.

Previewing Files

lf previews files on the preview pane by printing the file until the end or the
preview pane is filled. This output can be enhanced by providing a custom
preview script for filtering. This can be used to highlight source codes, list
contents of archive files or view pdf or image files as text to name few. For
coloring lf recognizes ansi escape codes.

In order to use this feature you need to set the value of "previewer" option to
the path of an executable file. lf passes the current file name as the first
argument and the height of the preview pane as the second argument when running
this file. Output of the execution is printed in the preview pane. You may want
to use the same script in your pager mapping as well if any:

    set previewer ~/.config/lf/pv.sh
    map i $~/.config/lf/pv.sh "$f" | less -R

Since this script is called for each file selection change it needs to be as
efficient as possible and this responsibility is left to the user. You may use
file extensions to determine the type of file more efficiently compared to
obtaining mime types from "file" command. Extensions can then be used to match
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
short startup times for preview. For this reason, "highlight" is recommended
over "pygmentize" for syntax highlighting. Besides, it is also important that
the application is processing the file on the fly rather than first reading it
to the memory and then do the processing afterwards. This is especially
relevant for big files. lf automatically closes the previewer script output
pipe with a SIGPIPE when enough lines are read. When everything else fails, you
can make use of the height argument to only feed the first portion of the file
to a program for preview.
*/
package main
