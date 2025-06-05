package keeper

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/mezo-org/mezod/x/evm/types"
)

// ExecuteContractCall executes an EVM contract call. Under the hood, it creates
// an EVM message based on the provided data and applies it to trigger a state
// transition for the given EVM contract. This call does not create a fully-fledged
// transaction so no gas is deducted from the sender's account.
func (k *Keeper) ExecuteContractCall(
	ctx sdk.Context,
	call types.ContractCall,
) (*types.MsgEthereumTxResponse, error) {
	nonce, err := k.accountKeeper.GetSequence(ctx, call.From().Bytes())
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to get account nonce")
	}

	consensusParamsResponse, err := k.consensusKeeper.Params(ctx, nil)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to get consensus params")
	}

	// Set the gas cap to the block's max gas limit.
	gasCap := uint64(consensusParamsResponse.GetParams().GetBlock().MaxGas) //nolint:gosec

	msg := core.Message{
		To:                call.To(),
		From:              call.From(),
		Nonce:             nonce,
		Value:             big.NewInt(0),
		GasLimit:          gasCap,
		GasPrice:          big.NewInt(0),
		GasFeeCap:         big.NewInt(0),
		GasTipCap:         big.NewInt(0),
		Data:              call.Data(),
		AccessList:        ethtypes.AccessList{},
		BlobGasFeeCap:     big.NewInt(0),
		BlobHashes:        []common.Hash{},
		SkipAccountChecks: false,
	}

	res, err := k.ApplyMessage(ctx, msg, &tracers.Tracer{}, true)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to apply EVM message")
	}

	if res.Failed() {
		return nil, errorsmod.Wrap(types.ErrVMExecution, res.VmError)
	}

	return res, nil
}
