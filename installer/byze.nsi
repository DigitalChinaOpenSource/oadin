!ifndef VERSION
  !define VERSION "0.0.0"
!endif
!define APP_NAME "Byze CLI"
!define COMPANY_NAME "Digital China"
!define INSTALL_DIR "$PROFILE\Byze"

Outfile "..\byze-installer.exe"
InstallDir "${INSTALL_DIR}"
RequestExecutionLevel user ; Install to user's profile, so no admin rights needed by default
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
  ReadRegStr $0 HKCU "Environment" "Path"

  # 如果当前 Path 不为空，添加分号作为分隔符
  StrCmp $0 "" 0 +2
    StrCpy $0 "$0;"

  # 直接追加安装目录路径
  StrCpy $0 "$0$INSTDIR"
  WriteRegStr HKCU "Environment" "Path" "$0"
  SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment"

  # Post-install
  ExecWait '"$INSTDIR\postinstall.bat"'
SectionEnd