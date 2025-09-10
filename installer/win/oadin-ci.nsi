!ifndef VERSION
  !define VERSION "0.0.0"
!endif

; CI/CD Compatible NSIS Script for 64-bit Installation
; Designed to work with 32-bit NSIS compiler in CI environment

; Include 64-bit support libraries
!include "x64.nsh"
!include "LogicLib.nsh"

!define APP_NAME "Oadin CLI"
!define COMPANY_NAME "Digital China"
; Use hard-coded 64-bit path to avoid CI environment issues
!define INSTALL_DIR "C:\Program Files\Oadin"

Outfile "..\..\oadin-installer.exe"
InstallDir "${INSTALL_DIR}"
RequestExecutionLevel admin
SetCompress auto
SetCompressor lzma

Name "${APP_NAME}"
Caption "${APP_NAME} ${VERSION} Setup"

; Force 64-bit installation - CI optimized
Function .onInit
  ; Verify 64-bit system
  ${IfNot} ${RunningX64}
    MessageBox MB_OK|MB_ICONSTOP "This application requires 64-bit Windows system."
    Abort
  ${EndIf}

  ; Enable 64-bit mode
  SetRegView 64
  ${DisableX64FSRedirection}

  ; Force 64-bit Program Files - CI compatible approach
  ; Priority 1: Use PROGRAMFILES64 if available
  StrCpy $R0 "$PROGRAMFILES64"
  ${If} $R0 != ""
    ${AndIf} $R0 != "\$PROGRAMFILES64"
    StrCpy $INSTDIR "$R0\Oadin"
  ${Else}
    ; Priority 2: Use ProgramW6432 environment variable
    ReadEnvStr $R1 "ProgramW6432"
    ${If} $R1 != ""
      StrCpy $INSTDIR "$R1\Oadin"
    ${Else}
      ; Priority 3: Hard-coded 64-bit path
      StrCpy $INSTDIR "C:\Program Files\Oadin"
    ${EndIf}
  ${EndIf}

  ; Validate we're not installing to x86 directory
  Push $INSTDIR
  Push "(x86)"
  Call StrStr
  Pop $R2
  ${If} $R2 != ""
    ; Found (x86) in path, force correct path
    StrCpy $INSTDIR "C:\Program Files\Oadin"
  ${EndIf}

  ; Log installation path for CI debugging
  DetailPrint "Target installation directory: $INSTDIR"
  
  ; No MessageBox - silent installation
FunctionEnd

; String search function
Function StrStr
  Exch $R1 ; st=haystack,old$R1, $R1=needle
  Exch    ; st=old$R1,haystack, $R1=needle
  Exch $R2 ; st=old$R1,old$R2, $R2=haystack, $R1=needle
  Push $R3
  Push $R4
  Push $R5
  StrLen $R3 $R1
  StrCpy $R4 0
  ; $R1=needle
  ; $R2=haystack
  ; $R3=len(needle)
  ; $R4=cnt
  ; $R5=tmp
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
  DetailPrint "PROGRAMFILES64: $PROGRAMFILES64"
  DetailPrint "PROGRAMFILES: $PROGRAMFILES"

  ; Create installation directory
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
  WriteRegStr HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}" "UninstallString" "$INSTDIR\uninstall.exe"
  
  WriteUninstaller "$INSTDIR\uninstall.exe"

  ; Execute installation scripts
  DetailPrint "Running pre-install script..."
  nsExec::ExecToLog '"$INSTDIR\preinstall.bat"'

  DetailPrint "Running post-install script..."
  nsExec::ExecToLog '"$INSTDIR\postinstall.bat" "$INSTDIR"'

  DetailPrint "Starting Oadin service..."
  nsExec::ExecToLog '"$INSTDIR\start-oadin.bat"'

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
  
  Delete "$INSTDIR\oadin.exe"
  Delete "$INSTDIR\preinstall.bat"
  Delete "$INSTDIR\postinstall.bat"
  Delete "$INSTDIR\start-oadin.bat"
  Delete "$INSTDIR\uninstall.exe"

  RMDir "$INSTDIR"

  DeleteRegKey HKLM "SOFTWARE\${COMPANY_NAME}\${APP_NAME}"

  ${EnableX64FSRedirection}
SectionEnd
