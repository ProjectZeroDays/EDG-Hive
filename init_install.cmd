@echo off
rem HIVE client helper launcher (Windows).
rem The server stack is Linux-only. This launches the PowerShell helper,
rem which preps your PC to reach the server (SSH tunnel + browser).
rem
rem To INSTALL the server, run over SSH instead:
rem   ssh root@<server> "bash <(curl -fsSL https://raw.githubusercontent.com/<you>/HIVE/main/init_install.sh)"
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0init_install.ps1" %*
