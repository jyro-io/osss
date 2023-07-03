#!/usr/bin/env bash

# arm, amd64
ARCH=${1:-amd64}

if [ $ARCH = "arm" ]; then
  ARCHOPT="GOARCH=$ARCH GOARM=5"
else
  ARCHOPT="GOARCH=$ARCH"
fi

killall osss-camera
killall osss-monitor

rm -v \
osss-camera \
osss-camera.json \

go mod tidy && \
env GOOS=linux $ARCHOPT go build -o osss-camera ./internal/app/camera && \
cd ../monitor && \
bash build.sh $ARCH && \
cd ../camera

if ! python test/camera.py ; then
  printf "tests failed\n"
  exit 1
fi
