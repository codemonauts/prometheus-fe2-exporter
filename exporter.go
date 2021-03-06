package main

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Exporter struct {
	Hostname  string
	AccessKey string
}

func NewExporter(hostname string, accessKey string) *Exporter {
	return &Exporter{
		Hostname:  hostname,
		AccessKey: accessKey,
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- scrapeDuration
	ch <- inputStatus
	ch <- cloudServiceStatus
	ch <- mqttServerStatus
	ch <- freeMemory
	ch <- freeDiskSpace
	ch <- systemStatus
	ch <- loggedErrors
	ch <- redundancyStatus
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
	inputResponse, err := QueryInputs(e.Hostname, e.AccessKey)
	if err != nil {
		fmt.Println(err)
		errors += 1
	} else {
		for _, input := range *inputResponse {
			for _, state := range []string{"OK", "ERROR", "NOT_USED"} {
				ch <- prometheus.MustNewConstMetric(
					inputStatus, prometheus.GaugeValue, CheckState(input.State, state), input.Name, input.Identifier, state,
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
	serviceResponse, err := QueryCloudServices(e.Hostname, e.AccessKey)
	if err != nil {
		fmt.Println(err)
		errors += 1
	} else {
		for _, service := range *serviceResponse {
			for _, state := range []string{"OK", "ERROR"} {
				ch <- prometheus.MustNewConstMetric(
					cloudServiceStatus, prometheus.GaugeValue, CheckState(service.State, state), service.Name, state,
				)
			}
		}
	}

	// System status
	statusResponse, err := QueryStatus(e.Hostname, e.AccessKey)
	if err != nil {
		fmt.Println(err)
		errors += 1
	} else {
		ch <- prometheus.MustNewConstMetric(loggedErrors, prometheus.GaugeValue, statusResponse.NbrOfLoggedErrors)

		for _, state := range []string{"OK", "WARN", "ERROR"} {
			ch <- prometheus.MustNewConstMetric(systemStatus, prometheus.GaugeValue, CheckState(statusResponse.State, state), state)
		}
		for _, state := range []string{"OK", "WARN"} {
			ch <- prometheus.MustNewConstMetric(redundancyStatus, prometheus.GaugeValue, CheckState(statusResponse.RedundancyState.State, state), state)
		}
	}

	// MQTT status
	mqttResponse, err := QueryMQTTServer(e.Hostname, e.AccessKey)
	if err != nil {
		fmt.Println(err)
		errors += 1
	} else {
		for _, server := range *mqttResponse {
			for _, state := range []string{"OK", "ERROR", "NOT_USED"} {
				ch <- prometheus.MustNewConstMetric(
					mqttServerStatus, prometheus.GaugeValue, CheckState(server.State, state), server.Name, state,
				)
			}
		}
	}

	// Storage/Memory status
	systemReponse, err := QuerySystem(e.Hostname, e.AccessKey)
	if err != nil {
		fmt.Println(err)
		errors += 1
	} else {
		ch <- prometheus.MustNewConstMetric(freeMemory, prometheus.GaugeValue, systemReponse.FreeMemory)
		for _, disk := range systemReponse.Disks {
			ch <- prometheus.MustNewConstMetric(freeDiskSpace, prometheus.GaugeValue, disk.FreeSpace, disk.DriveLetter)
		}
	}

	if errors == 0 {
		return 1
	} else {
		return 0
	}
}
