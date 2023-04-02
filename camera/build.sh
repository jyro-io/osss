#!/usr/bin/env bash

# arm, amd64
ARCH=$1

if [ $ARCH = "arm" ]; then
  ARCHOPT="GOARCH=$ARCH GOARM=5"
else
  ARCHOPT="GOARCH=$ARCH"
fi

go mod tidy && \
env GOOS=linux $ARCHOPT go build -o osss-camera ./internal/app/camera
