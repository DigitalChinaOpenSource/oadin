@echo off
set "USER_HOME=%USERPROFILE%"
set "OADIN_HOME=%LOCALAPPDATA%\Programs\Oadin"
set "PATH=%OADIN_HOME%;%PATH%"

echo Starting Oadin server...

REM 切换到安装目录
cd /d "%OADIN_HOME%"

REM 使用 start 命令独立启动 oadin，不依赖父进程
"%OADIN_HOME%\oadin.exe" server start -d > "%USER_HOME%\Oadin\oadin-server.log" 2>&1

echo Oadin server started successfully.
