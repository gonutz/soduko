set GOOS=windows
set GOARCH=386
go build -ldflags="-H=windowsgui -s -w"
if errorlevel 1 pause
