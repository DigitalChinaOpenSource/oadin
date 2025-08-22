!ifndef VERSION
  !define VERSION "0.0.0"
!endif

; 包含64位支持库
!include "x64.nsh"

!define APP_NAME "Oadin CLI"
!define COMPANY_NAME "Digital China"
; 明确指定64位Program Files目录
!define INSTALL_DIR "$PROGRAMFILES64\Oadin"

Outfile "..\..\oadin-installer.exe"
InstallDir "${INSTALL_DIR}"
RequestExecutionLevel admin ; Admin rights required for Program Files installation
SetCompress auto
SetCompressor lzma

; 安装器页面设置
Name "${APP_NAME}"
Caption "${APP_NAME} ${VERSION} 安装程序"

; 64位安装器设置
Target amd64-unicode

; 强制安装到64位Program Files目录
Function .onInit
  ; 检查是否为64位系统
  ${IfNot} ${RunningX64}
    MessageBox MB_OK|MB_ICONSTOP "此应用程序需要64位Windows系统。安装将退出。"
    Abort
  ${EndIf}

  ; 强制使用64位注册表视图和文件系统
  SetRegView 64
  ${DisableX64FSRedirection}

  ; 明确设置安装目录为64位Program Files
  ; 使用$PROGRAMFILES64确保安装到正确位置
  ${If} ${RunningX64}
    StrCpy $INSTDIR "$PROGRAMFILES64\Oadin"
  ${EndIf}

  ; 显示安装路径确认
  ; MessageBox MB_OK "安装路径: $INSTDIR"
FunctionEnd

Section "Install"
  ; 确保64位安装环境
  SetRegView 64
  ${DisableX64FSRedirection}

  ; 创建安装目录
  CreateDirectory "$INSTDIR"
  SetOutPath "$INSTDIR"

  ; 复制文件
  File "..\..\oadin.exe"
  File "preinstall.bat"
  File "postinstall.bat"
  File "start-oadin.bat"

  ; 写入安装信息到注册表（64位）
  WriteRegStr HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "InstallDir" "$INSTDIR"
  WriteRegStr HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "Version" "${VERSION}"

  # Pre-install silently
  DetailPrint "执行预安装脚本..."
  nsExec::Exec '"$INSTDIR\preinstall.bat"'

  # Post-install silently with argument
  DetailPrint "执行后安装脚本..."
  nsExec::Exec '"$INSTDIR\postinstall.bat" "$INSTDIR"'

  # start oadin server
  DetailPrint "启动 Oadin 服务..."
  nsExec::Exec '"$INSTDIR\start-oadin.bat"'

  ; 重新启用文件系统重定向
  ${EnableX64FSRedirection}

  DetailPrint "安装完成！安装路径: $INSTDIR"
SectionEnd

; 卸载器
Function un.onInit
  SetRegView 64
  ${DisableX64FSRedirection}
FunctionEnd

Section "Uninstall"
  ; 删除文件
  Delete "$INSTDIR\oadin.exe"
  Delete "$INSTDIR\preinstall.bat"
  Delete "$INSTDIR\postinstall.bat"
  Delete "$INSTDIR\start-oadin.bat"

  ; 删除目录
  RMDir "$INSTDIR"

  ; 清理注册表
  DeleteRegKey HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}"

  ${EnableX64FSRedirection}
SectionEnd