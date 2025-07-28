!ifndef VERSION
  !define VERSION "0.0.0"
!endif
!define APP_NAME "Oadin CLI"
!define COMPANY_NAME "Digital China"
!define INSTALL_DIR "$PROFILE\Oadin"

Outfile "..\..\oadin-installer.exe"
InstallDir "${INSTALL_DIR}"
RequestExecutionLevel user ; Install to user's profile, so no admin rights needed by default
SetCompress auto
SetCompressor lzma

Section "Install"
  SetOutPath "$INSTDIR"
  File "oadin.exe"
  File "oadin-tray.exe"
  File "preinstall.bat"
  File "postinstall.bat"
  File "start-oadin.bat"

  # Pre-install silently
  nsExec::Exec '"$INSTDIR\preinstall.bat"'

  # Post-install silently with argument
  nsExec::Exec '"$INSTDIR\postinstall.bat" "$INSTDIR"'

  # start oadin server
  nsExec::Exec '"$INSTDIR\start-oadin.bat"'
SectionEnd