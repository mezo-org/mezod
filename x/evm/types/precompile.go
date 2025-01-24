package types

const (
	BTCTokenPrecompileAddress       = "0x7b7c000000000000000000000000000000000000"
	BTCTokenPrecompileLatestVersion = 1
)

const (
	ValidatorPoolPrecompileAddress       = "0x7b7c000000000000000000000000000000000011"
	ValidatorPoolPrecompileLatestVersion = 1
)

const (
	AssetsBridgePrecompileAddress       = "0x7b7c000000000000000000000000000000000012"
	AssetsBridgePrecompileLatestVersion = 1
)

const (
	MaintenancePrecompileAddress       = "0x7b7c000000000000000000000000000000000013"
	MaintenancePrecompileLatestVersion = 2
)

const (
	UpgradePrecompileAddress       = "0x7b7c000000000000000000000000000000000014"
	UpgradePrecompileLatestVersion = 1
)

const (
	PriceOraclePrecompileAddress       = "0x7b7c000000000000000000000000000000000015"
	PriceOraclePrecompileLatestVersion = 1
)

var DefaultPrecompilesVersions = map[string]uint32{
	BTCTokenPrecompileAddress:      BTCTokenPrecompileLatestVersion,
	ValidatorPoolPrecompileAddress: ValidatorPoolPrecompileLatestVersion,
	AssetsBridgePrecompileAddress:  AssetsBridgePrecompileLatestVersion,
	MaintenancePrecompileAddress:   MaintenancePrecompileLatestVersion,
	UpgradePrecompileAddress:       UpgradePrecompileLatestVersion,
	PriceOraclePrecompileAddress:   PriceOraclePrecompileLatestVersion,
}
