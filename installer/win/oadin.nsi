!ifndef VERSION
  !define VERSION "0.0.0"
!endif

; 64位NSIS支持
Unicode True
Target amd64-unicode

!define APP_NAME "Oadin CLI"
!define COMPANY_NAME "Digital China"
!define INSTALL_DIR "$PROGRAMFILES\Oadin"

Outfile "..\..\oadin-installer.exe"
InstallDir "${INSTALL_DIR}"
RequestExecutionLevel admin ; Admin rights required for Program Files installation
SetCompress auto
SetCompressor lzma

; 64位系统检查
Function .onInit
  ; 检查是否为64位系统
  System::Call "kernel32::GetCurrentProcess() i .s"
  System::Call "kernel32::IsWow64Process(i s, *i .r0)"
  IntCmp $0 0 not64bit
  Goto is64bit
  not64bit:
    MessageBox MB_OK|MB_ICONSTOP "此安装程序需要64位Windows系统。"
    Abort
  is64bit:
    ; 设置64位安装路径
    StrCpy $INSTDIR "$PROGRAMFILES\Oadin"
FunctionEnd

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