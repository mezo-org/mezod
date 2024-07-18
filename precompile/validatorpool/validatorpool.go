package validatorpool

import (
	"embed"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
	poatypes "github.com/evmos/evmos/v12/x/poa/types"
)

//go:embed abi.json
var filesystem embed.FS

// EvmAddress is the EVM address of the validator pool precompile.
// EVM native precompiles reserve the addresses from 0x...01 to 0x...09.
// We use the opposite range (0x1... to 0x9...) for custom Mezo precompiles to
// avoid collisions.
const EvmAddress = "0x2000000000000000000000000000000000000000"

type ValidatorPool interface {
	// SubmitApplication submits a new application to the validator pool
	SubmitApplication(types.Context, types.AccAddress, poatypes.Validator) error
	// ApproveApplication (onlyOwner) approves a pending application and
	// promotes the applications candidate to validator
	ApproveApplication(types.Context, types.AccAddress, types.ValAddress) error
	// Leave removes the sender from the validator pool
	Leave(types.Context, types.AccAddress) error
	// Kick (onlyOwner) removes a validator from the pool
	Kick(types.Context, types.AccAddress, types.ValAddress) error
	// GetOwner returns the validator pool owner address
	GetOwner(types.Context) types.AccAddress
	// TransferOwnership (onlyOwner) starts ownership transfer flow with a pending
	// ownership transfer
	TransferOwnership(types.Context, types.AccAddress, types.AccAddress) error
	// AcceptOwnership accepts a pending ownership transfer
	AcceptOwnership(types.Context, types.AccAddress) error
}

// NewPrecompile creates a new validator pool precompile.
func NewPrecompile(vp ValidatorPool) (*precompile.Contract, error) {
	contractAbi, err := precompile.LoadAbiFile(filesystem, "abi.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load abi file: [%w]", err)
	}

	contract := precompile.NewContract(
		contractAbi,
		common.HexToAddress(EvmAddress),
	)

	methods := newPrecompileMethods(vp)
	contract.RegisterMethods(methods...)

	return contract, nil
}

// newPrecompileMethods builds the list of methods for the validator pool precompile.
// All methods returned by this function are registered in the validator pool precompile.
func newPrecompileMethods(vp ValidatorPool) []precompile.Method {
	return []precompile.Method{
		newSubmitApplicationMethod(vp),
		newApproveApplicationMethod(vp),
		newKickMethod(vp),
		newLeaveMethod(vp),
		newTransferOwnershipMethod(vp),
		newAcceptOwnershipMethod(vp),
	}
}
