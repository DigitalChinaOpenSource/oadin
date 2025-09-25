!ifndef VERSION
  !define VERSION "0.0.0"
!endif

; CI/CD Compatible NSIS Script for 64-bit Installation
; Designed to work with 32-bit NSIS compiler in CI environment

; Include 64-bit support libraries
!include "x64.nsh"
!include "LogicLib.nsh"
!include "MUI2.nsh"

!define APP_NAME "Oadin CLI"
!define COMPANY_NAME "Digital China"
!define DEFAULT_INSTALL_DIR "C:\Program Files\Oadin"

Outfile "..\..\oadin-installer.exe"
InstallDir "${DEFAULT_INSTALL_DIR}"
RequestExecutionLevel admin
SetCompress auto
SetCompressor lzma

Name "${APP_NAME}"
Caption "${APP_NAME} ${VERSION} Setup"

; MUI Settings
!define MUI_ABORTWARNING

; Installer pages
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES

; Finish page with checkboxes
!define MUI_FINISHPAGE_RUN
!define MUI_FINISHPAGE_RUN_TEXT "Start Oadin Service"
!define MUI_FINISHPAGE_RUN_FUNCTION LaunchOadinService
!define MUI_FINISHPAGE_SHOWREADME
!define MUI_FINISHPAGE_SHOWREADME_TEXT "Enable Oadin Auto-Start"
!define MUI_FINISHPAGE_SHOWREADME_FUNCTION EnableAutoStart
!define MUI_FINISHPAGE_RUN_NOTCHECKED
!define MUI_FINISHPAGE_SHOWREADME_NOTCHECKED
!insertmacro MUI_PAGE_FINISH

; Uninstaller pages
!insertmacro MUI_UNPAGE_WELCOME
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

; Language files
!insertmacro MUI_LANGUAGE "English"

; Initialization
Function .onInit
  ${IfNot} ${RunningX64}
    MessageBox MB_OK|MB_ICONSTOP "This application requires a 64-bit Windows system."
    Abort
  ${EndIf}

  SetRegView 64
  ${DisableX64FSRedirection}

  ; Check if OadinService exists
  nsExec::ExecToStack 'sc query "OadinService"'
  Pop $R0 ; Return code
  Pop $R1 ; Output

  ${If} $R0 == 0
    ; Service found, ask user
    MessageBox MB_YESNO|MB_ICONQUESTION "Oadin service is already installed. Do you want to uninstall the old version and continue installation?" IDYES do_uninstall IDNO cancel_install
    cancel_install:
      Abort
    do_uninstall:
      Call RemoveOldOadin
  ${Else}
    ; Service not found, check folder
    MessageBox MB_OK "Current Path: $PROGRAMFILES64\Oadin"
    IfFileExists "$PROGRAMFILES64\Oadin\*.*" folder_found no_folder

    folder_found:
      MessageBox MB_YESNO|MB_ICONQUESTION "Oadin service is already installed. Do you want to uninstall the old version and continue installation?" IDYES do_overwrite IDNO cancel_overwrite
      cancel_overwrite:
         Abort
      do_overwrite:
        ; Check if Oadin process is running
        nsExec::ExecToStack 'tasklist /FI "IMAGENAME eq oadin.exe"'
        Pop $R0  ; Return code
        Pop $R1  ; Command output

        StrCpy $R2 $R1
        StrCpy $R3 "oadin.exe"

        ; Search for "oadin.exe" in tasklist output
        Push $R3
        Push $R2
        Call StrStr
        Pop $R4
        StrCmp $R4 "" no_process

        DetailPrint "Oadin process detected, stopping it..."
        nsExec::ExecToLog '"$PROGRAMFILES64\Oadin\oadin.exe" server stop'

        no_process:
          do_remove_folder:
            ; Remove folder
            RMDir /r "$PROGRAMFILES64\Oadin"

            ; Read user PATH
            ReadRegStr $R2 HKCU "Environment" "Path"

            ; Check if PATH contains $PROGRAMFILES64\Oadin
            StrCpy $R0 "$PROGRAMFILES64\Oadin"
            Push $R0
            Push $R2
            Call StrStr
            Pop $R0
            StrCmp $R0 "" no_path ; if empty, PATH does not contain it

            ; Remove oadin path from PATH (handle different cases)
            ; Case 1: starts with path
            StrCpy $R0 "$PROGRAMFILES64\Oadin;"
            Push $R0
            Push ""
            Push $R2
            Call StrReplace
            Pop $R2

            ; Case 2: ends with path
            StrCpy $R0 ";$PROGRAMFILES64\Oadin"
            Push $R0
            Push ""
            Push $R2
            Call StrReplace
            Pop $R2

            ; Case 3: middle path
            StrCpy $R0 "$PROGRAMFILES64\Oadin"
            Push $R0
            Push ""
            Push $R2
            Call StrReplace
            Pop $R2
          no_path:
            ; if empty, PATH does not contain it
            DetailPrint "PATH does not contain it"
          no_folder:
            ; Do nothing
            DetailPrint "no folder"
  ${EndIf}

  ReadRegStr $R0 HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "InstallDir"
  ${If} $R0 != ""
    StrCpy $INSTDIR $R0
  ${Else}
    StrCpy $R1 "$PROGRAMFILES64"
    ${If} $R1 != ""
      ${AndIf} $R1 != "\$PROGRAMFILES64"
      StrCpy $INSTDIR "$R1\Oadin"
    ${Else}
      ReadEnvStr $R2 "ProgramW6432"
      ${If} $R2 != ""
        StrCpy $INSTDIR "$R2\Oadin"
      ${Else}
        StrCpy $INSTDIR "${DEFAULT_INSTALL_DIR}"
      ${EndIf}
    ${EndIf}

    Push $INSTDIR
    Push "(x86)"
    Call StrStr
    Pop $R3
    ${If} $R3 != ""
      StrCpy $INSTDIR "${DEFAULT_INSTALL_DIR}"
    ${EndIf}
  ${EndIf}

  DetailPrint "Target installation directory: $INSTDIR"
