# Autocompletion for PowerShell.
#
# You need to either copy the content of this file to $PROFILE
# or call this script directly.
#

using namespace System.Management.Automation

Register-ArgumentCompleter -Native -CommandName 'lf' -ScriptBlock {
    param($wordToComplete)
    $completions = @(
        [CompletionResult]::new('-command ', '-command', [CompletionResultType]::ParameterName, 'command to execute on client initialization')
        [CompletionResult]::new('-cpuprofile ', '-cpuprofile', [CompletionResultType]::ParameterName, 'path to the file to write the CPU profile')
        [CompletionResult]::new('-doc', '-doc', [CompletionResultType]::ParameterName, 'show documentation')
        [CompletionResult]::new('-last-dir-path ', '-last-dir-path', [CompletionResultType]::ParameterName, 'path to the file to write the last dir on exit (to use for cd)')
        [CompletionResult]::new('-memprofile ', '-memprofile', [CompletionResultType]::ParameterName, 'path to the file to write the memory profile')
        [CompletionResult]::new('-remote ', '-remote', [CompletionResultType]::ParameterName, 'send remote command to server')
        [CompletionResult]::new('-selection-path ', '-selection-path', [CompletionResultType]::ParameterName, 'path to the file to write selected files on open (to use as open file dialog)')
        [CompletionResult]::new('-server', '-server', [CompletionResultType]::ParameterName, 'start server (automatic)')
        [CompletionResult]::new('-version', '-version', [CompletionResultType]::ParameterName, 'show version')
    )

    if ($wordToComplete.StartsWith('-')) {
        $completions.Where{ $_.CompletionText -like "$wordToComplete*" } | Sort-Object -Property ListItemText
    }
}
