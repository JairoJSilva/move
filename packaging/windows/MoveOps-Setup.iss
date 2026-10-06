; =====================================================================
; Script Inno Setup Profissional para MoveOps - Enterprise Data Migration Engine
; Wizard Clássico "Avançar, Avançar, Concluir"
; =====================================================================

#define MyAppName "MoveOps"
#define MyAppVersion "1.0.0"
#define MyAppPublisher "MoveOps Engineering Team"
#define MyAppURL "https://moveops.io"
#define MyAppExeName "moveops-tray.exe"
#define MyAppConsoleExeName "moveops.exe"

[Setup]
AppId={{9F3B1A2C-4E5D-6F7A-8B9C-0D1E2F3A4B5C}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
AllowNoIcons=yes
OutputDir=..\..\dist
OutputBaseFilename=MoveOps-Setup
SetupIconFile=moveops.ico
UninstallDisplayIcon={app}\moveops.ico
Compression=lzma2/ultra64
SolidCompression=yes
WizardStyle=modern
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
PrivilegesRequired=admin
DisableDirPage=no
DisableProgramGroupPage=no

[Languages]
Name: "brazilianportuguese"; MessagesFile: "compiler:Languages\BrazilianPortuguese.isl"
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
Name: "startuptask"; Description: "Iniciar automaticamente com o Windows (Ícone na Bandeja do Sistema)"; GroupDescription: "Configurações do Sistema:"; Flags: unchecked
Name: "firewalltask"; Description: "Permitir conexões no Firewall do Windows (Portas 8080 e 8082)"; GroupDescription: "Configurações de Rede:"; Flags: checked

[Files]
Source: "..\..\bin\moveops-tray.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\bin\moveops.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\bin\hypersync.exe"; DestDir: "{app}"; Flags: ignoreversion; Tasks: 
Source: "moveops.ico"; DestDir: "{app}"; Flags: ignoreversion
Source: "moveops.png"; DestDir: "{app}"; Flags: ignoreversion
Source: "Iniciar-MoveOps.bat"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\MoveOps (Bandeja do Sistema)"; Filename: "{app}\{#MyAppExeName}"; Parameters: "-tray"; IconFilename: "{app}\moveops.ico"
Name: "{group}\MoveOps Console"; Filename: "{app}\{#MyAppConsoleExeName}"; Parameters: "-open"; IconFilename: "{app}\moveops.ico"
Name: "{group}\Abrir Painel MoveOps no Navegador"; Filename: "{app}\Iniciar-MoveOps.bat"; IconFilename: "{app}\moveops.ico"
Name: "{group}\{cm:UninstallProgram,{#MyAppName}}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Parameters: "-tray"; IconFilename: "{app}\moveops.ico"; Tasks: desktopicon
Name: "{autostartup}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Parameters: "-tray"; IconFilename: "{app}\moveops.ico"; Tasks: startuptask

[Run]
Filename: "netsh"; Parameters: "advfirewall firewall add rule name=""MoveOps Migration Engine"" dir=in action=allow protocol=TCP localport=8080,8082"; Flags: runhidden; Tasks: firewalltask
Filename: "{app}\{#MyAppExeName}"; Parameters: "-tray"; Description: "{cm:LaunchProgram,{#StringChange(MyAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent

[UninstallRun]
Filename: "netsh"; Parameters: "advfirewall firewall delete rule name=""MoveOps Migration Engine"""; Flags: runhidden; RunOnceId: "RemoveMoveOpsFirewallRule"
