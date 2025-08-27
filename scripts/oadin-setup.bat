@echo off
setlocal enabledelayedexpansion

REM ========================================
REM OADIN Auto Installation Script (No PowerShell)
REM ========================================

REM === Configuration section ===
set "SETUP_FILE=oadin-installer-test-2.0.40.exe"
REM Ensure complete Program Files path is used to prevent path parsing errors
if defined ProgramFiles (
    set "OADIN_INSTALL_DIR=%ProgramFiles%\Oadin"
    set "AI_SMARTVISION_INSTALL_DIR=%ProgramFiles%\ai-smartvision"
) else (
    set "OADIN_INSTALL_DIR=C:\Program Files\Oadin"
    set "AI_SMARTVISION_INSTALL_DIR=C:\Program Files\ai-smartvision"
)
set "SILENT_ARGS=/S"
set "LOG_FILE=%OADIN_INSTALL_DIR%\install.log"

set "AI_SMARTVISION_FILE=ai-smartvision-2.0.0-x64.exe"
set "AI_SMARTVISION_LOG=%OADIN_INSTALL_DIR%\ai-smartvision-install.log"

set "OLLAMA_ZIP=ipex-llm-ollama.zip"
set "OLLAMA_DIR=%OADIN_INSTALL_DIR%\ipex-llm-ollama"

set "MODEL_ZIP=models.zip"
set "MODEL_DIR=%ProgramData%\Oadin\engine\ollama"

set "TEMP_EXTRACT_DIR=%TEMP%\oadin_extract_%RANDOM%"

REM === Color definitions ===
REM Simplified to avoid special character issues
set "GREEN_PREFIX=[SUCCESS]"
set "RED_PREFIX=[ERROR]" 
set "YELLOW_PREFIX=[WARNING]"
set "BLUE_PREFIX=[INFO]"

REM === Jump to main function ===
goto :main

REM === Function definitions ===
:print_status
echo %BLUE_PREFIX% %~1
goto :eof

:print_success
echo %GREEN_PREFIX% %~1
goto :eof

:print_error
echo %RED_PREFIX% %~1
goto :eof

:print_warning
echo %YELLOW_PREFIX% %~1
goto :eof

:create_directory
if not exist "%~1" (
    call :print_status "Creating directory: %~1"
    mkdir "%~1" 2>nul
    if !errorlevel! NEQ 0 (
        call :print_error "Failed to create directory: %~1"
        exit /b 1
    )
    call :print_success "Directory created successfully"
) else (
    call :print_status "Directory already exists: %~1"
)
goto :eof

:validate_paths
REM === Validate path settings ===
call :print_status "Validating path settings..."

REM Check if OADIN_INSTALL_DIR contains complete Program Files path
echo "%OADIN_INSTALL_DIR%" | findstr /C:"Program Files" >nul
if %errorlevel% NEQ 0 (
    call :print_error "Error: OADIN_INSTALL_DIR path is incorrect: %OADIN_INSTALL_DIR%"
    call :print_error "Should contain 'Program Files'"
    exit /b 1
)

REM Check path length to ensure it's not truncated
set "PATH_LENGTH=0"
set "TEST_PATH=%OADIN_INSTALL_DIR%"
:count_loop
if defined TEST_PATH (
    set /a PATH_LENGTH+=1
    set "TEST_PATH=%TEST_PATH:~1%"
    goto count_loop
)

if %PATH_LENGTH% LSS 10 (
    call :print_error "Error: Path too short, might be truncated: %OADIN_INSTALL_DIR%"
    exit /b 1
)

call :print_success "Path validation passed"
goto :eof

:check_file_exists
if not exist "%~1" (
    call :print_error "File does not exist: %~1"
    exit /b 1
)
call :print_success "File check passed: %~1"
goto :eof

:extract_zip_native
set "ZIP_FILE=%~1"
set "DEST_DIR=%~2"
set "EXTRACT_TEMP=%TEMP_EXTRACT_DIR%\%RANDOM%"

call :print_status "Starting extraction of: %ZIP_FILE%"
call :print_status "Target directory: %DEST_DIR%"

REM Create temporary extraction directory
call :create_directory "%EXTRACT_TEMP%"

