!ifndef VERSION
  !define VERSION "0.0.0"
!endif

; CI/CD Compatible NSIS Script for 64-bit Installation
; 集成改进版start-oadin.bat脚本，支持服务和手动启动模式

!include "x64.nsh"
!include "LogicLib.nsh"
!include "nsDialogs.nsh"

!define APP_NAME "Oadin CLI"
!define COMPANY_NAME "Digital China"
!define SERVICE_NAME "OadinService"
!define SERVICE_DISPLAY_NAME "${APP_NAME} Service"
!define SERVICE_DESCRIPTION "Oadin CLI 后台服务"
!define DEFAULT_INSTALL_DIR "$PROGRAMFILES64\Oadin"

Outfile "..\..\oadin-installer.exe"
InstallDir "${DEFAULT_INSTALL_DIR}"
RequestExecutionLevel admin
SetCompress auto
SetCompressor lzma

Name "${APP_NAME}"
Caption "${APP_NAME} ${VERSION} Setup"

Page directory
Page custom ServicePageCreate ServicePageLeave
Page instfiles
UninstPage uninstConfirm
UninstPage instfiles

; 变量定义
Var CHECK_SERVICE
Var CHECK_AUTOSTART
Var TEMP_INSTDIR

; 初始化函数
Function .onInit
  ${IfNot} ${RunningX64}
    MessageBox MB_OK|MB_ICONSTOP "此应用程序需要64位Windows系统。"
    Abort
  ${EndIf}

  SetRegView 64
  ${DisableX64FSRedirection}

  ; 设置默认安装路径
  StrCpy $R0 "$PROGRAMFILES64"
  ${If} $R0 != ""
    StrCpy $INSTDIR "$R0\Oadin"
  ${Else}
    ReadEnvStr $R1 "ProgramW6432"
    ${If} $R1 != ""
      StrCpy $INSTDIR "$R1\Oadin"
    ${Else}
      StrCpy $INSTDIR "C:\Program Files\Oadin"
    ${EndIf}
  ${EndIf}

  StrCpy $TEMP_INSTDIR $INSTDIR
  DetailPrint "默认安装目录: $INSTDIR"
  ${EnableX64FSRedirection}
FunctionEnd

; 字符串搜索函数
Function StrStr
  Exch $R1
  Exch
  Exch $R2
  Push $R3
  Push $R4
  Push $R5
  StrLen $R3 $R1
  StrCpy $R4 0
  loop:
    StrCpy $R5 $R2 $R3 $R4
    StrCmp $R5 $R1 done
    StrCmp $R5 "" done
    IntOp $R4 $R4 + 1
    Goto loop
  done:
  StrCpy $R1 $R2 "" $R4
  Pop $R5
  Pop $R4
  Pop $R3
  Pop $R2
  Exch $R1
FunctionEnd

; 服务配置页面
Function ServicePageCreate
  nsDialogs::Create 1018
  Pop $0

  ${If} $0 == error
    Abort
  ${EndIf}

  ${NSD_CreateLabel} 0 0 100% 12u "服务配置"
  Pop $0
  ${NSD_SetFont} $0 "Arial" 10 true

  ${NSD_CreateCheckbox} 0 30u 100% 12u "将 Oadin 注册为 Windows 服务"
  Pop $CHECK_SERVICE
  ${NSD_Check} $CHECK_SERVICE

  ${NSD_CreateCheckbox} 0 50u 100% 12u "设置为开机自启（需要注册服务）"
  Pop $CHECK_AUTOSTART
  ${NSD_Check} $CHECK_AUTOSTART

  ${NSD_OnClick} $CHECK_SERVICE OnServiceCheckboxChange

  nsDialogs::Show
FunctionEnd

; 服务复选框事件
Function OnServiceCheckboxChange
  ${NSD_GetState} $CHECK_SERVICE $R0
  ${If} $R0 == 0
    ${NSD_SetState} $CHECK_AUTOSTART 0
    EnableWindow $CHECK_AUTOSTART 0
  ${Else}
    EnableWindow $CHECK_AUTOSTART 1
  ${EndIf}
FunctionEnd

; 离开服务配置页面
Function ServicePageLeave
  ${NSD_GetState} $CHECK_SERVICE $R0
  StrCpy $CHECK_SERVICE $R0

  ${NSD_GetState} $CHECK_AUTOSTART $R0
  StrCpy $CHECK_AUTOSTART $R0

  Push $INSTDIR
  Push "(x86)"
  Call StrStr
  Pop $R2
  ${If} $R2 != ""
    MessageBox MB_YESNO|MB_ICONWARNING "检测到32位目录。建议使用64位目录，是否继续？" IDYES continue_install IDNO change_dir
    Goto end_check
  ${EndIf}

  continue_install:
  Goto end_check

  change_dir:
  StrCpy $INSTDIR $TEMP_INSTDIR
  Abort

  end_check:
FunctionEnd

