package assetsbridge

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/holiman/uint256"
	keepbtc "github.com/keep-network/keep-core/pkg/bitcoin"
	"github.com/mezo-org/mezod/precompile"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// BridgeOutMethodName is the name of the bridgeOut method.
// It matches the name of the method in the contract ABI.
//
//nolint:gosec
const BridgeOutMethodName = "bridgeOut"

// SendMsgURL defines the authorization type for MsgSend
var SendMsgURL = sdk.MsgTypeURL(&banktypes.MsgSend{})

// BridgeOutMethod is the implementation of the bridgeOut method.
type BridgeOutMethod struct {
	bridgeKeeper BridgeKeeper
	bankKeeper   BankKeeper
	evmKeeper    EvmKeeper
	authzKeeper  AuthzKeeper
}

func newBridgeOutMethod(
	bridgeKeeper BridgeKeeper,
	bankKeeper BankKeeper,
	evmKeeper EvmKeeper,
	authzKeeper AuthzKeeper,
) *BridgeOutMethod {
	return &BridgeOutMethod{
		bridgeKeeper: bridgeKeeper,
		bankKeeper:   bankKeeper,
		evmKeeper:    evmKeeper,
		authzKeeper:  authzKeeper,
	}
}

func (m *BridgeOutMethod) MethodName() string {
	return BridgeOutMethodName
}

func (m *BridgeOutMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *BridgeOutMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *BridgeOutMethod) Payable() bool {
	return false
}

func (m *BridgeOutMethod) Run(
	context *precompile.RunContext,
	rawInputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	// extract inputs
	inputs, err := m.extractInputs(rawInputs)
	if err != nil {
		return precompile.MethodOutputs{false}, err
	}

	// run validation
	if err := m.validate(context, inputs); err != nil {
		return precompile.MethodOutputs{false}, err
	}

	return m.execute(context, inputs)
}

// execute will execute the burn of the token then send
// AssetsUnlocked events to the bridgeKeeper.
func (m *BridgeOutMethod) execute(
	context *precompile.RunContext,
	inputs *bridgeOutInputs,
) (precompile.MethodOutputs, error) {
	var (
		err            error
		assetsUnlocked *bridgetypes.AssetsUnlockedEvent
		isBTC          = bytes.Equal(
			common.HexToAddress(evmtypes.BTCTokenPrecompileAddress).Bytes(),
			inputs.Token.Bytes(),
		)
	)

	switch inputs.Chain {
	case TargetChainEthereum:
		if isBTC {
			assetsUnlocked, err = m.executeBitcoin(context, inputs)
		} else {
			assetsUnlocked, err = m.executeEthereum(context, inputs)
		}
	case TargetChainBitcoin:
		assetsUnlocked, err = m.executeBitcoin(context, inputs)
	}

	if assetsUnlocked != nil {
		err := context.EventEmitter().Emit(
			NewAssetsUnlockedEvent(
				context.MsgSender(),
				common.HexToAddress(assetsUnlocked.Token),
				assetsUnlocked.Recipient,
				uint8(assetsUnlocked.Chain), //nolint:gosec // G115: Safe conversion, Chain is validated elsewhere
				assetsUnlocked.UnlockSequence.BigInt(),
				assetsUnlocked.Amount.BigInt(),
			),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to emit AssetsUnlocked event: [%w]", err)
		}

	}

	return precompile.MethodOutputs{err == nil}, err
}

func (m *BridgeOutMethod) executeEthereum(
	context *precompile.RunContext,
	inputs *bridgeOutInputs,
) (*bridgetypes.AssetsUnlockedEvent, error) {
	var (
		sdkCtx          = context.SdkCtx()
		bridgeAddrBytes = common.HexToAddress(
			evmtypes.AssetsBridgePrecompileAddress,
		).Bytes()
		spenderAddr = context.MsgSender()
	)

	sdkAmount, err := precompile.TypesConverter.BigInt.ToSDK(inputs.Amount)
	if err != nil {
		return nil, fmt.Errorf("unable to convert amount to sdk type: %v", err)
	}

	call, err := evmtypes.NewERC20BurnFromCall(
		bridgeAddrBytes,
		inputs.Token.Bytes(),
		spenderAddr.Bytes(),
		inputs.Amount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ERC20 burnFrom call: %w", err)
	}

	_, err = m.evmKeeper.ExecuteContractCall(sdkCtx, call)
	if err != nil {
		return nil, fmt.Errorf("failed to execute ERC20 burnFrom call: %w", err)
	}

	assetsUnlocked, err := m.bridgeKeeper.AssetsUnlocked(
		sdkCtx,
		inputs.Token.Bytes(),
		sdkAmount,
		uint8(inputs.Chain),
		inputs.Recipient,
	)
	if err != nil {
		// TODO(JEREMY): error happened here, funds might have been
		// lost, should we panic?
		return nil, fmt.Errorf("failed to send AssetsUnlocked to bridge: %w", err)
	}

	return assetsUnlocked, nil
}

