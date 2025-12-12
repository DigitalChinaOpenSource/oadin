#!/bin/bash
set -e

APP_NAME="Oadin"
# "/Applications/Oadin.app/Contents/MacOS"
# "/Applications/Oadin.app/Contents/Resources"
APP_BUNDLE="$(pwd)/${APP_NAME}.app"
CONTENTS_DIR="${APP_BUNDLE}/Contents"
MACOS_DIR="${CONTENTS_DIR}/MacOS"
RESOURCES_DIR="${CONTENTS_DIR}/Resources"

echo "Creating Oadin.app bundle..."

# 清理旧的应用包
rm -rf "${APP_BUNDLE}"

# 创建应用包结构
mkdir -p "${MACOS_DIR}"
mkdir -p "${RESOURCES_DIR}"

# 检查文件是否存在
if [ ! -f "$(pwd)/oadin" ]; then
    echo "Error: oadin executable not found"
    exit 1
fi

if [ ! -f "$(pwd)/oadin-app" ]; then
    echo "Error: oadin-app executable not found"
    exit 1
fi

# 复制可执行文件
cp $(pwd)/oadin "${RESOURCES_DIR}/"
cp $(pwd)/oadin-app "${MACOS_DIR}/"

# 复制应用图标（如果存在）
if [ -f "tray/icon/oadin-icon.icns" ]; then
    cp "tray/icon/oadin-icon.icns" "${RESOURCES_DIR}/AppIcon.icns"
    echo "✅ Added application icon"
elif [ -f "tray/icon/oadin-icon.png" ]; then
    # 如果没有icns文件，可以从png转换
    cp "tray/icon/oadin-icon.png" "${RESOURCES_DIR}/AppIcon.png"
    echo "✅ Added PNG application icon"
fi

# 创建 Info.plist
cat > "Oadin.app/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleDevelopmentRegion</key>
    <string>en</string>
    <key>CFBundleDisplayName</key>
    <string>Oadin</string>
    <key>CFBundleExecutable</key>
    <string>oadin-app</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon.icns</string>
    <key>CFBundleIdentifier</key>
    <string>com.digitalchina.oadin</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
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
    <key>CFBundleShortVersionString</key>
    <string>${CI_COMMIT_TAG}</string>
    <key>CFBundleVersion</key>
    <string>${CI_COMMIT_TAG}</string>
    <key>LSMinimumSystemVersion</key>
    <string>11.0</string>
    <key>LSUIElement</key>
    <false/>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSHumanReadableCopyright</key>
    <string>Copyright © 2024 Digital China. All rights reserved.</string>
    <key>LSEnvironment</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin</string>
    </dict>
</dict>
</plist>
EOF

# 设置权限
chmod +x "${RESOURCES_DIR}/oadin"
chmod +x "${MACOS_DIR}/oadin-app"
# chown -R root:wheel "${APP_BUNDLE}"

echo "Oadin.app created successfully at ${APP_BUNDLE}"
