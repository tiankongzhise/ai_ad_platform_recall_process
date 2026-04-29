
<#
powershell 单行编译指令

$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED=0; go build -ldflags="-s -w" -trimpath -o ad_platform_recall_service cmd/server/main.go

#>

$env:GOOS="linux"
$env:GOARCH="amd64"
$env:CGO_ENABLED=0
go build -ldflags="-s -w" -trimpath -o ad_platform_recall_service cmd/server/main.go