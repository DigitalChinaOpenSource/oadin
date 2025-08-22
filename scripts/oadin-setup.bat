@echo off
setlocal enabledelayedexpansion

REM ========================================
REM OADIN Auto Installation Script (No PowerShell)
REM ========================================

REM === 配置部分 ===
set "SETUP_FILE=oadin-installer-test-2.0.40.exe"
set "OADIN_INSTALL_DIR=%ProgramFiles%\Oadin"
set "SILENT_ARGS=/S"
set "LOG_FILE=%OADIN_INSTALL_DIR%\install.log"

set "AI_SMARTVISION_FILE=ai-smartvision-2.0.0-x64.exe"
set "AI_SMARTVISION_INSTALL_DIR=%ProgramFiles%\ai-smartvision"
set "AI_SMARTVISION_LOG=%OADIN_INSTALL_DIR%\ai-smartvision-install.log"

set "OLLAMA_ZIP=ipex-llm-ollama.zip"
set "OLLAMA_DIR=%OADIN_INSTALL_DIR%\ipex-llm-ollama"

set "MODEL_ZIP=models.zip"
set "MODEL_DIR=%ProgramData%\Oadin\engine\ollama"

set "TEMP_EXTRACT_DIR=%TEMP%\oadin_extract_%RANDOM%"

REM === 颜色定义 ===
set "GREEN=[92m"
set "RED=[91m"
set "YELLOW=[93m"
set "BLUE=[94m"
set "NC=[0m"

REM === 跳转到主函数 ===
goto :main

REM === 函数定义 ===
:print_status
echo %BLUE%[INFO]%NC% %~1
goto :eof

:print_success
echo %GREEN%[SUCCESS]%NC% %~1
goto :eof

:print_error
echo %RED%[ERROR]%NC% %~1
goto :eof

:print_warning
echo %YELLOW%[WARNING]%NC% %~1
goto :eof

:create_directory
if not exist "%~1" (
    call :print_status "创建目录: %~1"
    mkdir "%~1" 2>nul
    if !errorlevel! NEQ 0 (
        call :print_error "无法创建目录: %~1"
        exit /b 1
    )
    call :print_success "目录创建成功"
) else (
    call :print_status "目录已存在: %~1"
)
goto :eof

:check_file_exists
if not exist "%~1" (
    call :print_error "文件不存在: %~1"
    exit /b 1
)
call :print_success "文件检查通过: %~1"
goto :eof

:extract_zip_native
set "ZIP_FILE=%~1"
set "DEST_DIR=%~2"
set "EXTRACT_TEMP=%TEMP_EXTRACT_DIR%\%RANDOM%"

call :print_status "开始解压: %ZIP_FILE% -> %DEST_DIR%"

REM 创建临时解压目录
call :create_directory "%EXTRACT_TEMP%"

REM 使用系统自带的解压工具 (Windows 10+)
if exist "%SystemRoot%\System32\tar.exe" (
    call :print_status "使用 tar.exe 解压..."
    "%SystemRoot%\System32\tar.exe" -xf "%ZIP_FILE%" -C "%EXTRACT_TEMP%" >nul 2>&1
    if !errorlevel! EQU 0 (
        call :move_extracted_files "%EXTRACT_TEMP%" "%DEST_DIR%"
        if !errorlevel! EQU 0 (
            call :print_success "使用 tar.exe 解压成功"
            goto :extract_cleanup
        )
    )
    call :print_warning "tar.exe 解压失败，尝试其他方法..."
)

REM 使用 VBScript 解压
call :print_status "使用 VBScript 解压..."
call :extract_zip_vbs "%ZIP_FILE%" "%EXTRACT_TEMP%"
if !errorlevel! EQU 0 (
    call :move_extracted_files "%EXTRACT_TEMP%" "%DEST_DIR%"
    if !errorlevel! EQU 0 (
        call :print_success "使用 VBScript 解压成功"
        goto :extract_cleanup
    )
)

REM 使用 7-Zip (如果安装了)
call :print_status "尝试使用 7-Zip 解压..."
call :extract_zip_7zip "%ZIP_FILE%" "%DEST_DIR%"
if !errorlevel! EQU 0 (
    call :print_success "使用 7-Zip 解压成功"
    goto :extract_cleanup
)

call :print_error "所有解压方法都失败了"
exit /b 1

:extract_cleanup
if exist "%EXTRACT_TEMP%" (
    rmdir /s /q "%EXTRACT_TEMP%" >nul 2>&1
)
goto :eof

:move_extracted_files
set "SRC_DIR=%~1"
set "DEST_DIR=%~2"

call :create_directory "%DEST_DIR%"

