# Open Source Security System

## Overview

This repository contains everything you need to configure
a self-contained camera security system using open source
software and the Raspberry Pi hardware platform.

## Dependencies

* PiSugar: https://github.com/PiSugar/PiSugar
* Motion: https://github.com/Motion-Project/motion

## Building the software

The build process has only been tested on Ubuntu 22.

To build the software, follow these steps:

* The monitor application needs to be built first: `bash build_image.sh monitor`
* Followed by the camera application: `bash build_image.sh camera`

## Building the hardware

For the monitor, you'll need:

* RPi 400 or Model B
* Screen, we used a touchscreen to simplify usage 

For the camera, you'll need:

* RPi Zero W or Zero 2 W
* PiSugar 3 battery
  * https://www.tindie.com/products/pisugar/pisugar3-battery-for-raspberry-pi-zero/
* A case, unless you prefer to live life on the edge. 
  We're currently 3D printing the PiSugar cases listed above, 
  and mounting the camera on the outside: https://github.com/PiSugar/PiSugar/tree/master/model2
* RPi Camera
  * We 3D printed this camera housing: https://www.thingiverse.com/thing:1707484
