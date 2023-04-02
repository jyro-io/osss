import yaml
import subprocess
import socket
import sys
import time

def parse_yaml_file(file_path):
  with open(file_path, 'r') as file:
    try:
      yaml_dict = yaml.safe_load(file)
      return yaml_dict
    except yaml.YAMLError as exc:
      print(f"error in parsing YAML file: {exc}")
  sys.exit(1)

def fail(monitor):
  monitor.terminate()
  sys.exit(1)

def succeed(monitor):
  monitor.terminate()
  sys.exit(0)

print('loading config...')
parsed_yaml = parse_yaml_file('configs/config.yaml')

print('running monitor tests...')

monitor_log_write = open('osss-monitor.json', 'w')
# start monitor in the background
monitor = subprocess.Popen(
  ['./osss-monitor'], 
  stdout=monitor_log_write, 
  stderr=monitor_log_write
)
time.sleep(1)  # wait for start

# bind to monitor feed socket
monitorSock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
monitorFeed = ('127.0.0.1', parsed_yaml['monitorPort'])
monitorSock.bind(monitorFeed)
monitorSock.settimeout(1000)

# send data to camera listener
# from two simulated cameras
sock1 = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
sock2 = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
cameraListener = ('127.0.0.1', parsed_yaml['cameraPort'])
message1 = 'message for testing from camera 1'  # 33 bytes
message2 = 'longer message for testing from camera 2'  # 40 bytes
try:
  sock1.sendto(message1.encode(), cameraListener)
  print(f"sent data on camera listener: \"{message1}\" to {cameraListener}")
  sock2.sendto(message2.encode(), cameraListener)
  print(f"sent data on camera listener: \"{message2}\" to {cameraListener}")
finally:
  sock1.close()
  sock2.close()

# wait for response from monitor feed

messages = []
while True:
  data, addr = monitorSock.recvfrom(2048)
  print(f"received data on monitor feed: {data.decode()} from {addr}")
  if len(data.decode()) > 0:
    messages.append(data.decode().strip('\x00'))
  if len(messages) == 2:
    break
if message1 not in messages or message2 not in messages:
  print('failed to receive data on monitor feed')
  fail(monitor)

succeed(monitor)