REM Use system built-in extraction tool (Windows 10+)
if exist "%SystemRoot%\System32\tar.exe" (
    call :print_status "Using tar.exe for extraction..."
    "%SystemRoot%\System32\tar.exe" -xf "%ZIP_FILE%" -C "%EXTRACT_TEMP%" >nul 2>&1
    if !errorlevel! EQU 0 (
        call :move_extracted_files "%EXTRACT_TEMP%" "%DEST_DIR%"
        if !errorlevel! EQU 0 (
            call :print_success "tar.exe extraction successful"
            goto :extract_cleanup
        )
    )
    call :print_warning "tar.exe extraction failed, trying other methods..."
)

REM Use VBScript extraction
call :print_status "Using VBScript for extraction..."
call :extract_zip_vbs "%ZIP_FILE%" "%EXTRACT_TEMP%"
if !errorlevel! EQU 0 (
    call :move_extracted_files "%EXTRACT_TEMP%" "%DEST_DIR%"
    if !errorlevel! EQU 0 (
        call :print_success "VBScript extraction successful"
        goto :extract_cleanup
    )
)

REM Use 7-Zip (if installed)
call :print_status "Trying 7-Zip extraction..."
call :extract_zip_7zip "%ZIP_FILE%" "%DEST_DIR%"
if !errorlevel! EQU 0 (
    call :print_success "7-Zip extraction successful"
    goto :extract_cleanup
)

call :print_error "All extraction methods failed"
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

REM Move all files and folders
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

REM Create VBScript extraction script
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

REM Check common 7-Zip installation locations
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

call :print_status "Starting installation of %PROGRAM_NAME%..."
call :check_file_exists "%INSTALLER%"
if !errorlevel! NEQ 0 exit /b !errorlevel!

REM Debug output
call :print_status "Executing command: %INSTALLER% %INSTALL_ARGS%"

if not "%LOG_FILE_PARAM%"=="" (
    "%INSTALLER%" %INSTALL_ARGS% /log="%LOG_FILE_PARAM%"
) else (
    "%INSTALLER%" %INSTALL_ARGS%
)

if !errorlevel! EQU 0 (
    call :print_success "%PROGRAM_NAME% installation completed!"
) else (
    call :print_error "%PROGRAM_NAME% installation failed, error code: !errorlevel!"
    exit /b !errorlevel!
)
goto :eof

:install_ai_smartvision
set "INSTALLER=%~1"
set "INSTALL_DIR=%~2"
set "LOG_FILE_PARAM=%~3"

call :print_status "Starting AI SmartVision installation..."
call :check_file_exists "%INSTALLER%"
if !errorlevel! NEQ 0 exit /b !errorlevel!

call :print_status "Installation path: %INSTALL_DIR%"
call :print_status "Log file: %LOG_FILE_PARAM%"

REM Use temporary batch file to avoid quote issues
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
    call :print_success "AI SmartVision installation completed!"
) else (
    call :print_error "AI SmartVision installation failed, error code: %INSTALL_RESULT%"
    exit /b %INSTALL_RESULT%
)
goto :eof

:cleanup_temp
if exist "%TEMP_EXTRACT_DIR%" (
    call :print_status "Cleaning up temporary files..."
    rmdir /s /q "%TEMP_EXTRACT_DIR%" >nul 2>&1
)
goto :eof

:main
REM === Script start ===
call :print_status "OADIN Auto Installation Script Starting"
call :print_status "Current time: %date% %time%"

REM === Debug environment variables ===
call :print_status "Debug information - Environment variables:"
call :print_status "  ProgramFiles = %ProgramFiles%"
call :print_status "  ProgramFiles(x86) = %ProgramFiles(x86)%"
call :print_status "  OADIN_INSTALL_DIR = %OADIN_INSTALL_DIR%"
call :print_status "  AI_SMARTVISION_INSTALL_DIR = %AI_SMARTVISION_INSTALL_DIR%"

REM === Check administrator privileges ===
call :print_status "Checking administrator privileges..."
openfiles >nul 2>&1
if errorlevel 1 (
    call :print_error "Please run this script as administrator!"
    pause
    exit /b 1
)
call :print_success "Administrator privileges verified"

