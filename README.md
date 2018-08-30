[![Build Status](https://travis-ci.org/opencopilot/packet-ip-sidecar.svg?branch=master)](https://travis-ci.org/opencopilot/packet-ip-sidecar)
### Packet IP Sidecar

Watches Packet metadata for elastic IP assignments to a device, adds/removes IPs to the loopback.


#### Usage

`docker run --net host --cap-add NET_ADMIN opencopilot/packet-ip-sidecar`
