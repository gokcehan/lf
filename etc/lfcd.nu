# Change working dir in shell to last dir in lf on exit (adapted from ranger).
#
# You need to add this to your Nushell Enviroment Config File
# (Execute 'config env' in the nushell to open it).

# You may also like to assign a key (Ctrl-O) to this command:
# See the documentation: https://www.nushell.sh/book/line_editor.html#keybindings
#
# keybindings: [
#   {
#     name: lfcd
#     modifier: control
#     keycode: char_o
#     mode: [emacs, vi_normal, vi_insert]
#     event: {
#       send: executehostcommand
#       cmd: "lfcd"
#     }
#   }
# ]

def-env lfcd [] {
  let tmp = (mktemp)
  lf -last-dir-path $tmp
  try {
    let target_dir = (open --raw $tmp)
    rm -f $tmp
    try {
        if ($target_dir != $env.PWD) { cd $target_dir }
    } catch { |e| print -e $'lfcd: Can not change to ($target_dir): ($e | get debug)' }
  } catch {
    |e| print -e $'lfcd: Reading ($tmp) returned an error: ($e | get debug)'
  }
}
