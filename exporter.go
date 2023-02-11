/*
Copyright 2023 Richard Kosegi

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"net/http"
	"time"
)

const (
	subsystem = "exporter"
	namespace = "openmeteo"
	baseUri   = "https://api.open-meteo.com/v1/forecast"
)

var (
	cache = map[string]CacheEntry{}
)

type exporter struct {
	logger            log.Logger
	tempDesc          *prometheus.GaugeVec
	scrapeErrors      prometheus.Counter
	totalScrapes      prometheus.Counter
	windSpeedDesc     *prometheus.GaugeVec
	windDirDesc       *prometheus.GaugeVec
	cacheHit          *prometheus.CounterVec
	httpFetchDuration prometheus.Summary
	httpTraffic       prometheus.Counter
	config            *Config
	client            http.Client
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	e.tempDesc.Describe(ch)
	e.windSpeedDesc.Describe(ch)
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
	level.Error(e.logger).Log("msg", "Error while fetching data", "error", err)
	e.scrapeErrors.Inc()
}

func (e *exporter) scrapeTarget(target Location, ch chan<- prometheus.Metric) {
	var respObj Response
	if target.TtlMinutes == 0 {
		target.TtlMinutes = 600
	}

	var fetch = true
	last, present := cache[target.Name]
	if present {
		if time.Now().Unix() < int64(target.TtlMinutes*60)+last.LastUpdate.Unix() {
			fetch = false
			e.cacheHit.WithLabelValues(target.Name).Inc()
		}
	}
	if fetch {
		var uri = fmt.Sprintf("%s?latitude=%.2f&longitude=%.2f&current_weather=true",
			baseUri, target.Latitude, target.Longitude)

		req, err := http.NewRequest(http.MethodGet, uri, nil)
		if err != nil {
			e.onError(err)
			return
		}
		req.Header.Set("accept", "application/json")
		req.Header.Set("content-type", "application/json")

		resp, err := e.client.Do(req)
		if err != nil {
			e.onError(err)
			return
		}

		defer func(body io.Closer) {
			_ = body.Close()
		}(resp.Body)

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			e.onError(err)
			return
		}
		e.httpTraffic.Add(float64(len(data)))
		err = json.Unmarshal(data, &respObj)
		if err != nil {
			e.onError(err)
			return
		}
		cache[target.Name] = CacheEntry{
			Response:   &respObj,
			LastUpdate: time.Now(),
		}
	} else {
		respObj = *last.Response
	}
	e.tempDesc.WithLabelValues(target.Name).Set(respObj.CurrentWeather.Temperature)
	e.windSpeedDesc.WithLabelValues(target.Name).Set(respObj.CurrentWeather.WindSpeed)
	e.windDirDesc.WithLabelValues(target.Name).Set(respObj.CurrentWeather.WindDirection)

	e.tempDesc.Collect(ch)
	e.windSpeedDesc.Collect(ch)
	e.windDirDesc.Collect(ch)
	e.cacheHit.Collect(ch)
}

func (e *exporter) scrape(ch chan<- prometheus.Metric) {
	start := time.Now().UnixMilli()
	for _, target := range e.config.Locations {
		e.scrapeTarget(target, ch)
	}
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
