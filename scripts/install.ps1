# Installs the StackDrift CLI on Windows. Downloads the release binary and puts
# it on your PATH for the current user. Run with:
#   irm https://raw.githubusercontent.com/digitalaffinity-au/stackdrift-cli/main/scripts/install.ps1 | iex

$ErrorActionPreference = "Stop"

$repo = "digitalaffinity-au/stackdrift-cli"
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
$binary = "stackdrift-windows-$arch.exe"
$installDir = Join-Path $env:LOCALAPPDATA "StackDrift"
$target = Join-Path $installDir "stackdrift.exe"

Write-Host "Installing the StackDrift CLI"

New-Item -ItemType Directory -Force -Path $installDir | Out-Null

$url = "https://github.com/$repo/releases/latest/download/$binary"
Write-Host "Downloading $url"
Invoke-WebRequest -Uri $url -OutFile $target

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    Write-Host "Added $installDir to your PATH. Open a new terminal for it to take effect."
}

Write-Host "Installed to $target"
Write-Host ""
Write-Host "Next: run 'stackdrift login' then 'stackdrift scan' in a project directory."