REM === Validate path configuration ===
call :validate_paths
if !errorlevel! NEQ 0 (
    pause
    exit /b !errorlevel!
)

REM === Display installation directory configuration ===
call :print_status "========================================="
call :print_status "Installation directory configuration:"
call :print_status "  - OADIN Main Program: %OADIN_INSTALL_DIR%"
call :print_status "  - AI SmartVision: %AI_SMARTVISION_INSTALL_DIR%"
call :print_status "  - Ollama: %OLLAMA_DIR%"
call :print_status "  - Model Data: %MODEL_DIR%"
call :print_status "  - Log File: %LOG_FILE%"
call :print_status "========================================="

REM === Create temporary directory ===
call :create_directory "%TEMP_EXTRACT_DIR%"

REM === Install OADIN main program ===
call :print_status "=== Step 1: Install OADIN Main Program ==="
call :print_status "Target installation directory: %OADIN_INSTALL_DIR%"
REM Ensure path is properly quoted to avoid space issues
call :install_program "%SETUP_FILE%" "/S /D=%OADIN_INSTALL_DIR%" "%LOG_FILE%" "OADIN"
if !errorlevel! NEQ 0 (
    call :cleanup_temp
    pause
    exit /b !errorlevel!
)

REM === Install AI SmartVision ===
call :print_status "=== Step 2: Install AI SmartVision ==="
call :install_ai_smartvision "%AI_SMARTVISION_FILE%" "%AI_SMARTVISION_INSTALL_DIR%" "%AI_SMARTVISION_LOG%"
if !errorlevel! NEQ 0 (
    call :cleanup_temp
    pause
    exit /b !errorlevel!
)

REM === Extract Ollama package ===
call :print_status "=== Step 3: Install Ollama ==="
call :extract_zip_native "%OLLAMA_ZIP%" "%OLLAMA_DIR%"
if !errorlevel! NEQ 0 (
    call :print_error "Ollama installation failed"
    call :cleanup_temp
    pause
    exit /b !errorlevel!
)

REM === Extract model package ===
call :print_status "=== Step 4: Install Model Files ==="
call :extract_zip_native "%MODEL_ZIP%" "%MODEL_DIR%"
if !errorlevel! NEQ 0 (
    call :print_error "Model files installation failed"
    call :cleanup_temp
    pause
    exit /b !errorlevel!
)

REM === Verify installation results ===
call :print_status "=== Verifying Installation Results ==="

set "VERIFICATION_FAILED=0"

if exist "%OADIN_INSTALL_DIR%\oadin.exe" (
    call :print_success "OADIN main program installation successful"
) else (
    call :print_error "OADIN main program installation verification failed"
    set "VERIFICATION_FAILED=1"
)

if exist "%AI_SMARTVISION_INSTALL_DIR%" (
    call :print_success "AI SmartVision installation successful"
) else (
    call :print_error "AI SmartVision installation verification failed"
    set "VERIFICATION_FAILED=1"
)

if exist "%OLLAMA_DIR%" (
    call :print_success "Ollama installation successful"
) else (
    call :print_error "Ollama installation verification failed"
    set "VERIFICATION_FAILED=1"
)

if exist "%MODEL_DIR%" (
    call :print_success "Model files installation successful"
) else (
    call :print_error "Model files installation verification failed"
    set "VERIFICATION_FAILED=1"
)

REM === Clean up temporary files ===
call :cleanup_temp

REM === Complete ===
if %VERIFICATION_FAILED% EQU 0 (
    call :print_success "========================================="
    call :print_success "All components installation completed!"
    call :print_success "========================================="
    call :print_status "Installation details:"
    call :print_status "  - OADIN: %OADIN_INSTALL_DIR%"
    call :print_status "  - AI SmartVision: %AI_SMARTVISION_INSTALL_DIR%"
    call :print_status "  - Ollama: %OLLAMA_DIR%"
    call :print_status "  - Model Files: %MODEL_DIR%"
    call :print_status "========================================="
) else (
    call :print_error "Errors occurred during installation, please check log files"
    call :print_error "Log location: %LOG_FILE%"
)

echo.
call :print_status "Press any key to exit..."
pause >nul
endlocal
