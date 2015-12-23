#!/usr/bin/env bash
set -e

export GOPATH=~/go
export PATH="${PATH}:${GOPATH}/bin"

## Create target directory
mkdir -p target/root/usr/bin
if [ $? -ne 0 ]; then
    exit 1
fi

## Download cross compiler
go get github.com/laher/goxc
if [ $? -ne 0 ]; then
    exit 1
fi

goxc -bc="linux,!arm windows darwin"
if [ $? -ne 0 ]; then
    exit 1
fi

## Build files
#GOOS=linux GOARCH=386 go build -o target/root/usr/bin/orachle-java-launcher main.go
#if [ $? -ne 0 ]; then
#    exit 1
#fi

cd target/root
if [ $? -ne 0 ]; then
    exit 1
fi
tar -cjf target/oracle-java-launcher.tar.b2 .
if [ $? -ne 0 ]; then
    exit 1
fi
