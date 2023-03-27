import subprocess
import socket
import sys

# Set up the UDP socket
sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)

# start monitor in the background
monitor_out = open('osss-monitor.log', 'w')
monitor = subprocess.Popen(
  ['./osss-monitor'], 
  stdout=subprocess.PIPE, 
  stderr=monitor_out
)

# init variables
cameraEndpoint = ('localhost', 7776)
monitorEndpoint = ('localhost', 7777)

message = b'testing'
sent = sock.sendto(message, cameraEndpoint)
data, server = sock.recvfrom(4096)
if data[0] != b'testing':
  print('camera request failed: '+str(data))
  sys.exit(1)

message = b'testing'
sent = sock.sendto(message, monitorEndpoint)
data, server = sock.recvfrom(4096)
if data[0] != b'testing':
  print('camera request failed: '+str(data))
  sys.exit(1)

monitor_out.close()
monitor.terminate()