#	Change working dir in shell to last dir in lf on exit.
#
#	You need to either copy the content of this file to your shell rc file
#	(e.g. ~/.tcshrc) or source this file directly:
#
#	setenv LF_HOME "${HOME}/.config/lf"
#	[ -e "${LF_HOME}/lfcd.csh" ] && source "${LF_HOME}/lfcd.csh"
#
#	You may also like to assign a key to this command:
#
#		bindkey -c "^O" lfcd
#
alias lfcd 'set _=`mktemp` && lf -last-dir-path=$_ "\!*" && set _=`cat $_ && rm -f $_` && [ -d "$_" ] && cd "$_"'
