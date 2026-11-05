@echo off
setlocal

:: Путь к корню проекта
set "ROOT=%~dp0"

echo ====================================================
echo   ConstructMaterialAI Launcher
echo   Project Root: %ROOT%
echo ====================================================

:: 1. Backend
echo [1/3] Launching Backend...
start "Backend Server" /D "%ROOT%backend" cmd /k go run cmd/service/main.go

:: 2. ML Service
echo [2/3] Launching ML Service...
if exist "%ROOT%data-service\.venv\Scripts\activate.bat" (
    start "ML Data Service" /D "%ROOT%data-service" cmd /k ".venv\Scripts\activate.bat && uvicorn api.api:app --reload"
) else (
    start "ML Data Service" /D "%ROOT%data-service" cmd /k uvicorn api.api:app --reload
)

:: 3. Desktop App
echo [3/3] Launching Desktop App...
start "Desktop App" /D "%ROOT%desktop" cmd /k go run .

echo.
echo ====================================================
echo   All services launched in separate windows.
echo ====================================================
pause
