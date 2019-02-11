#!/bin/sh

if [ -z "$1" ]
  then
    echo "A release version need to be specified"
fi

# clean/prepare
rm -rf .releases
mkdir -p .releases/linux/cloudsql-proxy-inject-$1/bin
mkdir -p .releases/darwin/cloudsql-proxy-inject-$1/bin

# build
make build-linux
mv ./cloudsql-proxy-inject-linux-amd64 .releases/linux/cloudsql-proxy-inject-$1/bin
make build-darwin
mv ./cloudsql-proxy-inject-darwin-amd64 .releases/darwin/cloudsql-proxy-inject-$1/bin

# compress
cd .releases/linux && tar -czvf cloudsql-proxy-inject-$1-linux.tar.gz cloudsql-proxy-inject-$1/bin && cd ../..
cd .releases/darwin && tar -czvf cloudsql-proxy-inject-$1-darwin.tar.gz cloudsql-proxy-inject-$1/bin && cd ../..

