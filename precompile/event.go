package precompile

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v12/x/evm/statedb"
)

type EventArgument struct {
	Indexed bool
	Value   interface{}
}

type Event interface {
	EventName() string
	Arguments() []*EventArgument
}

type EventEmitter struct {
	sdkCtx  sdk.Context
	abi     abi.ABI
	address common.Address
	stateDB *statedb.StateDB
}

func NewEventEmitter(
	sdkCtx sdk.Context,
	abi abi.ABI,
	address common.Address,
	stateDB *statedb.StateDB,
) *EventEmitter {
	return &EventEmitter{
		sdkCtx:  sdkCtx,
		abi:     abi,
		address: address,
		stateDB: stateDB,
	}
}

func (ee *EventEmitter) Emit(event Event) error {
	abiEvent := ee.abi.Events[event.EventName()]

	indexedArguments := make([]interface{}, 0)
	arguments := make([]interface{}, 0)

	for _, arg := range event.Arguments() {
		if arg.Indexed {
			indexedArguments = append(indexedArguments, arg.Value)
		} else {
			arguments = append(arguments, arg.Value)
		}
	}

	indexedArgumentsTopics, err := abi.MakeTopics(indexedArguments)
	if err != nil {
		return fmt.Errorf("failed to make topics: [%w]", err)
	}

	// The first topic is always the ID of the event.
	topics := append([]common.Hash{abiEvent.ID}, indexedArgumentsTopics[0]...)

	// Pack non-indexed arguments using the event ABI. Note that we need
	// to shift the arguments slice to exclude indexed arguments.
	data, err := abiEvent.Inputs[len(indexedArguments):].Pack(arguments)
	if err != nil {
		return err
	}

	ee.stateDB.AddLog(&ethtypes.Log{
		Address:     ee.address,
		Topics:      topics,
		Data:        data,
		BlockNumber: uint64(ee.sdkCtx.BlockHeight()),
	})

	return nil
}
