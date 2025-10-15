!define MUI_HEADERICON "oadin.ico"
!define MUI_UNICON "oadin.ico"

!include "LogicLib.nsh"
!include "MUI2.nsh"
!include Sections.nsh
!include FileFunc.nsh


!ifndef VERSION
  !define VERSION "0.0.0"
!endif

!define APP_NAME "Oadin CLI"
!define COMPANY_NAME "Digital China"
!define DEFAULT_INSTALL_DIR "C:\Program Files\Oadin"
!define DEFAULT_INSTALL_DATA_DIR "C:\ProgramData\Oadin"
!define DEFAULT_PARAMS_DATA_DIR "C:\ProgramData\oadin_phase.txt"
!define DEFAULT_PARAMS_DATA_DIR1 "C:\ProgramData\oadin_phase1.txt"

Outfile "..\..\oadin-installer.exe"
InstallDir "${DEFAULT_INSTALL_DIR}"
RequestExecutionLevel user
SetCompress auto
SetCompressor lzma



Name "${APP_NAME}"
Caption "${APP_NAME} ${VERSION} Setup"

; MUI Settings
!define MUI_ABORTWARNING

; ------------------ Install Pages ------------------
!define MUI_PAGE_CUSTOMFUNCTION_PRE SkipIfPhase
!insertmacro MUI_PAGE_WELCOME
Page custom SelectInstallScopePage SelectInstallScopePageLeave
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!define MUI_FINISHPAGE_RUN
!define MUI_FINISHPAGE_RUN_TEXT "Start Oadin"
!define MUI_FINISHPAGE_RUN_FUNCTION LaunchOadin
!define MUI_FINISHPAGE_SHOWREADME
!define MUI_FINISHPAGE_SHOWREADME_TEXT "Enable Auto-Start"
!define MUI_FINISHPAGE_SHOWREADME_FUNCTION EnableAutoStart
!define MUI_FINISHPAGE_RUN_NOTCHECKED
!define MUI_FINISHPAGE_SHOWREADME_NOTCHECKED
!insertmacro MUI_PAGE_FINISH

; ------------------ Uninstall Pages ------------------
!insertmacro MUI_UNPAGE_WELCOME
UninstPage custom un.SelectUninstallModePage un.SelectUninstallModePageLeave
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

!insertmacro MUI_LANGUAGE "English"

Var SILENT
Var INSTALL_SCOPE  ; 1=System-wide (HKLM), 0=Current user (HKCU)
Var ExecutionLevelPhase

Function SkipIfPhase
  ${If} $ExecutionLevelPhase == 2
    Abort
  ${EndIf}
FunctionEnd

; ------------------ Initialization ------------------
Function .onInit
  SetRegView 64
  StrCpy $SILENT 0
  IfSilent 0 +3
    StrCpy $SILENT 1
  
  Sleep 200
  ${If} ${FileExists} "${DEFAULT_PARAMS_DATA_DIR}"
    FileOpen $0 "${DEFAULT_PARAMS_DATA_DIR}" r
    FileClose $0
    Delete "${DEFAULT_PARAMS_DATA_DIR}"
    StrCpy $ExecutionLevelPhase 2
  ${Else}
    StrCpy $ExecutionLevelPhase 1
  ${EndIf}

  ; Check previous installation path
  ReadRegStr $R0 HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "InstallDir"
  ${If} $R0 != ""
    StrCpy $INSTDIR $R0
    StrCpy $INSTALL_SCOPE 1
  ${Else}
    ReadRegStr $R0 HKCU "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "InstallDir"
    ${If} $R0 != ""
      StrCpy $INSTDIR $R0
      StrCpy $INSTALL_SCOPE 0
    ${Else}
      StrCpy $INSTDIR "${DEFAULT_INSTALL_DIR}"
      StrCpy $INSTALL_SCOPE 1
    ${EndIf}
  ${EndIf}

  ${If} $R0 != ""
    ; Already installed
    ${If} $SILENT == 0
      ${If} $INSTALL_SCOPE == 1
        ; --- Dynamic elevation ---
        UserInfo::GetAccountType
        Pop $0
        StrCmp $0 "Admin" done_elevated  ; Already administrator, continue
        ExecShell "runas" "$EXEPATH" ""
        Quit  ; Exit current process so the elevated process can take over
        done_elevated:
          ;Already administrator,
      ${EndIf}
      MessageBox MB_YESNO|MB_ICONQUESTION "${APP_NAME} is already installed at $R0. Do you want to repair the installation?" IDYES do_fix IDNO cancel_install
      cancel_install:
        Abort
      do_fix:
        Call RemoveOldOadin
    ${Else}
      Call RemoveOldOadin
    ${EndIf}
  ${Else}
    ; Not installed, ask installation scope in non-silent mode
  ${EndIf}
