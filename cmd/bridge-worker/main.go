package main

import (
	"flag"

	bridgeworker "github.com/mezo-org/mezod/bridge-worker"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "config.json", "a path to the configuration file")
}

func main() {
	flag.Parse()
	bridgeworker.Start(configPath)
}
