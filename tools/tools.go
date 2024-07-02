// +build tools

package tools

import (
	// Packages needed for smart contract bindings generation.
	// `go mod tidy` considers these packages unneeded and removes them if they
	// are not included here.
	_ "github.com/cpuguy83/go-md2man/v2"
	_ "gopkg.in/yaml.v2"
	_ "github.com/peterh/liner"
	_ "github.com/graph-gophers/graphql-go"
	_ "github.com/influxdata/influxdb-client-go/v2"
	_ "github.com/influxdata/influxdb"
)