FunctionEnd

; ------------------ Install Section ------------------
Section "Install"
  SetRegView 64
  SetOutPath "$INSTDIR"
  CreateDirectory "$INSTDIR"

  File "..\..\oadin.exe"
  File "..\..\oadin-app.exe"
  File "oadin.ico"

  ; Write registry and PATH
  ${If} $INSTALL_SCOPE == 1
    WriteRegStr HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "InstallDir" "$INSTDIR"
    WriteRegStr HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "Version" "${VERSION}"
    WriteRegStr HKLM "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "DisplayName" "${APP_NAME}"
    WriteRegStr HKLM "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "UninstallString" "$INSTDIR\uninstall.exe"
    Call AddPathSystem
  ${Else}
    WriteRegStr HKCU "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "InstallDir" "$INSTDIR"
    WriteRegStr HKCU "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "Version" "${VERSION}"
    WriteRegStr HKCU "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "DisplayName" "${APP_NAME}"
    WriteRegStr HKCU "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "UninstallString" "$INSTDIR\uninstall.exe"
    Call AddPathUser
  ${EndIf}

  WriteUninstaller "$INSTDIR\uninstall.exe"
  ; nsExec::ExecToLog '"$INSTDIR\postinstall.bat" "$INSTDIR"'
SectionEnd

; ------------------ PATH Functions ------------------
Var SysStr
Var SysLastChar
Function AddPathSystem
  SetRegView 64
  ReadRegStr $R0 HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "Path"
  StrCpy $R1 "$INSTDIR"
  Push $R1
  Push $R0
  Push $R1
  Call StrContains
  Pop $R2
  Pop $R1
  ${If} $R2 == "false"
    ${If} $R0 == ""
      WriteRegExpandStr HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "Path" "$R1"
    ${Else}
      loop_trim:
        StrCpy $SysStr $R0
        StrLen $0 $SysStr
        IntOp $0 $0 - 1
        StrCpy $SysLastChar $SysStr 1 $0
        ${If} $SysLastChar == ";"
          StrCpy $R0 $R0 -1
          Goto loop_trim
        ${EndIf}
      WriteRegExpandStr HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "Path" "$R0;$R1"
    ${EndIf}
    ; SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment"
  ${EndIf}
FunctionEnd

Var UserStr
Var lastChar
Function AddPathUser
  SetRegView 64
  ReadRegStr $R0 HKCU "Environment" "Path"
  StrCpy $R1 "$INSTDIR"
  Push $R1
  Push $R0
  Push $R1
  Call StrContains
  Pop $R2
  Pop $R1
  ${If} $R2 == "false"
    ${If} $R0 == ""
      WriteRegExpandStr HKCU "Environment" "Path" "$R1"
    ${Else}
      loop_trim:
        StrCpy $UserStr $R0
        StrLen $0 $UserStr
        IntOp $0 $0 - 1
        StrCpy $lastChar $UserStr 1 $0
        ${If} $lastChar == ";"
          StrCpy $R0 $R0 -1
          Goto loop_trim
        ${EndIf}
      WriteRegExpandStr HKCU "Environment" "Path" "$R0;$R1"
    ${EndIf}
    ; SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment"
  ${EndIf}
FunctionEnd

; ------------------ String Contains ------------------
Function StrContains
  Exch $R1 ; Substring
  Exch
  Exch $R0 ; String
  Push $R2
  Push $R3
  StrLen $R2 $R1
  StrLen $R3 $R0
  StrCpy $R4 0
  loop:
    StrCpy $R5 $R0 $R2 $R4
    StrCmp $R5 $R1 found
    IntOp $R4 $R4 + 1
    StrCmp $R4 $R3 done
    Goto loop
  found:
    StrCpy $R1 "true"
    Goto end
  done:
    StrCpy $R1 "false"
  end:
  Pop $R3
  Pop $R2
  Exch $R1
