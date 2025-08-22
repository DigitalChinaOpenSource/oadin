!ifndef VERSION
  !define VERSION "0.0.0"
!endif

; 包含64位支持库
!include "x64.nsh"

!define APP_NAME "Oadin CLI"
!define COMPANY_NAME "Digital China"
!define INSTALL_DIR "$PROGRAMFILES\Oadin"

Outfile "..\..\oadin-installer.exe"
InstallDir "${INSTALL_DIR}"
RequestExecutionLevel admin ; Admin rights required for Program Files installation
SetCompress auto
SetCompressor lzma

; 强制安装到64位Program Files目录
Function .onInit
  ${If} ${RunningX64}
    ; 在64位系统上，强制使用真正的Program Files目录
    ; 而不是被重定向到Program Files (x86)
    SetRegView 64
    StrCpy $INSTDIR "$PROGRAMFILES\Oadin"
  ${Else}
    MessageBox MB_OK|MB_ICONSTOP "此应用程序需要64位Windows系统。"
    Abort
  ${EndIf}
FunctionEnd

Section "Install"
  ; 禁用WOW64文件系统重定向，确保安装到真正的Program Files
  ${DisableX64FSRedirection}
  
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
  
  ; 重新启用文件系统重定向
  ${EnableX64FSRedirection}
SectionEnd