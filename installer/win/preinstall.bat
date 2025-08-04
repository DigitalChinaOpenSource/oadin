@echo off
echo Pre-install: i am running

setlocal

:: 定义当前版本，你需要根据实际情况修改
set "THIS_VERSION=1.3.12"

:: 定义 Oadin 文件夹路径 - 使用新的标准路径
set "OADIN_FOLDER=%LOCALAPPDATA%\Programs\Oadin"

echo --- Oadin 安装前置检查 ---

:: 尝试执行 oadin --version 并捕获输出
set "OADIN_OUTPUT="
oadin --version 2>nul > "%TEMP%\oadin_version_output.txt"
if exist "%TEMP%\oadin_version_output.txt" (
    for /f "delims=" %%i in ('type "%TEMP%\oadin_version_output.txt"') do (
        set "OADIN_OUTPUT=%%i"
    )
    del "%TEMP%\oadin_version_output.txt"
)

:: 检查 oadin 命令是否成功执行
oadin --version >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo 执行 "oadin --version" 时可能发生错误或命令未找到，忽略错误并继续安装。
    echo.
) else (
    echo 当前 Oadin 版本输出: "%OADIN_OUTPUT%"

    :: 检查输出是否等于定义的版本
    if /i "%OADIN_OUTPUT%" NEQ "%THIS_VERSION%" (
        echo Oadin 版本 "%OADIN_OUTPUT%" 与预期版本 "%THIS_VERSION%" 不匹配。
        echo 正在删除用户目录下的 Oadin 文件夹及其内容: "%OADIN_FOLDER%"

        :: 检查文件夹是否存在
        if exist "%OADIN_FOLDER%" (
            rmdir /s /q "%OADIN_FOLDER%"
            if %ERRORLEVEL% NEQ 0 (
                echo.
                echo 错误: 删除 Oadin 文件夹时发生错误。
                echo 请手动删除 "%OADIN_FOLDER%" 文件夹并重试安装。
                exit /b 1
            ) else (
                echo Oadin 文件夹已成功删除。
            )
        ) else (
            echo Oadin 文件夹 "%OADIN_FOLDER%" 不存在，无需删除。
        )
    ) else (
        echo Oadin 版本 "%OADIN_OUTPUT%" 与预期版本 "%THIS_VERSION%" 匹配，继续安装。
    )
)

echo --- 前置检查完成 ---

endlocal
exit /b 0