FunctionEnd

; ------------------ Uninstall Section ------------------
Var UNINSTALL_PROGRAM
Var UNINSTALL_DATA

Function un.SelectUninstallModePage
  nsDialogs::Create 1018
  Pop $0
  ${If} $0 == error
    Abort
  ${EndIf}

  ${NSD_CreateLabel} 0 0 100% 12u "Select uninstall mode:"
  Pop $R0

  ; Checkbox for Program (disabled and checked)
  ${NSD_CreateCheckBox} 0 20 100% 12u "Program (Required)"
  Pop $R1
  ${NSD_SetState} $R1 ${BST_CHECKED}
  EnableWindow $R1 0
  StrCpy $UNINSTALL_PROGRAM 1

  ; Checkbox for Data
  ${NSD_CreateCheckBox} 0 40 100% 12u "Data (optional)"
  Pop $R2
  ${NSD_SetState} $R2 ${BST_UNCHECKED}
  StrCpy $UNINSTALL_DATA 0

  nsDialogs::Show
FunctionEnd

Function un.SelectUninstallModePageLeave
  StrCpy $UNINSTALL_PROGRAM 1

  ${NSD_GetState} $R2 $R0
  StrCmp $R0 ${BST_CHECKED} 0 +2
    StrCpy $UNINSTALL_DATA 1
FunctionEnd

Function un.onInit
  SetRegView 64
  ; Determine installation scope
  ReadRegStr $INSTDIR HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "InstallDir"
  ${If} $INSTDIR != ""
    StrCpy $INSTALL_SCOPE 1
  ${Else}
    ReadRegStr $INSTDIR HKCU "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "InstallDir"
    ${If} $INSTDIR != ""
      StrCpy $INSTALL_SCOPE 0
    ${Else}
      StrCpy $INSTDIR "${DEFAULT_INSTALL_DIR}"
      StrCpy $INSTALL_SCOPE 1
    ${EndIf}
  ${EndIf}
  ${If} $INSTALL_SCOPE == 1
    ; --- Dynamic elevation ---
    UserInfo::GetAccountType
    Pop $0
    StrCmp $0 "Admin" done_elevated  ; Already administrator, continue
    ExecShell "runas" "$EXEPATH" ""
    Quit  ; Exit current process so the elevated process can take over
    done_elevated:
      ;Already administrator,
  ${EndIf} 
FunctionEnd

Section "Uninstall"
  SetRegView 64
  ; stop oadin server
  nsExec::Exec '"$INSTDIR\oadin.exe" server stop'
  nsExec::ExecToStack 'taskkill /F /IM oadin.exe' 
  nsExec::ExecToStack 'taskkill /F /IM oadin-app.exe'  
  Sleep 200
  ; Remove based on uninstall mode
  ${If} $UNINSTALL_PROGRAM == 1
    RMDir /r "$INSTDIR"
  ${EndIf}
  ${If} $UNINSTALL_DATA == 1
    RMDir /r "${DEFAULT_INSTALL_DATA_DIR}"
  ${EndIf}
  Push $INSTDIR
  Delete "$DESKTOP\Oadin.lnk"

  Call un.RemovePathEnv

  DeleteRegKey HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}"
  DeleteRegKey HKLM "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}"
  DeleteRegValue HKLM "Software\Microsoft\Windows\CurrentVersion\Run" "Oadin"
  DeleteRegKey HKCU "SOFTWARE\${COMPANY_NAME}\${APP_NAME}"
  DeleteRegKey HKCU "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}"
  DeleteRegValue HKCU "Software\Microsoft\Windows\CurrentVersion\Run" "Oadin"
SectionEnd

