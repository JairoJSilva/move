<#
.SYNOPSIS
    Instalador 1-Click Nativo do MoveOps para Windows (PowerShell)
.DESCRIPTION
    Instala os binários do MoveOps em C:\Program Files\MoveOps, configura atalhos na
    Área de Trabalho e Menu Iniciar, cria regras no Windows Firewall e inicializa o serviço.
#>

[CmdletBinding()]
param(
    [string]$InstallDir = "$env:ProgramFiles\MoveOps",
    [int]$Port = 8082,
    [switch]$StartOnBoot = $false
)

# Garante privilégios de Administrador
$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Host "[!] Elevando privilégios para Administrador..." -ForegroundColor Yellow
    Start-Process powershell.exe -Verb RunAs -ArgumentList ("-NoProfile -ExecutionPolicy Bypass -File `"$PSCommandPath`"")
    exit
}

Write-Host "=========================================================" -ForegroundColor Cyan
Write-Host "       MoveOps - Instalador Oficial do Sistema           " -ForegroundColor Cyan
Write-Host "=========================================================" -ForegroundColor Cyan
Write-Host ""

$SourceDir = $PSScriptRoot
if (-not (Test-Path "$SourceDir\moveops.exe")) {
    # Tenta localizar a partir da pasta bin se rodando a partir da raiz do repositório
    if (Test-Path "$SourceDir\..\..\bin\moveops.exe") {
        $SourceDir = (Resolve-Path "$SourceDir\..\..\bin").Path
    }
}

Write-Host "==> Criando diretório de instalação: $InstallDir" -ForegroundColor Green
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# Arquivos a serem copiados
$filesToCopy = @(
    "moveops.exe",
    "moveops-tray.exe",
    "hypersync.exe",
    "moveops.ico",
    "moveops.png",
    "Iniciar-MoveOps.bat"
)

foreach ($f in $filesToCopy) {
    $srcPath = Join-Path $SourceDir $f
    if (-not (Test-Path $srcPath)) {
        # Busca nas pastas vizinhas
        $candidateRoot = (Resolve-Path "$PSScriptRoot\..\..").Path
        if (Test-Path "$candidateRoot\bin\$f") { $srcPath = "$candidateRoot\bin\$f" }
        elseif (Test-Path "$candidateRoot\packaging\windows\$f") { $srcPath = "$candidateRoot\packaging\windows\$f" }
        elseif (Test-Path "$candidateRoot\$f") { $srcPath = "$candidateRoot\$f" }
    }

    if (Test-Path $srcPath) {
        Copy-Item -Path $srcPath -Destination $InstallDir -Force
        Write-Host "    [+] Copiado: $f" -ForegroundColor Gray
    }
}

# 1. Configura Regra no Firewall do Windows
Write-Host "==> Configurando regras de rede no Firewall do Windows..." -ForegroundColor Green
try {
    netsh advfirewall firewall delete rule name="MoveOps Migration Engine" 2>$null | Out-Null
    netsh advfirewall firewall add rule name="MoveOps Migration Engine" dir=in action=allow protocol=TCP localport="8080,$Port" | Out-Null
    Write-Host "    [OK] Regra de entrada TCP (Portas 8080, $Port) autorizada." -ForegroundColor Green
} catch {
    Write-Host "    [!] Aviso ao configurar firewall: $_" -ForegroundColor Yellow
}

# 2. Cria Atalho na Área de Trabalho
Write-Host "==> Criando atalhos do sistema..." -ForegroundColor Green
$WshShell = New-Object -ComObject WScript.Shell
$DesktopPath = [Environment]::GetFolderPath("Desktop")
$DesktopShortcutPath = Join-Path $DesktopPath "MoveOps.lnk"

$Shortcut = $WshShell.CreateShortcut($DesktopShortcutPath)
if (Test-Path "$InstallDir\moveops-tray.exe") {
    $Shortcut.TargetPath = "$InstallDir\moveops-tray.exe"
    $Shortcut.Arguments = "-port $Port -tray"
} else {
    $Shortcut.TargetPath = "$InstallDir\moveops.exe"
    $Shortcut.Arguments = "-port $Port -open"
}
$Shortcut.WorkingDirectory = $InstallDir
if (Test-Path "$InstallDir\moveops.ico") {
    $Shortcut.IconLocation = "$InstallDir\moveops.ico, 0"
}
$Shortcut.Description = "MoveOps - Enterprise Data Migration Engine"
$Shortcut.Save()
Write-Host "    [OK] Atalho criado na Área de Trabalho." -ForegroundColor Green

# 3. Cria Atalho no Menu Iniciar
$StartMenuPath = [Environment]::GetFolderPath("CommonPrograms")
$MoveOpsProgramsPath = Join-Path $StartMenuPath "MoveOps"
if (-not (Test-Path $MoveOpsProgramsPath)) {
    New-Item -ItemType Directory -Path $MoveOpsProgramsPath -Force | Out-Null
}

$StartShortcutPath = Join-Path $MoveOpsProgramsPath "MoveOps.lnk"
$StartShortcut = $WshShell.CreateShortcut($StartShortcutPath)
if (Test-Path "$InstallDir\moveops-tray.exe") {
    $StartShortcut.TargetPath = "$InstallDir\moveops-tray.exe"
    $StartShortcut.Arguments = "-port $Port -tray"
} else {
    $StartShortcut.TargetPath = "$InstallDir\moveops.exe"
    $StartShortcut.Arguments = "-port $Port -open"
}
$StartShortcut.WorkingDirectory = $InstallDir
if (Test-Path "$InstallDir\moveops.ico") {
    $StartShortcut.IconLocation = "$InstallDir\moveops.ico, 0"
}
$StartShortcut.Description = "MoveOps - Enterprise Data Migration Engine"
$StartShortcut.Save()
Write-Host "    [OK] Atalho criado no Menu Iniciar." -ForegroundColor Green

# 4. Inicialização com o Windows (se solicitado)
if ($StartOnBoot) {
    $StartupDir = [Environment]::GetFolderPath("CommonStartup")
    $StartupShortcutPath = Join-Path $StartupDir "MoveOps.lnk"
    $StartupShortcut = $WshShell.CreateShortcut($StartupShortcutPath)
    $StartupShortcut.TargetPath = "$InstallDir\moveops-tray.exe"
    $StartupShortcut.Arguments = "-port $Port -tray"
    $StartupShortcut.WorkingDirectory = $InstallDir
    if (Test-Path "$InstallDir\moveops.ico") {
        $StartupShortcut.IconLocation = "$InstallDir\moveops.ico, 0"
    }
    $StartupShortcut.Save()
    Write-Host "    [OK] Inicialização automática na bandeja ativada com o Windows." -ForegroundColor Green
}

Write-Host ""
Write-Host "=========================================================" -ForegroundColor Cyan
Write-Host "       Instalação do MoveOps Concluída com Sucesso!      " -ForegroundColor Green
Write-Host "=========================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Iniciando MoveOps e abrindo painel no navegador..." -ForegroundColor Yellow

if (Test-Path "$InstallDir\moveops-tray.exe") {
    Start-Process "$InstallDir\moveops-tray.exe" -ArgumentList "-port $Port -tray -open" -WorkingDirectory $InstallDir
} else {
    Start-Process "$InstallDir\moveops.exe" -ArgumentList "-port $Port -open" -WorkingDirectory $InstallDir
}

Start-Sleep -Seconds 2
