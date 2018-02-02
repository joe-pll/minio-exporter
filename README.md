MINIO EXPORTER
==============

A Prometheus exporter for Minio cloud storage server

## Compatibility
| minio-exporter version | Minio version                   |
| ---------------------- | ------------------------------- |
| v0.1.0                 | <= RELEASE.2018-01-02T23-07-00Z |
| v0.2.0                 | == RELEASE.2018-01-18T20-33-21Z |

Build and launch the exporter as following.  
```bash
make
./minio_exporter [flags]
```
## Flags

```bash
./minio_exporter --help
```

| Flag | Description | Default |
| ---- | ------------| ------- |
| version | Print version number and leave | |
| web.listen-address | The address to listen on to expose metrics. | *:9290* |
| web.telemetry-path | The listening path for metrics. | */metrics* |
| minio.server | The URL of the minio server. Use HTTPS if Minio accepts secure connections only. | *http://localhost:9000* |
| minio.access-key | The value of the Minio access key. It is required in order to connect to the server | "" |
| minio.access-secret | The calue of the Minio access secret. It is required in order to connect to the server | "" |
| minio.bucket-stats | Collect statistics about the buckets and files in buckets. It requires more computation, use it carefully in case of large buckets. | false |

```bash
./minio_exporter -minio.server another-host:9000 -minio.access-key "login_name" -minio.access-secret "login_password"
```

## Run with Docker
You can execute the exporter using the Docker image.

```bash
docker pull joepll/minio-exporter
docker run -p 9290:9290 joepll/minio-exporter -minio.server "minio.host:9000" -minio.access-key "login_name" -minio.access-secret "login_secret"
```

The same result can be achieved with Enviroment variables.
* **LISTEN_ADDRESS**: is the exporter address, as the option *web.listen-address*
* **METRIC_PATH**: the telemetry path. It corresponds to *web.telemetry-path*
* **MINIO_URL**: the URL of the Minio server, as *minio.server*
* **MINIO_ACCESS_KEY**: the Minio access key (*minio.access-key*)
* **MINIO_ACCESS_SECRET**: the Minio access secret (*minio.access-secret*)


```bash
docker run \
       -p 9290:9290 \
       -e "MINIO_URL=http://host.local:9000" \
       -e "MINIO_ACCESS_KEY=loginname" \
       -e "MINIO_ACCESS_SECRET=password" \
       joepll/minio-exporter
```


# Exported metrics
| Metric | Description | Default |
| ------ | ----------- | ------- |
| minio_collector_success | Whether or not the collector succeeded | true |
| minio_collector_duration_seconds | Duration in second spent in collecting metrics | true |
| minio_uptime | Server uptime in seconds | true |
| minio_conn_total_input_bytes | The number of total bytes received in input from the server | true |
| minio_conn_total_output_bytes | The number of total bytes transmitted by the server | true |
| minio_storage_total_disk_space | The number of bytes of the total storage available in Minio | true |
| minio_storage_free_disk_space | The remaining disk space in bytes | true |
| minio_storage_online_disks | Count containing the available number of disks | true |
| minio_offline_disks | The number of disks not reachable by the servers | true |
| minio_read_quorum | Minio read quorum | true |
| minio_write_quorum | Minio write quorum | true |
| minio_storage_type | The type of storaged used by Minio. It can be 1 = FS or 2 = Erasure | true |
| minio_buckets_object_number | The number of objects for each bucket. The bucket name is expressed in a label | false |
| minio_objects_total_size | The total size of all the objects per bucket. | false |
| minio_max_object_size | The biggest element in size into the bucket | false |
| minio_incomplete_uploads_number | The number of incomplete updloads | false |
| minio_incomplete_uploads_total_size | Total size in bytes of all the objects incomplete upload per bucket | false |
