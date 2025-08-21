@echo off
set "OADIN_HOME=%ProgramFiles%\Oadin"
set "PATH=%OADIN_HOME%;%PATH%"

REM 数据文件和日志应该写入到 ProgramData 目录
set "DATA_DIR=%PROGRAMDATA%\Oadin"
if not exist "%DATA_DIR%" mkdir "%DATA_DIR%"

REM 使用 start 命令独立启动 oadin，不依赖父进程
oadin server start -d > "%DATA_DIR%\oadin-server.log" 2>&1
