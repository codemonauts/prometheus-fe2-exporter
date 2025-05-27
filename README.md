# prometheus-fe2-exporter

This exporter scrapes the REST-API of an [Alamos
FE2](https://www.alamos-gmbh.com/service/fe2/) server. Documentation (in German)
for the API endpoints can be found in their
[Confluence](https://alamos-support.atlassian.net/wiki/spaces/documentation/pages/1683226637/Monitoring-Schnittstelle).

# Usage

## Docker

```sh 
docker run -d \
    -p 9865:9865 \ 
    -e FE2_EXPORTER_HOST=alamos-fe2-server \
    -e FE2_EXPORTER_PORT=83 \
    -e FE2_EXPORTER_SSL=false \
    -e FE2_EXPORTER_HOST=Gz8NcB6aKTp2uQ+Y1oVhRA== \
    bouni/prometheus-fe2-exporter:latest 
```

## Docker Compose

```yaml
  prometheus-fe2-exporter:
    container_name: prometheus-fe2-exporter
    image: prometheus-fe2-exporter:latest
    environment:
      FE2_EXPORTER_HOST: "alamos-fe2-server"
      FE2_EXPORTER_PORT: 83
      FE2_EXPORTER_SSL: false
      FE2_EXPORTER_ACCESSKEY: "Gz8NcB6aKTp2uQ+Y1oVhRA=="
    restart: unless-stopped
    ports: 
      - 9865:9865
```

## Binay

```
Usage of ./prometheus-fe2-exporter:
  -accesskey string
    	Authorization key for the monitoring api
  -host string
    	Address of the FE2 server
  -port string
        Port of the FE2 server (default 83)
  -ssl
    	Use SSL to talk to the FE2 server (default true)
```
