package btctoken

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
)

// TransferEventName is the name of the Transfer event. It matches the name
// of the event in the contract ABI.
const TransferEventName = "Transfer"

// transferEvent is the implementation of the Transfer event that contains
// the following arguments:
// - from (indexed): the address from which the tokens are transferred,
// - to (indexed): the address to which the tokens are transferred,
// - value (non-indexed): the amount of tokens transferred.
type transferEvent struct {
	from, to common.Address
	value    *big.Int
}

func newTransferEvent(from, to common.Address, value *big.Int) *transferEvent {
	return &transferEvent{
		from:  from,
		to:    to,
		value: value,
	}
}

func (te *transferEvent) EventName() string {
	return TransferEventName
}

func (te *transferEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   te.from,
		},
		{
			Indexed: true,
			Value:   te.to,
		},
		{
			Indexed: false,
			Value:   te.value,
		},
	}
}
