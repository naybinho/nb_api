@echo off
title NB_Api Server
echo ===================================================
echo     NB_Api - Modo Unificado (porta unica 8080)
echo ===================================================
echo.
echo Compilando o Frontend...
echo.
cd client
call npm run build
if %errorlevel% neq 0 (
    echo [ERRO] Falha ao compilar o frontend.
    pause
    exit /b 1
)
cd ..

echo.
echo Iniciando o Servidor Unificado na porta 8080...
echo Frontend + Backend disponiveis em: http://localhost:8080
echo.
go run .\cmd\server -addr :8080

pause