Function RemovePathEnv
  SetRegView 64
  StrCpy $R1 "$INSTDIR"
  ${If} $INSTALL_SCOPE == 1
    ReadRegStr $R0 HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "Path"
    ${If} $R0 != ""
      Push $R0
      Push $R1
      Call StrRemove
      ; Pop $R2
      WriteRegExpandStr HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "Path" "$0"
    ${EndIf}
  ${Else}
    ReadRegStr $R0 HKCU "Environment" "Path"
    ${If} $R0 != ""
      Push $R0
      Push $R1
      Call StrRemove
      ; Pop $R2
      WriteRegExpandStr HKCU "Environment" "Path" "$0"
    ${EndIf}
  ${EndIf}
  ; SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment"
FunctionEnd

Function un.RemovePathEnv
  SetRegView 64
  StrCpy $R1 "$INSTDIR"
  ${If} $INSTALL_SCOPE == 1
    ReadRegStr $R0 HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "Path"
    ${If} $R0 != ""
      Push $R0
      Push $R1
      Call un.StrRemove
      ; Pop $R2
      WriteRegExpandStr HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "Path" "$0"
    ${EndIf}
  ${Else}
    ReadRegStr $R0 HKCU "Environment" "Path"
    ${If} $R0 != ""
      Push $R0
      Push $R1
      Call un.StrRemove
      ; Pop $R2
      WriteRegExpandStr HKCU "Environment" "Path" "$0"
    ${EndIf}
  ${EndIf}
  ; SendMessage ${HWND_BROADCAST} ${WM_SETTINGCHANGE} 0 "STR:Environment"
FunctionEnd

Function StrRemove
  Pop $R1  ; target string (second parameter)
  Pop $R0  ; source string (first parameter)

  Push $R2  ; target length
  Push $R3  ; position
  Push $R4  ; front part
  Push $R5  ; back part
  Push $R6  ; temporary position

  StrCpy $R0 $R0    ; source string
  StrCpy $R1 $R1    ; target string
  StrLen $R2 $R1    ; target length

  ${If} $R2 == 0
    StrCpy $0 $R0
    Goto exit
  ${EndIf}

  StrCpy $R3 0

  find_loop:
      StrCpy $R6 $R0 $R2 $R3
      ${If} $R6 == $R1
          ; Match found, split the string
          StrCpy $R4 $R0 $R3 0        ; front part
          IntOp $R3 $R3 + $R2
          StrLen $R6 $R0
          IntOp $R6 $R6 - $R3
          StrCpy $R5 $R0 $R6 $R3      ; back part
          StrCpy $R0 "$R4$R5"         ; recombine
          StrCpy $R3 0                ; restart search
      ${Else}
          IntOp $R3 $R3 + 1
      ${EndIf}

      StrLen $R6 $R0
      ${If} $R3 < $R6
          Goto find_loop
      ${EndIf}

  StrCpy $0 $R0

  exit:
  Pop $R6
  Pop $R5
  Pop $R4
  Pop $R3
  Pop $R2
  Pop $R1
  Pop $R0
FunctionEnd

Function un.StrRemove
  Pop $R1  ; target string (second parameter)
  Pop $R0  ; source string (first parameter)

  Push $R2  ; target length
  Push $R3  ; position
  Push $R4  ; front part
  Push $R5  ; back part
  Push $R6  ; temporary position

  StrCpy $R0 $R0    ; source string
  StrCpy $R1 $R1    ; target string
  StrLen $R2 $R1    ; target length

  ${If} $R2 == 0
    StrCpy $0 $R0
    Goto exit
  ${EndIf}

  StrCpy $R3 0

  find_loop:
      StrCpy $R6 $R0 $R2 $R3
      ${If} $R6 == $R1
          ; Match found, split the string
          StrCpy $R4 $R0 $R3 0        ; front part
          IntOp $R3 $R3 + $R2
          StrLen $R6 $R0
          IntOp $R6 $R6 - $R3
          StrCpy $R5 $R0 $R6 $R3      ; back part
          StrCpy $R0 "$R4$R5"         ; recombine
          StrCpy $R3 0                ; restart search
      ${Else}
          IntOp $R3 $R3 + 1
      ${EndIf}

      StrLen $R6 $R0
      ${If} $R3 < $R6
          Goto find_loop
      ${EndIf}

  StrCpy $0 $R0

  exit:
  Pop $R6
  Pop $R5
  Pop $R4
  Pop $R3
  Pop $R2
  Pop $R1
  Pop $R0
