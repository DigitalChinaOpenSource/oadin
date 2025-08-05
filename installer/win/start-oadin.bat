@echo off
set "USER_HOME=%USERPROFILE%"
set "OADIN_HOME=%LOCALAPPDATA%\Programs\Oadin"

echo Starting Oadin server...

REM 确保数据目录存在
if not exist "%USER_HOME%\Oadin" mkdir "%USER_HOME%\Oadin"

REM 切换到安装目录
cd /d "%OADIN_HOME%"

REM 使用完整路径启动 oadin
"%OADIN_HOME%\oadin.exe" server start -d > "%USER_HOME%\Oadin\oadin-server.log" 2>&1

echo Oadin server started successfully.
