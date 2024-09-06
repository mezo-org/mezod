package types

const (
	// ModuleName defines the module name
	ModuleName = "bridge"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName
)

var (
	ParamsKey = []byte{0x10} // standalone key for module params

	AssetsLockedSequenceTipKey = []byte{0x20} // standalone key for the assets locked sequence tip
)
