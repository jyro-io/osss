#!/usr/bin/env bash

# this script builds the osss-monitor application, 
# creates a custom RPi-OS image, 
# and installs the application to the custom image.

IMAGE=osss-monitor

# image build dependencies
sudo apt-get install -y \
  coreutils quilt parted qemu-user-static debootstrap zerofree zip \
  dosfstools libarchive-tools libcap2-bin grep rsync xz-utils file git curl bc \
  qemu-utils kpartx gpg pigz

if [ ! -d "pi-gen" ]; then
  git clone git@github.com:jyro-io/pi-gen.git
fi

# build osss-monitor
cd src
go mod tidy
go build -o ../../osss-monitor
cd ..

cd pi-gen
git checkout osss-monitor

# generate wifi credentials
#python image-config/generate_credentials.py

# build image
cp ../image-config/config config
printf "IMG_NAME=$IMAGE\n" >> config
touch ./stage4/SKIP ./stage5/SKIP
touch ./stage4/SKIP_IMAGES ./stage5/SKIP_IMAGES
sudo ./build.sh
cd ..

if [ ! -f "/usr/bin/rpi-imager" ]; then
  sudo apt update && sudo apt install -y rpi-imager
fi
read -p "Insert your destination sd card now. In RPi Imager, select the custom image file (${IMAGE}.img), and the sd card device. Press enter to begin."
rpi-imager