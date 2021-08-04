# goblackhole

![GoPher](gopher.png "GoPher")

Goblackhole downloads periodicaly a remote file and propagts these ips to a remote bgp peer.

It can be used to implement [rfc7999](https://datatracker.ietf.org/doc/html/rfc7999).

## Usage

### Configuration
The Config should be stored in `./config.yaml`
```yaml
---
Peers:
  - remote_as: 64512
    remote_ip: "172.17.0.2"
local_id: 192.168.34.169 
local_as: 65001
LogLevel: Debug 
Blocklist: http://network.pages.mgmtbi.ch/blacklist/blacklist.txt
Interval: 1min
Community: 65535:666 # For rfc7999
```