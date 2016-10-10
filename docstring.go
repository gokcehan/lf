// DO NOT EDIT! (AUTO-GENERATED)

package main

var genDocString = `
lf is a terminal file manager.

Source code can be found in the repository at
https://github.com/gokcehan/lf.

This documentation can either be read from terminal using "lf -doc" or
online at https://godoc.org/github.com/gokcehan/lf.


Reference

The following commands are provided by lf with default keybindings.

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
    toggle            (default "<space>")
    invert            (default "v")
    yank              (default "y")
    delete            (default "d")
    paste             (default "p")
    renew             (default "<c-l>")

The following commands are provided by lf without default keybindings.

    echo  prints its arguments to the message line
    cd    changes working directory to its argument
    push  simulate key pushes given in its argument

The following options can be used to customize the behavior of lf.

    hidden     bool    (default off)
    preview    bool    (default on)
    scrolloff  int     (default 0)
    tabstop    int     (default 8)
    ifs        string  (default "") (not exported if empty)
    previewer  string  (default "") (not filtered if empty)
    shell      string  (default "$SHELL")
    showinfo   string  (default "none")
    sortby     string  (default "name")
    ratios     string  (default "1:2:3")

The following variables are exported for shell commands.

    $f   current file
    $fs  marked file(s) separated with ':'
    $fx  current file or marked file(s) if any


Configuration

The configuration file should either be located in
"$XDG_CONFIG_HOME/lf/lfrc" or "~/.config/lf/lfrc". A sample configuration
file can be found at
https://github.com/gokcehan/lf/blob/master/etc/lfrc.example.


Prefixes

The following command prefixes are used by lf:

    :  read (default)
    $  read-shell
    !  read-shell-wait
    &  read-shell-async
    /  search
    ?  search-back

The same evaluator is used for the command line and the configuration file.
The difference is that prefixes are not necessary in the command line.
Instead different modes are provided to read corresponding commands. Note
that by default these modes are mapped to the prefix keys above.


Syntax

Characters from "#" to "\n" are comments and ignored.

There are three special commands for configuration.

"set" is used to set an option which could be bool (e.g. "set hidden", "set
nohidden", "set hidden!"), int (e.g. "set scrolloff 10"), or string (e.g.
"set sortby time").

"map" is used to bind a key to a command which could be built-in command
(e.g. "map gh cd ~"), custom command (e.g. "map D trash"), or shell command
(e.g. "map i $less "$f"", "map u !du -h . | less"). You can delete an
existing binding by leaving the expression empty (e.g. "map gh").

"cmd" is used to define a custom command or delete an existing command by
leaving the expression empty (e.g. "cmd trash").

If there is no prefix then ":" is assumed. An explicit ":" could be provided
to group statements until a "\n" occurs. This is especially useful for "map"
and "cmd" commands. If you need multiline you can wrap statements in "{{"
and "}}" after the proper prefix.


Mappings

The usual way to map a key sequence is to assign it to a named or unnamed
command. While this provides a clean way to remap builtin keys as well as
other commands, it can be limiting at times. For this reason "push" command
is provided by lf. This command is used to simulate key pushes given as its
arguments. You can "map" a key to a "push" command with an argument to
create various keybindings.

This is mainly useful for two purposes. First, it can be used to map a
command with a command count.

    map <c-j> push 10j

Second, it can be used to avoid typing the name when a command takes
arguments.

    map r :push :rename<space>

One thing to be careful is that since "push" command works with keys instead
of commands it is possible to accidentally create recursive bindings.

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

We check "$fs" to see if there are any marked files. Otherwise we just
delete the current file. Since this is such a common pattern, a separate
"$fx" variable is provided. We can use this variable to get rid of the
conditional.

    cmd trash ${{
        mkdir -p ~/.trash
        IFS=':'; mv --backup=numbered $fx $HOME/.trash
    }}

The trash directory is checked each time the command is executed. We can
move it outside of the command so it would only run once at startup.

    ${{ mkdir -p ~/.trash }}

    cmd trash ${{ IFS=':'; mv --backup=numbered $fx $HOME/.trash }}

Since these are one liners, we can drop "{{" and "}}".

    $mkdir -p ~/.trash

    cmd trash $IFS=':'; mv --backup=numbered $fx $HOME/.trash

Finally note that we set "IFS" variable accordingly in the command. Instead
we could use the "ifs" option to set it for all commands (e.g. "set ifs
':'"). This could be especially useful for interactive use (e.g. "rm $fs"
would simply work). This option is not set by default as things may behave
unexpectedly at other places.


File Operations

lf uses the underlying "cp" and "mv" shell commands for file operations. For
this purpose, when you "yank" (i.e. copy) a file, it doesn't actually copy
the file on the disk, but only records its name to memory. The actual file
operation takes place when you do the "paste" in which case the "cp" command
is used. Similarly the "mv" command is used for "delete" (i.e. cut or kill)
followed by "paste". These traditional names (e.g. "yank" and "delete") are
picked instead of the other common convention (e.g. copy and cut) to
resemble the default keybinds for these operations.


Opening Files

You can use "open-file" command to open a file. This is a special command
called by "open" when the current file is not a directory. Normally a user
maps the "open" command to a key (default "l") and customize "open-file"
command as desired. You can define it just as you would define any other
command.

    cmd open-file $IFS=':'; vim $fx

It is possible to use different command types.

    cmd open-file &xdg-open "$f"

You may want to use either file extensions or mime types from "file"
command.

    cmd open-file ${{
        case $(file --mime-type "$f" -b) in
            text/*) IFS=':'; vim $fx;;
            *) IFS=':'; for f in $fx; do xdg-open "$f" &> /dev/null & done;;
        esac
    }}

lf does not come bundled with a file opener. You can use any of the existing
file openers as you like. Possible options are "open" (for Mac OS X only),
"xdg-utils" (executable name is "xdg-open"), "libfile-mimeinfo-perl"
(executable name is "mimeopen"), "rifle" (ranger's default file opener), or
"mimeo" to name a few.


Previewing Files

lf previews files on the preview pane by printing the file until the end or
the preview pane is filled. This output can be enhanced by providing a
custom preview script for filtering. This can be used to highlight source
codes, list contents of archive files or view pdf or image files as text to
name few. For coloring lf recognizes ansi escape codes.

In order to use this feature you need to set the value of "previewer" option
to the path of an executable file. lf passes the current file name as the
first argument and the height of the preview pane as the second argument
when running this file. Output of the execution is printed in the preview
pane. You may want to use the same script in your pager mapping as well if
any.

    set previewer ~/.config/lf/pv.sh
    map i $~/.config/lf/pv.sh "$f" | less -R

Since this script is called for each file selection change it needs to be as
efficient as possible and this responsibility is left to the user. You may
use file extensions to determine the type of file more efficiently compared
to obtaining mime types from "file" command. Extensions can then be used to
match cleanly within a conditional.

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
over "pygmentize" for syntax highlighting. Besides, it is also important
that the application is processing the file on the fly rather than first
reading it to the memory and then do the processing afterwards. This is
especially relevant for big files. lf automatically closes the previewer
script output pipe with a SIGPIPE when enough lines are read. When
everything else fails, you can make use of the height argument to only feed
the first portion of the file to a program for preview.
`
