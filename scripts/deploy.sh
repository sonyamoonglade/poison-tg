#!/bin/bash
ARCH=amd64
OS=linux
IP=$(echo $VM_IP)
USER=aalexandrovich

# build
rm -rf ./build
mkdir build
GOOS=$OS GOARCH=$ARCH go build -o ./build/app cmd/app/main.go
cp templates.json ./build
cp -r videos ./build
echo "building..."

# transfer build folder 
scp -r -i ~/.ssh/vadim-shop build $USER@$IP:./
echo "copying build folder"

# stop existing session
ssh -i ~/.ssh/vadim-shop $USER@$IP "kill -9 \$(pidof app)"
echo "stopped running process"

# cleanup
rm -rf ./build
