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
  StrCpy $1 "$INSTDIR"

  # Check if $INSTDIR is already in $0
  StrStr $0 $1 +2
  Goto AddPath

  # Skip adding if it already exists
  Goto EndPath

  AddPath:
  # If $0 is not empty, check if a semicolon needs to be added
  StrCmp $0 "" 0 +3
  StrLen $2 $0
  StrCpy $3 $0 -1 1
  StrCmp $3 ";" 0 +2
  StrCpy $0 "$0;"

  # Append the installation directory path
  StrCpy $0 "$0$INSTDIR"
  WriteRegStr HKCU "Environment" "Path" "$0"
  SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment"

  EndPath:
  # Post-install
  ExecWait '"$INSTDIR\postinstall.bat"'
SectionEnd