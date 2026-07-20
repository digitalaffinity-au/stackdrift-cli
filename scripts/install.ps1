# Installs the StackDrift CLI on Windows. Downloads the release binary and
# places it in a directory that is already on your PATH, so no environment
# variable changes are needed. Run with:
#   irm https://raw.githubusercontent.com/digitalaffinity-au/stackdrift-cli/main/scripts/install.ps1 | iex

$ErrorActionPreference = "Stop"

$repo = "digitalaffinity-au/stackdrift-cli"
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
$binary = "stackdrift-windows-$arch.exe"
$url = "https://github.com/$repo/releases/latest/download/$binary"

Write-Host "Installing the StackDrift CLI"

function Test-OnPath($dir) {
    $target = $dir.TrimEnd('\')
    foreach ($part in ($env:Path -split ';')) {
        if ($part.TrimEnd('\') -ieq $target) { return $true }
    }
    return $false
}

$windowsApps = Join-Path $env:LOCALAPPDATA "Microsoft\WindowsApps"

if ((Test-Path $windowsApps) -and (Test-OnPath $windowsApps)) {
    $target = Join-Path $windowsApps "stackdrift.exe"
    Write-Host "Downloading $url"
    Invoke-WebRequest -Uri $url -OutFile $target
    Write-Host "Installed to $target"
    Write-Host "That directory is already on your PATH."
}
else {
    $installDir = Join-Path $env:LOCALAPPDATA "StackDrift"
    New-Item -ItemType Directory -Force -Path $installDir | Out-Null
    $target = Join-Path $installDir "stackdrift.exe"
    Write-Host "Downloading $url"
    Invoke-WebRequest -Uri $url -OutFile $target
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$installDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
        Write-Host "Added $installDir to your PATH. Open a new terminal for it to take effect."
    }
    Write-Host "Installed to $target"
}

Write-Host ""
Write-Host "Next: run 'stackdrift login' then 'stackdrift scan' in a project directory."
