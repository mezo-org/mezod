//go:build !testbed

package app

import (
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/mezo-org/mezod/precompile"
	bridgekeeper "github.com/mezo-org/mezod/x/bridge/keeper"
	evmkeeper "github.com/mezo-org/mezod/x/evm/keeper"
	poakeeper "github.com/mezo-org/mezod/x/poa/keeper"
	oracletypes "github.com/skip-mev/connect/v2/x/oracle/types"
)

// customEvmPrecompiles builds custom precompiles of the EVM module.
func customEvmPrecompiles(
	bankKeeper bankkeeper.Keeper,
	authzKeeper authzkeeper.Keeper,
	poaKeeper poakeeper.Keeper,
	evmKeeper evmkeeper.Keeper,
	upgradeKeeper upgradekeeper.Keeper,
	oracleQueryServer oracletypes.QueryServer,
	bridgeKeeper bridgekeeper.Keeper,
	chainID string,
) ([]*precompile.VersionMap, error) {
	return baseCustomEvmPrecompiles(bankKeeper, authzKeeper, poaKeeper, evmKeeper, upgradeKeeper, oracleQueryServer, bridgeKeeper, chainID)
}
