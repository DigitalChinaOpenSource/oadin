@echo off
echo Pre-install: i am running

setlocal

:: 定义当前版本，你需要根据实际情况修改
set "THIS_VERSION=1.3.12"

:: 定义 Oadin 文件夹路径
:: %ProgramFiles% 环境变量指向 Program Files 目录
set "OADIN_FOLDER=%ProgramFiles%\Oadin"
:: 定义数据文件夹路径
set "OADIN_DATA_FOLDER=%PROGRAMDATA%\Oadin"

echo --- Oadin 安装前置检查 ---

:: 尝试执行 oadin --version 并捕获输出
:: 使用 2>&1 将标准错误重定向到标准输出，然后用 findstr 检查输出
:: 注意：oadin --version 的输出格式必须是明确的版本号，例如 "1.3.12"
:: 如果 oadin --version 报错，它的错误信息会通过 findstr 管道，findstr 不会找到版本号，ERRORLEVEL 会被设置
set "OADIN_OUTPUT="
oadin --version 2>nul > "%TEMP%\oadin_version_output.txt"
if exist "%TEMP%\oadin_version_output.txt" (
    for /f "delims=" %%i in ('type "%TEMP%\oadin_version_output.txt"') do (
        set "OADIN_OUTPUT=%%i"
    )
    del "%TEMP%\oadin_version_output.txt"
)

:: 检查 oadin 命令是否成功执行（即是否有输出）
:: 这种方法无法精确判断是命令不存在还是命令执行了但输出不符合预期
:: 我们依赖 ERRORLEVEL 来判断命令是否“报错”
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
        echo 正在删除 Program Files 下的 Oadin 文件夹及其内容: "%OADIN_FOLDER%"
        echo 正在删除 ProgramData 下的 Oadin 数据文件夹: "%OADIN_DATA_FOLDER%"

        :: 检查程序文件夹是否存在并删除
        if exist "%OADIN_FOLDER%" (
            :: 使用 rmdir /s /q 删除文件夹及其内容
            :: /s 表示删除目录和所有子目录和文件
            :: /q 表示静默模式，不提示确认
            rmdir /s /q "%OADIN_FOLDER%"
            if %ERRORLEVEL% NEQ 0 (
                echo.
                echo 错误: 删除 Oadin 程序文件夹时发生错误。
                echo 请手动删除 "%OADIN_FOLDER%" 文件夹并重试安装。
                exit /b 1 :: 退出脚本，表示删除失败，阻止安装
            ) else (
                echo Oadin 程序文件夹已成功删除。
            )
        ) else (
            echo Oadin 程序文件夹 "%OADIN_FOLDER%" 不存在，无需删除。
        )

        :: 检查数据文件夹是否存在并删除
        if exist "%OADIN_DATA_FOLDER%" (
            rmdir /s /q "%OADIN_DATA_FOLDER%"
            if %ERRORLEVEL% NEQ 0 (
                echo.
                echo 警告: 删除 Oadin 数据文件夹时发生错误。
                echo 数据文件夹: "%OADIN_DATA_FOLDER%"
                echo 这不会阻止安装，但可能存在旧数据残留。
            ) else (
                echo Oadin 数据文件夹已成功删除。
            )
        ) else (
            echo Oadin 数据文件夹 "%OADIN_DATA_FOLDER%" 不存在，无需删除。
        )
    ) else (
        echo Oadin 版本 "%OADIN_OUTPUT%" 与预期版本 "%THIS_VERSION%" 匹配，继续安装。
    )
)

echo --- 前置检查完成 ---

endlocal
exit /b 0 :: 成功退出