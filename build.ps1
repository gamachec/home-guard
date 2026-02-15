param(
    [switch]$Dev,
    [string]$Version = "dev"
)

$env:GOOS = "windows"
$env:GOARCH = "amd64"

New-Item -ItemType Directory -Force -Path dist | Out-Null

if ($Dev) {
    go build -ldflags "-X main.version=$Version" -o dist\home-guard.exe .\cmd\agent
    go build -ldflags "-X main.version=$Version" -o dist\home-guard-updater.exe .\cmd\updater
} else {
    go build -ldflags "-H windowsgui -X main.version=$Version" -o dist\home-guard.exe .\cmd\agent
    go build -ldflags "-X main.version=$Version" -o dist\home-guard-updater.exe .\cmd\updater
}