FunctionEnd

; Verify directory selection
Function .onVerifyInstDir
  Push $INSTDIR
  Push "(x86)"
  Call StrStr
  Pop $R0
  ${If} $R0 != ""
    MessageBox MB_OK|MB_ICONEXCLAMATION "Please select a 64-bit installation directory (not under Program Files (x86))."
    Abort
  ${EndIf}

  ClearErrors
  CreateDirectory "$INSTDIR"
  ${If} ${Errors}
    MessageBox MB_OK|MB_ICONEXCLAMATION "Cannot create directory. Please choose another location or run as administrator."
    Abort
  ${EndIf}
  RMDir "$INSTDIR"
FunctionEnd

; String search function
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

; String replace
Function StrReplace
  Exch $R2 ; new substring
  Exch
  Exch $R1 ; old substring
  Exch 2
  Exch $R0 ; original string

  Push $R3
  Push $R4
  Push $R5

  StrCpy $R3 ""
loop:
  StrCpy $R4 $R0 "" 0
  StrCmp $R4 "" done
  StrCpy $R5 $R0 ${NSIS_MAX_STRLEN}
  StrCpy $R5 $R5 "" 0
  StrCpy $R5 $R5 "" 0
  StrCmp $R5 $R1 0 no_match
    StrCpy $R3 "$R3$R2"
    StrCpy $R0 $R0 "" ${NSIS_MAX_STRLEN}
    Goto loop
no_match:
  StrCpy $R3 "$R3$R4"
  StrCpy $R0 $R0 "" ${NSIS_MAX_STRLEN}
  Goto loop
done:
  Pop $R5
  Pop $R4
  Pop $R3
  Exch $R3
FunctionEnd


Function RemoveOldOadin
  DetailPrint "Stopping and removing existing Oadin service..."
  nsExec::ExecToLog 'sc stop "OadinService"'
  nsExec::ExecToLog 'sc delete "OadinService"'

  ; Clean old installation directory
  ReadRegStr $R3 HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "InstallDir"
  ${If} $R3 != ""
    DetailPrint "Removing old installation directory: $R3"
    RMDir /r "$R3"

    ${un.EnvVarUpdate} $0 "PATH" "R" "HKLM" "$R3"
    System::Call 'User32::SendMessageTimeoutA(i 0xffff, i 0x1A, i 0, t "Environment", i 0, i 1000, *i .r0)'
  ${EndIf}

  ; Clean registry entries
  DeleteRegKey HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}"
  DeleteRegKey HKLM "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}"
  DeleteRegValue HKLM "Software\Microsoft\Windows\CurrentVersion\Run" "Oadin"

