@echo off
setlocal
echo Building...
call build-all.cmd
if %ERRORLEVEL% neq 0 exit /b 1

echo.
echo Running tests...
go test ./... -count=1 -race
if %ERRORLEVEL% neq 0 (
    echo.
    echo TESTS FAILED
    exit /b 1
)

echo.
echo Smoke: starting server...
start /b bin\rank233-server.exe -addr 127.0.0.1:16320
timeout /t 2 /nobreak >nul

curl -sf http://127.0.0.1:16320/healthz >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo ERROR: healthz failed
    taskkill /f /im rank233-server.exe >nul 2>&1
    exit /b 1
)

curl -sf http://127.0.0.1:16320/version >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo ERROR: version failed
    taskkill /f /im rank233-server.exe >nul 2>&1
    exit /b 1
)

taskkill /f /im rank233-server.exe >nul 2>&1
echo.
echo CI PASSED
