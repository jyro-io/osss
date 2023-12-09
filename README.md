# Open Source Security System

## Note: this is unstable and currently under development 

## Overview

This repository contains everything you need to configure
a self-contained camera security system using open source
software and the Raspberry Pi hardware platform.

## Dependencies

* Gocv: https://gocv.io
* PiSugar: https://github.com/PiSugar/PiSugar

## Software

This build process has only been tested on Ubuntu 22.

### Testing

#### Dependencies

* Build/Install gocv: https://gocv.io/getting-started/linux/

#### Running the tests

The only comprehensive way to test this before building the RPi images
and deploying to hardware is to have a development machine with two cameras.

Use `ls /dev | grep video` to find relevant device files for your system. 

In `camera` directory, running `bash build.sh 0 2` should result in two
windows that contain the camera feed from the monitor application.
If not, fiddle with the two camera numbers to get the right combo
for your system. 

You should see movements outlined in red on your two camera feeds.

### Image Build

To build the software, follow these steps:

* Install test dependencies: `sudo apt-get install libcanberra-gtk-module`
* The monitor application needs to be built first: `bash build_image.sh monitor`
* Followed by the camera application: `bash build_image.sh camera`

## Hardware

For the monitor, you'll need:

* RPi 400 or Model B
* Screen, I used a touchscreen to simplify usage 

For the camera, you'll need:

* RPi Zero W or Zero 2 W
* PiSugar 3 battery
  * https://www.tindie.com/products/pisugar/pisugar3-battery-for-raspberry-pi-zero/
* A case, unless you prefer to live life on the edge. 
  I'm currently 3D printing the PiSugar cases listed above, 
  and mounting the camera on the outside: https://github.com/PiSugar/PiSugar/tree/master/model2
* RPi Camera
  * I 3D printed this camera housing: https://www.thingiverse.com/thing:1707484

## Development notes

* If you make modifications to the pi-gen process, 
  make sure that your code accounts for both starting 
  from scratch, or resuming the image build from an arbitrary
  stage.

## Future Development

* Monitor: Add support for capturing video clips on the monitor and saving to a mounted USB drive. [Issue](https://github.com/rory-linehan/osss/issues/1)
