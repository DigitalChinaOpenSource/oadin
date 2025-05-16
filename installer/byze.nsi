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

  # demo to add env var
  # WriteEnvStr HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "Path" "$%PATH%;$INSTDIR"
  WriteEnvStr HKCU "Environment" "Path" "$%Path%;$INSTDIR"
  SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment"

  # Post-install
  ExecWait '"$INSTDIR\postinstall.bat"'
SectionEnd