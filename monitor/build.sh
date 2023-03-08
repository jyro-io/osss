#!/usr/bin/env bash

ARCH=$1

if [ $ARCH = "arm" ]; then
  ARCHOPT="GOARCH=$ARCH GOARM=5"
else
  ARCHOPT="GOARCH=$ARCH"
fi

cd src
go mod tidy
env GOOS=linux $ARCHOPT go build -o ../osss-monitor