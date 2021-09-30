package main

import (
	"fmt"
	"net/http"

	"github.com/imroc/req"
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
		"Was the last scrape successful.",
		nil, nil,
	)
	inputStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "alarm_input_state"),
		"Current state of an alarm input",
		[]string{"name", "id"}, nil,
	)
)

type InputResponse struct {
	Name       string `json:"name"`
	Identifier string `json:"id"`
	State      string `json:"state"`
}

func (i InputResponse) GetValue() float64 {
	switch i.State {
	case "OK":
		return 1
	case "ERROR":
		return 0
	case "NOT_USED":
		return 2
	default:
		fmt.Printf("Unknown input state: %q\n", i.State)
		return 0
	}
}

type Exporter struct {
	Hostname  string
	accessKey string
}

func NewExporter(hostname string, accessKey string) *Exporter {
	return &Exporter{
		Hostname:  hostname,
		accessKey: accessKey,
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- inputStatus
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	status := e.Scrape(ch)
	ch <- prometheus.MustNewConstMetric(up, prometheus.GaugeValue, status)
}

func (e *Exporter) Scrape(ch chan<- prometheus.Metric) float64 {
	authHeader := req.Header{
		"Accept":        "application/json",
		"Authorization": e.accessKey,
	}
	r, err := req.Get(fmt.Sprintf("%s/rest/monitoring/input", e.Hostname), authHeader)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	var response []InputResponse
	err = r.ToJSON(&response)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	for _, input := range response {
		ch <- prometheus.MustNewConstMetric(
			inputStatus, prometheus.GaugeValue, input.GetValue(), input.Name, input.Identifier,
		)

	}
	return 1
}

func main() {
	accessKey := "topsecret"
	hostname := "http://10.10.0.1"

	exporter := NewExporter(hostname, accessKey)
	prometheus.MustRegister(exporter)

	fmt.Printf("Listening on %q\n", listenAddr)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(listenAddr, nil)
}
