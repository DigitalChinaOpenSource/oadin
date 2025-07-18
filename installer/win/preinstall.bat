@echo off
setlocal

:: 定义日志文件路径
set "LOG_FILE=%~dp0preinstall_log.txt"
:: 定义 Oadin 可执行文件路径
set "OADIN_EXE="%~dp0oadin.exe""
:: 定义停止命令
set "STOP_COMMAND=oadin server stop"

echo ====================================================== >> "%LOG_FILE%"
echo %DATE% %TIME% - Starting pre-installation script...    >> "%LOG_FILE%"
echo Current directory: %~dp0                               >> "%LOG_FILE%"
echo Oadin executable path: %OADIN_EXE%                     >> "%LOG_FILE%"

:: 检查 oadin.exe 进程是否正在运行
echo Checking for running Oadin process...                  >> "%LOG_FILE%"

:: 查找进程，/FI "IMAGENAME eq oadin.exe" 是筛选条件
tasklist /nh /fi "IMAGENAME eq oadin.exe" | find /i "oadin.exe" >nul
if %errorlevel% equ 0 (
    echo Oadin process found running. Attempting to stop... >> "%LOG_FILE%"
    :: 尝试停止 Oadin 服务器
    %STOP_COMMAND% >> "%LOG_FILE%" 2>&1
    echo Stop command executed.                             >> "%LOG_FILE%"

    :: 等待一段时间，给 Oadin 服务器时间来停止
    timeout /t 10 /nobreak >nul 2>&1
    echo Waited for 10 seconds.                             >> "%LOG_FILE%"

    :: 再次检查 Oadin 进程是否已停止
    tasklist /nh /fi "IMAGENAME eq oadin.exe" | find /i "oadin.exe" >nul
    if %errorlevel% equ 0 (
        echo ERROR: Oadin process is still running after stop attempt! >> "%LOG_FILE%"
        echo Please manually close Oadin CLI before proceeding.        >> "%LOG_FILE%"
        :: 可以选择在此处退出，让 NSIS 报错，或者继续但提醒用户
        :: exit /b 1
    ) else (
        echo Oadin process successfully stopped.             >> "%LOG_FILE%"
    )
) else (
    echo Oadin process not found running.                    >> "%LOG_FILE%"
)

echo Pre-installation script finished.                     >> "%LOG_FILE%"
echo ====================================================== >> "%LOG_FILE%"

endlocal
exit /b 0