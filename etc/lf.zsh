#compdef lf

# Autocompletion for zsh shell.
#
# You need to rename this file to _lf and add containing folder to $fpath in
# ~/.zshrc file:
#
#     fpath=(/path/to/directory/containing/the/file $fpath)
#     autoload -U compinit
#     compinit
#

local arguments

arguments=(
    '-command[command to execute on client initialization]'
    '-cpuprofile[path to the file to write the CPU profile]'
    '-doc[show documentation]'
    '-last-dir-path[path to the file to write the last dir on exit (to use for cd)]'
    '-memprofile[path to the file to write the memory profile]'
    '-remote[send remote command to server]'
    '-selection-path[path to the file to write selected files on open (to use as open file dialog)]'
    '-server[start server (automatic)]'
    '-version[show version]'
    '*:filename:_files'
)

_arguments -s $arguments
