#!/bin/bash
OS="linux"
export GOOS=${OS}
export GOARCH=amd64
export CGO_ENABLED=0

COMMIT_HASH=`git rev-parse HEAD 2>/dev/null`
BUILD_DATE=`date  +%Y-%m-%d-%H:%M`
BRANCH_NAME=`git symbolic-ref --short -q HEAD`

TARGET=../../bin/vpnedge_${OS}
SOURCE=../cmd/VpnEdge
cd ${SOURCE}
go build -ldflags "-X \"main.BuildVersion=${COMMIT_HASH}\" -X \"main.BuildDate=${BUILD_DATE}\" -X \"main.BuildBranch=${BRANCH_NAME}\"" -i -o ${TARGET}