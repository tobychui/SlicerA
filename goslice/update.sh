#/bin/bash
echo "Pulling latest update from GoSlice repo"
cd GoSlice/
git pull

echo "Compile started"
cd cmd/goslice/

echo "Building darwin"
GOOS=darwin GOARCH=amd64 go build
mv goslice ../../../goslice-macos-amd64.elf

echo "Building linux"
GOOS=linux GOARCH=amd64 go build
mv goslice ../../../goslice-linux-amd64.elf
GOOS=linux GOARCH=arm GOARM=6 go build
mv goslice ../../../goslice-linux-arm.elf
GOOS=linux GOARCH=arm GOARM=7 go build
mv goslice ../../../goslice-linux-armv7.elf
GOOS=linux GOARCH=arm64 go build
mv goslice ../../../goslice-linux-arm64.elf

echo "Building windows"
GOOS=windows GOARCH=amd64 go build
mv goslice.exe ../../../goslice-windows-amd64.exe

echo "Completed"