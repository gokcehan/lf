#
#	For using these completions you must load it from .tcshrc
#
#		source ~/.config/lf/lf.csh
#
set lf_args = "-command -cpuprofile -doc --help -last-dir-path -memprofile -remote -selection-path -server -version --version"

uncomplete lf
uncomplete lfcd

complete lf   "C/-*/(${lf_args})/"
complete lfcd "C/-*/(${lf_args})/"
