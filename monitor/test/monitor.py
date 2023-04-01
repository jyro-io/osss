import subprocess
import socket
import sys
import time

def end_with_error(handle):
  handle.close()
  sys.exit(1)

print('running monitor tests...')

monitor_log_write = open('osss-monitor.json', 'w')
# start monitor in the background
monitor = subprocess.Popen(
  ['./osss-monitor'], 
  stdout=monitor_log_write, 
  stderr=monitor_log_write
)
time.sleep(1)  # wait for start

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
monitorEndpoint = ('127.0.0.1', 7777)
message = 'message for testing'  # 19 bytes
try:
  sock.sendto(message.encode(), monitorEndpoint)
  print(f"sent message: \"{message}\" to {monitorEndpoint}")
finally:
  sock.close()
time.sleep(1)  # wait for server receive

monitor.terminate()

# examine output
monitor_log_read = open('osss-monitor.json', 'r')
if 'received 19 bytes from 127.0.0.1' not in monitor_log_read.read():
  print('monitor failed to receive test message')
  end_with_error(monitor_log_read)
