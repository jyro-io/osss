#!/usr/bin/env bash

# this script builds the osss applications, 
# creates a custom RPi-OS image, 
# and installs the application to the custom image.

# usage:
# bash build_image.sh monitor
# bash build_image.sh camera

APP=$1
DEV=true

if [ $APP = "monitor" ] || [ $APP = "camera" ]; then
  APPNAME=osss-$APP
  APPCONFIG=$APP.config
  ROOTDIR=$PWD

  # image build dependencies
  sudo apt-get install -y \
    coreutils quilt parted qemu-user-static debootstrap zerofree zip \
    dosfstools libarchive-tools libcap2-bin grep rsync xz-utils file git curl bc \
    qemu-utils kpartx gpg pigz binfmt-support

  if [ ! -d "pi-gen" ]; then
    git clone git@github.com:jyro-io/pi-gen.git
  fi

  # build app
  cd $APP
  bash build.sh
  cd $ROOTDIR

  cp $APP/$APPCONFIG pi-gen/$APPCONFIG
  cd pi-gen
  git checkout $APPNAME
  git pull

  # handle wifi credentials
  if [ $APP = "monitor" ]; then
    python wpa_credentials.py
  elif [ $APP = "camera" ] && [ -f "$ROOTDIR/.wpaenv" ]; then
    set -o allexport
    source .wpaenv
    set +o allexport
  else
    echo "error: wpa credentials (.wpaenv) not present, run 'bash build_image.sh monitor' first!"
    exit 1
  fi

  # build image
  touch ./stage4/SKIP ./stage5/SKIP
  touch ./stage4/SKIP_IMAGES ./stage5/SKIP_IMAGES
  if [ $APP = "monitor" ]; then
    touch ./stage3/EXPORT_IMAGE
  elif [ $APP = "camera" ]; then
    touch ./stage2/EXPORT_IMAGE ./stage3/SKIP ./stage3/SKIP_IMAGES
  fi
  printf "IMG_NAME=$APPNAME\n" >> $APPCONFIG
  if [ $DEV = true ]; then
    sudo CONTINUE=1 PRESERVE_CONTAINER=1 ./build-docker.sh -c $APPCONFIG
  else
    sudo ./build-docker.sh -c $APPCONFIG
  fi
  cd $ROOTDIR

  if [ ! -f "/usr/bin/rpi-imager" ]; then
    sudo apt update && sudo apt install -y rpi-imager
  fi
  read -p "Insert your destination sd card now. In RPi Imager, select the custom image file ($ROOTDIR/pi-gen/deploy/$APPNAME-lite.img), and the sd card device. Press enter to begin."
  rpi-imager
else
  echo "error: argument must be one of [monitor, camera]"
  exit 1
fi