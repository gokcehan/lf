# Change working dir in powershell to last dir in lf on exit.
#
# You need to put this file to a folder in $ENV:PATH variable.

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
