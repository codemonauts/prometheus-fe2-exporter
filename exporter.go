package main

import (
	"fmt"
	"time"

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
	ch <- scrapeDuration
	ch <- inputStatus
	ch <- cloudServiceStatus
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	status := e.Scrape(ch)
	duration := time.Since(start)

	ch <- prometheus.MustNewConstMetric(up, prometheus.GaugeValue, status)
	ch <- prometheus.MustNewConstMetric(scrapeDuration, prometheus.GaugeValue, duration.Seconds())
}

func (e *Exporter) Scrape(ch chan<- prometheus.Metric) float64 {
	errors := 0

	// Get alarm inputs
	inputResponse, err := QueryInputs(e.Hostname, e.accessKey)
	if err != nil {
		fmt.Println(err)
		errors += 1
	} else {
		for _, input := range *inputResponse {
			for _, state := range []string{"OK", "ERROR", "NOT_USED"} {
				ch <- prometheus.MustNewConstMetric(
					inputStatus, prometheus.GaugeValue, input.HasStatus(state), input.Name, input.Identifier, state,
				)
			}
		}

		for _, input := range *inputResponse {
			if v, err := input.GetValue(); err == nil {
				ch <- prometheus.MustNewConstMetric(
					inputValue, prometheus.GaugeValue, v, input.Name, input.Identifier,
				)
			}

		}
	}

	// Get cloud services
	serviceResponse, err := QueryCloudServices(e.Hostname, e.accessKey)
	if err != nil {
		fmt.Println(err)
		errors += 1
	} else {
		for _, service := range *serviceResponse {
			for _, state := range []string{"OK", "ERROR"} {
				ch <- prometheus.MustNewConstMetric(
					cloudServiceStatus, prometheus.GaugeValue, service.HasStatus(state), service.Name, state,
				)
			}
		}
	}

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
