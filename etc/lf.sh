# Change working dir in shell to last dir in lf on exit (adapted from ranger).
#
# You need to either copy the content of this file to your shell rc file
# (e.g. ~/.bashrc) or source this file directly using something like:
#
#     LFSH="$GOPATH/src/github.com/gokcehan/lf/etc/lf.sh"
#     if [ -f "$LFSH" ]; then
#         source "$LFSH"
#     fi
#
# You may also like to assign a key to this command:
#
#     bind '"\C-o":"\C-ulf\C-m"'
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
