#!/usr/bin/env bash
set -e

## Create target directory
mkdir -p target/root/usr/bin
if [ $? -ne 0 ]; then
    exit 1
fi

## Build files
GOOS=linux GOARCH=386 go build -o target/root/usr/bin/orachle-java-launcher launcher.go
if [ $? -ne 0 ]; then
    exit 1
fi

GOOS=linux GOARCH=386 go build -o target/root/usr/bin/orachle-java-launcher-installer installer.go
if [ $? -ne 0 ]; then
    exit 1
fi

cd target/root
if [ $? -ne 0 ]; then
    exit 1
fi
tar -cjf target/oracle-java-launcher.tar.b2 .
if [ $? -ne 0 ]; then
    exit 1
fi
