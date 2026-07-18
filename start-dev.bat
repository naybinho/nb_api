@echo off
title NB_Api Dev Server
echo ===================================================
echo   NB_Api - Modo Desenvolvimento (HMR + Proxy)
echo ===================================================
echo.
echo Iniciando o Backend (Go) na porta 8081...
echo.
start "NB_Api Backend" cmd /k "go run .\cmd\server -addr :8081"

echo Iniciando o Frontend (Vite Dev Server) na porta 5173...
echo.
cd client
start "NB_Api Frontend" cmd /k "npm run dev"
cd ..

echo.
echo Tudo iniciado!
echo Frontend disponivel em: http://localhost:5173 (com HMR)
echo Backend  disponivel em: http://localhost:8081
echo.
echo O Vite faz proxy das requisicoes /api para o backend na porta 8081.
echo.
pause
