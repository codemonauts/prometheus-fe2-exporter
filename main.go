package main

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	listenAddr = ":9509"
	namespace  = "fe2"
)

var (
	up = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the last scrape successful",
		nil, nil,
	)
	scrapeDuration = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "scrape_duration"),
		"Duration of last scrape",
		nil, nil,
	)
	inputStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "alarm_input_state"),
		"Current state of an alarm input",
		[]string{"name", "id", "state"}, nil,
	)
	inputValue = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "alarm_input_value"),
		"Current value of an alarm input",
		[]string{"name", "id"}, nil,
	)
	cloudServiceStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cloud_service_state"),
		"Current state of a cloud service",
		[]string{"name", "state"}, nil,
	)
	mqttServerStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "mqtt_server_state"),
		"Current state of the mqtt service",
		[]string{"name", "state"}, nil,
	)
)

func main() {
	accessKey := "topsecret"
	hostname := "http://10.10.0.1"

	exporter := NewExporter(hostname, accessKey)
	prometheus.MustRegister(exporter)

	fmt.Printf("Listening on %q\n", listenAddr)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(listenAddr, nil)
}
