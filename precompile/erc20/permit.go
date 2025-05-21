package erc20

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mezo-org/mezod/precompile"
	evmkeeper "github.com/mezo-org/mezod/x/evm/keeper"
)

const (
	PermitMethodName          = "permit"
	NonceMethodName           = "nonce"
	DomainSeparatorMethodName = "DOMAIN_SEPARATOR"
	PermitTypehashMethodName  = "PERMIT_TYPEHASH"
	PermitTypehash            = "Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)"
)

// PermitMethod is a precompile method that allows users to approve a spender to
// spend their tokens with a signature.
type PermitMethod struct {
	bankKeeper      bankkeeper.Keeper
	authzkeeper     authzkeeper.Keeper
	evmkeeper       evmkeeper.Keeper
	denom           string
	domainSeparator []byte
	nonceKey        []byte
}

func NewPermitMethod(
	bankKeeper bankkeeper.Keeper,
	authzkeeper authzkeeper.Keeper,
	evmkeeper evmkeeper.Keeper,
	denom string,
	domainSeparator []byte,
	nonceKey []byte,
) *PermitMethod {
	return &PermitMethod{
		bankKeeper:      bankKeeper,
		authzkeeper:     authzkeeper,
		evmkeeper:       evmkeeper,
		denom:           denom,
		domainSeparator: domainSeparator,
		nonceKey:        nonceKey,
	}
}

func (am *PermitMethod) MethodName() string {
	return PermitMethodName
}

func (am *PermitMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (am *PermitMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (am *PermitMethod) Payable() bool {
	return false
}

// EIP2612 approval made with secp256k1 signature.
// Users can authorize a transfer of their tokens with a signature
// conforming EIP712 standard, rather than an on-chain transaction
// from their address. Anyone can submit this signature on the
// user's behalf by calling the permit function, paying gas fees,
// and possibly performing other actions in the same transaction.
func (am *PermitMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	timestamp := context.SdkCtx().BlockTime().Unix() // Unix time in seconds

	if err := precompile.ValidateMethodInputsCount(inputs, 7); err != nil {
		return nil, err
	}

	owner, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid owner address: %v", inputs[0])
	}
	if isZeroAddress(owner) {
		return nil, fmt.Errorf("owner address cannot be empty")
	}

	spender, ok := inputs[1].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid spender address: %v", inputs[1])
	}
	if isZeroAddress(spender) {
		return nil, fmt.Errorf("spender address cannot be empty")
	}

	amount, ok := inputs[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid amount: %v", inputs[2])
	}

	if amount == nil {
		return nil, errors.New("amount is required")
	}

	if amount.Sign() < 0 {
		return nil, errors.New("amount cannot be negative")
	}

	deadline, ok := inputs[3].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid deadline: %v", inputs[3])
	}
	// Check if deadline has passed
	if deadline.Int64() < timestamp {
		return nil, fmt.Errorf("permit expired")
	}

	v, ok := inputs[4].(byte)
	if !ok {
		return nil, fmt.Errorf("invalid v value: %v", inputs[4])
	}
	if v != 27 && v != 28 {
		return nil, fmt.Errorf("invalid v value: %v", v)
	}
	// Only signatures `v` value of 27 or 28 are considered valid, however
	// ValidateSignatureValues assumes that the `v` value is already adjusted and
	// expects it to be 0 or 1.
	v -= 27

	rComponent, ok := inputs[5].([32]byte)
	if !ok {
		return nil, fmt.Errorf("invalid r component of the signature: %v", inputs[5])
	}
	r := new(big.Int).SetBytes(rComponent[:])

	sComponent, ok := inputs[6].([32]byte)
	if !ok {
		return nil, fmt.Errorf("invalid s component of the signature: %v", inputs[6])
	}
	s := new(big.Int).SetBytes(sComponent[:])

	// A boolean set to true checks the signature with `s` value against the lower
	// half of the secp256k1 curve's order and is considered valid.
	if !crypto.ValidateSignatureValues(v, r, s, true) {
		return nil, fmt.Errorf("invalid signature values")
	}

	nonce, _, err := getNonce(am.nonceKey, am.evmkeeper, owner, context.SdkCtx())
	if err != nil {
		return nil, err
	}

	digest, err := buildDigest(owner, spender, amount, new(big.Int).SetBytes(nonce.Bytes()), deadline, am.domainSeparator)
	if err != nil {
		return nil, fmt.Errorf("failed to build digest: %v", err)
	}

	// Concatenate r, s, and v to form the full signature
	signature := append(r.Bytes(), append(s.Bytes(), v)...)

	// Recover the public key from the signature
	recoveredPubKey, err := crypto.SigToPub(digest, signature)
	if err != nil {
		return nil, fmt.Errorf("failed to recover public key from signature: %v", err)
	}

	// The recovered pub key is verified to ensure that the owner has signed the message.
	if !bytes.Equal(crypto.PubkeyToAddress(*recoveredPubKey).Bytes(), owner.Bytes()) {
		return nil, fmt.Errorf("verification failed over the signed message")
	}

	authorization, expiration := am.authzkeeper.GetAuthorization(context.SdkCtx(), spender.Bytes(), owner.Bytes(), SendMsgURL)

	err = handleAuthorization(am.denom, authorization, spender, amount, context, owner, expiration, am.authzkeeper)
	if err != nil {
		return nil, err
	}

	// After the approval is successful, the nonce should be incremented by 1 so
	// that the signature can be used only once over the given message.
	err = am.incrementNonce(owner, context.SdkCtx())
	if err != nil {
		return nil, fmt.Errorf("failed to set nonce: %w", err)
	}

	err = context.EventEmitter().Emit(
		NewApprovalEvent(
			owner,
			spender,
			amount,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit approval event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil
}

func (am *PermitMethod) incrementNonce(address common.Address, ctx sdk.Context) error {
	nonce, key, err := getNonce(am.nonceKey, am.evmkeeper, address, ctx)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %w", err)
	}
	nonceBigInt := new(big.Int).SetBytes(nonce.Bytes())
	// Increment nonce by 1 so that the signature can be used once over the message
	nonceBigInt.Add(nonceBigInt, big.NewInt(1))
	// Set the new nonce value
	am.evmkeeper.SetStateExtension(ctx, address, common.HexToHash(hex.EncodeToString(key)), nonceBigInt.Bytes())

	return nil
}

