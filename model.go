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

import "time"

type CurrentWeather struct {
	Temperature   float64
	WindSpeed     float64 `json:"windspeed"`
	WindDirection float64 `json:"winddirection"`
	RHumidity float64 `json:"relative_humidity_2m"`
	Precipitation   float64 `json:"precipitation"`
	WeatherCode     string  `json:"weather_code"`
}

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

type Location struct {
	Name        string
	TtlMinutes  int
	Coordinates `yaml:",inline"`
}

type Response struct {
	Coordinates    `json:",inline"`
	CurrentWeather CurrentWeather `json:"current"`
}

type CacheEntry struct {
	Response   *Response
	LastUpdate time.Time
}

type Config struct {
	Locations []Location
}
