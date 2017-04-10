# Change working dir in Fish to last dir in lf on exit (adapted from Bash version)
#
# You need to copy the content of this file to ~/.config/fish/config.fish
#

function lf
    set tmp (mktemp)
    command lf -last-dir-path=$tmp $argv 
    if test -f $tmp
        set dir (cat $tmp)
        if [ $dir != "" && $dir != (pwd) ]
            cd $dir
        end
    end
    rm -f $tmp
end
