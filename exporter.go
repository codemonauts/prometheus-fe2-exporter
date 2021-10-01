package main

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

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
	errors := 0

	// Get alarm inputs
	response, err := QueryInputs(e.Hostname, e.accessKey)
	if err != nil {
		fmt.Println(err)
		errors += 1
	} else {
		for _, input := range *response {
			for _, state := range []string{"OK", "ERROR", "NOT_USED"} {
				ch <- prometheus.MustNewConstMetric(
					inputStatus, prometheus.GaugeValue, input.HasStatus(state), input.Name, input.Identifier, state,
				)
			}
		}

		for _, input := range *response {
			if v, err := input.GetValue(); err == nil {
				ch <- prometheus.MustNewConstMetric(
					inputValue, prometheus.GaugeValue, v, input.Name, input.Identifier,
				)
			}

		}
	}

	// Get cloud services
	// TODO

	// System status
	// TODO

	// MQTT status
	// TODO

	// Storage/Memory status
	// TODO

	if errors == 0 {
		return 1
	} else {
		return 0
	}
}
