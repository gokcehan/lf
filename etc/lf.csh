# Autocompletion for tcsh shell.
#
# You need to either copy the content of this file to your shell rc file
# (e.g. ~/.tcshrc) or source this file directly:
#
#     set LF_COMPLETE = "/path/to/lf.csh"
#     if ( -f "$LF_COMPLETE" ) then
#         source "$LF_COMPLETE"
#     endif
#

set LF_ARGS = "-command -config -cpuprofile -doc -last-dir -log -memprofile -remote -selection -server -single -version -help "

complete lf   "C/-*/(${LF_ARGS})/"
complete lfcd "C/-*/(${LF_ARGS})/"
