package main

import (
	"context"
	"log/slog"
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
	ch <- inputValue
	ch <- cloudServiceStatus
	ch <- mqttServerStatus
	ch <- freeMemory
	ch <- freeDiskSpace
	ch <- systemStatus
	ch <- loggedErrors
	ch <- redundancyStatus
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	start := time.Now()
	status := e.Scrape(ctx, ch)
	duration := time.Since(start)

	send(ch, up, prometheus.GaugeValue, status)
	send(ch, scrapeDuration, prometheus.GaugeValue, duration.Seconds())
}

func send(ch chan<- prometheus.Metric, desc *prometheus.Desc, valueType prometheus.ValueType, value float64, labelValues ...string) {
	m, err := prometheus.NewConstMetric(desc, valueType, value, labelValues...)
	if err != nil {
		slog.Error("failed to build metric", "desc", desc.String(), "error", err)
		return
	}
	ch <- m
}

func (e *Exporter) Scrape(ctx context.Context, ch chan<- prometheus.Metric) float64 {
	errorCount := 0

	// Get alarm inputs
	inputResponse, err := QueryInputs(ctx, e.Hostname, e.AccessKey)
	if err != nil {
		slog.Error("querying alarm inputs", "error", err)
		errorCount++
	} else {
		for _, input := range *inputResponse {
			for _, state := range []string{"OK", "ERROR", "NOT_USED"} {
				send(ch, inputStatus, prometheus.GaugeValue, CheckState(input.State, state), input.Name, input.Identifier, state)
			}
		}

		for _, input := range *inputResponse {
			if v, err := input.GetValue(); err == nil {
				send(ch, inputValue, prometheus.GaugeValue, v, input.Name, input.Identifier)
			}
		}
	}

	// Get cloud services
	serviceResponse, err := QueryCloudServices(ctx, e.Hostname, e.AccessKey)
	if err != nil {
		slog.Error("querying cloud services", "error", err)
		errorCount++
	} else {
		for _, service := range *serviceResponse {
			for _, state := range []string{"OK", "ERROR"} {
				send(ch, cloudServiceStatus, prometheus.GaugeValue, CheckState(service.State, state), service.Name, state)
			}
		}
	}

	// System status
	statusResponse, err := QueryStatus(ctx, e.Hostname, e.AccessKey)
	if err != nil {
		slog.Error("querying system status", "error", err)
		errorCount++
	} else {
		send(ch, loggedErrors, prometheus.GaugeValue, statusResponse.NbrOfLoggedErrors)

		for _, state := range []string{"OK", "WARN", "ERROR"} {
			send(ch, systemStatus, prometheus.GaugeValue, CheckState(statusResponse.State, state), state)
		}
		for _, state := range []string{"OK", "WARN"} {
			send(ch, redundancyStatus, prometheus.GaugeValue, CheckState(statusResponse.RedundancyState.State, state), state)
		}
	}

	// MQTT status
	mqttResponse, err := QueryMQTTServer(ctx, e.Hostname, e.AccessKey)
	if err != nil {
		slog.Error("querying mqtt servers", "error", err)
		errorCount++
	} else {
		for _, server := range *mqttResponse {
			for _, state := range []string{"OK", "ERROR", "NOT_USED"} {
				send(ch, mqttServerStatus, prometheus.GaugeValue, CheckState(server.State, state), server.Name, state)
			}
		}
	}

	// Storage/Memory status
	systemResponse, err := QuerySystem(ctx, e.Hostname, e.AccessKey)
	if err != nil {
		slog.Error("querying system resources", "error", err)
		errorCount++
	} else {
		send(ch, freeMemory, prometheus.GaugeValue, systemResponse.FreeMemory)
		for _, disk := range systemResponse.Disks {
			send(ch, freeDiskSpace, prometheus.GaugeValue, disk.FreeSpace, disk.DriveLetter)
		}
	}

	if errorCount == 0 {
		return 1
	}
	return 0
}
