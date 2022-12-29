:: ATENTION: You need to install "bat" and "xpdf-utils" in order for this script to work.
::The most comfortable way is through chocolatey.

@echo off
if "%~x1" == ".pdf" (
    pdftotext -f 1 -l 5 -layout "%f%" - | bat --paging=never --style=numbers -f
) else if "%~x1" == ".exe" (
    echo Executable file
) else (
    bat --paging=never --style=numbers,changes -f "%f%"
)