func buildDigest(owner, spender common.Address, amount, nonce, deadline *big.Int, domainSeparator []byte) ([]byte, error) {
	bytes32Type, err := abi.NewType("bytes32", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new type: %v", err)
	}
	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new type: %v", err)
	}
	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new type: %v", err)
	}

	// Convert to bytes32 (32-byte array)
	var PermitTypehashBytes32 [32]byte
	copy(PermitTypehashBytes32[:], crypto.Keccak256([]byte(PermitTypehash))[:32])

	// Encode the permit parameters
	encodedPermitParams, err := abi.Arguments{
		{Type: bytes32Type},
		{Type: addressType},
		{Type: addressType},
		{Type: uint256Type},
		{Type: uint256Type},
		{Type: uint256Type},
	}.Pack(
		PermitTypehashBytes32,
		owner,
		spender,
		amount,
		nonce,
		deadline,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to encode permit parameters: %v", err)
	}

	// A digest should consist of:
	// - The EIP712 Domain Separator (name, version, chainId, this precompile address, salt).
	//   See: https://eips.ethereum.org/EIPS/eip-712#definition-of-domainseparator
	encodedData := append([]byte("\x19\x01"), domainSeparator...)
	// - The Permit struct that includes: permit_typehash, owner, spender, value, nonce, deadline.
	//   See: https://eips.ethereum.org/EIPS/eip-2612#specification
	encodedData = append(encodedData, crypto.Keccak256(encodedPermitParams)...)
	// - The hash of the encoded data
	return crypto.Keccak256(encodedData), nil
}

