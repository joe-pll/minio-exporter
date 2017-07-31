package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"

	minio "github.com/minio/minio-go"
	"github.com/minio/minio/pkg/madmin"
)

const (
	// namespace for all the metrics
	namespace = "minio"
	program   = "minio_exporter"
)

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		"minio_exporter: Duration of a collector scrape.",
		nil,
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		"minio_exporter: Whether the collector succeeded.",
		nil,
		nil,
	)
)

// MinioExporter collects Minio statistics using the
// Prometheus metrics package
type MinioExporter struct {
	AdminClient *madmin.AdminClient
	MinioClient *minio.Client
	BucketStats bool
}

// NewMinioExporter inits and returns a MinioExporter
func NewMinioExporter(uri string, minioKey string, minioSecret string, bucketStats bool) (*MinioExporter, error) {
	secure := false
	newURI := uri

	if !strings.Contains(newURI, "://") {
		newURI = "http://" + newURI
	}

	urlMinio, err := url.Parse(newURI)
	if err != nil {
		return nil, fmt.Errorf("invalid Minio URI: %s with error <%s>", newURI, err)
	}

	if urlMinio.Scheme == "https" {
		secure = true
	}

	mdmClient, err := madmin.New(urlMinio.Host, minioKey, minioSecret, secure)
	if err != nil {
		return nil, fmt.Errorf("Minio admin client error %s", err)
	}

	minioClient, err := minio.New(urlMinio.Host, minioKey, minioSecret, secure)
	if err != nil {
		return nil, fmt.Errorf("Minio client error %s", err)
	}

	return &MinioExporter{
		AdminClient: mdmClient,
		MinioClient: minioClient,
		BucketStats: bucketStats,
	}, nil
}

// Describe implements the prometheus.Collector interface.
func (e *MinioExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}

// Collect implements the prometheus.Collector interface.
func (e *MinioExporter) Collect(ch chan<- prometheus.Metric) {
	begin := time.Now()
	err := execute(e, ch)
	duration := time.Since(begin)

	var success float64
	if err != nil {
		log.Errorf("ERROR: collector failed after %fs: %s", duration.Seconds(), err)
		success = 0
	} else {
		log.Debugf("OK: collector succeeded after %fs", duration.Seconds())
		success = 1
	}

	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds())
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success)
}

func execute(e *MinioExporter, ch chan<- prometheus.Metric) error {
	status, err := e.AdminClient.ServiceStatus()
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "uptime"),
			"Minio server uptime in seconds",
			nil,
			nil),
		prometheus.CounterValue,
		status.Uptime.Seconds())

	// Collect server admin statistics
	collectServerStats(e, ch)
	if e.BucketStats {
		collectBucketStats(e, ch)
	}
	return nil
}

func collectServerStats(e *MinioExporter, ch chan<- prometheus.Metric) {
	statsAll, _ := e.AdminClient.ServerInfo()

	for _, stats := range statsAll {
		connStats := stats.Data.ConnStats
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "conn", "total_input_bytes"),
				"Minio total input bytes received",
				nil,
				nil),
			prometheus.GaugeValue,
			float64(connStats.TotalInputBytes))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "conn", "total_output_bytes"),
				"Minio total output bytes received",
				nil,
				nil),
			prometheus.GaugeValue,
			float64(connStats.TotalOutputBytes))

		collectStorageInfo(stats.Data.StorageInfo, ch)
	}

}

func collectStorageInfo(si madmin.StorageInfo, ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "storage", "total_disk_space"),
			"Total Minio disk space in bytes",
			nil,
			nil),
		prometheus.GaugeValue,
		float64(si.Total))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "storage", "free_disk_space"),
			"Free Minio disk space in bytes",
			nil,
			nil),
		prometheus.GaugeValue,
		float64(si.Free))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "storage", "online_disks"),
			"Total number of Minio online disks",
			nil,
			nil),
		prometheus.GaugeValue,
		float64(si.Backend.OnlineDisks))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "storage", "offline_disks"),
			"Total number of Minio offline disks",
			nil,
			nil),
		prometheus.GaugeValue,
		float64(si.Backend.OfflineDisks))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "storage", "read_quorum"),
			"Minio read quorum",
			nil,
			nil),
		prometheus.GaugeValue,
		float64(si.Backend.ReadQuorum))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "storage", "write_quorum"),
			"Minio write quorum",
			nil,
			nil),
		prometheus.GaugeValue,
		float64(si.Backend.WriteQuorum))

	var fstype string
	switch fstypeN := si.Backend.Type; fstypeN {
	case 1:
		fstype = "FS"
	case 2:
		fstype = "Erasure"
	}
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "storage", "storage_type"),
			"Minio backend storage type used",
			[]string{"type"},
			nil),
		prometheus.GaugeValue,
		float64(si.Backend.Type), fstype)
}

