# goblackhole

![Gopher](./media/gopher.png "GoPher")

Goblackhole downloads periodicaly a remote file and propagates these ips to a remote bgp peer.

It can be used to implement [rfc7999](https://datatracker.ietf.org/doc/html/rfc7999).

## Usage

### Install

Install it with Go.
```
go get github.com/bedag/goblackhole
```

Or download it from the release page: https://github.com/bedag/goblackhole/releases

Or use our Docker Image bedag/goblackhole

### Docker

```
docker run -d --name gbh bedag/goblackhole:<version>
```

You should mount the configuration file under /etc/goblackhole/config.yaml

### Configuration
The Config should be stored in `./config.yaml`
```yaml
---
Peers:
  - remote_as: 64512
    remote_ip: "172.17.0.2"
    MultiHop: 2
local_id: 10.217.133.15
local_as: 65001
LogLevel: Info 
Blocklist: https://raw.githubusercontent.com/stamparm/ipsum/master/ipsum.txt
Community: 
- 666
NextHop: 192.168.0.1
```

### Kubernetes

You can find the offical Helm Chart here: https://github.com/bedag/helm-charts/tree/master/charts/goblackhole

# Contributing

We'd love to have you contribute! Please refer to our contribution guidelines for details.

By making a contribution to this project, you agree to and comply with the Developer's Certificate of Origin.

# Thanks

- Gopher: https://github.com/quasilyte/gopherkon
- Gobgp:  https://github.com/osrg/gobgp
