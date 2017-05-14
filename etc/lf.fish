# Change working dir in Fish to last dir in lf on exit (adapted from ranger).
#
# You need to either copy the content of this file to ~/.config/fish/config.fish
# or source this file directly using something like:
#
#     set lffish "$GOPATH/src/github.com/gokcehan/lf/etc/lf.fish"
#     if test -f "$lffish"
#         source "$lffish"
#     end
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
