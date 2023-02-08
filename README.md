# Open Meteo Exporter

This is prometheus exporter for [Open Meteo](https://open-meteo.com/) current conditions at particular place.

### Building

#### Docker build

```shell
make build-docker
```

#### Local build


```shell
make build-local
```

### Running

Here is example `config.yaml`:

```yaml
---
locations:
  - name: Vienna
    latitude: 48.2082
    longitude: 16.3738
```

Start exporter

```shell
./owm-exporter
ts=2023-02-08T17:31:29.678Z caller=main.go:76 level=info msg="Starting openmeteo_exporter" version="(version=, branch=, revision=7a038743ac2af96be06afd015ee88aad1e9d8376-modified)" config=config.yaml
ts=2023-02-08T17:31:29.678Z caller=main.go:79 level=info msg="Build context" build_context="(go=go1.19.5, platform=linux/amd64, user=, date=)"
ts=2023-02-08T17:31:29.678Z caller=main.go:86 level=info msg="Got 1 targets"
ts=2023-02-08T17:31:29.680Z caller=tls_config.go:232 level=info msg="Listening on" address=[::]:9113
ts=2023-02-08T17:31:29.680Z caller=tls_config.go:235 level=info msg="TLS is disabled." http2=false address=[::]:9113
```

Example collector output from  `http://localhost:9113/metrics`
```
# HELP openmeteo_current_temperature The current temperature.
# TYPE openmeteo_current_temperature gauge
openmeteo_current_temperature{location="Vienna"} -0.1
# HELP openmeteo_current_wind_dir The current wind direction.
# TYPE openmeteo_current_wind_dir gauge
openmeteo_current_wind_dir{location="Vienna"} 137
# HELP openmeteo_current_wind_speed The current wind speed.
# TYPE openmeteo_current_wind_speed gauge
openmeteo_current_wind_speed{location="Vienna"} 5.9
# HELP openmeteo_exporter_build_info A metric with a constant '1' value labeled by version, revision, branch, goversion from which openmeteo_exporter was built, and the goos and goarch for the build.
# TYPE openmeteo_exporter_build_info gauge
openmeteo_exporter_build_info{branch="",goarch="amd64",goos="linux",goversion="go1.19.5",revision="7a038743ac2af96be06afd015ee88aad1e9d8376-modified",version=""} 1
# HELP openmeteo_exporter_http_fetch_duration Total time spent on fetching data from api.open-meteo.com
# TYPE openmeteo_exporter_http_fetch_duration summary
openmeteo_exporter_http_fetch_duration_sum 260
openmeteo_exporter_http_fetch_duration_count 1
# HELP openmeteo_exporter_http_rx_bytes Total bytes received from api.open-meteo.com
# TYPE openmeteo_exporter_http_rx_bytes counter
openmeteo_exporter_http_rx_bytes 282
# HELP openmeteo_exporter_scrape_errors Total number of times an error occurred during scraping operation.
# TYPE openmeteo_exporter_scrape_errors counter
openmeteo_exporter_scrape_errors 0
# HELP openmeteo_exporter_total_scrapes Total number of times this exporter was scraped for metrics.
# TYPE openmeteo_exporter_total_scrapes counter
openmeteo_exporter_total_scrapes 1
```
