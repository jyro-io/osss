#!/usr/bin/env bash

# this script builds the osss applications, 
# creates a custom RPi-OS image, 
# and installs the application to the custom image.

# usage:
# bash build.sh monitor
# bash build.sh camera

APP=$1
APPNAME=osss-$APP
APPCONFIG=$APP.config
ROOTDIR=$PWD

# image build dependencies
sudo apt-get install -y \
  coreutils quilt parted qemu-user-static debootstrap zerofree zip \
  dosfstools libarchive-tools libcap2-bin grep rsync xz-utils file git curl bc \
  qemu-utils kpartx gpg pigz

if [ ! -d "pi-gen" ]; then
  git clone git@github.com:jyro-io/pi-gen.git
fi

# build app
cd $APP/src
go mod tidy
go build -o ../$APPNAME
cd $ROOTDIR

cp $APP/$APPCONFIG pi-gen/$APPCONFIG
cd pi-gen
git checkout $APPNAME

# generate wifi credentials
#python config_wifi_credentials.py $APP

# build image
printf "IMG_NAME=$APPNAME\n" >> $APPCONFIG
touch ./stage4/SKIP ./stage5/SKIP
touch ./stage4/SKIP_IMAGES ./stage5/SKIP_IMAGES
sudo ./build.sh -c $APPCONFIG
cd $ROOTDIR

if [ ! -f "/usr/bin/rpi-imager" ]; then
  sudo apt update && sudo apt install -y rpi-imager
fi
read -p "Insert your destination sd card now. In RPi Imager, select the custom image file (${IMAGE}.img), and the sd card device. Press enter to begin."
rpi-imager