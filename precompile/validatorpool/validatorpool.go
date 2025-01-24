package validatorpool

import (
	"embed"
	"fmt"

	evmtypes "github.com/mezo-org/mezod/x/evm/types"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
	poatypes "github.com/mezo-org/mezod/x/poa/types"
)

//go:embed abi.json
var filesystem embed.FS

// EvmAddress is the EVM address of the validatorpool precompile. The address is
// prefixed with 0x7b7c which was used to derive Mezo chain ID. This prefix is
// used to avoid potential collisions with EVM native precompiles.
const EvmAddress = evmtypes.ValidatorPoolPrecompileAddress

// Description is the validator description structure that contains information
// about the validator.
//
// Note: This struct mimics `poatypes.Description` however is declared with an
// `= struct`, this is required for correct deserialization within a precompile
// methods `Run` function
type Description = struct {
	// Moniker is the validator's name.
	Moniker string `json:"moniker"`
	// Identity is the optional identity signature (ex. UPort or Keybase).
	Identity string `json:"identity"`
	// Website is the optional website link.
	Website string `json:"website"`
	// SecurityContact is the optional security contact information.
	SecurityContact string `json:"securityContact"`
	// Details is the optional details about the validator.
	Details string `json:"details"`
}

// PoaKeeper interface used by the precompile
type PoaKeeper interface {
	// GetApplication returns the application for a operator
	GetApplication(types.Context, types.ValAddress) (poatypes.Application, bool)
	// GetAllApplications returns all applications
	GetAllApplications(types.Context) []poatypes.Application
	// SubmitApplication submits a new application to the validator pool
	SubmitApplication(types.Context, types.AccAddress, poatypes.Validator) error
	// ApproveApplication (onlyOwner) approves a pending application and
	// promotes the applications candidate to validator
	ApproveApplication(types.Context, types.AccAddress, types.ValAddress) error
	// CleanupApplications (onlyOwner) removes all pending applications
	CleanupApplications(types.Context, types.AccAddress) error
	// GetValidator returns the validator for a operator address
	GetValidator(types.Context, types.ValAddress) (poatypes.Validator, bool)
	// GetAllValidators returns all validators (in all states)
	GetAllValidators(types.Context) []poatypes.Validator
	// Leave removes the sender from the validator pool
	Leave(types.Context, types.AccAddress) error
	// Kick (onlyOwner) removes a validator from the pool
	Kick(types.Context, types.AccAddress, types.ValAddress) error
	// GetOwner returns the validator pool owner address
	GetOwner(types.Context) types.AccAddress
	// GetCandidateOwner returns the candidate validator pool owner address
	GetCandidateOwner(types.Context) types.AccAddress
	// TransferOwnership (onlyOwner) starts ownership transfer flow with a pending
	// ownership transfer
	TransferOwnership(types.Context, types.AccAddress, types.AccAddress) error
	// AcceptOwnership accepts a pending ownership transfer
	AcceptOwnership(types.Context, types.AccAddress) error
	// AddPrivilege (onlyOwner) adds a privilege to a set of operators.
	AddPrivilege(
		ctx types.Context,
		sender types.AccAddress,
		operators []types.ValAddress,
		privilege string,
	) error
	// RemovePrivilege (onlyOwner) removes a privilege from a set of operators.
	RemovePrivilege(
		ctx types.Context,
		sender types.AccAddress,
		operators []types.ValAddress,
		privilege string,
	) error
	// GetValidatorsOperatorsByPrivilege returns a list of validators with the
	// specified privilege.
	GetValidatorsOperatorsByPrivilege(
		ctx types.Context,
		privilege string,
	) []types.ValAddress
}

// NewPrecompileVersionMap creates a new version map for the validator pool precompile.
func NewPrecompileVersionMap(pk PoaKeeper) (*precompile.VersionMap, error) {
	contractV1, err := NewPrecompile(pk, false)
	if err != nil {
		return nil, err
	}

	return precompile.NewVersionMap(
		map[int]*precompile.Contract{
			0: contractV1, // returning v1 as v0 is legacy to support this precompile before versioning was introduced
			evmtypes.ValidatorPoolPrecompileLatestVersion: contractV1,
		},
	), nil
}

// NewPrecompile creates a new validator pool precompile.
func NewPrecompile(
	pk PoaKeeper,
	submitApplicationLegacyGas bool,
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

	methods := newPrecompileMethods(pk, submitApplicationLegacyGas)
	contract.RegisterMethods(methods...)

	return contract, nil
}

// newPrecompileMethods builds the list of methods for the validator pool precompile.
// All methods returned by this function are registered in the validator pool precompile.
func newPrecompileMethods(pk PoaKeeper, submitApplicationLegacyGas bool) []precompile.Method {
	return []precompile.Method{
		newSubmitApplicationMethod(pk, submitApplicationLegacyGas),
		newApproveApplicationMethod(pk),
		newCleanupApplicationsMethod(pk),
		newKickMethod(pk),
		newLeaveMethod(pk),
		newOwnerMethod(pk),
		newCandidateOwnerMethod(pk),
		newTransferOwnershipMethod(pk),
		newAcceptOwnershipMethod(pk),
		newValidatorMethod(pk),
		newValidatorsMethod(pk),
		newApplicationMethod(pk),
		newApplicationsMethod(pk),
		newAddPrivilegeMethod(pk),
		newRemovePrivilegeMethod(pk),
		newValidatorsByPrivilegeMethod(pk),
		newPrivilegesMethod(),
	}
}
