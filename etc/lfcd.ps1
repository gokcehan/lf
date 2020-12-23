# Change working dir in powershell to last dir in lf on exit.
#
# You need to put this file to a folder in $ENV:PATH variable.
#
# You may also like to assign a key to this command:
#
# You may put this in one of the profiles found in $PROFILE.
#
# Set-PSReadLineKeyHandler -Chord Ctrl+o -ScriptBlock {
#     [Microsoft.PowerShell.PSConsoleReadLine]::RevertLine()
#     [Microsoft.PowerShell.PSConsoleReadLine]::Insert('lfcd.ps1')
#     [Microsoft.PowerShell.PSConsoleReadLine]::AcceptLine()
# }
#

$tmp = [System.IO.Path]::GetTempFileName()
lf -last-dir-path="$tmp" $args
if (test-path -pathtype leaf "$tmp") {
    $dir = type "$tmp"
    remove-item -force "$tmp"
    if (test-path -pathtype container "$dir") {
        if ("$dir" -ne "$pwd") {
            cd "$dir"
        }
    }
}
