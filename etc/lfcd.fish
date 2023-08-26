# Change working dir in fish to last dir in lf on exit (adapted from ranger).
#
# You may put this file to a directory in $fish_function_path variable:
#
#     mkdir -p ~/.config/fish/functions
#     ln -s "/path/to/lfcd.fish" ~/.config/fish/functions
#
# You may also like to assign a key (Ctrl-O) to this command:
#
#     bind \co 'set old_tty (stty -g); stty sane; lfcd; stty $old_tty; commandline -f repaint'
#
# You may put this in a function called fish_user_key_bindings.

function lfcd
    cd (command lf -last-dir $argv)
end
