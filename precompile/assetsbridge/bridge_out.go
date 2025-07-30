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
	authzKeeper  AuthzKeeper
}

func newBridgeOutMethod(
	bridgeKeeper BridgeKeeper,
	authzKeeper AuthzKeeper,
) *BridgeOutMethod {
	return &BridgeOutMethod{
		bridgeKeeper: bridgeKeeper,
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
		err   error
		isBTC = bytes.Equal(
			common.HexToAddress(evmtypes.BTCTokenPrecompileAddress).Bytes(),
			inputs.Token.Bytes(),
		)
	)

	sdkAmount, err := precompile.TypesConverter.BigInt.ToSDK(inputs.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to convert amount: [%w]", err)
	}
	assetsUnlocked, err := m.bridgeKeeper.SaveAssetsUnlocked(
		context.SdkCtx(),
		inputs.Recipient,
		inputs.Token.Bytes(),
		context.MsgSender().Bytes(),
		sdkAmount,
		uint8(inputs.Chain),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send AssetsUnlocked to bridge: %w", err)
	}

	err = context.EventEmitter().Emit(
		NewAssetsUnlockedEvent(
			assetsUnlocked.UnlockSequence.BigInt(),
			assetsUnlocked.Recipient,
			common.HexToAddress(assetsUnlocked.Token),
			context.MsgSender(),
			assetsUnlocked.Amount.BigInt(),
			uint8(assetsUnlocked.Chain), //nolint:gosec // G115: Safe conversion, Chain is validated elsewhere
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit AssetsUnlocked event: [%w]", err)
	}

	switch inputs.Chain {
	case TargetChainEthereum:
		if isBTC {
			err = m.burnBitcoin(context, inputs)
		} else {
			err = m.burnERC20(context, inputs)
		}
	case TargetChainBitcoin:
		err = m.burnBitcoin(context, inputs)
	default:
		panic(fmt.Sprintf("unreachable, unsupported target chain: %v", inputs.Chain))
	}

	return precompile.MethodOutputs{err == nil}, err
}

func (m *BridgeOutMethod) burnERC20(
	context *precompile.RunContext,
	inputs *bridgeOutInputs,
) error {
	var (
		sdkCtx   = context.SdkCtx()
		fromAddr = context.MsgSender()
	)

	return m.bridgeKeeper.BurnERC20(
		sdkCtx, inputs.Token.Bytes(), fromAddr.Bytes(), inputs.Amount)
}

func (m *BridgeOutMethod) burnBitcoin(
	context *precompile.RunContext,
	inputs *bridgeOutInputs,
) error {
	bridgeAddr := sdk.AccAddress(common.HexToAddress(
		evmtypes.AssetsBridgePrecompileAddress,
	).Bytes())
	senderAddr := sdk.AccAddress(context.MsgSender().Bytes())

	authorization, expiration := m.authzKeeper.GetAuthorization(
		context.SdkCtx(), bridgeAddr.Bytes(), senderAddr.Bytes(), SendMsgURL,
	)
	if authorization == nil {
		return fmt.Errorf("%s authorization type does not exist or is expired for address %s", SendMsgURL, senderAddr)
	}

	sendAuth, ok := authorization.(*banktypes.SendAuthorization)
	if !ok {
		return fmt.Errorf(
			"expected authorization to be a %T", banktypes.SendAuthorization{},
		)
	}

	sdkAmount, err := precompile.TypesConverter.BigInt.ToSDK(inputs.Amount)
	if err != nil {
		return fmt.Errorf("failed to convert amount: [%w]", err)
	}
	coin := sdk.Coin{Denom: evmtypes.DefaultEVMDenom, Amount: sdkAmount}

	// now update the authorization to spend for the AssetsBridge
	msg := banktypes.NewMsgSend(senderAddr.Bytes(), bridgeAddr.Bytes(), sdk.NewCoins(coin))
	resp, err := sendAuth.Accept(context.SdkCtx(), msg)
	if err != nil {
		return fmt.Errorf("couldn't accept authorization: %w", err)
	}

	if resp.Delete {
		// Authorization fully consumed, delete it
		err = m.authzKeeper.DeleteGrant(context.SdkCtx(), bridgeAddr, senderAddr, SendMsgURL)
	} else if resp.Updated != nil {
		err = m.authzKeeper.SaveGrant(context.SdkCtx(), bridgeAddr, senderAddr, resp.Updated, expiration)
	}
	if err != nil {
		return fmt.Errorf("couldn't update authorization BTC: %w", err)
	}

	if !resp.Accept {
		return errors.New("bridge is not authorized to burn BTC")
	}

	if err := m.bridgeKeeper.BurnBTC(
		context.SdkCtx(),
		senderAddr.Bytes(),
		sdkAmount,
	); err != nil {
		return fmt.Errorf("couldn't burn BTC: %w", err)
	}

	// finally update the journal entries to propagate the changes
	// done to the gas token (BTC in our case)
	balanceDelta, overflow := uint256.FromBig(inputs.Amount)
	if overflow {
		return fmt.Errorf("conversion from big.Int to uint256.Int overflowed: %v", inputs.Amount)
	}

	// only one side of the transfer to update here as we
	// burnt funds
	journal := context.Journal()
	journal.SubBalance(common.BytesToAddress(senderAddr.Bytes()), balanceDelta, tracing.BalanceChangeTransfer)

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
	btcToken := common.HexToAddress(evmtypes.BTCTokenPrecompileAddress)
	if bytes.Equal(btcToken.Bytes(), token.Bytes()) {
		return nil
	}

	switch chain {
	case TargetChainEthereum:
		if _, ok := m.bridgeKeeper.GetERC20TokenMapping(sdkCtx, token.Bytes()); ok {
			return nil
		}

		return fmt.Errorf("unsupported token: %v for ethereum target chain", token)
	case TargetChainBitcoin:
		return fmt.Errorf("unsupported token: %v for bitcoin target chain", token)
	default:
		panic("unreachable: unknown chain type")
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
	amount *big.Int,
	chain uint8,
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
