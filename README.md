# Open Source Security System

## Overview

This repository contains everything you need to configure
a self-contained camera security system using open source
software and the Raspberry Pi hardware platform.

## Building the software

The build process has only been tested on Ubuntu 22.

To build the software, follow these steps:

* The monitor application needs to be built first: `bash build_image.sh monitor`
* Followed by the camera application: `bash build_image.sh camera`

## Building the hardware

You'll need a full RPi for the monitor (we use the 400, Model B should work fine as well),
along with a screen. We used a touchscreen that works with RPi out of the box, in order
to avoid the need for a mouse.

For the camera, you'll need:

* RPi Zero W or Zero 2 W
* Battery pack
* Camera
