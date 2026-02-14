param(
    [switch]$Dev
)

$env:GOOS = "windows"
$env:GOARCH = "amd64"

New-Item -ItemType Directory -Force -Path dist | Out-Null

if ($Dev) {
    go build -o dist\home-guard.exe .\cmd\agent
} else {
    go build -ldflags "-H windowsgui" -o dist\home-guard.exe .\cmd\agent
}
