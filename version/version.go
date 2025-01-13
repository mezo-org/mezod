// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE
package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

var (
	AppVersion  = ""
	GitCommit   = ""
	GitModified = ""
	BuildDate   = ""

	GoVersion = ""
	GoArch    = ""
)

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				GitCommit = setting.Value
			case "vcs.modified":
				GitModified = setting.Value
			case "vcs.time":
				BuildDate = setting.Value

			}
		}
	}

	// if we are not in a git repository, this value will be 0
	if len(AppVersion) == 0 {
		AppVersion = "dev"
	} else {
		// only set in case we are in a git repository, then we know
		// the code on the commit have been altered
		if GitModified == "true" {
			GitCommit += "-modified"
		}
	}

	GoVersion = runtime.Version()
	GoArch = runtime.GOARCH

}

func Version() string {
	return fmt.Sprintf(
		"Version %s (%s) Compiled at %s using Go %s (%s)",
		AppVersion,
		GitCommit,
		BuildDate,
		GoVersion,
		GoArch,
	)
}
