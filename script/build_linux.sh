#!/bin/bash
SHELL_FOLDER=$(cd "$(dirname "$0")";pwd)
PROJECT_FOLDER=${SHELL_FOLDER}/..
echo ${SHELL_FOLDER}
OS="linux"
export GOOS=${OS}
export GOARCH=amd64
export CGO_ENABLED=1

COMMIT_HASH=`git rev-parse --verify --short=8 HEAD 2>/dev/null`
BUILD_DATE=`date  +%m-%d-%H:%M`
BRANCH_NAME=`git symbolic-ref --short -q HEAD`

TARGET_DIR=${PROJECT_FOLDER}/bin/VpnApp_${BRANCH_NAME}_${COMMIT_HASH}
if [ ! -d ${TARGET_DIR} ]; then
  mkdir ${TARGET_DIR}
fi
TARGET=${TARGET_DIR}/vpnedge_${OS}
SOURCE=${PROJECT_FOLDER}/cmd/VpnMultiProto

cd ${SOURCE}
go build -gcflags "-N -l" -ldflags "-X \"main.BuildVersion=${COMMIT_HASH}\" -X \"main.BuildDate=${BUILD_DATE}\" -X \"main.BuildBranch=${BRANCH_NAME}\"" -i -o ${TARGET}