import os
import random, string

ssid = ''.join(random.choices(string.ascii_letters + string.digits, k=32))
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