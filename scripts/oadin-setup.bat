@echo off
setlocal enabledelayedexpansion

REM === 配置部分 ===
set "SETUP_FILE=oadin-installer-test-2.0.40.exe"
set "SILENT_ARGS=/S"
set "LOG_FILE=install.log"

set "AI_SMARTVISION_FILE=ai-smartvision-2.0.0-x64.exe"
set "AI_SMARTVISION_INSTALL_DIR=%ProgramFiles%\ai-smartvision"
set "AI_SMARTVISION_LOG=ai-smartvision-install.log"

set "OLLAMA_ZIP=ipex-llm-ollama.zip"
set "OLLAMA_DIR=%ProgramFiles%\Oadin"

set "MODEL_ZIP=models.zip"
set "MODEL_DIR=%ProgramFiles%/Oadin/engine/ollama"

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

REM === 安装 AI SmartVision ===
if not exist "%AI_SMARTVISION_FILE%" (
    echo 没有找到 AI SmartVision 安装包: %AI_SMARTVISION_FILE%
    pause
    exit /b 7
)
echo 正在静默安装 AI SmartVision ...
start /wait "" "%AI_SMARTVISION_FILE%" /S /D="%AI_SMARTVISION_INSTALL_DIR%" /WAIT /LOG="%AI_SMARTVISION_LOG%"
if %errorlevel% EQU 0 (
    echo AI SmartVision 安装完成！
) else (
    echo AI SmartVision 安装失败，错误码: %errorlevel%
    pause
    exit /b 8
)

REM === 解压ollama的zip包到指定目录 ===
if not exist "%OLLAMA_ZIP%" (
    echo 没有找到Ollama安装包: %OLLAMA_ZIP%
    pause
    exit /b 9
)
echo 正在解压Ollama压缩包到 %OLLAMA_DIR% ...
powershell -command "Expand-Archive -Path '%OLLAMA_ZIP%' -DestinationPath '%OLLAMA_DIR%' -Force"
if %errorlevel% EQU 0 (
    echo Ollama 解压成功！
) else (
    echo Ollama 解压失败！
    pause
    exit /b 10
)

REM === 解压模型包到指定目录 ===
if not exist "%MODEL_ZIP%" (
    echo 没有找到模型包: %MODEL_ZIP%
    pause
    exit /b 11
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
    exit /b 12
)

echo 所有操作完成！
pause
endlocal
