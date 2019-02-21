#!/bin/bash
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

COMMIT_HASH=`git rev-parse HEAD 2>/dev/null`
BUILD_DATE=`date  +%Y-%m-%d-%H:%M`
TARGET=../../bin/vpnedge_linux_release
SOURCE=../cmd/VpnEdge
cd ${SOURCE}
go build -ldflags "-X \"main.BuildVersion=${COMMIT_HASH}\" -X \"main.BuildDate=${BUILD_DATE}\"" -i -o ${TARGET}