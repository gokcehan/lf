# change the directory to last dir on exit
# adapted from the similar script for ranger
#
# you need to add something like the following to your shell rc file (e.g. ~/.bashrc):
#
# LFSH="$GOPATH/src/github.com/gokcehan/lf/etc/lf.sh"
# if [ -f "$LFSH" ]; then
#     source "$LFSH"
#     bind '"\C-o":"\C-ulf\C-m"'
# fi
#

lf () {
    tmp="$(mktemp)"
    command lf -last-dir-path="$tmp" "$@"
    if [ -f "$tmp" ]; then
        dir="$(cat "$tmp")"
        [ "$dir" != "$(pwd)" ] && cd "$dir"
    fi
    rm -f "$tmp"
}
