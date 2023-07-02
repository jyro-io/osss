#!/usr/bin/env bash

# this script builds the osss applications, 
# creates a custom RPi-OS image, 
# and installs the application to the custom image.

# usage:
# bash build_image.sh monitor
# bash build_image.sh camera

APP=$1

if [ -f ".buildenv" ]; then
  source .buildenv
fi

if [ $DOCKER_BUILD = true ]; then
  BUILD="bash build-docker.sh"
  if [ $DEV = false ]; then
    # true -> false means a lingering container is present unless manually removed
    docker rm pigen_work
  fi
else
  BUILD="sudo bash build.sh"
fi

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
    git clone git@github.com:rory-linehan/pi-gen.git
  else
    if [ $DEV = false ]; then
      cd $ROOTDIR/pi-gen
      git clean -fd
      git restore .
      cd $ROOTDIR
    fi
  fi

  # switch to app branch in pi-gen
  cd $ROOTDIR/pi-gen
  git checkout $APPNAME
  if [ $DEV = false ]; then
    git pull
  fi
  cd $ROOTDIR

  # copy pi-gen config
  cp -v $APP/$APPCONFIG $ROOTDIR/pi-gen/$APPCONFIG

  # handle wifi credentials for camera
  if [ $APP = "monitor" ]; then
    cd $ROOTDIR
    python $APP/scripts/configure.py  # configure monitor
  elif [ $APP = "camera" ] && [ -f .wpaenv ]; then
    cd $ROOTDIR
    cat .wpaenv >> $ROOTDIR/pi-gen/$APPCONFIG
  else
    echo "error: wpa credentials (.wpaenv) not present, run 'bash build_image.sh monitor' first!"
    exit 1
  fi

  # build app
  cd $ROOTDIR/$APP
  if [ $APP = "monitor" ]; then
    source "venv/bin/activate"
  fi
  bash build.sh arm
  if [ $APP = "monitor" ]; then
    source deactivate
  fi
  cd $ROOTDIR

  # setup configuration files
  INSTALLDIRFILES=./$APPNAME/00-install/files/
  mkdir -p $INSTALLDIRFILES
  cp -v $ROOTDIR/$APP/$APPNAME $INSTALLDIRFILES
  cp -v $ROOTDIR/$APP/configs/config.yaml $INSTALLDIRFILES
  cp -v $ROOTDIR/$APP/etc/$APPNAME.service $INSTALLDIRFILES
  printf "IMG_NAME=$APPNAME\n" >> $APP/$APPCONFIG

  # misc configuration files
  if [ $APP = "monitor" ]; then
    cp -v $ROOTDIR/$APP/etc/camera-stream.desktop $INSTALLDIRFILES
  elif [ $APP = "camera" ]; then
    cp -v $ROOTDIR/$APP/etc/motion.conf $INSTALLDIRFILES
    cd $ROOTDIR
    python $APP/scripts/configure.py  # configure camera
    cd $ROOTDIR/pi-gen
  fi

  # build image
  cd $ROOTDIR/pi-gen
  if [ $DEV = true ]; then
    CONTINUE=1 PRESERVE_CONTAINER=1 $BUILD -c $APPCONFIG
  else
    $BUILD -c $APPCONFIG
  fi

  # write image to sd card
  if [ ! -f "/usr/bin/rpi-imager" ]; then
    sudo apt update && sudo apt install -y rpi-imager
  fi
  read -p "Insert your destination sd card now. In RPi Imager, select the custom image file ($ROOTDIR/pi-gen/deploy/$(date +%Y-%m-%d)-$APPNAME.img), and the sd card device. Press enter to begin."
  rpi-imager
else
  echo "error: argument must be one of [monitor, camera]"
  exit 1
fi