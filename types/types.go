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

package types

import "time"

type CurrentWeatherDefault struct {
	Temperature   float64
	WindSpeed     float64 `json:"windspeed"`
	WindDirection float64 `json:"winddirection"`
}

type CurrentWeatherAlt struct {
	Temperature         *float64 `json:"temperature_2m"`
	ApparentTemperature *float64 `json:"apparent_temperature"`
	RelativeHumidity    *float64 `json:"relative_humidity_2m"`
	Precipitation       *float64 `json:"precipitation"`
	Rain                *float64 `json:"rain"`
	Showers             *float64 `json:"showers"`
	Snowfall            *float64 `json:"snowfall"`
	CloudCover          *float64 `json:"cloud_cover"`
	SurfacePressure     *float64 `json:"surface_pressure"`
	PressureMsl         *float64 `json:"pressure_msl"`
	WindSpeed           *float64 `json:"wind_speed_10m"`
	WindDirection       *float64 `json:"wind_direction_10m"`
	WindGusts           *float64 `json:"wind_gusts_10m"`
	WeatherCode         *float64 `json:"weather_code"`
}

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

type FetchMethod string

const (
	FetchMethodDefault = "default"
	FetchMethodAlt     = "alt"
)

type Location struct {
	Name        string
	FetchMethod *FetchMethod `yaml:"method,omitempty"`
	TtlMinutes  int
	Coordinates `yaml:",inline"`
}

type Response struct {
	Coordinates    `json:",inline"`
	CurrentWeather CurrentWeatherDefault `json:"current_weather"`
}

type ResponseAlt struct {
	Coordinates    `json:",inline"`
	CurrentWeather CurrentWeatherAlt `json:"current"`
}

type CacheEntry struct {
	Response   interface{}
	LastUpdate time.Time
}

type Config struct {
	Locations []Location
}
