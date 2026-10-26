@echo off
REM SMS Server Deploy Batch Wrapper
REM This script launches the PowerShell deployment script

setlocal enabledelayedexpansion

set "SCRIPT_DIR=%~dp0"
cd "%SCRIPT_DIR%"

powershell -NoProfile -ExecutionPolicy Bypass -File "!SCRIPT_DIR!deploy.ps1" %*

set "EXIT_CODE=!ERRORLEVEL!"
exit /b !EXIT_CODE!
