package assetsbridge

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/txscript"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/holiman/uint256"
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
	evmKeeper    EvmKeeper
	authzKeeper  AuthzKeeper
}

func newBridgeOutMethod(
	bridgeKeeper BridgeKeeper,
	evmKeeper EvmKeeper,
	authzKeeper AuthzKeeper,
) *BridgeOutMethod {
	return &BridgeOutMethod{
		bridgeKeeper: bridgeKeeper,
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
				assetsUnlocked.UnlockSequence.BigInt(),
				assetsUnlocked.Recipient,
				common.HexToAddress(assetsUnlocked.Token),
				context.MsgSender(),
				uint8(assetsUnlocked.Chain), //nolint:gosec // G115: Safe conversion, Chain is validated elsewhere
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

	assetsUnlocked, err := m.bridgeKeeper.SaveAssetsUnlocked(
		sdkCtx,
		inputs.Token.Bytes(),
		sdkAmount,
		uint8(inputs.Chain),
		inputs.Recipient,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send AssetsUnlocked to bridge: %w", err)
	}

	return assetsUnlocked, nil
}

func (m *BridgeOutMethod) executeBitcoin(
	context *precompile.RunContext,
	inputs *bridgeOutInputs,
) (*bridgetypes.AssetsUnlockedEvent, error) {
	bridgeAddr := sdk.AccAddress(common.HexToAddress(
		evmtypes.AssetsBridgePrecompileAddress,
	).Bytes())
	senderAddr := sdk.AccAddress(context.MsgSender().Bytes())

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

	if err := m.validateAuthorizationLimits(sendAuth, coins); err != nil {
		return nil, fmt.Errorf("authorization validation failed: %w", err)
	}

	if err := m.bridgeKeeper.BurnBTC(
		context.SdkCtx(),
		senderAddr.Bytes(),
		sdkAmount,
	); err != nil {
		return nil, fmt.Errorf("couldn't burn BTC: %w", err)
	}

	// now update the authorization to spend for the AssetsBridge
	msg := banktypes.NewMsgSend(senderAddr.Bytes(), bridgeAddr.Bytes(), coins)
	resp, err := sendAuth.Accept(context.SdkCtx(), msg)
	if err != nil {
		return nil, fmt.Errorf("couldn't update authorization: %w", err)
	}

	if resp.Updated != nil {
		err = m.authzKeeper.SaveGrant(context.SdkCtx(), bridgeAddr, senderAddr, resp.Updated, expiration)
	} else {
		// Authorization fully consumed, delete it
		m.authzKeeper.DeleteGrant(context.SdkCtx(), bridgeAddr, senderAddr, SendMsgURL)
	}

	assetsUnlocked, err := m.bridgeKeeper.SaveAssetsUnlocked(
		context.SdkCtx(),
		inputs.Token.Bytes(),
		sdkAmount,
		uint8(inputs.Chain),
		inputs.Recipient,
	)
	if err != nil {
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
	journal.SubBalance(common.BytesToAddress(senderAddr.Bytes()), balanceDelta, tracing.BalanceChangeTransfer)

	return assetsUnlocked, nil
}

func (m *BridgeOutMethod) validateAuthorizationLimits(
	sendAuth *banktypes.SendAuthorization,
	requestedCoins sdk.Coins,
) error {
	if sendAuth.SpendLimit == nil || sendAuth.SpendLimit.Empty() {
		return fmt.Errorf("no allowance for %v", requestedCoins[0].Denom)
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

	if inputs.Amount == nil {
		return errors.New("amount is required")
	}

	if inputs.Amount.Sign() <= 0 {
		return errors.New("amount must be positive")
	}

	_, ok := inputs.Chain.Validate()
	if !ok {
		return fmt.Errorf("unsupported chain: %v", inputs.Chain)
	}

	if err := m.validateRecipientForChain(inputs.Chain, inputs.Recipient); err != nil {
		return err
	}

	// first check that the token is either BTC or a supported ERC20
	if err := m.validateToken(sdkCtx, inputs.Token, inputs.Chain); err != nil {
		return err
	}

	return nil
}

func (m *BridgeOutMethod) validateToken(
	sdkCtx sdk.Context,
	token common.Address,
	chain TargetChain,
) error {
	switch chain {
	case TargetChainEthereum:
		btcToken := common.HexToAddress(evmtypes.BTCTokenPrecompileAddress)
		if bytes.Equal(btcToken.Bytes(), token.Bytes()) {
			return nil
		}

		if _, ok := m.bridgeKeeper.GetERC20TokenMapping(sdkCtx, token.Bytes()); ok {
			return nil
		}

		return fmt.Errorf("unsupported token: %v for ethereum target chain", token)

	case TargetChainBitcoin:
		btcToken := common.HexToAddress(evmtypes.BTCTokenPrecompileAddress)
		if bytes.Equal(btcToken.Bytes(), token.Bytes()) {
			return nil
		}

		return fmt.Errorf("unsupported token: %v for bitcoin target chain", token)

	default:
		panic("unreachable")
	}
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

	chainRaw, ok := inputs[2].(uint8)
	if !ok {
		return nil, fmt.Errorf("invalid chain: %v", inputs[2])
	}

	chain := TargetChain(chainRaw)

	recipient, ok := inputs[3].([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid recipient address: %v", inputs[3])
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
		// check the length and for the zero address
		if len(recipient) != 20 || bytes.Equal(recipient, (common.Address{}).Bytes()) {
			return fmt.Errorf("invalid recipient address for Ethereum chain: %v", hex.EncodeToString(recipient))
		}
	case TargetChainBitcoin:
		if !isSupportedBitcoinScriptType(recipient) {
			return fmt.Errorf("invalid recipient address for Bitcoin: %v", hex.EncodeToString(recipient))
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
// - unlockSequenceNumber (indexed): the sequenceNumber of this AssetsUnlocked
// - recipient (indexed): the address to which the tokens are transferred
// - token (indexed): the token being bridged out.
// - sender (non-indexed): the address from which the tokens are bridged out,
// - amount (non-indexed): the amount of tokens transferred
// - chain (non-indexed): the destination chain
type AssetsUnlockedEvent struct {
	unlockSequenceNumber *big.Int
	recipient            []byte
	token                common.Address
	sender               common.Address
	amount               *big.Int
	chain                uint8
}

func NewAssetsUnlockedEvent(
	sequenceNumber *big.Int,
	recipient []byte,
	token common.Address,
	sender common.Address,
	chain uint8,
	amount *big.Int,
) *AssetsUnlockedEvent {
	return &AssetsUnlockedEvent{
		unlockSequenceNumber: sequenceNumber,
		recipient:            recipient,
		token:                token,
		sender:               sender,
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

func isSupportedBitcoinScriptType(script []byte) bool {
	switch txscript.GetScriptClass(script) {
	case txscript.PubKeyHashTy, txscript.WitnessV0PubKeyHashTy,
		txscript.ScriptHashTy, txscript.WitnessV0ScriptHashTy:
		return true
	default:
		return false
	}
}
