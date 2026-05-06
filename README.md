# Open Virtual Switch (OVS) Exporter

<a href="https://github.com/webnifico/ovs_exporter/actions/" target="_blank"><img src="https://github.com/webnifico/ovs_exporter/actions/workflows/main.yml/badge.svg?branch=main"></a>

Export Open Virtual Switch (OVS) data to Prometheus.

## Introduction

This exporter exports metrics from the following OVS components:
* OVS `vswitchd` service
* `Open_vSwitch` database
* OVN `ovn-controller` service

## Getting Started

Run the following commands to install it:

```bash
wget https://github.com/webnifico/ovs_exporter/releases/download/v1.0.6/ovs-exporter-1.0.6.linux-amd64.tar.gz
tar xvzf ovs-exporter-1.0.6.linux-amd64.tar.gz
cd ovs-exporter*
./install.sh
cd ..
rm -rf ovs-exporter-1.0.6.linux-amd64*
systemctl status ovs-exporter -l
curl -s localhost:9475/metrics | grep server_id
```

Run the following commands to build and test it:

```bash
git clone https://github.com/webnifico/ovs_exporter.git
cd ovs_exporter
make
go test ./... -run TestParseUpcallShowMetrics
```

For full host-level testing against a running Open vSwitch instance, use:

```bash
make test
```

## Exported Metrics

| Metric | Meaning | Labels |
| ------ | ------- | ------ |
| `ovs_up` |  Is OVS stack up (1) or is it down (0). | `system_id` |

For example:

```bash
$ curl localhost:9475/metrics | grep ovn
# HELP ovs_up Is OVS stack up (1) or is it down (0).
# TYPE ovs_up gauge
ovs_up 1
```

## Flags

```bash
./bin/ovs-exporter --help
```

## Development Notes

Run the following command to build `arm64`:

```bash
make BUILD_OS="linux" BUILD_ARCH="arm64"
```

Next, package the binary:

```bash
make BUILD_OS="linux" BUILD_ARCH="arm64" dist
```

The `qtest` target starts the exporter with `sudo` and expects local Open vSwitch sockets, so it is intended for host-level validation on an OVS node:

```bash
make qtest
```

After a successful release, upload packages to Github:

```bash
owner=$(cat .git/config  | egrep "^\s+url" | cut -d":" -f2 | cut -d"/" -f1)
repo=$(cat .git/config  | egrep "^\s+url" | cut -d":" -f2 | cut -d"/" -f2 | sed 's/.git$//')
tag="v$(< VERSION)"
github_api_token="PASTE_TOKEN_HERE"
filename="./dist/${repo}-$(< VERSION).linux-amd64.tar.gz"
upload-github-release-asset.sh github_api_token=${github_api_token} owner=${owner} repo=${repo} tag=${tag} filename=dist/ovs-exporter-$(< VERSION).linux-amd64.tar.gz
upload-github-release-asset.sh github_api_token=${github_api_token} owner=${owner} repo=${repo} tag=${tag} filename=dist/ovs-exporter-$(< VERSION).linux-arm64.tar.gz
```
