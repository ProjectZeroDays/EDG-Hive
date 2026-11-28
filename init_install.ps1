# HIVE client-side helper (Windows).
#
# The HIVE server stack is Linux-only (binary + docker + ollama). This script
# does NOT install the server on Windows - it prepares your PC to reach it:
#   1. checks for OpenSSH
#   2. builds the SSH local-port-forward command for your server
#   3. offers a desktop shortcut that opens the tunnel AND your browser
#
# Run the actual server install over SSH:
#   ssh root@<server> "bash <(curl -fsSL https://raw.githubusercontent.com/<you>/HIVE/main/init_install.sh)"
# or clone the repo on the server and run:  bash init_install.sh

Write-Host "HIVE client helper (Windows)" -ForegroundColor Cyan
Write-Host "The server stack runs on a Linux box; this prepares your PC to reach it.`n"

# 1. OpenSSH present? (ships with Windows 10 1809+)
$ssh = Get-Command ssh.exe -ErrorAction SilentlyContinue
if (-not $ssh) {
  Write-Host "ssh.exe not found. Install 'OpenSSH Client':"
  Write-Host "  Settings -> System -> Optional features -> Add -> OpenSSH Client" -ForegroundColor Yellow
  exit 1
}
Write-Host "[ok] OpenSSH client found"

# 2. gather server details
$server = Read-Host "Server IP or hostname (e.g. 43.106.75.221)"
$user   = Read-Host "SSH user [root]"; if (-not $user) { $user = "root" }

Write-Host "`nUse this in PowerShell to reach HIVE through the tunnel on the server:" -ForegroundColor Cyan
Write-Host "  ssh -L 8787:127.0.0.1:8787 -L 11434:127.0.0.1:11434 ${user}@${server}" -ForegroundColor White
Write-Host "then open  http://localhost:8787  (LLM API also on localhost:11434)`n"

Write-Host "Run the server-side installer (paste in the SSH session):"
Write-Host "  bash <(curl -fsSL https://raw.githubusercontent.com/<you>/HIVE/main/init_install.sh)" -ForegroundColor White

# 3. optional desktop shortcut
$mk = Read-Host "`nCreate a desktop shortcut that opens tunnel + browser? [y/N]"
if ($mk -match '^[yY]') {
  $ps = "$PSCommandPath"
  $ws = New-Object -ComObject WScript.Shell
  $sc = $ws.CreateShortcut("$env:USERPROFILE\Desktop\HIVE Tunnel.lnk")
  $sc.TargetPath       = "powershell.exe"
  $sc.Arguments        = "-NoProfile -ExecutionPolicy Bypass -File `"$ps`" -Tunnel"
  $sc.WorkingDirectory = Split-Path $ps
  $sc.Save()
  Write-Host "Shortcut created. It runs this script with -Tunnel."
}

# -Tunnel mode: open SSH tunnel, then browser; Ctrl+C to close
if ($args -contains "-Tunnel") {
  Write-Host "Opening tunnel to ${user}@${server} ... keep this window open." -ForegroundColor Cyan
  Start-Process "http://localhost:8787"
  ssh -L 8787:127.0.0.1:8787 -L 11434:127.0.0.1:11434 "${user}@${server}"
}
