@echo off
echo ===================================================
echo             Iniciando NB_Api
echo ===================================================

echo.
echo Iniciando o Backend (Go) na porta 8080...
start "NB_Api Backend" cmd /k "go run .\cmd\server -addr :8080"

echo.
echo Iniciando o Frontend (React Dev Server)...
start "NB_Api Frontend" cmd /k "cd client && npm run dev"

echo.
echo Tudo iniciado! Duas novas janelas do terminal foram abertas.
echo Backend disponivel em: http://localhost:8080
echo Frontend disponivel em: http://localhost:5173 (modo dev)
echo.
pause