FunctionEnd

; ------------------ Finish Page Functions ------------------
Function LaunchOadin
  ;nsExec::Exec '"$INSTDIR\oadin.exe" server start -d'
  ExecShell "open" "$INSTDIR\oadin-app.exe" "" SW_HIDE
FunctionEnd

Function EnableAutoStart
  SetRegView 64
  ${If} $INSTALL_SCOPE == 1
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Run" "Oadin" '"$INSTDIR\oadin-app.exe"'
  ${Else}
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Run" "Oadin" '"$INSTDIR\oadin-app.exe"'
  ${EndIf}
FunctionEnd

; ------------------ Remove Previous Installation ------------------
Function RemoveOldOadin
  SetRegView 64
  ReadRegStr $R3 HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "InstallDir"
  ${If} $R3 != ""
    ; stop oadin server
    nsExec::Exec '"$R3\oadin.exe" server stop'
    nsExec::ExecToStack 'taskkill /F /IM oadin.exe' 
    nsExec::ExecToStack 'taskkill /F /IM oadin-app.exe'
    Sleep 200
    RMDir /r "$R3"
    Delete "$DESKTOP\Oadin.lnk"
    Push $R3
    Call RemovePathEnv
    DeleteRegKey HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}"
    DeleteRegKey HKLM "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}"
  ${EndIf}
  ReadRegStr $R3 HKCU "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "InstallDir"
  ${If} $R3 != ""
    nsExec::Exec '"$R3\oadin.exe" server stop'
    nsExec::ExecToStack 'taskkill /F /IM oadin.exe' 
    nsExec::ExecToStack 'taskkill /F /IM oadin-app.exe'
    Sleep 200
    RMDir /r "$R3"
    Delete "$DESKTOP\Oadin.lnk"
    Push $R3
    Call RemovePathEnv
    DeleteRegKey HKCU "SOFTWARE\${COMPANY_NAME}\${APP_NAME}"
    DeleteRegKey HKCU "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}"
  ${EndIf}
FunctionEnd

Function SelectInstallScopePage
  ${If} $ExecutionLevelPhase == 2
    StrCpy $INSTALL_SCOPE 1
    Abort
  ${EndIf}

  nsDialogs::Create 1018
  Pop $0

  ${If} $0 == error
    Abort
  ${EndIf}

  ; Label text
  ${NSD_CreateLabel} 0 0 100% 12u "Select installation scope:"
  Pop $R0

  ; Radio button: Current User
  ${NSD_CreateRadioButton} 0 20 100% 12u "Install for Current User"
  Pop $R1

  ; Radio button: All Users
  ${NSD_CreateRadioButton} 0 40 100% 12u "Install for All Users"
  Pop $R2

  ; Default selection: "All Users"
  ${NSD_SetState} $R2 ${BST_CHECKED}

  ; Save control handles
  StrCpy $R3 $R1
  StrCpy $R4 $R2

  nsDialogs::Show
FunctionEnd

Function SelectInstallScopePageLeave
  ; Check which radio button is selected
  ${NSD_GetState} $R4 $R0
  StrCmp $R0 ${BST_CHECKED} all_users selected_current

  selected_current:
    StrCpy $INSTALL_SCOPE 0
    UserInfo::GetAccountType
    Pop $0
    StrCmp $0 "Admin" 0 done_elevated_check
      MessageBox MB_OK "Current permissions do not match the required permissions.Please rerun the installer."
      Quit
      done_elevated_check:
      ; user permission, continue 
    Goto done_scope

  all_users:
    StrCpy $INSTALL_SCOPE 1

    ; --- Dynamic elevation ---
    UserInfo::GetAccountType
    Pop $0
    StrCmp $0 "Admin" done_elevated  ; Already administrator, continue

    ; Not an administrator, restart self with elevated privileges
    FileOpen $0 "${DEFAULT_PARAMS_DATA_DIR}" w
    FileWrite $0 "1"
    FileClose $0
    ExecShell "runas" "$EXEPATH" ""
    Quit  ; Exit current process so the elevated process can take over

    done_elevated:
    ; Already elevated, continue installation

  done_scope:
FunctionEnd
