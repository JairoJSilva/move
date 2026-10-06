@echo off
setlocal
title MoveOps - Enterprise Data Migration Engine

echo =========================================================
echo   MoveOps - Enterprise Data Migration Engine v1.0.0
echo =========================================================
echo.
echo Iniciando servidor e abrindo painel de controle...

set "PORT=8082"
set "APP_DIR=%~dp0"

if exist "%APP_DIR%moveops-tray.exe" (
    echo Iniciando no Modo Bandeja do Sistema [moveops-tray.exe]...
    start "" "%APP_DIR%moveops-tray.exe" -port %PORT% -tray -open
) else if exist "%APP_DIR%moveops.exe" (
    echo Iniciando console [moveops.exe]...
    start "" "%APP_DIR%moveops.exe" -port %PORT% -open
) else if exist "%APP_DIR%hypersync.exe" (
    echo Iniciando [hypersync.exe]...
    start "" "%APP_DIR%hypersync.exe" -port %PORT% -open
) else (
    echo ERRO: Executavel do MoveOps nao encontrado em: %APP_DIR%
    pause
    exit /b 1
)

echo.
echo MoveOps iniciado com sucesso na porta %PORT%!
echo O painel sera aberto automaticamente no seu navegador padrao.
echo Caso nao abra, acesse manualmente: http://localhost:%PORT%
echo.
timeout /t 3 >nul
exit /b 0
