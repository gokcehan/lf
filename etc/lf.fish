# Change working dir in fish to last dir in lf on exit (adapted from ranger).
#
# You need to either copy the content of this file to ~/.config/fish/config.fish
# or put the file to ~/.config/fish/functions using something like:
#
#     mkdir -p ~/.config/fish/functions
#     ln -s "$GOPATH/src/github.com/gokcehan/lf/etc/lf.fish" ~/.config/fish/functions
#

function lf
    set tmp (mktemp)
    command lf -last-dir-path=$tmp $argv
    if test -f "$tmp"
        set dir (cat $tmp)
        if test -n "$dir"
            if test "$dir" != (pwd)
                cd $dir
            end
        end
    end
    rm -f $tmp
end
