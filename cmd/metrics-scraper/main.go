package main

import (
	"flag"

	metricsscraper "github.com/mezo-org/mezod/metrics-scraper"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "config.json", "a path to the configuration file")
}

func main() {
	flag.Parse()

	metricsscraper.Start(configPath)
}
