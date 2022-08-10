# Prometheus lolMiner Exporter

[![GitHub release](https://img.shields.io/github/v/release/dewkul/prom-lolminer-exporter?label=Version)](https://github.com/dewkul/prom-lolminer-exporter/releases)
[![CI](https://github.com/dewkul/prom-lolminer-exporter/workflows/CI/badge.svg?branch=main)](https://github.com/dewkul/prom-lolminer-exporter/actions?query=workflow%3ACI)
[![Docker pulls](https://img.shields.io/docker/pulls/dewkul/prom-lolminer-exporter?label=Docker%20Hub)](https://hub.docker.com/r/dewkul/prom-lolminer-exporter)

![Dashboard](https://grafana.com/api/dashboards/14296/images/10340/image)

## What's new
- Support new lolminer API format

## Usage

### lolMiner

Specify `--apiport=<port>` for lolMiner to enable the API server on the specified port.

### Exporter (Docker)

Example `docker-compose.yml`:

```yaml
version: "3.7"

services:
  lolminer-exporter:
    image: dewkul/prom-lolminer-exporter:1
    #command:
    #  - '--endpoint=:8080'
    #  - '--debug'
    user: 1000:1000
    environment:
      - TZ=Europe/Oslo
    ports:
      - "8080:8080/tcp"
```

### Prometheus

Example `prometheus.yml`:

```yaml
global:
    scrape_interval: 15s
    scrape_timeout: 10s

scrape_configs:
  - job_name: "lolminer"
    static_configs:
      # Insert lolminer address here
      - targets: ["lolminer:3493"]
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        # Insert lolMiner exporter address here
        replacement: lolminer-exporter:8080
```

### Grafana

[Example dashboard](https://grafana.com/grafana/dashboards/14296).

## Configuration

### Docker Image Versions

Use `1` for stable v1.Y.Z releases and `latest` for bleeding/unstable releases.

## Metrics

See the [example output](examples/output.txt) (I'm too lazy to create a pretty table).

### Docker

See the dev/example Docker Compose file: [docker-compose.yml](dev/docker-compose.yml)

## Development

- Build (Go): `go build -o prom-lolminer-exporter`
- Lint: `golint ./..`
- Build and run along Traefik (Docker Compose): `docker-compose -f dev/docker-compose.yml up --force-recreate --build`

## License

GNU General Public License version 3 (GPLv3).
