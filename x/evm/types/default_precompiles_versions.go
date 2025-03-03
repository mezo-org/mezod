//go:build !testbed

package types

func getDefaultPrecompilesVersions() []*PrecompileVersionInfo {
	return DefaultPrecompilesVersions
}
