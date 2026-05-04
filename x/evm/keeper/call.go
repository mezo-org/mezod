package keeper

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/mezo-org/mezod/x/evm/statedb"
	"github.com/mezo-org/mezod/x/evm/types"
)

// ExecuteContractCall executes an EVM contract call. Under the hood, it creates
// an EVM message based on the provided data and applies it to trigger a state
// transition for the given EVM contract. This call does not create a fully-fledged
// transaction so no gas is deducted from the sender's account.
func (k *Keeper) ExecuteContractCall(
	ctx sdk.Context,
	call types.ContractCall,
) (*types.MsgEthereumTxResponse, []statedb.StateChange, error) {
	nonce, err := k.accountKeeper.GetSequence(ctx, call.From().Bytes())
	if err != nil {
		return nil, nil, errorsmod.Wrap(err, "failed to get account nonce")
	}

	consensusParamsResponse, err := k.consensusKeeper.Params(ctx, nil)
	if err != nil {
		return nil, nil, errorsmod.Wrap(err, "failed to get consensus params")
	}

	// Use the call's gas limit if set, otherwise fall back to the
	// block's max gas limit.
	gasLimit := call.GasLimit()
	if gasLimit == 0 {
		gasLimit = uint64(consensusParamsResponse.GetParams().GetBlock().MaxGas) //nolint:gosec
	}

	msg := core.Message{
		To:                    call.To(),
		From:                  call.From(),
		Nonce:                 nonce,
		Value:                 big.NewInt(0),
		GasLimit:              gasLimit,
		GasPrice:              big.NewInt(0),
		GasFeeCap:             big.NewInt(0),
		GasTipCap:             big.NewInt(0),
		Data:                  call.Data(),
		AccessList:            ethtypes.AccessList{},
		BlobGasFeeCap:         big.NewInt(0),
		BlobHashes:            []common.Hash{},
		SkipTransactionChecks: false,
	}

	res, changes, err := k.ApplyMessage(ctx, msg, &tracers.Tracer{}, true)
	if err != nil {
		return nil, nil, errorsmod.Wrap(err, "failed to apply EVM message")
	}

	if res.Failed() {
		return nil, nil, errorsmod.Wrap(types.ErrVMExecution, res.VmError)
	}

	if len(changes) > 0 {
		ctx.Logger().Debug(
			"inner EVM execution committed state changes",
			"contract", call.To().Hex(),
			"changesCount", len(changes),
		)
	}

	return res, changes, nil
}
