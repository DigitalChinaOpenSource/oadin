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

  # Add installation directory to the user's Path environment variable
  SetRegView 64
  ReadRegStr $0 HKCU "Environment" "Path"
  StrCmp $0 "" 0 +2
  StrCpy $0 "$0;"
  StrCpy $0 "$0$INSTDIR"
  WriteRegStr HKCU "Environment" "Path" "$0"
  SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment"

  # Post-install
  ExecWait '"$INSTDIR\postinstall.bat"'
SectionEnd