func collectBucketStats(e *MinioExporter, ch chan<- prometheus.Metric) {
	buckets, _ := e.MinioClient.ListBuckets()
	wg := sync.WaitGroup{}
	wg.Add(len(buckets))
	for _, bucket := range buckets {
		go func(bucket minio.BucketInfo, e *MinioExporter, ch chan<- prometheus.Metric) {
			bucketStats(bucket, e, ch)
			wg.Done()
		}(bucket, e, ch)
	}
	wg.Wait()
}

func bucketStats(bucket minio.BucketInfo, e *MinioExporter, ch chan<- prometheus.Metric) {
	location, _ := e.MinioClient.GetBucketLocation(bucket.Name)
	var (
		objNum               int64
		bucketSize           int64
		maxObjectSize        int64
		incompleteUploads    int64
		incompleteUploadSize int64
	)
	for objStat := range e.MinioClient.ListObjects(bucket.Name, "", true, nil) {
		objNum = objNum + 1
		bucketSize = bucketSize + objStat.Size
		if objStat.Size > maxObjectSize {
			maxObjectSize = objStat.Size
		}
	}

	for upload := range e.MinioClient.ListIncompleteUploads(bucket.Name, "", true, nil) {
		incompleteUploads = incompleteUploads + 1
		incompleteUploadSize = incompleteUploadSize + upload.Size
	}
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bucket", "objects_number"),
			"The number of objects in to the bucket",
			[]string{"bucket", "location"},
			nil),
		prometheus.GaugeValue,
		float64(objNum), bucket.Name, location)
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bucket", "objects_total_size"),
			"The total size of all object in to the bucket",
			[]string{"bucket", "location"},
			nil),
		prometheus.GaugeValue,
		float64(bucketSize), bucket.Name, location)
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bucket", "max_object_size"),
			"The maximum object size per bucket",
			[]string{"bucket", "location"},
			nil),
		prometheus.GaugeValue,
		float64(maxObjectSize), bucket.Name, location)
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bucket", "incomplete_uploads_number"),
			"The total number of incomplete uploads per bucket",
			[]string{"bucket", "location"},
			nil),
		prometheus.GaugeValue,
		float64(incompleteUploads), bucket.Name, location)
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bucket", "incomplete_uploads_total_size"),
			"The total size of incomplete upload per bucket",
			[]string{"bucket", "location"},
			nil),
		prometheus.GaugeValue,
		float64(incompleteUploadSize), bucket.Name, location)
}

func init() {
	prometheus.MustRegister(version.NewCollector(program))
}

func getEnv(key string, defaultVal string) string {
	if env, ok := os.LookupEnv(key); ok {
		return env
	}
	return defaultVal
}

func main() {
	var (
		printVersion  = flag.Bool("version", false, "Print version information.")
		listenAddress = flag.String("web.listen-address", getEnv("LISTEN_ADDRESS", ":9290"), "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", getEnv("MINIO_METRIC_PATH", "/metrics"), "Path under which to expose metrics.")
		minioURI      = flag.String("minio.server", getEnv("MINIO_URL", "http://localhost:9000"), "HTTP address of the Minio server")
		minioKey      = flag.String("minio.access-key", getEnv("MINIO_ACCESS_KEY", ""), "The access key used to login in to Minio.")
		minioSecret   = flag.String("minio.access-secret", getEnv("MINIO_ACCESS_SECRET", ""), "The access secret used to login in to Minio")
		bucketStats   = flag.Bool("minio.bucket-stats", false, "Collect bucket statistics. It can take long.")
	)

	flag.Parse()

	if *printVersion {
		fmt.Fprintln(os.Stdout, version.Print("minio_exporter"))
		os.Exit(0)
	}

	exporter, err := NewMinioExporter(*minioURI, *minioKey, *minioSecret, *bucketStats)
	if err != nil {
		log.Fatalln(err)
	}

	log.Infoln("Starting minio_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
                        <head><title>Minio Exporter</title></head>
                        <body>
                        <h1>Minio Exporter</h1>
                        <p><a href='` + *metricsPath + `'>Metrics</a></p>
                        </body>
                        </html>`))
	})

	log.Infoln("Listening on", *listenAddress)
	err = http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}
