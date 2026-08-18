# prometheus-fe2-exporter

This exporter scrapes the REST-API of an [Alamos FE2](https://www.alamos-gmbh.com/service/fe2/) server. Documentation (in German)
for the API endpoints can be found in their [Confluence](https://alamos-support.atlassian.net/wiki/spaces/documentation/pages/1683226637/Monitoring-Schnittstelle).

# Usage

## Docker

```sh
docker run -d \
    -p 9865:9865 \
    -e FE2_EXPORTER_URL=http://alamos-fe2-server:83 \
    -e FE2_EXPORTER_ACCESSKEY=Gz8NcB6aKTp2uQ+Y1oVhRA== \
    codemonauts/prometheus-fe2-exporter:latest
```

## Docker Compose

```yaml
  prometheus-fe2-exporter:
    container_name: prometheus-fe2-exporter
    image: codemonauts/prometheus-fe2-exporter:latest
    environment:
      FE2_EXPORTER_URL: "http://alamos-fe2-server:83"
      FE2_EXPORTER_ACCESSKEY: "Gz8NcB6aKTp2uQ+Y1oVhRA=="
    restart: unless-stopped
    ports:
      - 9865:9865
```

## Binary

```
Usage of ./prometheus-fe2-exporter:
  -accesskey string
        Authorization key for the monitoring api
  -listen string
        Address to listen on for the metrics endpoint (default ":9865")
  -url string
        Address of the FE2 server (for example http://alamos-fe2-server:83)
```

All flags can also be set via environment variables: `FE2_EXPORTER_URL`,
`FE2_EXPORTER_ACCESSKEY` and `FE2_EXPORTER_LISTEN`.

## Endpoints

- `/metrics` – Prometheus metrics
- `/healthz` – liveness probe (returns `ok`)
