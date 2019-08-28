@echo off
rem Change working dir in cmd.exe to last dir in lf on exit.
rem
rem You need to put this file to a folder in %PATH% variable.

:tmploop
set tmpfile="%tmp%\lf.%random%.tmp"
if exist %tmpfile% goto:tmploop
lf -last-dir-path=%tmpfile% %*
if not exist %tmpfile% exit
set /p dir=<%tmpfile%
del /f %tmpfile%
if not exist "%dir%" exit
if "%dir%" == "%cd%" exit
cd /d "%dir%"
