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
    '-config[path to the config file (instead of the usual paths)]'
    '-cpuprofile[path to the file to write the CPU profile]'
    '-doc[show documentation]'
    '-last-dir-path[path to the file to write the last dir on exit (to use for cd)]'
    '-log[path to the log file to write messages]'
    '-memprofile[path to the file to write the memory profile]'
    '-print-last-dir[print the last dir to stdout on exit (to use for cd)]'
    '-print-selection[print the selected files to stdout on open (to use as open file dialog)]'
    '-remote[send remote command to server]'
    '-selection-path[path to the file to write selected files on open (to use as open file dialog)]'
    '-server[start server (automatic)]'
    '-single[start a client without server]'
    '-version[show version]'
    '-help[show help]'
    '*:filename:_files'
)

_arguments -s $arguments
