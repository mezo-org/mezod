//go:build tools
// +build tools

package tools

import (
	// Packages needed for smart contract bindings generation.
	// `go mod tidy` considers these packages unneeded and removes them.
	// This workaround ensures the packages are not removed.
	_ "github.com/cpuguy83/go-md2man/v2"
	_ "github.com/graph-gophers/graphql-go"
	_ "github.com/influxdata/influxdb"
	_ "github.com/influxdata/influxdb-client-go/v2"
	_ "github.com/peterh/liner"
	_ "gopkg.in/yaml.v2"
)
