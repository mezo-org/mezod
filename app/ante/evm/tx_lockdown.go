package evm

import (
	"slices"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// EthTxLockdownDecorator rejects transactions that do not pass the
// transaction lockdown rule while the lockdown is active.
type EthTxLockdownDecorator struct {
	evmKeeper EVMKeeper
	poaKeeper PoaKeeper
}

// NewEthTxLockdownDecorator creates a new EthTxLockdownDecorator.
func NewEthTxLockdownDecorator(ek EVMKeeper, pk PoaKeeper) EthTxLockdownDecorator {
	return EthTxLockdownDecorator{
		evmKeeper: ek,
		poaKeeper: pk,
	}
}

// AnteHandle rejects every transaction whose sender and target do not pass the
// transaction lockdown rule. The check runs only while the lockdown is active.
func (etld EthTxLockdownDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	params := etld.evmKeeper.GetParams(ctx)
	if !params.TxLockdownEnabled {
		return next(ctx, tx, simulate)
	}

	// The roles come from x/poa at check time, so a role rotation applies at
	// once. An unset emergency team matches nothing.
	roles := []common.Address{
		common.BytesToAddress(etld.poaKeeper.GetOwner(ctx)),
	}
	team := common.BytesToAddress(etld.poaKeeper.GetEmergencyTeam(ctx))
	if team != (common.Address{}) {
		roles = append(roles, team)
	}

	extraSenders := txLockdownAddresses(params.TxLockdownSenders)
	extraTargets := txLockdownAddresses(params.TxLockdownTargets)

	for i, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, errorsmod.Wrapf(errortypes.ErrUnknownRequest, "invalid message type %T, expected %T", msg, (*evmtypes.MsgEthereumTx)(nil))
		}

		txData, err := evmtypes.UnpackTxData(msgEthTx.Data)
		if err != nil {
			return ctx, errorsmod.Wrapf(err, "failed to unpack tx data any for tx %d", i)
		}

		sender := common.BytesToAddress(msgEthTx.GetFrom())
		target := txData.GetTo()

		if TxLockdownAllowed(sender, target, roles, extraSenders, extraTargets) {
			continue
		}

		targetForLog := "nil (contract creation)"
		if target != nil {
			targetForLog = target.Hex()
		}

		return ctx, errorsmod.Wrapf(
			errortypes.ErrUnauthorized,
			"transaction from %s to %s rejected by the transaction lockdown",
			sender,
			targetForLog,
		)
	}

	return next(ctx, tx, simulate)
}

// TxLockdownAllowed reports whether the transaction lockdown rule passes for
// the given sender and target. A nil target means contract creation.
func TxLockdownAllowed(
	sender common.Address,
	target *common.Address,
	roles []common.Address,
	extraSenders []common.Address,
	extraTargets []common.Address,
) bool {
	// The role clause. It holds whatever the extra allowlists say.
	if slices.Contains(roles, sender) {
		return true
	}
	if target != nil && slices.Contains(roles, *target) {
		return true
	}

	// Two empty dimensions leave the extra clause inactive. Only the role
	// clause passes then.
	if len(extraSenders) == 0 && len(extraTargets) == 0 {
		return false
	}

	// An empty dimension matches every address. A non-empty target dimension
	// never matches contract creation.
	senderAllowed := len(extraSenders) == 0 || slices.Contains(extraSenders, sender)
	targetAllowed := len(extraTargets) == 0 ||
		(target != nil && slices.Contains(extraTargets, *target))

	return senderAllowed && targetAllowed
}

// txLockdownAddresses converts hex-encoded addresses from the EVM parameters
// to EVM addresses.
func txLockdownAddresses(hexAddresses []string) []common.Address {
	addresses := make([]common.Address, len(hexAddresses))
	for i, hexAddress := range hexAddresses {
		addresses[i] = common.BytesToAddress(
			evmtypes.HexAddressToBytes(hexAddress),
		)
	}

	return addresses
}
