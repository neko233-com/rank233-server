@echo off
setlocal
echo Running tests...
go test ./... -count=1 -v
if %ERRORLEVEL% neq 0 (
    echo.
    echo TESTS FAILED
    exit /b 1
)
echo.
echo ALL TESTS PASSED