REM 移动所有文件和文件夹
for /d %%i in ("%SRC_DIR%\*") do (
    xcopy "%%i" "%DEST_DIR%\%%~ni" /E /I /H /Y >nul 2>&1
)

for %%i in ("%SRC_DIR%\*.*") do (
    copy "%%i" "%DEST_DIR%\" >nul 2>&1
)

goto :eof

:extract_zip_vbs
set "ZIP_FILE=%~1"
set "DEST_DIR=%~2"
set "VBS_FILE=%TEMP%\extract_%RANDOM%.vbs"

REM 创建VBScript解压脚本
(
echo Set objShell = CreateObject^("Shell.Application"^)
echo Set objFolder = objShell.NameSpace^("%ZIP_FILE%"^)
echo If Not objFolder Is Nothing Then
echo     Set objDestFolder = objShell.NameSpace^("%DEST_DIR%"^)
echo     If objDestFolder Is Nothing Then
echo         Set objFSO = CreateObject^("Scripting.FileSystemObject"^)
echo         objFSO.CreateFolder^("%DEST_DIR%"^)
echo         Set objDestFolder = objShell.NameSpace^("%DEST_DIR%"^)
echo     End If
echo     objDestFolder.CopyHere objFolder.Items, 256
echo     WScript.Echo "VBScript extraction completed"
echo Else
echo     WScript.Echo "Failed to open ZIP file"
echo     WScript.Quit 1
echo End If
) > "%VBS_FILE%"

cscript //nologo "%VBS_FILE%" >nul 2>&1
set "VBS_RESULT=%errorlevel%"

if exist "%VBS_FILE%" del "%VBS_FILE%" >nul 2>&1

exit /b %VBS_RESULT%

:extract_zip_7zip
set "ZIP_FILE=%~1"
set "DEST_DIR=%~2"

REM 检查常见的7-Zip安装位置
set "SEVENZIP_EXE="
for %%p in (
    "%ProgramFiles%\7-Zip\7z.exe"
    "%ProgramFiles(x86)%\7-Zip\7z.exe"
    "%ProgramW6432%\7-Zip\7z.exe"
    "%LOCALAPPDATA%\Programs\7-Zip\7z.exe"
) do (
    if exist "%%p" (
        set "SEVENZIP_EXE=%%p"
        goto :found_7zip
    )
)

exit /b 1

:found_7zip
call :create_directory "%DEST_DIR%"
"%SEVENZIP_EXE%" x "%ZIP_FILE%" -o"%DEST_DIR%" -y >nul 2>&1
exit /b %errorlevel%

:install_program
set "INSTALLER=%~1"
set "INSTALL_ARGS=%~2"
set "LOG_FILE_PARAM=%~3"
set "PROGRAM_NAME=%~4"

call :print_status "开始安装 %PROGRAM_NAME%..."
call :check_file_exists "%INSTALLER%"
if !errorlevel! NEQ 0 exit /b !errorlevel!

REM 调试输出
call :print_status "执行命令: %INSTALLER% %INSTALL_ARGS%"

if not "%LOG_FILE_PARAM%"=="" (
    "%INSTALLER%" %INSTALL_ARGS% /log="%LOG_FILE_PARAM%"
) else (
    "%INSTALLER%" %INSTALL_ARGS%
)

if !errorlevel! EQU 0 (
    call :print_success "%PROGRAM_NAME% 安装完成！"
) else (
    call :print_error "%PROGRAM_NAME% 安装失败，错误码: !errorlevel!"
    exit /b !errorlevel!
)
goto :eof

:install_ai_smartvision
set "INSTALLER=%~1"
set "INSTALL_DIR=%~2"
set "LOG_FILE_PARAM=%~3"

call :print_status "开始安装 AI SmartVision..."
call :check_file_exists "%INSTALLER%"
if !errorlevel! NEQ 0 exit /b !errorlevel!

call :print_status "安装路径: %INSTALL_DIR%"
call :print_status "日志文件: %LOG_FILE_PARAM%"

REM 使用临时批处理文件来避免引号问题
set "TEMP_BAT=%TEMP%\ai_smartvision_install_%RANDOM%.bat"
(
echo @echo off
echo "%INSTALLER%" /S "/D=%INSTALL_DIR%" /WAIT /log="%LOG_FILE_PARAM%"
echo exit /b %%errorlevel%%
) > "%TEMP_BAT%"

call "%TEMP_BAT%"
set "INSTALL_RESULT=%errorlevel%"

if exist "%TEMP_BAT%" del "%TEMP_BAT%" >nul 2>&1

if %INSTALL_RESULT% EQU 0 (
    call :print_success "AI SmartVision 安装完成！"
) else (
    call :print_error "AI SmartVision 安装失败，错误码: %INSTALL_RESULT%"
    exit /b %INSTALL_RESULT%
)
goto :eof

