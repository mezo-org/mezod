module github.com/mezo-org/mezod/infrastructure/terraform/mezo-staging/gcf/faucet

go 1.22.5

require (
	github.com/GoogleCloudPlatform/functions-framework-go v1.9.0
	github.com/ethereum/go-ethereum v1.14.8
)

require (
	cloud.google.com/go/functions v1.16.6 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.4 // indirect
	github.com/cloudevents/sdk-go/v2 v2.15.2 // indirect
	github.com/deckarep/golang-set v1.8.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/holiman/uint256 v1.3.1 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v0.0.0-20180701023420-4b7aa43c6742 // indirect
	github.com/rjeczalik/notify v0.9.1 // indirect
	github.com/shirou/gopsutil v3.21.4-0.20210419000835-c7a38de76ee5+incompatible // indirect
	github.com/stretchr/testify v1.8.2 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/crypto v0.25.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
)

// use Evmos geth fork
replace github.com/ethereum/go-ethereum => github.com/evmos/go-ethereum v1.10.26-evmos-rc2
