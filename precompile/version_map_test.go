package precompile

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var mockAbi = abi.ABI{}
var mockBytecode = "001122"

func TestSingleVersionMap(t *testing.T) {
	contract := NewContract(mockAbi, common.HexToAddress("0x123"), mockBytecode)

	vm := NewSingleVersionMap(contract)

	assert.Equal(t, contract, vm.GetByHeight(100))
	assert.Equal(t, contract, vm.GetByVersion(1))
	assert.Equal(t, contract, vm.GetLatest())
	assert.Equal(t, common.HexToAddress("0x123"), vm.Address())
}

func TestMultiVersionMap(t *testing.T) {
	contract1 := NewContract(mockAbi, common.HexToAddress("0x123"), mockBytecode)
	contract2 := NewContract(mockAbi, common.HexToAddress("0x123"), mockBytecode)
	rules := func(height int64) int {
		if height < 100 {
			return 1
		}
		return 2
	}

	vm := NewMultiVersionMap([]*Contract{contract1, contract2}, rules)

	assert.Equal(t, contract1, vm.GetByHeight(99))
	assert.Equal(t, contract2, vm.GetByHeight(100))
	assert.Equal(t, contract1, vm.GetByVersion(1))
	assert.Equal(t, contract2, vm.GetByVersion(2))
	assert.Equal(t, contract2, vm.GetLatest())
	assert.Equal(t, common.HexToAddress("0x123"), vm.Address())
}

