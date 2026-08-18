package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"flag"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	defaultListenAddr = ":9865"
	namespace         = "fe2"
)

var (
	url        string
	accessKey  string
	listenAddr string
	up         = prometheus.NewDesc(
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
	envListen := os.Getenv("FE2_EXPORTER_LISTEN")
	if envListen == "" {
		envListen = defaultListenAddr
	}

	flag.StringVar(&url, "url", envUrl, "Address of the FE2 server")
	flag.StringVar(&accessKey, "accesskey", envAccesskey, "Authorization key for the monitoring api")
	flag.StringVar(&listenAddr, "listen", envListen, "Address to listen on for the metrics endpoint")
}

func main() {
	flag.Parse()

	if url == "" || accessKey == "" {
		slog.Error("url and accesskey are both required parameters")
		os.Exit(1)
	}

	slog.Info("starting exporter", "server", url, "listen", listenAddr)

	exporter := NewExporter(url, accessKey)
	prometheus.MustRegister(exporter)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`<html>
<head><title>FE2 Exporter</title></head>
<body><h1>FE2 Exporter</h1><p><a href="/metrics">Metrics</a></p></body>
</html>`))
	})

	server := &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Shut down cleanly on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("error starting exporter", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("error during shutdown", "error", err)
	}
}