:cleanup_temp
if exist "%TEMP_EXTRACT_DIR%" (
    call :print_status "清理临时文件..."
    rmdir /s /q "%TEMP_EXTRACT_DIR%" >nul 2>&1
)
goto :eof

:main
REM === 脚本开始 ===
call :print_status "OADIN 自动化安装脚本启动"
call :print_status "当前时间: %date% %time%"

REM === 检查管理员权限 ===
call :print_status "检查管理员权限..."
openfiles >nul 2>&1
if errorlevel 1 (
    call :print_error "请以管理员方式运行本脚本！"
    pause
    exit /b 1
)
call :print_success "管理员权限验证通过"

REM === 显示安装目录配置 ===
call :print_status "========================================="
call :print_status "安装目录配置:"
call :print_status "  - OADIN 主程序: %OADIN_INSTALL_DIR%"
call :print_status "  - AI SmartVision: %AI_SMARTVISION_INSTALL_DIR%"
call :print_status "  - Ollama: %OLLAMA_DIR%"
call :print_status "  - 模型数据: %MODEL_DIR%"
call :print_status "  - 日志文件: %LOG_FILE%"
call :print_status "========================================="

REM === 创建临时目录 ===
call :create_directory "%TEMP_EXTRACT_DIR%"

REM === 安装 OADIN 主程序 ===
call :print_status "=== 第1步: 安装 OADIN 主程序 ==="
call :print_status "目标安装目录: %OADIN_INSTALL_DIR%"
call :install_program "%SETUP_FILE%" "/S /D=\"%OADIN_INSTALL_DIR%\"" "%LOG_FILE%" "OADIN"
if !errorlevel! NEQ 0 (
    call :cleanup_temp
    pause
    exit /b !errorlevel!
)

REM === 安装 AI SmartVision ===
call :print_status "=== 第2步: 安装 AI SmartVision ==="
call :install_ai_smartvision "%AI_SMARTVISION_FILE%" "%AI_SMARTVISION_INSTALL_DIR%" "%AI_SMARTVISION_LOG%"
if !errorlevel! NEQ 0 (
    call :cleanup_temp
    pause
    exit /b !errorlevel!
)

REM === 解压 Ollama 包 ===
call :print_status "=== 第3步: 安装 Ollama ==="
call :extract_zip_native "%OLLAMA_ZIP%" "%OLLAMA_DIR%"
if !errorlevel! NEQ 0 (
    call :print_error "Ollama 安装失败"
    call :cleanup_temp
    pause
    exit /b !errorlevel!
)

REM === 解压模型包 ===
call :print_status "=== 第4步: 安装模型文件 ==="
call :extract_zip_native "%MODEL_ZIP%" "%MODEL_DIR%"
if !errorlevel! NEQ 0 (
    call :print_error "模型文件安装失败"
    call :cleanup_temp
    pause
    exit /b !errorlevel!
)

REM === 验证安装结果 ===
call :print_status "=== 验证安装结果 ==="

set "VERIFICATION_FAILED=0"

if exist "%OADIN_INSTALL_DIR%\oadin.exe" (
    call :print_success "OADIN 主程序安装成功"
) else (
    call :print_error "OADIN 主程序安装验证失败"
    set "VERIFICATION_FAILED=1"
)

if exist "%AI_SMARTVISION_INSTALL_DIR%" (
    call :print_success "AI SmartVision 安装成功"
) else (
    call :print_error "AI SmartVision 安装验证失败"
    set "VERIFICATION_FAILED=1"
)

if exist "%OLLAMA_DIR%" (
    call :print_success "Ollama 安装成功"
) else (
    call :print_error "Ollama 安装验证失败"
    set "VERIFICATION_FAILED=1"
)

if exist "%MODEL_DIR%" (
    call :print_success "模型文件安装成功"
) else (
    call :print_error "模型文件安装验证失败"
    set "VERIFICATION_FAILED=1"
)

REM === 清理临时文件 ===
call :cleanup_temp

REM === 完成 ===
if %VERIFICATION_FAILED% EQU 0 (
    call :print_success "========================================="
    call :print_success "所有组件安装完成！"
    call :print_success "========================================="
    call :print_status "安装详情:"
    call :print_status "  - OADIN: %OADIN_INSTALL_DIR%"
    call :print_status "  - AI SmartVision: %AI_SMARTVISION_INSTALL_DIR%"
    call :print_status "  - Ollama: %OLLAMA_DIR%"
    call :print_status "  - 模型文件: %MODEL_DIR%"
    call :print_status "========================================="
) else (
    call :print_error "安装过程中存在错误，请检查日志文件"
    call :print_error "日志位置: %LOG_FILE%"
)

echo.
call :print_status "按任意键退出..."
pause >nul
endlocal
