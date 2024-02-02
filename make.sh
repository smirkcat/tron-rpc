#!/bin/sh
COMMIT_HASH=`git rev-parse HEAD 2>/dev/null` 
BUILD_DATE=`(date "+%Y-%m-%d %H:%M:%S")`
# export GOPROXY="https://goproxy.io"
go build -trimpath -ldflags "-w -s"  -ldflags "-X \"main.BuildVersion=${COMMIT_HASH}\" -X \"main.BuildDate=${BUILD_DATE}\""