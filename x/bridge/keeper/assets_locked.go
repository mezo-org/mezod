package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"fmt"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/tracers"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"math/big"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/types"
)

// GetAssetsLockedSequenceTip returns the current sequence tip for the
// AssetsLocked events. The tip denotes the sequence number of the last event
// processed by the x/bridge module.
func (k Keeper) GetAssetsLockedSequenceTip(ctx sdk.Context) math.Int {
	bz := ctx.KVStore(k.storeKey).Get(types.AssetsLockedSequenceTipKey)

	var sequenceTip math.Int
	err := sequenceTip.Unmarshal(bz)
	if err != nil {
		panic(err)
	}

	if sequenceTip.IsNil() {
		sequenceTip = math.ZeroInt()
	}

	return sequenceTip
}

// SetAssetsLockedSequenceTip sets the current sequence tip for the AssetsLocked
// events. The tip denotes the sequence number of the last event processed by
// the x/bridge module.
func (k Keeper) setAssetsLockedSequenceTip(
	ctx sdk.Context,
	sequenceTip math.Int,
) {
	bz, err := sequenceTip.Marshal()
	if err != nil {
		panic(err)
	}

	ctx.KVStore(k.storeKey).Set(types.AssetsLockedSequenceTipKey, bz)
}

// AcceptAssetsLocked processes the given AssetsLocked events sequence by minting
// the corresponding amount of coins for each event and sending them to the
// recipient address.
//
// Requirements:
//  1. The AssetsLocked sequence must not be empty.
//  2. The AssetsLocked sequence must be valid (i.e. all events in the slice
//     pass the AssetsLockedEvent.IsValid test AND sequence numbers of events
//     form a sequence strictly increasing by 1).
//  3. The sequence number of the first event in the slice must be exactly one
//     greater than the current sequence tip held in the state.
//
// The function returns an error if any of the requirements is not met.
// Checking the mentioned requirements is crucial to ensure state consistency
// regardless of the guarantees provided by the upstream code.
//
// If all requirements are met and x/bank interactions are all successful, the
// current sequence tip in the state is updated to the sequence number of the
// last event in the slice.
func (k Keeper) AcceptAssetsLocked(
	ctx sdk.Context,
	events types.AssetsLockedEvents,
) error {
	if len(events) == 0 {
		return fmt.Errorf("empty AssetsLocked sequence")
	}

	if !events.IsValid() {
		return fmt.Errorf("invalid AssetsLocked sequence")
	}

	currentSequenceTip := k.GetAssetsLockedSequenceTip(ctx)
	expectedSequenceStart := currentSequenceTip.AddRaw(1)
	if sequenceStart := events[0].Sequence; !expectedSequenceStart.Equal(sequenceStart) {
		return fmt.Errorf(
			"unexpected AssetsLocked sequence start; expected %s, got %s",
			expectedSequenceStart,
			sequenceStart,
		)
	}

	for _, event := range events {
		recipient, err := sdk.AccAddressFromBech32(event.Recipient)
		if err != nil {
			return fmt.Errorf("failed to parse recipient address: %w", err)
		}

		err = k.MintERC20(ctx, common.Address(recipient), event.Amount.BigInt())
		if err != nil {
			return fmt.Errorf("failed to mint ERC20: %w", err)
		}
	}

	k.setAssetsLockedSequenceTip(ctx, events[len(events)-1].Sequence)

	// TODO: Revisit this in the context of bridging events observability.
	//  From state's perspective, it's enough to update the sequence tip
	//  based on processed events to avoid double-bridging. Storing all
	//  processed events in the state is redundant, increases state management
	//  complexity, and negatively impacts the blockchain size in the long run.
	//  A sane alternative is using an opt-in EVM tx indexer (kv_indexer.go)
	//  to capture processed AssetsLocked events (they are part of the injected
	//  pseudo-tx and are available in the indexer) and expose them through
	//  a custom JSON-RPC API namespace (e.g. mezo_assetsLocked).

	return nil
}

func (k Keeper) MintERC20(ctx sdk.Context, account common.Address, amount *big.Int) error {
	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		return fmt.Errorf("failed to create address type: %w", err)
	}

	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return fmt.Errorf("failed to create uint256 type: %w", err)
	}

	methodAbi := abi.Method{
		Name: "mint",
		ID:   []byte{0x40, 0xc1, 0x0f, 0x19}, // 0x40c10f19 is the function selector for mint(address,uint256)
		Type: abi.Function,
		Inputs: []abi.Argument{
			{Name: "account", Type: addressType},
			{Name: "amount", Type: uint256Type},
		},
		Outputs: []abi.Argument{},
	}
	contractAbi := abi.ABI{
		Methods: map[string]abi.Method{
			"mint": methodAbi,
		},
	}

	data, err := contractAbi.Pack("mint", account, amount)
	if err != nil {
		return fmt.Errorf("failed to pack mint data: %w", err)
	}

	// Bridge's module EVM address.
	moduleAddress := common.BytesToAddress(authtypes.NewModuleAddress(types.ModuleName).Bytes())

	// Resolved while doing precompile/hardhat/deploy/01_deploy_test_erc20.ts
	erc20Address := common.HexToAddress("0xd17653E34f1E561019149D70C13A33B784c20cd1")

	res, err := k.CallContract(ctx, moduleAddress, &erc20Address, data)
	if err != nil {
		return fmt.Errorf("failed to mint ERC20: %w", err)
	}

	ctx.Logger().Info(
		"minted ERC20",
		"account",
		account,
		"amount",
		amount,
		"hash",
		res.Hash,
		"ret",
		res.Ret,
		"vmError",
		res.VmError,
		"gasUsed",
		res.GasUsed,
	)

	return nil
}

func (k Keeper) CallContract(
	ctx sdk.Context,
	from common.Address,
	contract *common.Address,
	data []byte,
) (*evmtypes.MsgEthereumTxResponse, error) {
	nonce, err := k.accountKeeper.GetSequence(ctx, from.Bytes())
	if err != nil {
		return nil, err
	}

	gasCap := uint64(10_000_000) // Block limit.

	msg := core.Message{
		To:                contract,
		From:              from,
		Nonce:             nonce,
		Value:             big.NewInt(0),
		GasLimit:          gasCap,
		GasPrice:          big.NewInt(0),
		GasFeeCap:         big.NewInt(0),
		GasTipCap:         big.NewInt(0),
		Data:              data,
		AccessList:        ethtypes.AccessList{},
		BlobGasFeeCap:     big.NewInt(0),
		BlobHashes:        []common.Hash{},
		SkipAccountChecks: false,
	}

	res, err := k.evmKeeper.ApplyMessage(ctx, msg, &tracers.Tracer{}, true)
	if err != nil {
		return nil, err
	}

	if res.Failed() {
		return nil, errorsmod.Wrap(evmtypes.ErrVMExecution, res.VmError)
	}

	return res, nil
}
