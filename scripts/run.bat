@echo off
pushd %~dp0
pushd ..
echo -- build started
go build -o ".build\issue1website.exe" -i -v "cmd\issue1.website\main.go"
echo -- build completed
echo -- enter "k" to stop the server.
echo -- enter "r" to refresh templates from disk.
.build\issue1website.exe
del .build\issue1website.exe
popd
popd