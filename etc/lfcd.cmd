@echo off
rem Change working dir in cmd.exe to last dir in lf on exit.
rem
rem You need to put this file to a folder in %PATH% variable.

:tmploop
set tmpfile="%tmp%\lf.%random%.tmp"
if exist %tmpfile% goto:tmploop
lf -last-dir-path=%tmpfile% %*
if not exist %tmpfile% exit 1
set /p dir=<%tmpfile%
del /f %tmpfile%
if not exist "%dir%" exit 1
if "%dir%" == "%cd%" exit 1
cd /d "%dir%"
