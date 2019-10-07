#compdef lf

# For using these completions you must: 
# - rename this file to _lf
# - add the containing folder to $fpath in .zshrc, like this:
# '''
# fpath=(/path/to/folder/containing_lf $fpath)
# autoload -U compinit
# compinit
# '''
#
# zsh completions for 'lf'
# automatically generated with http://github.com/RobSis/zsh-completion-generator
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
