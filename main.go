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
	"fmt"
	"net/http"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"gopkg.in/yaml.v3"
)

const (
	name = "openmeteo_exporter"
)

var (
	webConfig = webflag.AddFlags(kingpin.CommandLine, ":9113")
	cfgFile   = kingpin.Flag(
		"config-file",
		"Path to config file.",
	).Default("config.yaml").String()
	metricsPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()

	exp = &exporter{}
)

func loadConfig(cfgFile string) (*Config, error) {
	var cfg Config
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {
	exp.init()
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print(name))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(promlogConfig)
	level.Info(logger).Log("msg", fmt.Sprintf("Starting %s", name),
		"version", version.Info(),
		"config", *cfgFile)
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	config, err := loadConfig(*cfgFile)
	if err != nil {
		panic(err)
	}

	level.Info(logger).Log("msg", fmt.Sprintf("Got %d targets", len(config.Locations)))

	var landingPage = []byte(`<html>
<head><title>Open Meteo exporter</title></head>
<body>
<h1>Open Meteo exporter</h1>
<p><a href='` + *metricsPath + `'>Metrics</a></p>
</body>
</html>
`)

	prometheus.MustRegister(exp)
	prometheus.MustRegister(version.NewCollector(name))

	exp.config = config
	exp.logger = logger
	exp.init()
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err = w.Write(landingPage); err != nil {
			level.Error(logger).Log("msg", "Unable to write page content", "err", err)
		}
	})

	srv := &http.Server{}
	if err = web.ListenAndServe(srv, webConfig, logger); err != nil {
		panic(err)
	}
}
