# Autocompletion for Fish shell
# Put this file somewhere in $fish_complete_path
# for example in ~/.config/fish/completions/

complete -c lf -o cpuprofile  -r -d 'path to the file to write the CPU profile'
complete -c lf -o doc  -d 'show documentation'
complete -c lf -o last-dir-path  -r -d 'path to the file to write the last dir on exit (to use for cd)'
complete -c lf -o memprofile  -r -d 'path to the file to write the memory profile'
complete -c lf -o remote  -x -d 'send remote command to server'
complete -c lf -o selection-path  -r -d 'path to the file to write selected files on open (to use as open file dialog)'
complete -c lf -o server  -d 'start server (automatic)'
complete -c lf -o version  -d 'show version'
complete -c lf -s h -d "short help"
