@echo off
REM *****************************************************************************
REM Copyright 2025 Intel Corporation
REM
REM Licensed under the Apache License, Version 2.0 (the "License");
REM you may not use this file except in compliance with the License.
REM You may obtain a copy of the License at
REM
REM     http://www.apache.org/licenses/LICENSE-2.0
REM
REM Unless required by applicable law or agreed to in writing, software
REM distributed under the License is distributed on an "AS IS" BASIS,
REM WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
REM See the License for the specific language governing permissions and
REM limitations under the License.
REM *****************************************************************************

setlocal enabledelayedexpansion

echo [INFO] AOG Control Panel Frontend Build Script (Windows)

set "PROJECT_ROOT=%~dp0"
set "PROJECT_ROOT=%PROJECT_ROOT:~0,-1%"

echo [INFO] Project root: %PROJECT_ROOT%

echo [INFO] Checking if Node.js is installed...
call node --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Node.js not found, please install Node.js first
    echo [INFO] Download from: https://nodejs.org/
    pause
    exit /b 1
)

call npm --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] npm not found, please reinstall Node.js
    pause
    exit /b 1
)

echo [SUCCESS] Node.js and npm are available

echo [INFO] Checking if yarn is installed...
call yarn --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] yarn command not found, please install yarn first
    echo [INFO] You can install it with: npm install -g yarn
    pause
    exit /b 1
)

echo [SUCCESS] yarn check passed

set "FRONTEND_DIR=%PROJECT_ROOT%\frontend\control_panel"
set "CONSOLE_DIR=%PROJECT_ROOT%\console"

if not exist "%FRONTEND_DIR%" (
    echo [ERROR] Frontend directory not found: %FRONTEND_DIR%
    pause
    exit /b 1
)

if not exist "%CONSOLE_DIR%" (
    echo [ERROR] Console directory not found: %CONSOLE_DIR%
    pause
    exit /b 1
)

echo [SUCCESS] Directory structure check passed

echo [INFO] Entering frontend directory: %FRONTEND_DIR%
cd /d "%FRONTEND_DIR%"

:: ----------------------------------------------------------------------
:: Proxy injection for Yarn
:: ----------------------------------------------------------------------
set /p "USER_INPUT=[INFO] Do you want to set proxy values for Yarn using system environment variables? [Y/N]: "

if /i "%USER_INPUT%"=="Y" (
    echo [INFO] Checking for system proxy environment variables...
    set "HAS_PROXY_CONFIG=false"

    if defined http_proxy (
        echo [INFO] Detected http_proxy: %http_proxy%
        set "HAS_PROXY_CONFIG=true"
    )
    if defined https_proxy (
        echo [INFO] Detected https_proxy: %https_proxy%
        set "HAS_PROXY_CONFIG=true"
    )

    if "!HAS_PROXY_CONFIG!"=="true" (
        if defined http_proxy (
            call yarn config set httpProxy "%http_proxy%"
        )
        if defined https_proxy (
            call yarn config set httpsProxy "%https_proxy%"
        )
    ) else (
        echo [INFO] No proxy environment variables found, skipping Yarn proxy config.
    )
    goto AfterProxy
)

if /i "%USER_INPUT%"=="N" (
    echo [INFO] Skipping Yarn proxy config.
    goto AfterProxy
)

echo [WARN] Invalid input. Please enter Y or N next time.

:AfterProxy


echo [INFO] Installing frontend dependencies...
call yarn install

if errorlevel 1 (
    echo [ERROR] Failed to install dependencies
    pause
    exit /b 1
)

echo [SUCCESS] Dependencies installed successfully

echo [INFO] Building frontend...
call yarn build

if errorlevel 1 (
    echo [ERROR] Frontend build failed
    pause
    exit /b 1
)

echo [SUCCESS] Frontend build completed

set "DIST_DIR=%FRONTEND_DIR%\dist"
if not exist "%DIST_DIR%" (
    echo [ERROR] Build output directory not found: %DIST_DIR%
    pause
    exit /b 1
)

set "CONSOLE_DIST_DIR=%CONSOLE_DIR%\dist"
if exist "%CONSOLE_DIST_DIR%" (
    echo [INFO] Cleaning existing dist directory...
    rmdir /s /q "%CONSOLE_DIST_DIR%" 2>nul
)

echo [INFO] Deploying build artifacts to console directory...
move "%DIST_DIR%" "%CONSOLE_DIST_DIR%" >nul

if errorlevel 1 (
    echo [ERROR] Deployment failed
    pause
    exit /b 1
)

echo [SUCCESS] Deployment completed

if exist "%CONSOLE_DIST_DIR%\index.html" (
    echo [SUCCESS] Verification passed: index.html file exists
) else (
    echo [ERROR] Verification failed: index.html file not found
    pause
    exit /b 1
)

echo [SUCCESS] Control Panel frontend build and deployment completed!
echo [INFO] You can now start AOG service and visit http://127.0.0.1:16688/dashboard

pause
