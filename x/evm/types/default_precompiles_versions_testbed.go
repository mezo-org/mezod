//go:build testbed

package types

func getDefaultPrecompilesVersions() []*PrecompileVersionInfo {
	return append(DefaultPrecompilesVersions, &PrecompileVersionInfo{
		TestBedStrippedERC20PrecompileAddress, TestBedStrippedERC20PrecompileLatestVersion,
	})
}
