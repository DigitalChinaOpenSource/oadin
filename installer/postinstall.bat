@echo off

REM The installation directory is passed as the first argument
set "INSTALL_PATH=%~1"

REM Check if INSTALL_PATH is provided
if "%INSTALL_PATH%"=="" (
    echo Error: Installation path not provided to postinstall.bat
    exit /b 1
)

REM Get the current user's Path variable
for /f "tokens=*" %%i in ('reg query "HKCU\Environment" /v Path ^| findstr /i "Path"') do (
    set "CURRENT_PATH=%%i"
)

REM Extract the actual path value (remove "    Path    REG_SZ    ")
set "CURRENT_PATH=%CURRENT_PATH:*REG_SZ    =%"

REM Ensure the CURRENT_PATH is not empty before checking for trailing semicolon
if "%CURRENT_PATH%"=="" (
    set "PATH_TO_CHECK="
) else (
    REM Add a leading and trailing semicolon to make matching more robust (e.g., ;path; or ;path)
    set "PATH_TO_CHECK=;%CURRENT_PATH%;"
)

REM Normalize the INSTALL_PATH for consistent checking (add trailing backslash if not present)
REM This helps in cases where the path might be 'C:\byze' or 'C:\byze\'
set "NORMALIZED_INSTALL_PATH=%INSTALL_PATH%"
if not "%NORMALIZED_INSTALL_PATH:~-1%"=="\" (
    set "NORMALIZED_INSTALL_PATH=%NORMALIZED_INSTALL_PATH%\"
)

REM Check if the INSTALL_PATH already exists in the CURRENT_PATH (case-insensitive)
echo "%PATH_TO_CHECK%" | findstr /i /c:";%NORMALIZED_INSTALL_PATH%;" > nul
if %errorlevel% equ 0 (
    echo Byze CLI path already exists in user's Path environment variable. Skipping.
) else (
    REM Add the path if it doesn't exist
    REM Prepend with a semicolon if CURRENT_PATH is not empty and doesn't already end with a semicolon
    if not "%CURRENT_PATH%"=="" (
        if not "%CURRENT_PATH:~-1%"==";" (
            setx Path "%CURRENT_PATH%;%INSTALL_PATH%;"
        ) else (
            setx Path "%CURRENT_PATH%%INSTALL_PATH%;"
        )
    )

    echo Byze CLI path added to user's Path environment variable.
)

REM Force a refresh of environment variables for currently running applications
REM This is equivalent to SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment" in NSIS
call :RefreshEnv

exit /b 0

:RefreshEnv
REM The WM_SETTINGCHANGE message with "Environment" string should trigger a refresh.
REM No need to query Path again here as we just set it.
set WM_SETTINGCHANGE=0x001A
set HWND_BROADCAST=0xFFFF
powershell -Command "$signature = '[DllImport(\"user32.dll\", SetLastError = true)]public static extern IntPtr SendMessageTimeout(IntPtr hWnd, uint Msg, IntPtr wParam, string lParam, uint fuFlags, uint uTimeout, out IntPtr lpdwResult);'; $type = Add-Type -MemberDefinition $signature -Name 'Win32' -Namespace 'User32' -PassThru; [IntPtr]$lpdwResult = [IntPtr]::Zero; $null = $type::SendMessageTimeout([IntPtr]%HWND_BROADCAST%, %WM_SETTINGCHANGE%, [IntPtr]0, 'Environment', 0x0002, 1000, [ref]$lpdwResult);"
goto :eof