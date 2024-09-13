package faucet

import (
	"fmt"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

// amount: 0.01 BTC = 10000000000000000 abtc
// privkey: 0x....
// btctoken address = 0x7b7c000000000000000000000000000000000000

func init() {
	functions.HTTP("Distribute", distribute)
}

func distribute(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}