func (m *BridgeOutMethod) executeBitcoin(
	context *precompile.RunContext,
	inputs *bridgeOutInputs,
) (*bridgetypes.AssetsUnlockedEvent, error) {
	// first check if the bridge is authorized to spend the funds.
	bridgeAddrBytes := common.HexToAddress(
		evmtypes.AssetsBridgePrecompileAddress,
	).Bytes()
	bridgeAddr := sdk.AccAddress(bridgeAddrBytes)
	senderAddr := context.MsgSender()

	authorization, expiration := m.authzKeeper.GetAuthorization(
		context.SdkCtx(), bridgeAddr.Bytes(), senderAddr.Bytes(), SendMsgURL,
	)
	if authorization == nil {
		return nil, fmt.Errorf("%s authorization type does not exist or is expired for address %s", SendMsgURL, senderAddr)
	}

	if expiration != nil && expiration.Before(context.SdkCtx().BlockTime()) {
		return nil, fmt.Errorf("authorization expired at %v", expiration)
	}

	sendAuth, ok := authorization.(*banktypes.SendAuthorization)
	if !ok {
		return nil, fmt.Errorf(
			"expected authorization to be a %T", banktypes.SendAuthorization{},
		)
	}

	sdkAmount, err := precompile.TypesConverter.BigInt.ToSDK(inputs.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to convert amount: [%w]", err)
	}
	coins := sdk.Coins{{Denom: evmtypes.DefaultEVMDenom, Amount: sdkAmount}}
	msg := banktypes.NewMsgSend(bridgeAddrBytes, senderAddr.Bytes(), coins)

	if err := m.validateAuthorizationLimits(sendAuth, coins); err != nil {
		return nil, fmt.Errorf("authorization validation failed: %w", err)
	}

	_, err = m.authzKeeper.DispatchActions(context.SdkCtx(), bridgeAddr, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}

	assetsUnlocked, err := m.bridgeKeeper.AssetsUnlocked(
		context.SdkCtx(),
		inputs.Token.Bytes(),
		sdkAmount,
		uint8(inputs.Chain),
		inputs.Recipient,
	)
	if err != nil {
		// TODO(JEREMY): error happened here, funds might have been
		// lost, should we panic?
		return nil, fmt.Errorf("failed to send AssetsUnlocked to bridge: %w", err)
	}

	// finally update the journal entries to propagate the changes
	// done to the gas token (BTC in our case)
	balanceDelta, overflow := uint256.FromBig(inputs.Amount)
	if overflow {
		return nil, fmt.Errorf("conversion from big.Int to uint256.Int overflowed: %v", inputs.Amount)
	}

	// only one side of the transfer to update here as we
	// burnt funds
	journal := context.Journal()
	journal.SubBalance(senderAddr, balanceDelta, tracing.BalanceChangeTransfer)

	return assetsUnlocked, nil
}

