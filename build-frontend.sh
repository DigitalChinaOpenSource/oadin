#!/bin/bash
#*****************************************************************************
# Copyright 2025 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#*****************************************************************************


# AOG Control Panel Frontend Build Script
# This script automates frontend build and deployment to console directory

set -e  # Exit on error

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if command exists
check_command() {
    if ! command -v $1 &> /dev/null; then
        print_error "$1 command not found, please install $1 first"
        exit 1
    fi
}

# 获取脚本所在目录（项目根目录）
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR"

print_info "AOG Control Panel Frontend Build Script"
print_info "Project root: $PROJECT_ROOT"

# Check required commands
print_info "Checking required dependencies..."
check_command "yarn"

# Check directory structure
FRONTEND_DIR="$PROJECT_ROOT/frontend/control_panel"
CONSOLE_DIR="$PROJECT_ROOT/console"

if [ ! -d "$FRONTEND_DIR" ]; then
    print_error "Frontend directory not found: $FRONTEND_DIR"
    exit 1
fi

if [ ! -d "$CONSOLE_DIR" ]; then
    print_error "Console directory not found: $CONSOLE_DIR"
    exit 1
fi

print_success "Directory structure check passed"

# Enter frontend directory
print_info "Entering frontend directory: $FRONTEND_DIR"
cd "$FRONTEND_DIR"

# ----------------------------------------------------------------------
# Proxy injection for Yarn
# ----------------------------------------------------------------------
read -p "[INFO] Do you want to set proxy values for Yarn using system environment variables? [Y/N]: " USER_INPUT
if [[ "$USER_INPUT" == "Y" || "$USER_INPUT" == "y" ]]; then
    print_info "Checking for system proxy environment variables..."
    HAS_PROXY_CONFIG=false

    if [ ! -z "$http_proxy" ]; then
        print_info "Detected http_proxy: $http_proxy"
        yarn config set httpProxy "$http_proxy"
        HAS_PROXY_CONFIG=true
    fi

    if [ ! -z "$https_proxy" ]; then
        print_info "Detected https_proxy: $https_proxy"
        yarn config set httpsProxy "$https_proxy"
        HAS_PROXY_CONFIG=true
    fi

    if [ "$HAS_PROXY_CONFIG" = false ]; then
        print_info "No proxy environment variables found, skipping Yarn proxy config."
    fi
elif [[ "$USER_INPUT" == "N" || "$USER_INPUT" == "n" ]]; then
    print_info "Skipping Yarn proxy config."
else
    print_warning "Invalid input. Please enter Y or N next time."
fi

# Install dependencies
print_info "Installing frontend dependencies..."
yarn install

if [ $? -ne 0 ]; then
    print_error "Failed to install dependencies"
    exit 1
fi

print_success "Dependencies installed successfully"

# Build frontend
print_info "Building frontend..."
yarn build

if [ $? -ne 0 ]; then
    print_error "Frontend build failed"
    exit 1
fi

print_success "Frontend build completed"

# Check build artifacts
DIST_DIR="$FRONTEND_DIR/dist"
if [ ! -d "$DIST_DIR" ]; then
    print_error "Build output directory not found: $DIST_DIR"
    exit 1
fi

# Clean existing dist directory (if exists)
CONSOLE_DIST_DIR="$CONSOLE_DIR/dist"
if [ -d "$CONSOLE_DIST_DIR" ]; then
    print_info "Cleaning existing dist directory..."
    rm -rf "$CONSOLE_DIST_DIR"
fi

# Move build artifacts to console directory
print_info "Deploying build artifacts to console directory..."
mv "$DIST_DIR" "$CONSOLE_DIST_DIR"

if [ $? -ne 0 ]; then
    print_error "Deployment failed"
    exit 1
fi

print_success "Deployment completed"

# Verify deployment result
if [ -f "$CONSOLE_DIST_DIR/index.html" ]; then
    print_success "Verification passed: index.html file exists"
else
    print_error "Verification failed: index.html file not found"
    exit 1
fi

print_success "🎉 Control Panel frontend build and deployment completed!"
print_info "You can now start AOG service and visit http://127.0.0.1:16688/dashboard"

