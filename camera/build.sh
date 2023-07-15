#!/usr/bin/env bash

# arm, amd64
ARCH=${1:-amd64}

if [ $ARCH = "arm" ]; then
  ARCHOPT="GOARCH=$ARCH GOARM=5"
else
  ARCHOPT="GOARCH=$ARCH"
fi

rm -v osss-camera
go mod tidy && \
env GOOS=linux $ARCHOPT go build -o osss-camera ./internal/app/camera && \
cd ../monitor && \
bash build.sh $ARCH
cd ../camera

rm -v osss-camera.json

if ! python test/camera.py ; then
  printf "tests failed\n"
  killall osss-camera
  killall osss-monitor
  exit 1
fi

killall osss-camera
killall osss-monitor
