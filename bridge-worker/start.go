package bridgeworker

import (
	"crypto/ecdsa"
)

func Start(configPath string) {
	// TODO: Read the following from config
	var providerURL string = ""
	var ethereumNetwork string = ""
	var privateKey *ecdsa.PrivateKey = nil

	RunBridgeWorker(
		providerURL,
		ethereumNetwork,
		privateKey,
	)
}
