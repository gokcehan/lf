# Tutorial

## Configuration

The configuration file should be located in `~/.config/lf/lfrc`.
A sample configuration file can be found [here](/etc/lfrc.example).

## Prefixes

The following command prefixes are used by `lf`:

    :  read (default)
    $  read-shell
    !  read-shell-wait
    &  read-shell-async
    /  search
    ?  search-back

The same evaluator is used for the command line and the configuration file.
The difference is that prefixes are not necessary in the command line.
Instead different modes are provided to read corresponding commands.
Note that by default these modes are mapped to the prefix keys above.

## Syntax

Characters from `#` to `\n` are comments and ignored.

There are three special commands for configuration.

`set` is used to set an option which could be:

- bool (e.g. `set hidden`, `set nohidden`, `set hidden!`)
- int (e.g. `set scrolloff 10`)
- string (e.g. `set sortby time`)

`map` is used to bind a key to a command which could be:

- built-in command (e.g. `map gh cd ~`)
- custom command (e.g. `map dD trash`)
- shell command (e.g. `map i $less "$f"`, `map u !du -h . | less`)

`cmd` is used to define a custom command.

If there is no prefix then `:` is assumed.
An explicit `:` could be provided to group statements until a `\n` occurs.
This is especially useful for `map` and `cmd` commands.
If you need multiline you can wrap statements in `{{` and `}}` after the proper prefix.

## Custom Commands

To wrap up let us write a shell command to move selected file(s) to trash.

A first attempt to write such a command may look like this:

    cmd trash ${{
        mkdir -p ~/.trash
        if [ -z $fs ]; then
            mv --backup=numbered "$f" $HOME/.trash
        else
            IFS=':'; mv --backup=numbered $fs $HOME/.trash
        fi
    }}

We check `$fs` to see if there are any marked files.
Otherwise we just delete the current file.
Since this is such a common pattern, a separate `$fx` variable is provided.
We can use this variable to get rid of the conditional.

    cmd trash ${{
        mkdir -p ~/.trash
        IFS=':'; mv --backup=numbered $fx $HOME/.trash
    }}

The trash directory is checked each time the command is executed.
We can move it outside of the command so it would only run once at startup.

    ${{ mkdir -p ~/.trash }}

    cmd trash ${{ IFS=':'; mv --backup=numbered $fx $HOME/.trash }}

Since these are one liners, we can drop `{{` and `}}`.

    $mkdir -p ~/.trash

    cmd trash $IFS=':'; mv --backup=numbered $fx $HOME/.trash

Finally note that we set `IFS` variable accordingly in the command.
Instead we could use the `ifs` option to set it for all commands (e.g. `set ifs ':'`).
This could be especially useful for interactive use (e.g. `rm $fs` would simply work).
This option is not set by default as things may behave unexpectedly at other places.

## Opening Files

You can use `open-file` command to open a file.
This is a special command called by `open` when the current file is not a directory.
Normally a user maps the `open` command to a key (default `l`) and customize `open-file` command as desired.
You can define it just as you would define any other command.

    cmd open-file $IFS=':'; vim $fx

It is possible to use different command types.

    cmd open-file &xdg-open "$f"

You may want to use either file extensions or mime types with `file`.

    cmd open-file ${{
        case $(file --mime-type "$f" -b) in
            text/*) IFS=':'; vim $fx;;
            *) IFS=':'; for f in $fx; do xdg-open "$f" &> /dev/null & done;;
        esac
    }}

`lf` does not come bundled with a file opener.
Below are a few different file openers you can use.

- [xdg-utils](https://www.freedesktop.org/wiki/Software/xdg-utils/) (executable name is `xdg-open`)
- [libfile-mimeinfo-perl](https://metacpan.org/release/File-MimeInfo) (executable name is `mimeopen`)
- [rifle](http://ranger.nongnu.org/) (ranger's default file opener)
- [mimeo](http://xyne.archlinux.ca/projects/mimeo/)
