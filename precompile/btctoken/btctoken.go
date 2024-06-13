package btctoken

import (
	"embed"
	"fmt"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
)

//go:embed abi.json
var filesystem embed.FS

// EvmAddress is the EVM address of the BTC token precompile.
// EVM native precompiles reserve the addresses from 0x...01 to 0x...09.
// We use the 0x...1XXX range for custom Mezo precompiles to avoid collisions.
const EvmAddress = "0x0000000000000000000000000000000000001000"

func NewPrecompile(bankKeeper bankkeeper.Keeper) (*precompile.Contract, error){
	contractAbi, err := precompile.LoadAbiFile(filesystem, "abi.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load abi file: [%w]", err)
	}

	contract := precompile.NewContract(
		contractAbi,
		common.HexToAddress(EvmAddress),
	)

	methods := newPrecompileMethods(bankKeeper)
	contract.RegisterMethods(methods...)

	return contract, nil
}

func newPrecompileMethods(bankKeeper bankkeeper.Keeper) []precompile.Method {
	return []precompile.Method{
		newMintMethod(bankKeeper),
	}
}
