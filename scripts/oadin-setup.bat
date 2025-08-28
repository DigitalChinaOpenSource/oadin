@echo off
setlocal enabledelayedexpansion

REM ========================================
REM OADIN Auto Installation Script (Simple)
REM This script will install:
REM 1. OADIN Main Program
REM 2. AI SmartVision
REM 3. Ollama (IPEX-LLM version)
REM 4. Model Files
REM ========================================

REM === Configuration Section ===
set "SETUP_FILE=oadin-installer-latest.exe"
set "OADIN_INSTALL_DIR=%ProgramFiles%\Oadin"

set "AI_SMARTVISION_FILE=ai-smartvision-2.0.0-x64.exe"
set "AI_SMARTVISION_INSTALL_DIR=%ProgramFiles%\ai-smartvision"

set "OLLAMA_ZIP=ipex-llm-ollama.zip"
set "OLLAMA_DIR=%OADIN_INSTALL_DIR%\ipex-llm-ollama"

set "MODEL_ZIP=models.zip"
set "MODEL_DIR=%ProgramData%\Oadin\engine\ollama"

REM === Start Installation ===
echo ========================================
echo OADIN Auto Installation Script
echo Current Time: %date% %time%
echo ========================================

REM Check administrator privileges
echo [INFO] Checking administrator privileges...
openfiles >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Please run this script as administrator!
    pause
    exit /b 1
)
echo [OK] Administrator privileges verified

echo.
echo ========================================
echo Starting Installation Process
echo ========================================

REM === Step 1: Install OADIN Main Program ===
echo.
echo === Step 1: Install OADIN Main Program ===
echo [INFO] Checking installer: %SETUP_FILE%
if not exist "%SETUP_FILE%" (
    echo [ERROR] Installer not found: %SETUP_FILE%
    goto :install_failed
)
echo [OK] Installer check passed

echo [INFO] Starting OADIN installation...
"%SETUP_FILE%" /S "/D=%OADIN_INSTALL_DIR%"
if !errorlevel! NEQ 0 (
    echo [ERROR] OADIN installation failed, error code: !errorlevel!
    goto :install_failed
)
echo [OK] OADIN installation successful

REM === Step 2: Install AI SmartVision ===
echo.
echo === Step 2: Install AI SmartVision ===
echo [INFO] Checking installer: %AI_SMARTVISION_FILE%
if not exist "%AI_SMARTVISION_FILE%" (
    echo [ERROR] Installer not found: %AI_SMARTVISION_FILE%
    goto :install_failed
)
echo [OK] Installer check passed

echo [INFO] Starting AI SmartVision installation...
"%AI_SMARTVISION_FILE%" /S "/D=%AI_SMARTVISION_INSTALL_DIR%"
if !errorlevel! NEQ 0 (
    echo [ERROR] AI SmartVision installation failed, error code: !errorlevel!
    goto :install_failed
)
echo [OK] AI SmartVision installation successful

REM === Step 3: Extract Ollama ===
echo.
echo === Step 3: Install Ollama ===
echo [INFO] Checking archive: %OLLAMA_ZIP%
if not exist "%OLLAMA_ZIP%" (
    echo [ERROR] Archive not found: %OLLAMA_ZIP%
    goto :install_failed
)
echo [OK] Archive check passed

echo [INFO] Creating directory: %OLLAMA_DIR%
if not exist "%OLLAMA_DIR%" mkdir "%OLLAMA_DIR%"

echo [INFO] Starting Ollama extraction...
tar -xf "%OLLAMA_ZIP%" -C "%OLLAMA_DIR%" --strip-components=0
if !errorlevel! NEQ 0 (
    echo [ERROR] Ollama extraction failed, error code: !errorlevel!
    goto :install_failed
)
echo [OK] Ollama extraction successful

REM === Step 4: Extract Model Files ===
echo.
echo === Step 4: Install Model Files ===
echo [INFO] Checking archive: %MODEL_ZIP%
if not exist "%MODEL_ZIP%" (
    echo [ERROR] Archive not found: %MODEL_ZIP%
    goto :install_failed
)
echo [OK] Archive check passed

echo [INFO] Creating directory: %MODEL_DIR%
if not exist "%MODEL_DIR%" mkdir "%MODEL_DIR%"

echo [INFO] Starting model files extraction...
tar -xf "%MODEL_ZIP%" -C "%MODEL_DIR%" --strip-components=0
if !errorlevel! NEQ 0 (
    echo [ERROR] Model files extraction failed, error code: !errorlevel!
    goto :install_failed
)
echo [OK] Model files extraction successful

REM === Verify Installation Results ===
echo.
echo === Verify Installation Results ===
set "FAILED=0"

echo [INFO] Checking OADIN main program...
if exist "%OADIN_INSTALL_DIR%\oadin.exe" (
    echo [OK] OADIN main program installation successful
) else (
    echo [ERROR] OADIN main program not found
    set "FAILED=1"
)

echo [INFO] Checking AI SmartVision...
if exist "%AI_SMARTVISION_INSTALL_DIR%" (
    echo [OK] AI SmartVision installation successful
) else (
    echo [ERROR] AI SmartVision directory not found
    set "FAILED=1"
)

echo [INFO] Checking Ollama...
if exist "%OLLAMA_DIR%" (
    echo [OK] Ollama installation successful
) else (
    echo [ERROR] Ollama directory not found
    set "FAILED=1"
)

echo [INFO] Checking model files...
if exist "%MODEL_DIR%" (
    echo [OK] Model files installation successful
) else (
    echo [ERROR] Model files directory not found
    set "FAILED=1"
)

if %FAILED% NEQ 0 goto :install_failed

REM === Installation Successful ===
echo.
echo ========================================
echo [OK] All components installed successfully!
echo ========================================
echo Installation details:
echo   - OADIN: %OADIN_INSTALL_DIR%
echo   - AI SmartVision: %AI_SMARTVISION_INSTALL_DIR%
echo   - Ollama: %OLLAMA_DIR%
echo   - Model Files: %MODEL_DIR%
echo ========================================
echo.
echo Press any key to exit...
pause >nul
exit /b 0

REM === Installation Failed Handler ===
:install_failed
echo.
echo ========================================
echo [ERROR] An error occurred during installation!
echo ========================================
echo Please check the error messages above and try again.
echo.
echo Press any key to exit...
pause >nul
exit /b 1
