@echo off
pushd %~dp0
pushd ..
echo -- build started
go build -o ".build\issue1website.exe" -i -v "cmd\issue1.website\main.go"
echo -- build completed
echo -- enter "kill" to stop the server.
.build\issue1website.exe
popd
popd