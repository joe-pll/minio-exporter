package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

// MinioExporter collects Minio statistics using the
// Prometheus metrics package
type MinioExporter struct{}

// NewMinioExporter inits and returns a MinioExporter
func NewMinioExporter() (*MinioExporter, error) {
	return &MinioExporte{}, nil
}

// Describe implements the prometheus.Collector interface.
func (e *MinioExporter) Describe(ch chan<- *prometheus.Desc) {

}

// Collect implements the prometheus.Collector interface.
func (e *MinioExporter) Collect(ch chan<- *prometheus.Desc) {

}

func main() {
	var (
		printVersion  = flag.Bool("version", false, "Print version information.")
		listenAddress = flag.String("web.listen-address", ":9107", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	)

	flag.Parse()

	if *printVersion {
		fmt.Fprintln(os.Stdout, version.Print("minio_exporter"))
		os.Exit(0)
	}

	exporter, err := NewMinioExporter()
	if err != nil {
		log.Fatalln(err)
	}

	log.Infoln("Starting minio_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())
}
