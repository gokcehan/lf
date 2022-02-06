# Autocompletion for fish shell.
#
# You may put this file to a directory in $fish_complete_path variable:
#
#     mkdir -p ~/.config/fish/completions
#     ln -s "/path/to/lf.fish" ~/.config/fish/completions
#

complete -c lf -o command -r -d 'command to execute on client initialization'
complete -c lf -o config -r -d 'path to the config file (instead of the usual paths)'
complete -c lf -o cpuprofile -r -d 'path to the file to write the CPU profile'
complete -c lf -o doc -d 'show documentation'
complete -c lf -o last-dir-path -r -d 'path to the file to write the last dir on exit (to use for cd)'
complete -c lf -o log -r -d 'path to the log file to write messages'
complete -c lf -o memprofile -r -d 'path to the file to write the memory profile'
complete -c lf -o remote -x -d 'send remote command to server'
complete -c lf -o selection-path -r -d 'path to the file to write selected files on open (to use as open file dialog)'
complete -c lf -o server -d 'start server (automatic)'
complete -c lf -o single -d 'start a client without server'
complete -c lf -o version -d 'show version'
complete -c lf -o help -d 'show help'
