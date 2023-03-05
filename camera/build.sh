#!/usr/bin/env bash

# this script builds the camera application and installs it to an RPi OS image.

if [ ! -f "2023-02-21-raspios-bullseye-armhf-lite.img.xz" ]; then
  wget https://downloads.raspberrypi.org/raspios_lite_armhf/images/raspios_lite_armhf-2023-02-22/2023-02-21-raspios-bullseye-armhf-lite.img.xz
fi

if [ ! -f "2023-02-21-raspios-bullseye-armhf-lite.img" ]; then
  xz --decompress --keep 2023-02-21-raspios-bullseye-armhf-lite.img.xz
fi

mount -o loop 2023-02-21-raspios-bullseye-armhf-lite.img img/

umount img/