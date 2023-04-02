import os
import random, string

# pi-gen configuration file
config_path = os.path.join('pi-gen', 'monitor.config')
with open(config_path, 'r') as handle:
  contents = handle.read()
  contents = contents.replace('USERNAME', input('\nEnter your login username: '))
  contents = contents.replace('PASSWORD', input('\nEnter your login password: '))
with open(config_path, 'w') as handle:
  handle.write(contents)

# WPA configuration
ssid = ''.join(random.choices(string.ascii_letters + string.digits, k=24))
wpa_password = ''.join(random.choices(string.ascii_letters + string.digits, k=63))
hostapd_path = os.path.join('pi-gen', 'stage2', '02-net-tweaks', 'files', 'hostapd.conf')
with open(hostapd_path, 'r') as handle:
  contents = handle.read()
  contents = contents.replace('NETWORK', ssid)
  contents = contents.replace('PASSWORD', wpa_password)
with open(hostapd_path, 'w') as handle:
  handle.write(contents)
with open('.wpaenv', 'w') as handle:
  handle.write('WPA_ESSID='+ssid+'\n')  
  handle.write('WPA_PASSWORD='+wpa_password+'\n')
