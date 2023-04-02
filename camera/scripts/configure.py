import os
import random, string

# motion configuration file $APP/etc/motion.conf
config_path = os.path.join('pi-gen', 'osss-camera', '00-install', 'files', 'motion.conf')
with open(config_path, 'r') as handle:
  contents = handle.read()
  contents = contents.replace('CAMERANAME', input('\nEnter camera name: '))
with open(config_path, 'w') as handle:
  handle.write(contents)