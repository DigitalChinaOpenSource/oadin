!ifndef VERSION
  !define VERSION "0.0.0"
!endif
!define APP_NAME "Oadin CLI"
!define COMPANY_NAME "Digital China"
!define INSTALL_DIR "$LOCALAPPDATA\Programs\Oadin"


Icon "oadin-icon.ico"
UninstallIcon "oadin-icon.ico"

Outfile "oadin-installer.exe"
InstallDir "${INSTALL_DIR}"
RequestExecutionLevel user
SetCompress auto
SetCompressor lzma

Section "Install"
  SetOutPath "$INSTDIR"
  
  File "oadin.exe"
  File /nonfatal "oadin-tray.exe"
  File "preinstall.bat"
  File "postinstall.bat"
  File "start-oadin.bat"


  File /nonfatal "oadin-icon.ico"

  nsExec::Exec '"$INSTDIR\preinstall.bat"'
  nsExec::Exec '"$INSTDIR\postinstall.bat" "$INSTDIR"'
  nsExec::Exec '"$INSTDIR\start-oadin.bat"'
  

  CreateDirectory "$SMPROGRAMS\Oadin"
  CreateShortCut "$SMPROGRAMS\Oadin\Oadin Tray.lnk" "$INSTDIR\oadin-tray.exe" "" "$INSTDIR\oadin-tray.exe" 0
  CreateShortCut "$SMPROGRAMS\Oadin\Oadin CLI.lnk" "$INSTDIR\oadin.exe" "" "$INSTDIR\oadin.exe" 0
SectionEnd