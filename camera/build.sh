#!/usr/bin/env bash

# amd64
ARCH=amd64

if [ $ARCH = "arm" ]; then
  ARCHOPT="GOARCH=$ARCH GOARM=5"
else
  ARCHOPT="GOARCH=$ARCH"
fi

killall osss-camera
rm -v osss-camera osss-camera.json
go mod tidy && \
env GOOS=linux $ARCHOPT go build -o osss-camera ./internal/app/camera

if ! python test/camera.py ; then
  printf "camera tests failed\n"
  exit 1
fi