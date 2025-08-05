#!/bin/bash
set -e

APP_NAME="Oadin"
APP_BUNDLE="/Applications/${APP_NAME}.app"
CONTENTS_DIR="${APP_BUNDLE}/Contents"
MACOS_DIR="${CONTENTS_DIR}/MacOS"
RESOURCES_DIR="${CONTENTS_DIR}/Resources"

echo "Creating Oadin.app bundle..."

# 清理旧的应用包
sudo rm -rf "${APP_BUNDLE}"

# 创建应用包结构
sudo mkdir -p "${MACOS_DIR}"
sudo mkdir -p "${RESOURCES_DIR}"

# 检查文件是否存在
if [ ! -f "oadin" ]; then
    echo "Error: oadin executable not found"
    exit 1
fi

if [ ! -f "oadin-tray" ]; then
    echo "Error: oadin-tray executable not found"
    exit 1
fi

# 复制可执行文件
sudo cp oadin "${MACOS_DIR}/"
sudo cp oadin-tray "${MACOS_DIR}/"

# 创建 Info.plist
sudo cat > "${CONTENTS_DIR}/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>oadin-tray</string>
    <key>CFBundleIdentifier</key>
    <string>com.digitalchina.oadin</string>
    <key>CFBundleName</key>
    <string>Oadin</string>
    <key>CFBundleDisplayName</key>
    <string>Oadin AI Service Manager</string>
    <key>CFBundleVersion</key>
    <string>1.0</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>LSUIElement</key>
    <true/>
</dict>
</plist>
EOF

# 设置权限
sudo chmod +x "${MACOS_DIR}/oadin"
sudo chmod +x "${MACOS_DIR}/oadin-tray"
sudo chown -R root:wheel "${APP_BUNDLE}"

echo "Oadin.app created successfully at ${APP_BUNDLE}"
