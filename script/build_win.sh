#!/bin/bash
SHELL_FOLDER=$(cd "$(dirname "$0")";pwd)
PROJECT_FOLDER=${SHELL_FOLDER}/..
echo ${SHELL_FOLDER}
OS="windows"
export GOOS=${OS}
export GOARCH=amd64
export CGO_ENABLED=0

COMMIT_HASH=`git rev-parse HEAD 2>/dev/null`
BUILD_DATE=`date  +%Y-%m-%d-%H:%M`
BRANCH_NAME=`git symbolic-ref --short -q HEAD`

TARGET_DIR=${PROJECT_FOLDER}/bin/VpnApp_${BRANCH_NAME}
if [ ! -d ${TARGET_DIR} ]; then
  mkdir ${TARGET_DIR}
fi
TARGET=${TARGET_DIR}/vpnedge_${OS}.exe
SOURCE=${PROJECT_FOLDER}/cmd/VpnEdge

cd ${SOURCE}
go build -ldflags "-X \"main.BuildVersion=${COMMIT_HASH}\" -X \"main.BuildDate=${BUILD_DATE}\" -X \"main.BuildBranch=${BRANCH_NAME}\"" -i -o ${TARGET}