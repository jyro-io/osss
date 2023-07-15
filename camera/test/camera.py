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


print('running camera tests...')

print('loading monitor config...')
monitor_yaml = parse_yaml_file('../monitor/configs/config-dev.yaml')

monitor_log_write = open('../monitor/osss-monitor.json', 'w')
print('start monitor in the background...')
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

print('loading camera config...')
camera_yaml = parse_yaml_file('configs/config-dev.yaml')

camera_0_log_write = open('osss-camera-0.json', 'w')
print('start camera 0 in the background...')
camera0 = subprocess.Popen(
  [
    './osss-camera',
    '--config-file', 'configs/config-dev.yaml',
    '--camera-device', '0'
  ],
  stdout=camera_0_log_write,
  stderr=camera_0_log_write
)

camera_1_log_write = open('osss-camera-1.json', 'w')
print('start camera 1 in the background...')
camera1 = subprocess.Popen(
  [
    './osss-camera',
    '--config-file', 'configs/config-dev.yaml',
    '--camera-device', '2'  # increment until it works
  ],
  stdout=camera_1_log_write,
  stderr=camera_1_log_write
)

while True:
  time.sleep(1)