; 安装服务函数 - 关键改进：使用start-oadin.bat作为服务启动程序
Function InstallService
  ${If} $CHECK_SERVICE == 1
    DetailPrint "正在注册 Windows 服务..."

    ; 构建服务安装命令，使用改进后的start-oadin.bat作为服务启动程序
    ; 注意：服务模式需要传递-service参数
    StrCpy $R0 '"$INSTDIR\start-oadin.bat" -service'
    StrCpy $R1 'sc create ${SERVICE_NAME} binPath= "$R0" start= ${If} $CHECK_AUTOSTART == 1 auto ${Else} demand ${EndIf} DisplayName= "${SERVICE_DISPLAY_NAME}"'

    ; 执行服务安装命令
    nsExec::ExecToLog $R1
    Pop $R2
    ${If} $R2 != 0
      DetailPrint "服务注册失败，错误代码: $R2"
      MessageBox MB_OK|MB_ICONWARNING "服务注册失败，您可以手动运行以下命令：`$R1`"
    ${Else}
      ; 设置服务描述
      nsExec::ExecToLog '"sc description ${SERVICE_NAME} "${SERVICE_DESCRIPTION}""'

      ; 启动服务
      nsExec::ExecToLog '"sc start ${SERVICE_NAME}"'
      DetailPrint "Windows 服务注册成功"
    ${EndIf}
  ${EndIf}
FunctionEnd

; 卸载服务函数
Function UninstallService
  DetailPrint "正在卸载 Windows 服务..."

  ; 停止服务
  nsExec::ExecToLog '"sc stop ${SERVICE_NAME}"'

  ; 删除服务
  nsExec::ExecToLog '"sc delete ${SERVICE_NAME}"'
  Pop $R0

  ${If} $R0 == 0
    DetailPrint "Windows 服务卸载成功"
  ${Else}
    DetailPrint "服务卸载失败，可能服务未安装"
  ${EndIf}
FunctionEnd

; 安装部分 - 关键改进：正确处理start-oadin.bat脚本
Section "Install"
  SetRegView 64
  ${DisableX64FSRedirection}

  DetailPrint "正在安装到: $INSTDIR"

  ; 创建安装目录
  CreateDirectory "$INSTDIR"
  SetOutPath "$INSTDIR"

  IfFileExists "$INSTDIR" 0 install_error
  DetailPrint "安装目录创建成功"

  ; 复制文件（确保包含改进后的start-oadin.bat）
  File "..\..\oadin.exe"
  File "preinstall.bat"
  File "postinstall.bat"
  File "start-oadin.bat" ; 复制改进后的启动脚本

  ; 设置启动脚本的执行权限
  nsExec::ExecToLog 'icacls "$INSTDIR\start-oadin.bat" /grant:r "Users:(RX)"'

  ; 写入注册表
  WriteRegStr HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "InstallDir" "$INSTDIR"
  WriteRegStr HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "Version" "${VERSION}"
  WriteRegStr HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "Architecture" "x64"
  WriteRegStr HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "UninstallString" "$INSTDIR\uninstall.exe"

  WriteUninstaller "$INSTDIR\uninstall.exe"

  ; 执行安装脚本
  DetailPrint "运行预安装脚本..."
  nsExec::ExecToLog '"$INSTDIR\preinstall.bat"'

  DetailPrint "运行后安装脚本..."
  nsExec::ExecToLog '"$INSTDIR\postinstall.bat" "$INSTDIR"'

  ; 安装服务
  Call InstallService

  ; 如果未注册服务，使用启动脚本手动启动（普通模式）
  ${If} $CHECK_SERVICE == 0
    DetailPrint "启动 Oadin 应用程序..."
    nsExec::ExecToLog '"$INSTDIR\start-oadin.bat"' ; 不传递参数，使用普通模式
  ${EndIf}

  ${EnableX64FSRedirection}

  DetailPrint "安装成功，路径: $INSTDIR"
  Goto install_end

  install_error:
  DetailPrint "错误: 无法创建安装目录: $INSTDIR"
  MessageBox MB_OK|MB_ICONSTOP "安装失败: 无法创建目录 $INSTDIR"
  Abort

  install_end:
SectionEnd

; 卸载初始化
Function un.onInit
  SetRegView 64
  ${DisableX64FSRedirection}
FunctionEnd

; 卸载部分
Section "Uninstall"
  SetRegView 64
  ${DisableX64FSRedirection}

  ; 先卸载服务
  Call UninstallService

  ; 删除文件
  Delete "$INSTDIR\oadin.exe"
  Delete "$INSTDIR\preinstall.bat"
  Delete "$INSTDIR\postinstall.bat"
  Delete "$INSTDIR\start-oadin.bat" ; 删除启动脚本
  Delete "$INSTDIR\uninstall.exe"

  RMDir "$INSTDIR"
  DeleteRegKey HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}"

  ${EnableX64FSRedirection}
SectionEnd
