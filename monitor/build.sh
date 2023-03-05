#!/usr/bin/env bash

# this script builds the monitor application and installs it to an RPi OS image.

if [ ! -f "2023-02-21-raspios-bullseye-arm64.img.xz" ]; then
  wget https://downloads.raspberrypi.org/raspios_arm64/images/raspios_arm64-2023-02-22/2023-02-21-raspios-bullseye-arm64.img.xz
fi

if [ ! -f "2023-02-21-raspios-bullseye-arm64.img" ]; then
  xz --decompress --keep 2023-02-21-raspios-bullseye-arm64.img.xz
fi

mount -o loop 2023-02-21-raspios-bullseye-arm64.img img/

umount img/