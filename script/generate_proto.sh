#!/bin/bash

protoc -I ../proto ../proto/*.proto  --go_out=../proto

cd ../proto
sed -i "s/golang.org\/x\/net\/context/context/g" *.pb.go