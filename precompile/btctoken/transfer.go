package btctoken

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
	"math/big"
)

const TransferEventName = "Transfer"

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

