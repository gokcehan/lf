# Change working dir in shell to last dir in lf on exit (adapted from ranger).
#
# You need to add this to your Nushell Enviroment Config File
# (Execute 'config env' in the nushell to open it).

# You may also like to assign a key (Ctrl-O) to this command:
# See the documentation: https://www.nushell.sh/book/line_editor.html#keybindings
#
# keybindings: [
#    {
#      name: lfcd
#      modifier: control
#      keycode: char_o
#      mode: [emacs , vi_normal, vi_insert]
#      event: {
#        send: executehostcommand
#        cmd: "lfcd"
#      }
#    }
#  ]


def-env lfcd [] {
  let tmp = $"(mktemp)";
  let cmd = $'-last-dir-path=($tmp)';
  run-external 'lf' $cmd;
  if ($tmp | path exists) {
    let dir = $"(cat $tmp)";
    rm -f $"$tmp";
    if ($dir | path exists) {
      if ( $dir != $"pwd" ) {
        cd $dir;
      }
    }
  }
}