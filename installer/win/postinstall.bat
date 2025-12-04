@echo off

REM The installation directory is passed as the first argument
set "INSTALL_PATH=%~1"

REM Check if INSTALL_PATH is provided
if "%INSTALL_PATH%"=="" (
    echo Error: Installation path not provided to postinstall.bat
    exit /b 1
)

echo Adding Oadin to user PATH environment variable...

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
    REM Add a leading and trailing semicolon to make matching more robust
    set "PATH_TO_CHECK=;%CURRENT_PATH%;"
)

REM Normalize the INSTALL_PATH for consistent checking
set "NORMALIZED_INSTALL_PATH=%INSTALL_PATH%"
if not "%NORMALIZED_INSTALL_PATH:~-1%"=="\" (
    set "NORMALIZED_INSTALL_PATH=%NORMALIZED_INSTALL_PATH%\"
)

REM Check if the INSTALL_PATH already exists in the CURRENT_PATH (case-insensitive)
echo "%PATH_TO_CHECK%" | findstr /i /c:";%NORMALIZED_INSTALL_PATH%;" > nul
if %errorlevel% equ 0 (
    echo Oadin CLI path already exists in user's Path environment variable. Skipping.
) else (
    REM Add the path if it doesn't exist
    if not "%CURRENT_PATH%"=="" (
        if not "%CURRENT_PATH:~-1%"==";" (
            setx Path "%CURRENT_PATH%;%INSTALL_PATH%" > nul
        ) else (
            setx Path "%CURRENT_PATH%%INSTALL_PATH%" > nul
        )
    ) else (
        setx Path "%INSTALL_PATH%" > nul
    )

    echo Oadin CLI path added to user's Path environment variable.
)

REM Force a refresh of environment variables
call :RefreshEnv

echo Oadin installation completed successfully.
exit /b 0

:RefreshEnv
set WM_SETTINGCHANGE=0x001A
set HWND_BROADCAST=0xFFFF
powershell -Command "$signature = '[DllImport(\"user32.dll\", SetLastError = true)]public static extern IntPtr SendMessageTimeout(IntPtr hWnd, uint Msg, IntPtr wParam, string lParam, uint fuFlags, uint uTimeout, out IntPtr lpdwResult);'; $type = Add-Type -MemberDefinition $signature -Name 'Win32' -Namespace 'User32' -PassThru; [IntPtr]$lpdwResult = [IntPtr]::Zero; $null = $type::SendMessageTimeout([IntPtr]%HWND_BROADCAST%, %WM_SETTINGCHANGE%, [IntPtr]0, 'Environment', 0x0002, 1000, [ref]$lpdwResult);" > nul 2>&1
goto :eof