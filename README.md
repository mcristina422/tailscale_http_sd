# tailscale_http_sd

tailscale_http_sd is a webserver exposing a [Prometheus HTTP Service Discovery](https://prometheus.io/docs/prometheus/latest/http_sd/) for services behind a [tailscale](https://tailscale.com/) VPN. Inspired by the [Services BETA](https://tailscale.com/kb/1100/services/) offered by tailscale and the [recent introduction of HTTP SD in prometheus](https://github.com/prometheus/prometheus/pull/8839). 

My use case is running [node-exporter](https://github.com/prometheus/node_exporter) to expose system metrics on each compute device. A monitoring stack (such as [prometheus](https://github.com/prometheus/prometheus)) can then scrape remote devices (ie on different LANs) to have a centralized view of each devices stats. Since Tailscale is handling the connection for us, we don't have to worry about firewalls, port forwarding, TLS or securing each individual device. 

This [Grafana dashboard](https://grafana.com/grafana/dashboards/1860) is recommended for viewing node exporter metrics.

HTTP Service Discovery was added recently (and bugs fixed) so use at least prometheus 2.28.1

## Quickstart

You must provide the hostnames of the devices you wish to enable
`./tailscale_http_sd --host server-1 --host pi4`

TODO: expose more ways of enabling a host. `--all` or maybe using tailscale tags etc 

# prometheus endpoint
tailscale_http_sd will return the enabled devices as targets in [HTTP SD format](https://prometheus.io/docs/prometheus/latest/http_sd/#http_sd-format) at `localhost:8773/prometheus`
```json
[
  {
    "targets": [
      "100.49.16.20:9100"
    ],
    "labels": {
      "hostname": "server-1"
    }
  },
  {
    "targets": [
      "100.75.50.85:9100"
    ],
    "labels": {
      "hostname": "pi4"
    }
  }
]
```

# status endpoint
`localhost:8773/status` will provide the tailscale status endpoint for debugging


## Example Files
// TODO commit these to an examples/ folder

`docker-compose.yml`

```yaml
services:
  prometheus:
    image: prom/prometheus:v2.28.1
    volumes:
      - ./prometheus/:/etc/prometheus/
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/usr/share/prometheus/console_libraries'
      - '--web.console.templates=/usr/share/prometheus/consoles'
    network_mode: host # needed to hit targets over tailscale 
  tailscale_http:
    image: ghcr.io/mcristina422/tailscale_http_sd:latest
    network_mode: host
    volumes: # tailscale socket is needed to lookup hosts
      - /var/run/tailscale/tailscaled.sock:/var/run/tailscale/tailscaled.sock

```

`prometheus/prometheus.yaml`

```yaml
global:
  scrape_interval: 15s

scrape_configs:
- job_name: httpsd
  http_sd_configs:
    - url: 'http://localhost:8773/prometheus'
```
