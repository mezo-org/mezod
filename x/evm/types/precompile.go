package types

const (
	BTCTokenPrecompileAddress       = "0x7b7c000000000000000000000000000000000000"
	BTCTokenPrecompileLatestVersion = 1
)

const (
	MEZOTokenPrecompileAddress       = "0x7b7c000000000000000000000000000000000001"
	MEZOTokenPrecompileLatestVersion = 1
)

const (
	ValidatorPoolPrecompileAddress       = "0x7b7c000000000000000000000000000000000011"
	ValidatorPoolPrecompileLatestVersion = 1
)

const (
	AssetsBridgePrecompileAddress       = "0x7b7c000000000000000000000000000000000012"
	AssetsBridgePrecompileLatestVersion = 2
)

const (
	MaintenancePrecompileAddress       = "0x7b7c000000000000000000000000000000000013"
	MaintenancePrecompileLatestVersion = 3
)

const (
	UpgradePrecompileAddress       = "0x7b7c000000000000000000000000000000000014"
	UpgradePrecompileLatestVersion = 1
)

const (
	PriceOraclePrecompileAddress       = "0x7b7c000000000000000000000000000000000015"
	PriceOraclePrecompileLatestVersion = 1
)

// Start the TestBed precompile address with the prefix 0x7b7c1 in order
// to not conflict with any production precompile.
const (
	TestBedPrecompileAddress       = "0x7b7c100000000000000000000000000000000000"
	TestBedPrecompileLatestVersion = 1
)

// DefaultPrecompilesVersions is a list of default precompiles and their versions.
// Order of precompiles is important. If changed on a live network, it will break
// the consensus.
var DefaultPrecompilesVersions = []*PrecompileVersionInfo{
	{BTCTokenPrecompileAddress, BTCTokenPrecompileLatestVersion},
	{ValidatorPoolPrecompileAddress, ValidatorPoolPrecompileLatestVersion},
	{AssetsBridgePrecompileAddress, AssetsBridgePrecompileLatestVersion},
	{MaintenancePrecompileAddress, MaintenancePrecompileLatestVersion},
	{UpgradePrecompileAddress, UpgradePrecompileLatestVersion},
	{PriceOraclePrecompileAddress, PriceOraclePrecompileLatestVersion},
	{TestBedPrecompileAddress, TestBedPrecompileLatestVersion},
}
