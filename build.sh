go get ./...
echo "Successfully got all dependencies"
if [ arch == "arm" ]; then
  echo "Building for arm"
  CC=arm-linux-gnueabihf-gcc CXX=arm-linux-gnueabihf-g++ CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 go build
else
  echo "Building for different architecture"
  go build
fi