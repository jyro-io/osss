import subprocess
import socket
import sys

print('running monitor tests...')

monitor_log_write = open('osss-monitor.json', 'w')
# start monitor in the background
monitor = subprocess.Popen(
  ['./osss-monitor'], 
  stdout=subprocess.PIPE, 
  stderr=monitor_log_write
)

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
monitorEndpoint = ('127.0.0.1', 7777)
message = 'message for testing'
try:
  sock.sendto(message.encode(), monitorEndpoint)
  print(f"sent message: \"{message}\" to {monitorEndpoint}")
finally:
  sock.close()

monitor.terminate()

# examine output
monitor_log_read = open('osss-monitor.json', 'r')
if message not in monitor_log_read.read():
  print('monitor failed to receive test message')
  sys.exit(1)
