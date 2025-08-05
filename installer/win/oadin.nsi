!ifndef VERSION
  !define VERSION "0.0.0"
!endif
!define APP_NAME "Oadin CLI"
!define COMPANY_NAME "Digital China"
!define INSTALL_DIR "$LOCALAPPDATA\Programs\Oadin"

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

  nsExec::Exec '"$INSTDIR\preinstall.bat"'
  nsExec::Exec '"$INSTDIR\postinstall.bat" "$INSTDIR"'
  nsExec::Exec '"$INSTDIR\start-oadin.bat"'
SectionEnd