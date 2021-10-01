# prometheus-fe2-exporter

This exporter scrapes the REST-API of an [Alamos
FE2](https://www.alamos-gmbh.com/service/fe2/) server. Documentation (in German)
for the API endpoints can be found in their
[Confluence](https://alamos-support.atlassian.net/wiki/spaces/documentation/pages/1683226637/Monitoring-Schnittstelle).

# Usage
```
Usage of ./prometheus-fe2-exporter:
  -accesskey string
    	Authorization key for the monitoring api
  -host string
    	Address of the FE2 server
  -ssl
    	Use SSL to talk to the FE2 server (default true)
```