$ErrorActionPreference = "Stop"

$version = if ($env:VERSION) { $env:VERSION } else { "dev" }
$commit = try { git rev-parse --short HEAD 2>$null } catch { "unknown" }
$date = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

$ldflags = "-s -w -X github.com/yourname/rank233-server/internal/version.Version=$version -X github.com/yourname/rank233-server/internal/version.Commit=$commit -X github.com/yourname/rank233-server/internal/version.Date=$date"

New-Item -ItemType Directory -Force -Path bin | Out-Null
Write-Host "Building rank233-server..."
go build -ldflags="$ldflags" -o bin\rank233-server.exe .\cmd\rank233-server
if ($LASTEXITCODE -ne 0) { exit 1 }
Write-Host "Done: bin\rank233-server.exe"
