package bridgeworker

type Config struct {
	ProviderURL     string `json:"provider_url"`     // e.g. http://127.0.0.1:8545
	EthereumNetwork string `json:"ethereum_network"` // e.g. "sepolia" or "mainnet"
}