func (m *BridgeOutMethod) validateAuthorizationLimits(
	sendAuth *banktypes.SendAuthorization,
	requestedCoins sdk.Coins,
) error {
	if sendAuth.SpendLimit == nil || sendAuth.SpendLimit.Empty() {
		return fmt.Errorf("no allowance for for %v", requestedCoins[0].Denom)
	}

	for _, requestedCoin := range requestedCoins {
		allowedAmount := sendAuth.SpendLimit.AmountOf(requestedCoin.Denom)
		if allowedAmount.IsZero() {
			return fmt.Errorf("no allowance for %s", requestedCoin.Denom)
		}
		if requestedCoin.Amount.GT(allowedAmount) {
			return fmt.Errorf(
				"requested amount %s exceeds allowed amount %s for %s",
				requestedCoin.Amount,
				allowedAmount,
				requestedCoin.Denom,
			)
		}
	}

	// Check allowed list if it exists
	// It shouldn't be set seeing that we don't really use cosmos-sdk
	// but just in case?
	if len(sendAuth.AllowList) > 0 {
		found := false
		senderAddrStr := sdk.AccAddress(common.HexToAddress(
			evmtypes.AssetsBridgePrecompileAddress,
		).Bytes()).String()

		for _, allowedAddr := range sendAuth.AllowList {
			if allowedAddr == senderAddrStr {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("recipient address not in authorization allow list")
		}
	}

	return nil
}

// validate applies more extensive validation to the inputs, not only
// the semantic validation but also verify balances, and tokens etc.
func (m *BridgeOutMethod) validate(
	context *precompile.RunContext,
	inputs *bridgeOutInputs,
) error {
	sdkCtx := context.SdkCtx()
	// first check that the token is either BTC or a supported ERC20
	if !m.isValidToken(sdkCtx, inputs.Token, inputs.Chain) {
		return fmt.Errorf("unsupported token: %v", inputs.Token)
	}

	owner := context.MsgSender()
	if err := m.isAmountSpendableByBridge(sdkCtx, owner, inputs.Token, inputs.Amount); err != nil {
		return err
	}

	return m.isAmountAvailable(sdkCtx, owner, inputs.Token, inputs.Amount)
}

func (m *BridgeOutMethod) isAmountSpendableByBridge(
	sdkCtx sdk.Context,
	owner common.Address,
	token common.Address,
	amount *big.Int,
) error {
	// We use the ERC20 allowance call for both the BTC token
	// and ERC20s as both implement the ERC20 interface.

	bridgeAddrBytes := common.HexToAddress(
		evmtypes.AssetsBridgePrecompileAddress,
	).Bytes()

	call, err := evmtypes.NewERC20AllowanceCall(
		bridgeAddrBytes,
		token.Bytes(),
		owner.Bytes(),
		// spendable by the asset bridge
		common.HexToAddress(
			evmtypes.AssetsBridgePrecompileAddress,
		).Bytes(),
	)
	if err != nil {
		return fmt.Errorf("failed to create ERC20 allowance call: %w", err)
	}

	resp, err := m.evmKeeper.ExecuteContractCall(sdkCtx, call)
	if err != nil {
		return fmt.Errorf("failed to execute ERC20 allowance call: %w", err)
	}

	// this would be non empty if the call reverted
	if len(resp.VmError) > 0 {
		// TODO(Jeremy): Should we handle the revert properly
		// to return an error message? For now just returning
		// a generic error message
		return fmt.Errorf("ERC20 allowance call reverted")
	}

	allowance, err := extractBigIntFromEVMResult(resp.Ret)
	if err != nil {
		return fmt.Errorf("unable to unpack EVM result: %v", err)
	}

	// if the amount <= to the balance, then OK
	// else we do not have enough funds and we need
	// to return an error
	if amount.Cmp(allowance) > 0 {
		return fmt.Errorf("asset bridge allowance too low")
	}

	return nil
}

func (m *BridgeOutMethod) isAmountAvailable(
	sdkCtx sdk.Context,
	owner common.Address,
	token common.Address,
	amount *big.Int,
) error {
	// we use the ERC20BalanceOf call for either the BTC token
	// or the ERC20 as both implements the ERC20 interface

	bridgeAddrBytes := common.HexToAddress(
		evmtypes.AssetsBridgePrecompileAddress,
	).Bytes()

	call, err := evmtypes.NewERC20BalanceOfCall(
		bridgeAddrBytes,
		token.Bytes(),
		owner.Bytes(),
	)
	if err != nil {
		return fmt.Errorf("failed to create ERC20 balanceOf call: %w", err)
	}

	resp, err := m.evmKeeper.ExecuteContractCall(sdkCtx, call)
	if err != nil {
		return fmt.Errorf("failed to execute ERC20 balanceOf call: %w", err)
	}

	// this would be non empty if the call reverted
	if len(resp.VmError) > 0 {
		// TODO(Jeremy): Should we handle the revert properly
		// to return an error message? For now just returning
		// a generic error message
		return fmt.Errorf("ERC20 balanceOf call reverted")
	}

	balance, err := extractBigIntFromEVMResult(resp.Ret)
	if err != nil {
		return fmt.Errorf("unable to unpack EVM result: %v", err)
	}

	// if the amount <= to the balance, then OK
	// else we do not have enough funds and we need
	// to return an error
	if amount.Cmp(balance) > 0 {
		return fmt.Errorf("not enough funds to bridgeOut")
	}

	return nil
}

func (m *BridgeOutMethod) isValidToken(
	sdkCtx sdk.Context,
	token common.Address,
	chain TargetChain,
) bool {
	if chain == TargetChainEthereum { // both BTC and supported ERC20 are valid
		btcToken := common.HexToAddress(evmtypes.BTCTokenPrecompileAddress)
		if bytes.Equal(btcToken.Bytes(), token.Bytes()) {
			return true
		}

		if _, ok := m.bridgeKeeper.GetERC20TokenMapping(sdkCtx, token.Bytes()); ok {
			return true
		}
	} else if chain == TargetChainBitcoin { // only BTC is valid
		btcToken := common.HexToAddress(evmtypes.BTCTokenPrecompileAddress)
		if bytes.Equal(btcToken.Bytes(), token.Bytes()) {
			return true
		}
	}

	return false
}

// extractInputs extract the inputs from the precompile.MethodInputs and apply
// basic validation on the inputs
func (m *BridgeOutMethod) extractInputs(inputs precompile.MethodInputs) (*bridgeOutInputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 4); err != nil {
		return nil, err
	}

	token, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid token address: %v", inputs[0])
	}

	amount, ok := inputs[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid amount: %v", inputs[1])
	}

	if amount == nil {
		return nil, errors.New("amount is required")
	}

	if amount.Sign() <= 0 {
		return nil, errors.New("amount must be positive")
	}

	chainRaw, ok := inputs[2].(uint8)
	if !ok {
		return nil, fmt.Errorf("invalid chain: %v", inputs[2])
	}

	chain, ok := TargetChain(chainRaw).Validate()
	if !ok {
		return nil, fmt.Errorf("unsupported chain: %v", inputs[2])
	}

	recipient, ok := inputs[3].([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid recipient address: %v", inputs[3])
	}

	if err := m.validateRecipientForChain(chain, recipient); err != nil {
		return nil, err
	}

	return &bridgeOutInputs{
		Token:     token,
		Amount:    amount,
		Chain:     chain,
		Recipient: recipient,
	}, nil
}

