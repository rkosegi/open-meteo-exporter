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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rkosegi/open-meteo-exporter/types"
)

func (e *exporter) handleDefault(loc types.Location, ch chan<- prometheus.Metric) {
	var respObj types.Response
	if loc.TtlMinutes == 0 {
		loc.TtlMinutes = 10
	}

	var fetch = true
	last, present := e.cache[loc.Name]
	if present {
		if time.Now().Unix() < int64(loc.TtlMinutes*60)+last.LastUpdate.Unix() {
			fetch = false
			e.cacheHit.WithLabelValues(loc.Name).Inc()
		}
	}
	if fetch {
		var uri = fmt.Sprintf("%s?latitude=%.2f&longitude=%.2f&current_weather=true",
			baseUri, loc.Latitude, loc.Longitude)

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
		e.cache[loc.Name] = types.CacheEntry{
			Response:   &respObj,
			LastUpdate: time.Now(),
		}
	} else {
		respObj = *last.Response.(*types.Response)
	}
	e.tempDesc.WithLabelValues(loc.Name).Set(respObj.CurrentWeather.Temperature)
	e.windSpeedDesc.WithLabelValues(loc.Name).Set(respObj.CurrentWeather.WindSpeed)
	e.windDirDesc.WithLabelValues(loc.Name).Set(respObj.CurrentWeather.WindDirection)

	e.tempDesc.Collect(ch)
	e.windSpeedDesc.Collect(ch)
	e.windDirDesc.Collect(ch)
	e.cacheHit.Collect(ch)
}

func (e *exporter) handleAlt(loc types.Location, ch chan<- prometheus.Metric) {
	var respObj types.ResponseAlt
	if loc.TtlMinutes == 0 {
		loc.TtlMinutes = 10
	}

	var fetch = true
	entry, present := e.cache[loc.Name]
	if present {
		if time.Now().Unix() < int64(loc.TtlMinutes*60)+entry.LastUpdate.Unix() {
			fetch = false
			e.cacheHit.WithLabelValues(loc.Name).Inc()
		}
	}
	if fetch {
		var uri = fmt.Sprintf("%s?latitude=%.2f&longitude=%.2f&current=temperature_2m,relative_humidity_2m,"+
			"apparent_temperature,is_day,precipitation,rain,showers,snowfall,weather_code,cloud_cover,pressure_msl,"+
			"surface_pressure,wind_speed_10m,wind_direction_10m,wind_gusts_10m",
			baseUri, loc.Latitude, loc.Longitude)

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
		e.cache[loc.Name] = types.CacheEntry{
			Response:   &respObj,
			LastUpdate: time.Now(),
		}
	} else {
		respObj = *entry.Response.(*types.ResponseAlt)
	}
	if respObj.CurrentWeather.Temperature != nil {
		e.tempDesc.WithLabelValues(loc.Name).Set(*respObj.CurrentWeather.Temperature)
	}
	if respObj.CurrentWeather.ApparentTemperature != nil {
		e.tempApparentDesc.WithLabelValues(loc.Name).Set(*respObj.CurrentWeather.ApparentTemperature)
	}
	if respObj.CurrentWeather.RelativeHumidity != nil {
		e.relHumidityDesc.WithLabelValues(loc.Name).Set(*respObj.CurrentWeather.RelativeHumidity)
	}
	if respObj.CurrentWeather.Precipitation != nil {
		e.precipitationDesc.WithLabelValues(loc.Name).Set(*respObj.CurrentWeather.Precipitation)
	}
	if respObj.CurrentWeather.Rain != nil {
		e.rainDesc.WithLabelValues(loc.Name).Set(*respObj.CurrentWeather.Rain)
	}
	if respObj.CurrentWeather.Showers != nil {
		e.showersDesc.WithLabelValues(loc.Name).Set(*respObj.CurrentWeather.Showers)
	}
	if respObj.CurrentWeather.Snowfall != nil {
		e.snowfallDesc.WithLabelValues(loc.Name).Set(*respObj.CurrentWeather.Snowfall)
	}
	if respObj.CurrentWeather.CloudCover != nil {
		e.cloudCoverDesc.WithLabelValues(loc.Name).Set(*respObj.CurrentWeather.CloudCover)
	}
	if respObj.CurrentWeather.SurfacePressure != nil {
		e.surfacePressureDesc.WithLabelValues(loc.Name).Set(*respObj.CurrentWeather.SurfacePressure)
	}
	if respObj.CurrentWeather.PressureMsl != nil {
		e.pressureMslDesc.WithLabelValues(loc.Name).Set(*respObj.CurrentWeather.PressureMsl)
	}
	if respObj.CurrentWeather.WindSpeed != nil {
		e.windSpeedDesc.WithLabelValues(loc.Name).Set(*respObj.CurrentWeather.WindSpeed)
	}
	if respObj.CurrentWeather.WindDirection != nil {
		e.windDirDesc.WithLabelValues(loc.Name).Set(*respObj.CurrentWeather.WindDirection)
	}
	if respObj.CurrentWeather.WindGusts != nil {
		e.windGustsDesc.WithLabelValues(loc.Name).Set(*respObj.CurrentWeather.WindGusts)
	}
}
