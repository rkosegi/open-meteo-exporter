/*
 * Copyright 2024 Richard Kosegi
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package internal

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/rkosegi/open-meteo-exporter/types"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	subsystem = "exporter"
	namespace = "openmeteo"
	baseUri   = "https://api.open-meteo.com/v1/forecast"
)

type exporter struct {
	logger       *slog.Logger
	scrapeErrors prometheus.Counter
	totalScrapes prometheus.Counter

	tempDesc            *prometheus.GaugeVec
	tempApparentDesc    *prometheus.GaugeVec
	relHumidityDesc     *prometheus.GaugeVec
	precipitationDesc   *prometheus.GaugeVec
	rainDesc            *prometheus.GaugeVec
	showersDesc         *prometheus.GaugeVec
	snowfallDesc        *prometheus.GaugeVec
	cloudCoverDesc      *prometheus.GaugeVec
	surfacePressureDesc *prometheus.GaugeVec
	pressureMslDesc     *prometheus.GaugeVec
	windSpeedDesc       *prometheus.GaugeVec
	windDirDesc         *prometheus.GaugeVec
	windGustsDesc       *prometheus.GaugeVec
	cacheHit            *prometheus.CounterVec
	httpFetchDuration   prometheus.Summary
	httpTraffic         prometheus.Counter
	config              *types.Config
	client              http.Client
	cache               map[string]types.CacheEntry
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	e.tempDesc.Describe(ch)
	e.tempApparentDesc.Describe(ch)
	e.relHumidityDesc.Describe(ch)
	e.precipitationDesc.Describe(ch)
	e.rainDesc.Describe(ch)
	e.showersDesc.Describe(ch)
	e.snowfallDesc.Describe(ch)
	e.cloudCoverDesc.Describe(ch)
	e.surfacePressureDesc.Describe(ch)
	e.pressureMslDesc.Describe(ch)
	e.windSpeedDesc.Describe(ch)
	e.windDirDesc.Describe(ch)
	e.windGustsDesc.Describe(ch)

	e.httpFetchDuration.Describe(ch)
	e.httpTraffic.Describe(ch)
	e.cacheHit.Describe(ch)
	e.totalScrapes.Describe(ch)
	e.scrapeErrors.Describe(ch)
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	e.totalScrapes.Inc()
	e.scrape(ch)
	e.totalScrapes.Collect(ch)
	e.scrapeErrors.Collect(ch)
}

func (e *exporter) onError(err error) {
	e.logger.Error("Error while fetching data", "error", err)
	e.scrapeErrors.Inc()
}

func (e *exporter) scrapeTarget(target types.Location, ch chan<- prometheus.Metric) {
	if target.FetchMethod == nil || *target.FetchMethod == types.FetchMethodDefault {
		e.handleDefault(target, ch)
	} else if *target.FetchMethod == types.FetchMethodAlt {
		e.handleAlt(target, ch)
	}

}

func (e *exporter) scrape(ch chan<- prometheus.Metric) {
	start := time.Now().UnixMilli()
	for _, target := range e.config.Locations {
		e.scrapeTarget(target, ch)
	}
	e.tempDesc.Collect(ch)
	e.tempApparentDesc.Collect(ch)
	e.relHumidityDesc.Collect(ch)
	e.precipitationDesc.Collect(ch)
	e.rainDesc.Collect(ch)
	e.showersDesc.Collect(ch)
	e.snowfallDesc.Collect(ch)
	e.cloudCoverDesc.Collect(ch)
	e.surfacePressureDesc.Collect(ch)
	e.pressureMslDesc.Collect(ch)
	e.windSpeedDesc.Collect(ch)
	e.windDirDesc.Collect(ch)
	e.windGustsDesc.Collect(ch)

	e.httpFetchDuration.Observe(float64(time.Now().UnixMilli() - start))
	e.httpFetchDuration.Collect(ch)
	e.httpTraffic.Collect(ch)
}

func (e *exporter) init() {
	e.tempDesc = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "current",
		Name:      "temperature",
		Help:      "The current temperature.",
	}, []string{"location"})

	e.tempApparentDesc = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "current",
		Name:      "apparent_temperature",
		Help:      "The apparent temperature.",
	}, []string{"location"})

	e.relHumidityDesc = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "current",
		Name:      "relative_humidity",
		Help:      "The relative humidity.",
	}, []string{"location"})

	e.precipitationDesc = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "current",
		Name:      "precipitation",
		Help:      "Probability of precipitation.",
	}, []string{"location"})

	e.rainDesc = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "current",
		Name:      "rain",
		Help:      "Rain from large scale weather systems",
	}, []string{"location"})

	e.showersDesc = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "current",
		Name:      "showers",
		Help:      "Showers from convective precipitation",
	}, []string{"location"})

	e.snowfallDesc = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "current",
		Name:      "snowfall",
		Help:      "The snowfall.",
	}, []string{"location"})

	e.cloudCoverDesc = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "current",
		Name:      "cloud_cover",
		Help:      "Total cloud cover as an area fraction.",
	}, []string{"location"})

	e.surfacePressureDesc = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "current",
		Name:      "surface_pressure",
		Help:      "Atmospheric air pressure at surface",
	}, []string{"location"})

	e.pressureMslDesc = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "current",
		Name:      "pressure_msl",
		Help:      "Atmospheric air pressure reduced to mean sea level",
	}, []string{"location"})

	e.windSpeedDesc = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "current",
		Name:      "wind_speed",
		Help:      "The current wind speed.",
	}, []string{"location"})

	e.windDirDesc = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "current",
		Name:      "wind_dir",
		Help:      "The current wind direction.",
	}, []string{"location"})

	e.windGustsDesc = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "current",
		Name:      "wind_gusts",
		Help:      "Wind gusts at 10 meters above ground",
	}, []string{"location"})

	e.totalScrapes = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "total_scrapes",
		Help:      "Total number of times this exporter was scraped for metrics.",
	})

	e.scrapeErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "scrape_errors",
		Help:      "Total number of times an error occurred during scraping operation.",
	})

	e.httpFetchDuration = prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "http_fetch_duration",
		Help:      "Total time spent on fetching data from api.open-meteo.com",
	})

	e.httpTraffic = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "http_rx_bytes",
		Help:      "Total bytes received from api.open-meteo.com",
	})

	e.cacheHit = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "cache_hit",
		Help:      "Total number of times cache was hit",
	}, []string{"location"})

	e.client = http.Client{
		Timeout: time.Second * 30,
	}
}

func NewExporter(config *types.Config, logger *slog.Logger) prometheus.Collector {
	e := &exporter{
		logger: logger,
		config: config,
		cache:  map[string]types.CacheEntry{},
	}
	e.init()
	return e
}