// This functions implements the EIP712 domain separator for the permit function
// and produces the same result as the Solidity code seen e.g. in the OpenZeppelin
// lib https://github.com/OpenZeppelin/openzeppelin-contracts/blob/master/contracts/utils/cryptography/EIP712.sol#L89
// that is used by tBTC token https://github.com/keep-network/tbtc-v2/blob/main/solidity/contracts/token/TBTC.sol#L8
func BuildDomainSeparator(
	chainID *big.Int,
	name string,
	version string,
	verifyingContract common.Address,
) ([]byte, error) {
	// Hash the domain type
	domainType := "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"

	var DomainTypeHashBytes32 [32]byte
	var NameHashBytes32 [32]byte
	var VersionHashBytes32 [32]byte

	// Convert to bytes32 (32-byte array)
	copy(DomainTypeHashBytes32[:], crypto.Keccak256([]byte(domainType))[:32])
	copy(NameHashBytes32[:], crypto.Keccak256([]byte(name))[:32])
	copy(VersionHashBytes32[:], crypto.Keccak256([]byte(version))[:32])

	bytes32Type, err := abi.NewType("bytes32", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new type: %v", err)
	}
	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new type: %v", err)
	}
	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new type: %v", err)
	}

	// Encode the permit parameters
	encodedDomainSeparator, err := abi.Arguments{
		{Type: bytes32Type},
		{Type: bytes32Type},
		{Type: bytes32Type},
		{Type: uint256Type},
		{Type: addressType},
	}.Pack(
		DomainTypeHashBytes32,
		NameHashBytes32,
		VersionHashBytes32,
		chainID,
		verifyingContract,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to encode domain separator: %v", err)
	}

	return crypto.Keccak256(encodedDomainSeparator), nil
}

func getNonce(key []byte, evmkeeper evmkeeper.Keeper, address common.Address, ctx sdk.Context) (common.Hash, []byte, error) {
	if len(key) > 32 {
		return common.Hash{}, nil, fmt.Errorf("key %s is longer than 32 bytes", key)
	}
	nonce := evmkeeper.GetStateExtension(ctx, address, common.HexToHash(hex.EncodeToString(key)))
	if len(nonce) == 0 {
		return common.Hash{}, nil, fmt.Errorf("failed to get nonce for address %s", address.Hex())
	}
	return nonce, key, nil
}

type NonceMethod struct {
	evmkeeper evmkeeper.Keeper
	nonceKey  []byte
}

func NewNonceMethod(
	evmkeeper evmkeeper.Keeper,
	nonceKey []byte,
) *NonceMethod {
	return &NonceMethod{
		evmkeeper: evmkeeper,
		nonceKey:  nonceKey,
	}
}

func (nm *NonceMethod) MethodName() string {
	return NonceMethodName
}

func (nm *NonceMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (nm *NonceMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (nm *NonceMethod) Payable() bool {
	return false
}

// Returns the nonce of the given account.
func (nm *NonceMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	account, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("account argument must be common.Address")
	}

	nonce, _, err := getNonce(nm.nonceKey, nm.evmkeeper, account, context.SdkCtx())
	if err != nil {
		return nil, err
	}

	return precompile.MethodOutputs{
		new(big.Int).SetBytes(nonce.Bytes()),
	}, nil
}

type DomainSeparatorMethod struct {
	domainSeparator []byte
}

func NewDomainSeparatorMethod(
	domainSeparator []byte,
) *DomainSeparatorMethod {
	return &DomainSeparatorMethod{
		domainSeparator: domainSeparator,
	}
}

func (dsm *DomainSeparatorMethod) MethodName() string {
	return DomainSeparatorMethodName
}

func (dsm *DomainSeparatorMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (dsm *DomainSeparatorMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (dsm *DomainSeparatorMethod) Payable() bool {
	return false
}

// Returns the domain separator.
func (dsm *DomainSeparatorMethod) Run(
	_ *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	var domainSeparator [32]byte
	copy(domainSeparator[:], dsm.domainSeparator)
	return precompile.MethodOutputs{
		domainSeparator,
	}, nil
}

type PermitTypehashMethod struct{}

func NewPermitTypehashMethod() *PermitTypehashMethod {
	return &PermitTypehashMethod{}
}

func (ptm *PermitTypehashMethod) MethodName() string {
	return PermitTypehashMethodName
}

func (ptm *PermitTypehashMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (ptm *PermitTypehashMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (ptm *PermitTypehashMethod) Payable() bool {
	return false
}

// Returns the permit typehash.
func (ptm *PermitTypehashMethod) Run(
	_ *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	var permitTypehash [32]byte
	copy(permitTypehash[:], crypto.Keccak256([]byte(PermitTypehash))[:32])
	return precompile.MethodOutputs{
		permitTypehash,
	}, nil
}
