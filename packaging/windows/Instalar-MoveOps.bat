@echo off
setlocal
title Instalador MoveOps - Enterprise Data Migration Engine

echo =========================================================
echo       MoveOps - Instalador Rapido para Windows
echo =========================================================
echo.
echo Executando instalador nativo do sistema...
echo.

:: Verifica se o PowerShell esta disponivel
where powershell >nul 2>nul
if %errorlevel% neq 0 (
    echo ERRO: PowerShell nao encontrado no sistema.
    pause
    exit /b 1
)

:: Executa o script PowerShell com politicas de execucao liberadas
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0Install-MoveOps.ps1"

if %errorlevel% neq 0 (
    echo.
    echo Ocorreu um erro durante a instalacao do MoveOps.
    pause
    exit /b %errorlevel%
)

exit /b 0
