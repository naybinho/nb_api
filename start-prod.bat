@echo off
echo ===================================================
echo        Iniciando NB_Api (Modo Producao)
echo ===================================================

echo.
echo Iniciando o Servidor Unificado na porta 8080...
echo O frontend devera ser compilado previamente com "npm run build".
echo.
go run .\cmd\server -addr :8080

pause
