@echo off
echo Pre-install: i am running

setlocal

REM Define the expected version (modify as needed)
set "THIS_VERSION=1.3.12"

REM Define Oadin folder path
set "OADIN_FOLDER=%ProgramFiles%\Oadin"

echo --- Oadin Version Check ---

REM Step 2: Check Oadin version and cleanup if needed
echo.
echo [INFO] Checking Oadin version...
REM Attempt to run 'oadin --version' and capture output
REM Redirect stderr to stdout with 2>nul (separate capture file) then read it
REM NOTE: 'oadin --version' must output a plain version string like "1.3.12"
REM If the command fails the file may be empty and ERRORLEVEL will be set
set "OADIN_OUTPUT="
oadin --version 2>nul > "%TEMP%\oadin_version_output.txt"
if exist "%TEMP%\oadin_version_output.txt" (
    for /f "delims=" %%i in ('type "%TEMP%\oadin_version_output.txt"') do (
        set "OADIN_OUTPUT=%%i"
    )
    del "%TEMP%\oadin_version_output.txt"
)

REM Check whether command executed (any output captured)
REM This cannot distinguish between missing command vs unexpected output precisely
REM Rely on ERRORLEVEL to detect execution failure
oadin --version >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo WARNING: "oadin --version" may have failed or command not found. Continuing installation.
    echo.
) else (
    echo Current Oadin version output: "%OADIN_OUTPUT%"

    REM Compare output with expected version
    if /i "%OADIN_OUTPUT%" NEQ "%THIS_VERSION%" (
        echo Oadin version "%OADIN_OUTPUT%" does NOT match expected "%THIS_VERSION%".
        echo Deleting Oadin folder under Program Files: "%OADIN_FOLDER%"

        REM Check folder exists then delete
        if exist "%OADIN_FOLDER%" (
            REM Use rmdir /s /q to remove folder recursively (/s all subdirs, /q quiet)
            rmdir /s /q "%OADIN_FOLDER%"
            if %ERRORLEVEL% NEQ 0 (
                echo.
                echo ERROR: Failed to delete Oadin program folder.
                echo Please manually delete "%OADIN_FOLDER%" and retry installation.
                exit /b 1 REM Abort install due to failed cleanup
            ) else (
                echo Oadin program folder deleted successfully.
            )
        ) else (
            echo Oadin program folder "%OADIN_FOLDER%" does not exist, nothing to delete.
        )
    ) else (
        echo Oadin version "%OADIN_OUTPUT%" matches expected "%THIS_VERSION%". Proceeding.
    )
)

echo --- Pre-check Finished ---

endlocal
exit /b 0 REM Successful exit