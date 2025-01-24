package precompile

import (
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var (
	mockAbi      = abi.ABI{}
	mockBytecode = "001122"
)

func TestVersionMap(t *testing.T) {
	contract1 := NewContract(mockAbi, common.HexToAddress("0x123"), mockBytecode)
	contract2 := NewContract(mockAbi, common.HexToAddress("0x123"), mockBytecode)

	vm := NewVersionMap(map[int]*Contract{
		0: contract1,
		1: contract1,
		2: contract2,
	})

	actualContract, ok := vm.GetByVersion(0)
	assert.Equal(t, contract1, actualContract)
	assert.True(t, ok)

	actualContract, ok = vm.GetByVersion(1)
	assert.Equal(t, contract1, actualContract)
	assert.True(t, ok)

	actualContract, ok = vm.GetByVersion(2)
	assert.Equal(t, contract2, actualContract)
	assert.True(t, ok)

	actualContract, ok = vm.GetByVersion(3)
	assert.Nil(t, actualContract)
	assert.False(t, ok)

	actualLatestContract := vm.GetLatest()
	assert.Equal(t, contract2, actualLatestContract)

	actualLatestVersion := vm.GetLatestVersion()
	assert.Equal(t, 2, actualLatestVersion)

	assert.Equal(t, common.HexToAddress("0x123"), vm.Address())
}

func TestVersionMap_UnhappyPaths(t *testing.T) {
	contract := NewContract(mockAbi, common.HexToAddress("0x123"), mockBytecode)
	otherContract := NewContract(mockAbi, common.HexToAddress("0x124"), mockBytecode)

	assert.PanicsWithValue(t, "no contracts provided", func() {
		NewVersionMap(nil)
	})

	assert.PanicsWithValue(t, "no contracts provided", func() {
		NewVersionMap(map[int]*Contract{})
	})

	assert.PanicsWithValue(t, "versions must be strictly increasing by 1", func() {
		NewVersionMap(map[int]*Contract{
			0: contract,
			2: contract,
		})
	})

	assert.PanicsWithValue(t, "nil contract provided", func() {
		NewVersionMap(map[int]*Contract{
			1: contract,
			2: nil,
			3: contract,
		})
	})

	assert.PanicsWithValue(t, "all contracts must have the same address", func() {
		NewVersionMap(map[int]*Contract{
			0: contract,
			1: otherContract,
		})
	})
}
