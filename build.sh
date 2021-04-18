# /bin/sh
echo "Building darwin"
GOOS=darwin GOARCH=amd64 go build
mv "${PWD##*/}" "${PWD##*/}_darwin_amd64"

echo "Building linux"
GOOS=linux GOARCH=amd64 go build
mv "${PWD##*/}" "${PWD##*/}_linux_amd64"
GOOS=linux GOARCH=arm go build
mv "${PWD##*/}" "${PWD##*/}_linux_arm"
GOOS=linux GOARCH=arm64 go build
mv "${PWD##*/}" "${PWD##*/}_linux_arm64"

echo "Building freebsd"
GOOS=freebsd GOARCH=amd64 go build
mv "${PWD##*/}" "${PWD##*/}_freebsd_amd64"

echo "Building windows"
GOOS=windows GOARCH=amd64 go build


echo "Completed"
