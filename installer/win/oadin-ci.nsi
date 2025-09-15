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
!define SERVICE_NAME "OadinService"
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
!insertmacro MUI_PAGE_FINISH

; Uninstaller pages
!insertmacro MUI_UNPAGE_WELCOME
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

; Language files
!insertmacro MUI_LANGUAGE "English"

; Force 64-bit installation - CI optimized
Function .onInit
  ; Verify 64-bit system
  ${IfNot} ${RunningX64}
    MessageBox MB_OK|MB_ICONSTOP "This application requires 64-bit Windows system."
    Abort
  ${EndIf}

  SetRegView 64
  ${DisableX64FSRedirection}

  ; Check for previous installation
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
        ; Priority 3: Hard-coded 64-bit path
        StrCpy $INSTDIR "${DEFAULT_INSTALL_DIR}"
      ${EndIf}
    ${EndIf}

    ; Validate we're not installing to x86 directory
    Push $INSTDIR
    Push "(x86)"
    Call StrStr
    Pop $R3
    ${If} $R3 != ""
      StrCpy $INSTDIR "${DEFAULT_INSTALL_DIR}"
    ${EndIf}
  ${EndIf}

  ; Log installation path for CI debugging
  DetailPrint "Target installation directory: $INSTDIR"
  IfSilent end_init
  end_init:
FunctionEnd

; Directory page validation function
Function .onVerifyInstDir
  Push $INSTDIR
  Push "(x86)"
  Call StrStr
  Pop $R0
  ${If} $R0 != ""
    MessageBox MB_OK|MB_ICONEXCLAMATION "Please select a 64-bit installation directory (not in Program Files (x86))."
    Abort
  ${EndIf}

  ; Ensure the directory is writable
  ClearErrors
  CreateDirectory "$INSTDIR"
  ${If} ${Errors}
    MessageBox MB_OK|MB_ICONEXCLAMATION "Cannot create directory. Please choose a different location or run as administrator."
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

Section "Install"
  ; Ensure 64-bit environment
  SetRegView 64
  ${DisableX64FSRedirection}

  ; Log actual installation path
  DetailPrint "Installing to: $INSTDIR"
  CreateDirectory "$INSTDIR"
  SetOutPath "$INSTDIR"

  ; Verify directory creation success
  IfFileExists "$INSTDIR" 0 install_error
  DetailPrint "SUCCESS: 64-bit installation directory created"

  ; Copy files
  File "..\..\oadin.exe"
  File "preinstall.bat"
  File "postinstall.bat"
  File "start-oadin.bat"

  ; Write registry (64-bit view)
  WriteRegStr HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "InstallDir" "$INSTDIR"
  WriteRegStr HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "Version" "${VERSION}"
  WriteRegStr HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "Architecture" "x64"
  WriteRegStr HKLM "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "DisplayName" "${APP_NAME}"
  WriteRegStr HKLM "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "UninstallString" "$INSTDIR\uninstall.exe"
  WriteRegStr HKLM "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "DisplayVersion" "${VERSION}"
  WriteRegStr HKLM "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "Publisher" "${COMPANY_NAME}"
  WriteRegStr HKLM "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}" "DisplayIcon" "$INSTDIR\oadin.exe"

  WriteUninstaller "$INSTDIR\uninstall.exe"

  ; Execute installation scripts
  DetailPrint "Running pre-install script..."
  nsExec::ExecToLog '"$INSTDIR\preinstall.bat"'

  DetailPrint "Running post-install script..."
  nsExec::ExecToLog '"$INSTDIR\postinstall.bat" "$INSTDIR"'

  ; Register Windows service
  DetailPrint "Registering Windows Service: ${SERVICE_NAME}"
  nsExec::ExecToLog 'sc.exe create "${SERVICE_NAME}" binPath= "\"$INSTDIR\oadin.exe\" server start -d" DisplayName= "Oadin Service" start= demand'

  ; Ask if user wants to start service now
  MessageBox MB_YESNO "Do you want to start the Oadin service now?" IDNO skip_start
    DetailPrint "Starting ${SERVICE_NAME}..."
    nsExec::ExecToLog 'sc.exe start "${SERVICE_NAME}"'
  skip_start:

  ; Ask if user wants to enable auto start
  MessageBox MB_YESNO "Do you want the Oadin service to start automatically at boot?" IDNO skip_autostart
    DetailPrint "Configuring ${SERVICE_NAME} for auto start..."
    nsExec::ExecToLog 'sc.exe config "${SERVICE_NAME}" start= auto'
  skip_autostart:

  ${EnableX64FSRedirection}

  DetailPrint "Installation completed successfully to: $INSTDIR"
  Goto install_end

  install_error:
  DetailPrint "ERROR: Failed to create installation directory: $INSTDIR"
  MessageBox MB_OK|MB_ICONSTOP "Installation failed: Unable to create directory $INSTDIR"
  Abort

  install_end:
SectionEnd

Function un.onInit
  SetRegView 64
  ${DisableX64FSRedirection}
FunctionEnd

Section "Uninstall"
  SetRegView 64
  ${DisableX64FSRedirection}

  ; Read installation directory from registry
  ReadRegStr $INSTDIR HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "InstallDir"
  ${If} $INSTDIR == ""
    StrCpy $INSTDIR "${DEFAULT_INSTALL_DIR}"
  ${EndIf}

  DetailPrint "Uninstalling from: $INSTDIR"

  ; Stop and delete Windows service
  DetailPrint "Stopping and deleting service ${SERVICE_NAME}..."
  nsExec::ExecToLog 'sc.exe stop "${SERVICE_NAME}"'
  nsExec::ExecToLog 'sc.exe delete "${SERVICE_NAME}"'

  Delete "$INSTDIR\oadin.exe"
  Delete "$INSTDIR\preinstall.bat"
  Delete "$INSTDIR\postinstall.bat"
  Delete "$INSTDIR\start-oadin.bat"
  Delete "$INSTDIR\uninstall.exe"

  RMDir "$INSTDIR"

  DeleteRegKey HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}"
  DeleteRegKey HKLM "SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}"

  ${EnableX64FSRedirection}
SectionEnd
