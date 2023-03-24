#!/bin/bash

IP=$(echo $VM_IP)
USER=aalexandrovich

# build
rm -rf ./build
mkdir build
go build -o ./build/bot cmd/bot/main.go
cp templates.json ./build

echo "building..."

# stop existing session
ssh -i ~/.ssh/vadim-shop $USER@$IP "kill -9 \$(pidof bot)"

echo "stopped running process"

# transfer build folder 
scp -r -i ~/.ssh/vadim-shop build $USER@$IP:./

echo "copying build folder"

# run new version
ssh -i ~/.ssh/vadim-shop $USER@$IP "./run.sh </dev/null >~/backup-log.log 2>&1 &" 

echo "started new process"

# cleanup
rm -rf ./build
