!ifndef VERSION
  !define VERSION "0.0.0"
!endif

; Compatible with 32-bit NSIS but forces 64-bit installation behavior
; Remove Target amd64-unicode for CI/CD compatibility
; Unicode True

; 包含64位支持库
!include "x64.nsh"
!include "LogicLib.nsh"

!define APP_NAME "Oadin CLI"
!define COMPANY_NAME "Digital China"
; 明确指定64位Program Files目录 - 使用PROGRAMFILES64确保不会重定向
!define INSTALL_DIR "$PROGRAMFILES64\Oadin"

Outfile "..\..\oadin-installer.exe"
InstallDir "${INSTALL_DIR}"
RequestExecutionLevel admin ; Admin rights required for Program Files installation
SetCompress auto
SetCompressor lzma

; 安装器页面设置
Name "${APP_NAME}"
Caption "${APP_NAME} ${VERSION} 安装程序"

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

  ; 强制设置安装目录为64位Program Files - CI/CD兼容版本
  ; 多重保障防止安装到x86目录
  
  ; 方法1: 使用PROGRAMFILES64
  StrCpy $INSTDIR "$PROGRAMFILES64\Oadin"
  
  ; 方法2: 如果PROGRAMFILES64为空，使用备选方案
  ${If} $INSTDIR == "\Oadin"
    ; PROGRAMFILES64变量为空，使用其他方式
    ReadEnvStr $R0 "ProgramW6432"
    ${If} $R0 != ""
      StrCpy $INSTDIR "$R0\Oadin"
    ${Else}
      ; 最后的保障：硬编码64位路径
      StrCpy $INSTDIR "C:\Program Files\Oadin"
    ${EndIf}
  ${EndIf}

  ; 静默安装 - 无需弹窗确认
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
  WriteRegStr HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "UninstallString" "$INSTDIR\uninstall.exe"
  
  ; 生成卸载器
  WriteUninstaller "$INSTDIR\uninstall.exe"

  ; Pre-install silently
  DetailPrint "执行预安装脚本..."
  nsExec::Exec '"$INSTDIR\preinstall.bat"'

  ; Post-install silently with argument
  DetailPrint "执行后安装脚本..."
  nsExec::Exec '"$INSTDIR\postinstall.bat" "$INSTDIR"'

  ; start oadin server
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
  ; 确保64位环境
  SetRegView 64
  ${DisableX64FSRedirection}
  
  ; 删除文件
  Delete "$INSTDIR\oadin.exe"
  Delete "$INSTDIR\preinstall.bat"
  Delete "$INSTDIR\postinstall.bat"
  Delete "$INSTDIR\start-oadin.bat"
  Delete "$INSTDIR\uninstall.exe"

  ; 删除目录
  RMDir "$INSTDIR"

  ; 清理注册表
  DeleteRegKey HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}"

  ; 恢复文件系统重定向
  ${EnableX64FSRedirection}
  
  MessageBox MB_OK "卸载完成！"
SectionEnd