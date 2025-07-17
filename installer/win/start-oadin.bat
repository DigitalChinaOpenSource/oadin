@echo off
set "USER_HOME=%USERPROFILE%"
set "OADIN_HOME=%USER_HOME%\Oadin"
set "PATH=%OADIN_HOME%;%PATH%"

REM 使用 start 命令独立启动 oadin，不依赖父进程
oadin server start -d > "%USER_HOME%\Oadin\oadin-server.log" 2>&1
