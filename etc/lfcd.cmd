@echo off
rem Change working dir in cmd.exe to last dir in lf on exit.
rem
rem You need to put this file to a folder in %PATH% variable.

for /f "usebackq tokens=*" %%d in (`lf -print-last-dir %*`) do cd %%d