FunctionEnd

; Install Section
Section "Install"
  SetRegView 64
  ${DisableX64FSRedirection}

  DetailPrint "Installing to: $INSTDIR"
  CreateDirectory "$INSTDIR"
  SetOutPath "$INSTDIR"

  IfFileExists "$INSTDIR" 0 install_error
  File "..\..\oadin.exe"
  File "preinstall.bat"
  File "postinstall.bat"
  File "start-oadin.bat"

  WriteRegStr HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "InstallDir" "$INSTDIR"
  WriteRegStr HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "Version" "${VERSION}"
  WriteRegStr HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "Architecture" "x64"
  WriteRegStr HKLM "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "DisplayName" "${APP_NAME}"
  WriteRegStr HKLM "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "UninstallString" "$INSTDIR\uninstall.exe"
  WriteRegStr HKLM "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "DisplayVersion" "${VERSION}"
  WriteRegStr HKLM "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "Publisher" "${COMPANY_NAME}"
  WriteRegStr HKLM "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "DisplayIcon" "$INSTDIR\oadin.exe"

  DetailPrint "Registering Oadin service..."
  nsExec::ExecToLog 'sc create "OadinService" binPath= "\"$INSTDIR\oadin.exe\" server start" start= auto DisplayName= "Oadin Service"'

  WriteUninstaller "$INSTDIR\uninstall.exe"

  ; Execute post-install script (for PATH setup)
  DetailPrint "Running post-install script..."
  nsExec::ExecToLog '"$INSTDIR\postinstall.bat" "$INSTDIR"'

  ${EnableX64FSRedirection}
  Goto install_end

  install_error:
  DetailPrint "ERROR: Failed to create installation directory: $INSTDIR"
  MessageBox MB_OK|MB_ICONSTOP "Installation failed: Unable to create directory $INSTDIR"
  Abort

  install_end:
SectionEnd

; Finish Page Functions
Function LaunchOadinService

  DetailPrint "Starting Oadin service in background..."
  ; nsExec::Exec 'sc start "OadinService"'
  Exec 'sc start "OadinService"'
FunctionEnd

Function EnableAutoStart
  DetailPrint "Enabling Oadin auto-start..."
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Run" "Oadin" '"$INSTDIR\oadin.exe" server start'
FunctionEnd

; Uninstaller
Function un.onInit
  SetRegView 64
  ${DisableX64FSRedirection}
FunctionEnd

Section "Uninstall"
  SetRegView 64
  ${DisableX64FSRedirection}

  ReadRegStr $INSTDIR HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "InstallDir"
  ${If} $INSTDIR == ""
    StrCpy $INSTDIR "${DEFAULT_INSTALL_DIR}"
  ${EndIf}

  DetailPrint "Uninstalling from: $INSTDIR"

  ; Stop and delete service if it exists
  DetailPrint "Oadin process detected, stopping it..."
  nsExec::ExecToLog '"$INSTDIR\oadi n.exe" server stop'
  nsExec::ExecToLog 'sc stop "OadinService"'
  nsExec::ExecToLog 'sc delete "OadinService"'

  Delete "$INSTDIR\oadin.exe"
  Delete "$INSTDIR\preinstall.bat"
  Delete "$INSTDIR\postinstall.bat"
  Delete "$INSTDIR\start-oadin.bat"
  Delete "$INSTDIR\uninstall.exe"
  RMDir "$INSTDIR"

  ${un.EnvVarUpdate} $0 "PATH" "R" "HKLM" "$INSTDIR"
  System::Call 'User32::SendMessageTimeoutA(i 0xffff, i 0x1A, i 0, t "Environment", i 0, i 1000, *i .r0)'

  DeleteRegKey HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}"
  DeleteRegKey HKLM "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}"
  DeleteRegValue HKLM "Software\Microsoft\Windows\CurrentVersion\Run" "Oadin"

  ${EnableX64FSRedirection}
SectionEnd