func (m *BridgeOutMethod) validateRecipientForChain(chain TargetChain, recipient []byte) error {
	if len(recipient) == 0 {
		return fmt.Errorf("recipient can't be empty")
	}

	switch chain {
	case TargetChainEthereum:
		// here we just check the length, the zero address have been tested before already
		if len(recipient) != 20 {
			return fmt.Errorf("invalid recipient address format for Ethereum chain: %v", recipient)
		}
	case TargetChainBitcoin:
		if keepbtc.GetScriptType(recipient) == keepbtc.NonStandardScript {
			return fmt.Errorf("invalid recipient address format for Bitcoin: %v", recipient)
		}
	}

	return nil
}

type TargetChain uint8

const (
	TargetChainEthereum = iota
	TargetChainBitcoin
)

func (t TargetChain) Validate() (TargetChain, bool) {
	if t == TargetChainEthereum || t == TargetChainBitcoin {
		return t, true
	}

	return t, false
}

type bridgeOutInputs struct {
	Token     common.Address
	Amount    *big.Int
	Chain     TargetChain
	Recipient []byte
}

// AssetsUnlockedEventName is the name of the AssetsUnlocked event. It matches the name
// of the event in the contract ABI.
const AssetsUnlockedEventName = "AssetsUnlocked"

// AssetsUnlockedEvent is the implementation of the AssetsUnlocked event that contains
// the following arguments:
// - from (indexed): the address from which the tokens are bridged out,
// - recipient (indexed): the address to which the tokens are transferred
// - amount (non-indexed): the amount of tokens transferred
// - chain (non-indexed): the destination chain
// - sequenceNumber (non-indexed): the sequenceNumber of this AssetsUnlocked
// - token (non-indexed): the token being bridged out.
type AssetsUnlockedEvent struct {
	sender, token        common.Address
	recipient            []byte
	chain                uint8
	unlockSequenceNumber *big.Int
	amount               *big.Int
}

func NewAssetsUnlockedEvent(
	from, token common.Address,
	recipient []byte,
	chain uint8,
	sequenceNumber *big.Int,
	amount *big.Int,
) *AssetsUnlockedEvent {
	return &AssetsUnlockedEvent{
		sender:               from,
		recipient:            recipient,
		token:                token,
		unlockSequenceNumber: sequenceNumber,
		amount:               amount,
		chain:                chain,
	}
}

func (te *AssetsUnlockedEvent) EventName() string {
	return AssetsUnlockedEventName
}

func (te *AssetsUnlockedEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   te.unlockSequenceNumber,
		},
		{
			Indexed: true,
			Value:   te.recipient,
		},
		{
			Indexed: true,
			Value:   te.token,
		},
		{
			Indexed: false,
			Value:   te.sender,
		},
		{
			Indexed: false,
			Value:   te.amount,
		},
		{
			Indexed: false,
			Value:   te.chain,
		},
	}
}

// helpers for extracting data out of the EVM response

func extractBigIntFromEVMResult(retData []byte) (*big.Int, error) {
	if len(retData) != 32 {
		return nil, fmt.Errorf("invalid return data length")
	}

	balance := new(big.Int).SetBytes(retData)
	return balance, nil
}
