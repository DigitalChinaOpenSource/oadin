@echo off
set "USER_HOME=%USERPROFILE%"
set "BYZE_HOME=%USER_HOME%\Byze"
set "PATH=%BYZE_HOME%;%PATH%"

REM 使用 start 命令独立启动 byze，不依赖父进程
start "" /b byze server start -d
