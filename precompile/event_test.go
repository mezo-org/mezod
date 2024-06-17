package precompile

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/x/evm/statedb"
	"math/big"
	"reflect"
	"testing"
)

func TestEventEmitter_Emit(t *testing.T) {
	sdkCtx := sdk.Context{}
	sdkCtx = sdkCtx.WithBlockHeight(100)

	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	bytesType, err := abi.NewType("bytes", "bytes", nil)

	eventAbi := abi.Event{
		Name: "testEvent",
		ID: common.HexToHash(
			"0xa30d69e43099d8b4ea596246f71c10a8b994728d6d2232389b5b640bc04caa4d",
		),
		Inputs: []abi.Argument{
			{Name: "inArg1", Type: uint256Type, Indexed: true},
			{Name: "inArg2", Type: bytesType},
			{Name: "inArg3", Type: uint256Type},
		},
	}
	contactAbi := abi.ABI{
		Events: map[string]abi.Event{
			"testEvent": eventAbi,
		},
	}

	address := common.HexToAddress("0x1")

	stateDB := statedb.New(sdkCtx, nil, statedb.TxConfig{})

	eventEmitter := NewEventEmitter(sdkCtx, contactAbi, address, stateDB)

	event := &mockEvent{
		eventName: "testEvent",
		arguments: []*EventArgument{
			{Indexed: true, Value: big.NewInt(100)},
			{Indexed: false, Value: []byte{0xFF, 0xAA}},
			{Indexed: false, Value: big.NewInt(200)},
		},
	}

	err = eventEmitter.Emit(event)
	if err != nil {
		t.Fatal(err)
	}

	actualLogs := stateDB.Logs()

	if len(actualLogs) != 1 {
		t.Errorf("expected 1 log, got %d", len(actualLogs))
	}

	actualLog := actualLogs[0]

	if address != actualLog.Address {
		t.Errorf(
			"unexpected log address\n expected: %v\n actual:   %v",
			address,
			actualLog.Address,
		)
	}

	expectedTopics := []common.Hash{
		eventAbi.ID,
		// The value of the first indexed argument.
		common.BytesToHash(big.NewInt(100).Bytes()),
	}
	if !reflect.DeepEqual(expectedTopics, actualLog.Topics) {
		t.Errorf(
			"unexpected log topics\n expected: %v\n actual:   %v",
			expectedTopics,
			actualLog.Topics,
		)
	}

	var expectedData []byte
	// First 32-byte word denotes the location of the data part of the second
	// argument (inArg2). This argument is of dynamic type (bytes) hence the
	// data part is moved to the end of the log data. In this particular case,
	// the data part starts at the 64th byte (32 bytes for the location part
	// of the second argument (inArg2) + 32 bytes for the third argument (inArg3)
	// of static type (uint256) encoded in place). For reference, see:
	// https://docs.soliditylang.org/en/latest/abi-spec.html#use-of-dynamic-types
	expectedData = append(expectedData, common.LeftPadBytes(big.NewInt(64).Bytes(), 32)...)
	// Second 32-byte word is the third argument (inArg3) of static type
	// (uint256) encoded in place.
	expectedData = append(expectedData, common.LeftPadBytes(big.NewInt(200).Bytes(), 32)...)
	// Third 32-byte word is the first data part of the second argument (inArg2)
	// of dynamic type (bytes). The value of 2 is the number of bytes in the argument.
	expectedData = append(expectedData, common.LeftPadBytes(big.NewInt(2).Bytes(), 32)...)
	// Fourth 32-byte word is the second data part of the second argument (inArg2)
	// of dynamic type (bytes). The value of 0xFFAA is the actual value of the argument.
	expectedData = append(expectedData, common.RightPadBytes([]byte{0xFF, 0xAA}, 32)...)

	if !reflect.DeepEqual(expectedData, actualLog.Data) {
		t.Errorf(
			"unexpected log data\n expected: %v\n actual:   %v",
			expectedData,
			actualLog.Data,
		)
	}

	expectedBlockNumber := uint64(100)
	if expectedBlockNumber != actualLog.BlockNumber {
		t.Errorf(
			"unexpected log block\n expected: %v\n actual:   %v",
			expectedBlockNumber,
			actualLog.BlockNumber,
		)
	}
}

type mockEvent struct {
	eventName string
	arguments []*EventArgument
}

func (me *mockEvent) EventName() string {
	return me.eventName
}

func (me *mockEvent) Arguments() []*EventArgument {
	return me.arguments
}
