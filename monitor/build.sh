#!/usr/bin/env bash

# amd64
ARCH=amd64

if [ $ARCH = "arm" ]; then
  ARCHOPT="GOARCH=$ARCH GOARM=5"
else
  ARCHOPT="GOARCH=$ARCH"
fi

killall osss-monitor
rm -v osss-monitor osss-monitor.json
go mod tidy && \
env GOOS=linux $ARCHOPT go build -o osss-monitor ./internal/app/monitor

if ! python test/monitor.py ; then
  printf "monitor tests failed\n"
  exit 1
fi