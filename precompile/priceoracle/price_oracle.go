package priceoracle

import (
	"context"
	"embed"
	"fmt"

	evmtypes "github.com/mezo-org/mezod/x/evm/types"

	oracletypes "github.com/skip-mev/connect/v2/x/oracle/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
)

//go:embed abi.json
var filesystem embed.FS

// EvmAddress is the EVM address of the Price Oracle precompile. The address is
// prefixed with 0x7b7c which was used to derive Mezo chain ID. This prefix is
// used to avoid potential collisions with EVM native precompiles.
const EvmAddress = evmtypes.PriceOraclePrecompileAddress

// NewPrecompileVersionMap creates a new version map for the price oracle precompile.
func NewPrecompileVersionMap(
	oracleQueryServer OracleQueryServer,
) (*precompile.VersionMap, error) {
	contractV1, err := NewPrecompile(oracleQueryServer)
	if err != nil {
		return nil, err
	}

	return precompile.NewVersionMap(
		map[int]*precompile.Contract{
			evmtypes.PriceOraclePrecompileLatestVersion: contractV1,
		},
	), nil
}

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
