package priceoracle

import (
	"context"
	"embed"
	"fmt"

	oracletypes "github.com/skip-mev/connect/v2/x/oracle/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
)

//go:embed abi.json
var filesystem embed.FS

// EvmAddress is the EVM address of the Price Oracle precompile. The address is
// prefixed with 0x7b7c which was used to derive Mezo chain ID. This prefix is
// used to avoid potential collisions with EVM native precompiles.
const EvmAddress = "0x7b7c000000000000000000000000000000000015"

// NewPrecompile creates a new Price Oracle precompile.
func NewPrecompile(
	oracleQueryServer OracleQueryServer,
) (*precompile.Contract, error) {
	contractAbi, err := precompile.LoadAbiFile(filesystem, "abi.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load abi file: [%w]", err)
	}

	contract := precompile.NewContract(
		contractAbi,
		common.HexToAddress(EvmAddress),
		EvmByteCode,
	)

	methods := newPrecompileMethods(oracleQueryServer)
	contract.RegisterMethods(methods...)

	return contract, nil
}

// newPrecompileMethods builds the list of methods for the Price Oracle precompile.
// All methods returned by this function are registered in the Price Oracle precompile.
func newPrecompileMethods(
	oracleQueryServer OracleQueryServer,
) []precompile.Method {
	return []precompile.Method{
		newDecimalsMethod(),
		newLatestRoundDataMethod(oracleQueryServer),
	}
}

type OracleQueryServer interface {
	GetPrice(
		ctx context.Context,
		req *oracletypes.GetPriceRequest,
	) (*oracletypes.GetPriceResponse, error)
}
