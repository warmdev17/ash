#define MyAppName "ash"
#define MyCompany "warmdev"
#define MyAppVersion GetStringParam("MyAppVersion", "1.0.0")
#define InputExe GetStringParam("InputExe", "ash-windows-amd64.exe")
#define OutputDir GetStringParam("OutputDir", "..\..\dist")

[Setup]
AppId={{6B5B1F8E-5E2C-4D49-BF7A-9F9C1A9EDEAD}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyCompany}
DefaultDirName={pf}\Ash
DefaultGroupName=Ash
OutputDir={#OutputDir}
OutputBaseFilename=ash-installer
ArchitecturesInstallIn64BitMode=x64
UninstallDisplayIcon={app}\ash.exe
DisableDirPage=yes
DisableProgramGroupPage=yes
Compression=lzma
SolidCompression=yes
WizardStyle=modern
; Enable silent install (required by winget)
SilentInstall=Yes
DisableStartupPrompt=yes

[Files]
Source: "{#OutputDir}\{#InputExe}"; DestDir: "{app}"; DestName: "ash.exe"; Flags: ignoreversion

[Icons]
Name: "{group}\Ash CLI"; Filename: "{app}\ash.exe"

[Run]
; Optional: Run after install when not silent
Filename: "{app}\ash.exe"; Description: "Run ash"; Flags: nowait postinstall skipifsilent

[Registry]
; Add to system PATH (machine-wide) if not present
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; \
    ValueType: expandsz; ValueName: "Path"; \
    ValueData: "{olddata};{app}"; \
    Check: NeedsAddPath('{app}'); Flags: preservestringtype

[Code]
function NeedsAddPath(AppDir: string): Boolean;
var
  OrigPath: string;
begin
  if not RegQueryStringValue(HKLM, 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment',
    'Path', OrigPath) then
  begin
    Result := True;
    exit;
  end;
  if Pos(';' + UpperCase(AppDir) + ';', ';' + UpperCase(OrigPath) + ';') > 0 then
    Result := False
  else
    Result := True;
end
