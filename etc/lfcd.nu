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

# For nushell version >= 0.87.0
def --env --wrapped lfcd [...args: string] { 
  cd (lf -print-last-dir ...$args)
}
