@echo off
REM install.ps1 - Install kube-triage on Windows

param(
    [string]$InstallDir = "$env:USERPROFILE\.local\bin",
    [switch]$AddToPath
)

$ErrorActionPreference = "Stop"

$REPO = "Jani-shiv/kube-triage"
$BINARY = "kube-triage.exe"

function Write-Log {
    param($Message, $Level = "INFO")
    $timestamp = Get-Date -Format "HH:mm:ss"
    $color = switch ($Level) {
        "SUCCESS" { "Green" }
        "WARN"    { "Yellow" }
        "ERROR"   { "Red" }
        default   { "Cyan" }
    }
    Write-Host "[$timestamp] [$Level] $Message" -ForegroundColor $color
}

function Get-LatestVersion {
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$REPO/releases/latest" -ErrorAction Stop
        return $response.tag_name
    } catch {
        Write-Log "Failed to get latest version: $($_.Exception.Message)" "ERROR"
        exit 1
    }
}

function Download-AndInstall {
    param($Version, $InstallDir)

    $osArch = "windows_amd64"
    $archiveName = "${BINARY}_${Version}_${osArch}.zip"
    $downloadUrl = "https://github.com/$REPO/releases/download/$Version/$archiveName"
    $tempDir = [System.IO.Path]::GetTempPath()
    $archivePath = Join-Path $tempDir $archiveName

    Write-Log "Downloading $archiveName..."
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $archivePath -ErrorAction Stop
    } catch {
        Write-Log "Download failed: $($_.Exception.Message)" "ERROR"
        exit 1
    }

    Write-Log "Extracting..."
    $extractDir = Join-Path $tempDir "kube-triage-extract"
    if (Test-Path $extractDir) { Remove-Item $extractDir -Recurse -Force }
    New-Item -ItemType Directory -Path $extractDir | Out-Null

    Expand-Archive -Path $archivePath -DestinationPath $extractDir -Force

    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    $binarySource = Join-Path $extractDir $BINARY
    $binaryDest = Join-Path $InstallDir $BINARY

    Write-Log "Installing to $InstallDir..."
    Copy-Item $binarySource $binaryDest -Force

    Write-Log "Cleaning up..."
    Remove-Item $archivePath -Force
    Remove-Item $extractDir -Recurse -Force

    Write-Log "Installed $BINARY $Version to $InstallDir" "SUCCESS"
}

function Add-ToUserPath {
    param($InstallDir)
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -notlike "*$InstallDir*") {
        $newPath = $currentPath + ";" + $InstallDir
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-Log "Added $InstallDir to user PATH" "SUCCESS"
    }
}

# Main
Write-Log "Installing kube-triage..."

$version = Get-LatestVersion
if (-not $version) {
    Write-Log "Failed to get latest version" "ERROR"
    exit 1
}

Write-Log "Latest version: $version"
Download-AndInstall $version $InstallDir

if ($AddToPath) {
    Add-ToUserPath $InstallDir
}

Write-Log "Installation complete!" "SUCCESS"
Write-Host ""
Write-Host "Run 'kube-triage --help' to get started." -ForegroundColor Cyan
Write-Host ""
Write-Host "Quick start:" -ForegroundColor Cyan
Write-Host "  kube-triage analyze pod <pod-name> -n <namespace>" -ForegroundColor Cyan
Write-Host ""
Write-Host "For AI-powered analysis, set one of these environment variables:" -ForegroundColor Cyan
Write-Host "  `$env:NVIDIA_API_KEY='nvapi-xxxxx'     # NVIDIA Nemotron (free tier)" -ForegroundColor Cyan
Write-Host "  `$env:OLLAMA_HOST='http://localhost:11434'  # Local Ollama" -ForegroundColor Cyan
Write-Host "  `$env:OPENAI_API_KEY='sk-xxxxx'        # OpenAI" -ForegroundColor Cyan
Write-Host "  `$env:ANTHROPIC_API_KEY='sk-ant-xxxxx' # Anthropic" -ForegroundColor Cyan