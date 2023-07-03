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


def fail(camera):
  camera.terminate()
  sys.exit(1)


def succeed(camera):
  camera.terminate()
  sys.exit(0)


print('loading config...')
parsed_yaml = parse_yaml_file('configs/config-dev.yaml')

print('running camera tests...')

monitor_log_write = open('../monitor/osss-monitor.json', 'w')
# start monitor in the background
monitor = subprocess.Popen(
  [
    '../monitor/osss-monitor',
    '--config-file',
    '../monitor/configs/config-dev.yaml'
  ],
  stdout=monitor_log_write,
  stderr=monitor_log_write
)
time.sleep(1)  # wait for start

camera_log_write = open('osss-camera.json', 'w')
# start camera in the background
camera = subprocess.Popen(
  [
    './osss-camera',
    '--config-file',
    'configs/config-dev.yaml'
  ],
  stdout=camera_log_write, 
  stderr=camera_log_write
)

# connect to camera first to avoid potential race conditions
cameraSock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
cameraFeed = ('127.0.0.1', parsed_yaml['port'])
cameraSock.bind(cameraFeed)

fail(camera)

# # send data to camera listener
# # from two simulated cameras
# sock1 = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
# sock2 = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
# cameraListener = ('127.0.0.1', parsed_yaml['port'])
# sock1.connect(cameraListener)
# sock2.connect(cameraListener)
# message1 = 'testing message from camera 1'  # 29 bytes
# message2 = 'longer testing message from camera 2'  # 36 bytes
# try:
#   sock1.sendall(message1.encode())
#   print(f"sent data on camera listener: \"{message1}\" to {cameraListener}")
#   sock2.sendall(message2.encode())
#   print(f"sent data on camera listener: \"{message2}\" to {cameraListener}")
# finally:
#   sock1.close()
#   sock2.close()
#
# # wait for response from camera feed
#
# messages = []
# while True:
#   data, addr = cameraSock.recvfrom(2048)
#   print(f"received data on camera feed: {data.decode()} from {addr}")
#   if len(data.decode()) > 0:
#     messages.append(data.decode().strip('\x00'))
#   if len(messages) == 2:
#     break
# if message1 not in messages or message2 not in messages:
#   print('failed to receive data on camera feed')
#   fail(camera)
#
# succeed(camera)
