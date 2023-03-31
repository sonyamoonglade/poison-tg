#!/bin/bash

export $(xargs < .env)
# go build -o ./build/app cmd/app/main.go
nohup ./build/app -strict=false -config-path=./config.yml >/dev/null 2>&1 &
