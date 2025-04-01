package main

import (
	"flag"

	"github.com/mezo-org/mezod/monitoring"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "config.json", "a path to the configuration file")
}

func main() {
	flag.Parse()

	monitoring.Start(configPath)
}
