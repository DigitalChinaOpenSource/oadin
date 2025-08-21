!ifndef VERSION
  !define VERSION "0.0.0"
!endif
!define APP_NAME "Oadin CLI"
!define COMPANY_NAME "Digital China"
!define INSTALL_DIR "$PROGRAMFILES\Oadin"

Outfile "..\..\oadin-installer.exe"
InstallDir "${INSTALL_DIR}"
RequestExecutionLevel admin ; Admin rights required for Program Files installation
SetCompress auto
SetCompressor lzma

Section "Install"
  SetOutPath "$INSTDIR"
  File "..\..\oadin.exe"
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