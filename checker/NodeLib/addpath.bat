@echo off
chcp 65001 >nul  & rem 强制UTF-8编码
setlocal enabledelayedexpansion

:: 参数处理（支持任意格式路径）
set "raw_path=%~1"
if "%raw_path%"=="" (
    echo 错误：未提供路径参数
    exit /b 1
)

:: 路径标准化（确保无内层引号）
set "clean_path=%raw_path:"=%"
set "formatted_path="%clean_path%""

:: 从注册表获取PATH（兼容所有语言环境）
set "user_path="
for /f "tokens=2*" %%a in ('reg query "HKCU\Environment" /v "Path" 2^>nul ^| findstr /i "REG_"') do (
    set "user_path=%%b"
)

:: 检查路径是否已存在
set "exists=0"
if defined user_path (
    echo ;!user_path!; | findstr /i /L /C:";!formatted_path!;" >nul
    if !errorlevel! equ 0 set "exists=1"
)

:: 更新PATH变量
if !exists! equ 0 (
    if defined user_path (
        set "new_path=!user_path!;!formatted_path!"
    ) else (
        set "new_path=!formatted_path!"
    )

    :: 关键修复：SETX参数处理
    set "new_path=!new_path:"=!"
    setx Path "!new_path!" >nul
    if !errorlevel! neq 0 (
        echo [错误] SETX执行失败（代码：!errorlevel!）
        exit /b 1
    )
    echo 成功添加路径：%raw_path%
) else (
    echo 路径已存在：%raw_path%
)

endlocal