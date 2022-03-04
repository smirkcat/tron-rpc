#!/bin/sh

COMMIT_HASH=`git rev-parse HEAD 2>/dev/null` 
BUILD_DATE=`(date "+%Y-%m-%d %H:%M:%S")`
# TARGET=./bin/robot 
# SOURCE=./main.go
#-o ${TARGET} ${SOURCE} 
# export GOPROXY="https://goproxy.io"
export GOPATH="/mnt/d/go:/home/smirkcat/go"
go build -trimpath -ldflags "-X \"main.BuildVersion=${COMMIT_HASH}\" -X \"main.BuildDate=${BUILD_DATE}\""