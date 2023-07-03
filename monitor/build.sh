#!/usr/bin/env bash

# arm, amd64
ARCH=${1:-amd64}

if [ $ARCH = "arm" ]; then
  ARCHOPT="GOARCH=$ARCH GOARM=5"
else
  ARCHOPT="GOARCH=$ARCH"
fi

killall osss-monitor
rm -v osss-monitor osss-monitor.json
go mod tidy && \
env GOOS=linux $ARCHOPT go build -o osss-monitor ./internal/app/monitor
