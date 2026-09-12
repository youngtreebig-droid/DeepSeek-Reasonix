# Lean NSIS installer for the reduced Windows 7 Reasonix CLI.
#
# This is intentionally minimal and independent of the heavy desktop installer
# pipeline (desktop/build/windows/installer/project.nsi). It installs the
# single CLI executable into Program Files and adds it to the system PATH.
#
# Build with: makensis -DVERSION=<version> scripts/win7-installer.nsi
# (invoked automatically by scripts/build-win7.sh when makensis is present).
# Run from the repository root so the relative paths below resolve.

!ifndef VERSION
  !define VERSION "dev"
!endif

!include "MUI2.nsh"

Name "Reasonix (Windows 7 CLI)"
OutFile "dist\reasonix-win7-setup.exe"
Unicode true
InstallDir "$PROGRAMFILES64\Reasonix"
InstallDirRegKey HKLM "Software\Reasonix" "InstallDir"
RequestExecutionLevel admin

# EnvVarUpdate helpers require the NSIS environment; we edit PATH via the
# registry and broadcast the change so new shells pick it up.
!define REG_ENV 'HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"'

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_LANGUAGE "English"

Section "Reasonix CLI" SecCLI
  SetOutPath "$INSTDIR"
  File "dist\reasonix-win7-amd64.exe"
  File "dist\README-WIN7.txt"

  WriteRegStr HKLM "Software\Reasonix" "InstallDir" "$INSTDIR"
  WriteUninstaller "$INSTDIR\uninstall.exe"

  # Register in Add/Remove Programs.
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Reasonix" \
    "DisplayName" "Reasonix (Windows 7 CLI)"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Reasonix" \
    "DisplayVersion" "${VERSION}"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Reasonix" \
    "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Reasonix" \
    "InstallLocation" "$INSTDIR"

  # Add the install dir to the system PATH (append if not already present).
  ReadRegStr $0 ${REG_ENV} "Path"
  StrCpy $1 "$0;$INSTDIR"
  WriteRegExpandStr ${REG_ENV} "Path" "$1"
  SendMessage ${HWND_BROADCAST} ${WM_WININICHANGE} 0 "STR:Environment" /TIMEOUT=5000
SectionEnd

Section "Uninstall"
  Delete "$INSTDIR\reasonix-win7-amd64.exe"
  Delete "$INSTDIR\README-WIN7.txt"
  Delete "$INSTDIR\uninstall.exe"
  RMDir "$INSTDIR"

  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Reasonix"
  DeleteRegKey HKLM "Software\Reasonix"
SectionEnd
