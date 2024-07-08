@echo off
rem Change working dir in cmd.exe to last dir in lf on exit.
rem
rem You need to put this file to a folder in %PATH% variable.

chcp 65001 > nul 2>&1
for /f "usebackq tokens=*" %%d in (`lf -print-last-dir %*`) do cd /d %%d
