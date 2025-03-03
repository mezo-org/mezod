//go:build testbed

package types

// start these with the prefix: 0x7b7c1 in order to not conflict with any
// production ones
const (
	TestBedStrippedERC20PrecompileAddress       = "0x7b7c100000000000000000000000000000000000"
	TestBedStrippedERC20PrecompileLatestVersion = 1
)

func getDefaultPrecompilesVersions() []*PrecompileVersionInfo {
	return append(DefaultPrecompilesVersions, &PrecompileVersionInfo{
		TestBedStrippedERC20PrecompileAddress, TestBedStrippedERC20PrecompileLatestVersion,
	})
}
