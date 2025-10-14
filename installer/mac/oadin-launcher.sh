#!/bin/bash

# 获取脚本目录
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
APP_NAME="oadin-app"

# 获取当前控制台用户
CURRENT_USER=$(stat -f "%Su" /dev/console)
USER_HOME=$(dscl . -read "/Users/$CURRENT_USER" NFSHomeDirectory | awk '{print $2}')

# 切换到用户目录
cd "$USER_HOME"

# 加载用户的完整 shell 环境
if [ -n "$ZSH_VERSION" ]; then
    # zsh
    [ -f ~/.zshenv ] && source ~/.zshenv
    [ -f ~/.zprofile ] && source ~/.zprofile
    [ -f ~/.zshrc ] && source ~/.zshrc
elif [ -n "$BASH_VERSION" ]; then
    # bash
    [ -f ~/.bash_profile ] && source ~/.bash_profile
    [ -f ~/.bashrc ] && source ~/.bashrc
fi

# 导出所有重要环境变量
export USER="$CURRENT_USER"
export HOME="$USER_HOME"
export LOGNAME="$CURRENT_USER"

# 执行主程序，继承当前完整环境
exec "$SCRIPT_DIR/$APP_NAME" "$@"