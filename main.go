package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	listenAddr = ":9865"
	namespace  = "fe2"
)

var (
	url       string
	accessKey string
	up        = prometheus.NewDesc(
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
	freeMemory = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "free_memory"),
		"Free memory of host system",
		nil, nil,
	)
	freeDiskSpace = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "free_disk_space"),
		"Free space on storage disks",
		[]string{"drive_letter"}, nil,
	)
	systemStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "system_status"),
		"Current state of the system",
		[]string{"state"}, nil,
	)
	loggedErrors = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "logged_errors"),
		"Number of errors in the last 60 minutes",
		nil, nil,
	)
	redundancyStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "redundancy_status"),
		"Current redundancy state of the system",
		[]string{"state"}, nil,
	)
)

func init() {
	envUrl := os.Getenv("FE2_EXPORTER_URL")
	envAccesskey := os.Getenv("FE2_EXPORTER_ACCESSKEY")

	flag.StringVar(&url, "url", envUrl, "Address of the FE2 server")
	flag.StringVar(&accessKey, "accesskey", envAccesskey, "Authorization key for the monitoring api")
}

func main() {

	flag.Parse()

	if url == "" || accessKey == "" {
		fmt.Println("Url and accesskey are both required parameters")
		os.Exit(1)
	}

	fmt.Printf("FE2 server address is %q\n", url)

	exporter := NewExporter(url, accessKey)
	prometheus.MustRegister(exporter)

	fmt.Printf("Listening on %q\n", listenAddr)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(listenAddr, nil)
}
