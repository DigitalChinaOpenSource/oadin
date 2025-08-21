@echo off
setlocal enabledelayedexpansion

REM === 配置部分 ===
set "SETUP_FILE=oadin-installer-test-2.0.8.exe"
set "SILENT_ARGS=/S"
set "LOG_FILE=install.log"

set "OLLAMA_ZIP=ollama-win.zip"
set "OLLAMA_DIR=C:\ollama"

set "MODEL_ZIP=model.zip"
set "MODEL_DIR=C:\ollama\models"

REM === 检查管理员权限 ===
openfiles >nul 2>&1
if errorlevel 1 (
    echo 请以管理员方式运行本脚本！
    pause
    exit /b 1
)

REM === 检查安装包是否存在 ===
if not exist "%SETUP_FILE%" (
    echo 没有找到安装包^: %SETUP_FILE%
    pause
    exit /b 2
)
echo 正在静默安装 %SETUP_FILE% ...
start /wait "" "%SETUP_FILE%" %SILENT_ARGS% /log=%LOG_FILE%
REM === 检查安装是否成功 ===
if %errorlevel% EQU 0 (
    echo 安装完成！
) else (
    echo 安装过程返回错误码^: %errorlevel%
)

REM === 解压ollama的zip包到指定目录 ===
if not exist "%OLLAMA_ZIP%" (
    echo 没有找到Ollama安装包: %OLLAMA_ZIP%
    pause
    exit /b 3
)
echo 正在解压Ollama压缩包到 %OLLAMA_DIR% ...
powershell -command "Expand-Archive -Path '%OLLAMA_ZIP%' -DestinationPath '%OLLAMA_DIR%' -Force"
if %errorlevel% EQU 0 (
    echo Ollama 解压成功！
) else (
    echo Ollama 解压失败！
    pause
    exit /b 4
)

REM === 解压模型包到指定目录 ===
if not exist "%MODEL_ZIP%" (
    echo 没有找到模型包: %MODEL_ZIP%
    pause
    exit /b 5
)
if not exist "%MODEL_DIR%" (
    mkdir "%MODEL_DIR%"
)
echo 正在解压模型包到 %MODEL_DIR% ...
powershell -command "Expand-Archive -Path '%MODEL_ZIP%' -DestinationPath '%MODEL_DIR%' -Force"
if %errorlevel% EQU 0 (
    echo 模型解压完成！
) else (
    echo 模型解压失败！
    pause
    exit /b 6
)

echo 所有操作完成！
pause
endlocal
