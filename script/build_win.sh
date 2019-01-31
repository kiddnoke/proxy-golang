#!/bin/bash

COMMIT_HASH=`git rev-parse HEAD 2>/dev/null`
BUILD_DATE=`date  +%Y-%m-%d-%H:%M`

TARGET=../../bin/vpnedge.exe
SOURCE=../cmd/VpnEdge
cd ${SOURCE}
go build -ldflags "-X \"main.BuildVersion=${COMMIT_HASH}\" -X \"main.BuildDate=${BUILD_DATE}\"" -i -o ${TARGET}