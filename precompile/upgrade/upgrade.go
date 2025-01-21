package upgrade

import (
	"context"
	"embed"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
)

//go:embed abi.json
var filesystem embed.FS

// EvmAddress is the EVM address of the upgrade precompile. The address is
// prefixed with 0x7b7c which was used to derive Mezo chain ID. This prefix is
// used to avoid potential collisions with EVM native precompiles.
const EvmAddress = "0x7b7c000000000000000000000000000000000014"

// NewPrecompile creates a new upgrade precompile.
func NewPrecompile(upgradeKeeper UpgradeKeeper, poaKeeper PoaKeeper) (*precompile.Contract, error) {
	contractAbi, err := precompile.LoadAbiFile(filesystem, "abi.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load abi file: [%w]", err)
	}

	contract := precompile.NewContract(
		contractAbi,
		common.HexToAddress(EvmAddress),
		EvmByteCode,
	)

	methods := newPrecompileMethods(upgradeKeeper, poaKeeper)
	contract.RegisterMethods(methods...)

	return contract, nil
}

// newPrecompileMethods builds the list of methods for the upgrade precompile.
// All methods returned by this function are registered in the upgrade precompile.
func newPrecompileMethods(upgradeKeeper UpgradeKeeper, poaKeeper PoaKeeper) []precompile.Method {
	return []precompile.Method{
		newSubmitPlanMethod(upgradeKeeper, poaKeeper),
		newCancelPlanMethod(upgradeKeeper, poaKeeper),
		newPlanMethod(upgradeKeeper),
	}
}

type PoaKeeper interface {
	CheckOwner(ctx sdk.Context, sender sdk.AccAddress) error
}

//nolint:all
type UpgradeKeeper interface {
	ClearUpgradePlan(ctx context.Context) error
	GetUpgradePlan(ctx context.Context) (upgradetypes.Plan, error)
	ScheduleUpgrade(ctx context.Context, plan upgradetypes.Plan) error
}
