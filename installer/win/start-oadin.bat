@echo off
set "OADIN_HOME=%ProgramFiles%\Oadin"
set "PATH=%OADIN_HOME%;%PATH%"

REM 使用 start 命令独立启动 oadin，不依赖父进程
oadin server start -d > "%OADIN_HOME%\oadin-server.log" 2>&1