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
  StrCpy $1 "$INSTDIR"

  # 检查 $INSTDIR 是否已经存在于 $0 中
  StrStr $0 $1 +2
  Goto AddPath

  # 如果已存在，跳过追加
  Goto EndPath

  AddPath:
  # 如果 $0 不为空，检查是否需要添加分号
  StrCmp $0 "" 0 +3
  StrLen $2 $0
  StrCpy $3 $0 -1 1
  StrCmp $3 ";" 0 +2
  StrCpy $0 "$0;"

  # 追加安装目录路径
  StrCpy $0 "$0$INSTDIR"
  WriteRegStr HKCU "Environment" "Path" "$0"
  SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment"

  EndPath:
  # Post-install
  ExecWait '"$INSTDIR\postinstall.bat"'
SectionEnd