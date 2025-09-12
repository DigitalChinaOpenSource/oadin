@echo off
REM 改进版 Oadin 启动脚本
REM 支持从安装目录动态获取路径，兼容服务模式和手动启动模式

REM 尝试从环境变量获取安装路径（服务模式）
if defined OADIN_HOME (
    set "CURRENT_OADIN_HOME=%OADIN_HOME%"
) else (
    REM 尝试从注册表获取安装路径（手动启动模式）
    for /f "tokens=2*" %%a in ('reg query "HKLM\SOFTWARE\Digital China\Oadin CLI" /v "InstallDir" 2^>nul ^| findstr "InstallDir"') do (
        set "CURRENT_OADIN_HOME=%%b"
    )
)

REM 如果以上方法都失败，使用默认路径
if not defined CURRENT_OADIN_HOME (
    set "CURRENT_OADIN_HOME=%ProgramFiles%\Oadin"
)

REM 设置环境变量
set "PATH=%CURRENT_OADIN_HOME%;%PATH%"
set "LOG_FILE=%CURRENT_OADIN_HOME%\oadin-server.log"

REM 创建日志文件目录（如果不存在）
if not exist "%CURRENT_OADIN_HOaME%" (
    mkdir "%CURRENT_OADIN_HOME%" >nul 2>&1
)

REM 记录启动信息
echo [%date% %time%] Oadin server starting... >> "%LOG_FILE%"
echo [%date% %time%] OADIN_HOME: %CURRENT_OADIN_HOME% >> "%LOG_FILE%"

REM 检查程序是否存在
if not exist "%CURRENT_OADIN_HOME%\oadin.exe" (
    echo 错误：未找到 oadin.exe 在 %CURRENT_OADIN_HOME%
    echo [%date% %time%] 错误：未找到 oadin.exe >> "%LOG_FILE%"
    pause
    exit /b 1
)

REM 启动服务器（根据参数决定是否后台运行）
if "%1"=="-service" (
    REM 服务模式：不使用 start 命令，让服务管理器控制进程
    echo [%date% %time%] Starting in service mode >> "%LOG_FILE%"
    "%CURRENT_OADIN_HOME%\oadin.exe" server start -d >> "%LOG_FILE%" 2>&1
) else (
    REM 手动模式：使用 start 命令创建独立窗口
    echo [%date% %time%] Starting in normal mode >> "%LOG_FILE%"
    start "Oadin Server" "%CURRENT_OADIN_HOME%\oadin.exe" server start
    echo Oadin 服务器已启动
    echo 日志文件：%LOG_FILE%
)

exit /b 0