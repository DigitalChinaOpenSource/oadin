!ifndef VERSION
  !define VERSION "0.0.0"
!endif
!define APP_NAME "Byze CLI"
!define COMPANY_NAME "Digital China"
!define INSTALL_DIR "$PROGRAMFILES64\${COMPANY_NAME}\Byze"

Outfile "..\byze-installer.exe"
InstallDir "${INSTALL_DIR}"
RequestExecutionLevel admin
SetCompress auto
SetCompressor lzma

Section "Install"
  SetOutPath "$INSTDIR"
  File "..\byze.exe"
  File "preinstall.bat"
  File "postinstall.bat"

  # Pre-install
  ExecWait '"$INSTDIR\preinstall.bat"'

  # 添加环境变量（示例）
  # WriteEnvStr HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "Path" "$%PATH%;$INSTDIR"

  # Post-install
  ExecWait '"$INSTDIR\postinstall.bat"'
SectionEnd