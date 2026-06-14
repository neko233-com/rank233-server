param(
    [string]$Ver = ""
)

$ErrorActionPreference = "Stop"

$Repo = "yourname/rank233-server"
$Binary = "rank233-server"
$InstallDir = if ($env:RANK233_INSTALL) { $env:RANK233_INSTALL } else { "$env:LOCALAPPDATA\rank233-server" }

function Get-Platform {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64"   { return "windows-amd64" }
        "ARM64"   { return "windows-arm64" }
        default   { Write-Error "unsupported arch: $arch"; exit 1 }
    }
}

function Get-LatestVersion {
    $resp = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    return $resp.tag_name
}

function Install-Rank233Server {
    $platform = Get-Platform

    if (-not $Ver) {
        $Ver = Get-LatestVersion
    }
    if (-not $Ver) {
        Write-Error "Could not determine version. Pass explicitly: Install-Rank233Server -Ver v0.1.0"
        exit 1
    }

    $url = "https://github.com/$Repo/releases/download/$Ver/$Binary-$platform.exe"
    Write-Host "Installing $Binary $Ver ($platform)..."
    Write-Host "  Download: $url"

    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    }

    $outFile = Join-Path $InstallDir "$Binary.exe"
    Invoke-WebRequest -Uri $url -OutFile $outFile -UseBasicParsing

    Write-Host ""
    Write-Host "Installed: $outFile"
    Write-Host ""
    Write-Host "Quick start:"
    Write-Host "  $Binary                    # start server on :6320"
    Write-Host "  $Binary -addr :8080        # custom port"
    Write-Host "  $Binary -version           # print version"
    Write-Host ""
    Write-Host "Docker:"
    Write-Host "  docker run -p 6320:6320 ghcr.io/${Repo}:${Ver}"
}

Install-Rank233Server
