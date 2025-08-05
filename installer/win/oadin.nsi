!ifndef VERSION
  !define VERSION "0.0.0"
!endif
!define APP_NAME "Oadin CLI"
!define COMPANY_NAME "Digital China"
!define INSTALL_DIR "$LOCALAPPDATA\Programs\Oadin"

Outfile "..\..\oadin-installer.exe"
InstallDir "${INSTALL_DIR}"
RequestExecutionLevel user
SetCompress auto
SetCompressor lzma

Section "Install"
  SetOutPath "$INSTDIR"
  
  ; 必需文件
  File "..\..\oadin.exe"
  
  ; 可选文件 - 托盘应用（如果流水线构建了的话）
  IfFileExists "..\..\oadin-tray.exe" 0 +2
    File "..\..\oadin-tray.exe"
  
  ; 安装脚本
  File "preinstall.bat"
  File "postinstall.bat"
  File "start-oadin.bat"

  ; 执行安装步骤
  nsExec::Exec '"$INSTDIR\preinstall.bat"'

  # Post-install silently with argument
  nsExec::Exec '"$INSTDIR\postinstall.bat" "$INSTDIR"'

  # start oadin server
  nsExec::Exec '"$INSTDIR\start-oadin.bat"'
SectionEnd