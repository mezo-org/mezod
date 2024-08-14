package btctoken

import (
	"bytes"
	"fmt"
	"math/big"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mezo-org/mezod/precompile"
	evmkeeper "github.com/mezo-org/mezod/x/evm/keeper"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// PermitMethodName is the name of the permit method that should match the name
// in the contract ABI.
const (
	PermitMethodName = "permit"
	PermitTypehash   = "Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)"
)

// DomainSeparator was generated by combination of encoding and hashing different
// params shown in the pseudo code below.
//
// keccak256(encode(
//
//			keccak256(
//				"EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"
//			),
//			keccak256(BTC),
//			keccak256(1),
//			chainId, // e.g 31612
//			0x7b7C000000000000000000000000000000000000
//		)
//	)
var DomainSeparator []byte

type permitMethod struct {
	bankKeeper  bankkeeper.Keeper
	authzkeeper authzkeeper.Keeper
	evmkeeper   evmkeeper.Keeper
}

func newPermitMethod(
	bankKeeper bankkeeper.Keeper,
	authzkeeper authzkeeper.Keeper,
	evmkeeper evmkeeper.Keeper,
) *permitMethod {
	var err error
	DomainSeparator, err = buildDomainSeparator(chainID)
	if err != nil {
		return nil
	}

	return &permitMethod{
		bankKeeper:  bankKeeper,
		authzkeeper: authzkeeper,
		evmkeeper:   evmkeeper,
	}
}

func (am *permitMethod) MethodName() string {
	return PermitMethodName
}

func (am *permitMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (am *permitMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (am *permitMethod) Payable() bool {
	return false
}

// EIP2612 approval made with secp256k1 signature.
// Users can authorize a transfer of their tokens with a signature
// conforming EIP712 standard, rather than an on-chain transaction
// from their address. Anyone can submit this signature on the
// user's behalf by calling the permit function, paying gas fees,
// and possibly performing other actions in the same transaction.
func (am *permitMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	timestamp := time.Now().Unix() // Unix time in seconds

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
	if amount == nil || amount.Sign() < 0 {
		amount = big.NewInt(0)
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

	nonce, _, err := getNonce(am.evmkeeper, owner, context.SdkCtx())
	if err != nil {
		return nil, err
	}

	digest, err := buildDigest(owner, spender, amount, new(big.Int).SetBytes(nonce.Bytes()), deadline)
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

	err = handleAuthorization(authorization, spender, amount, context, owner, expiration, am.authzkeeper)
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

func (am *permitMethod) incrementNonce(address common.Address, ctx sdk.Context) error {
	nonce, key, err := getNonce(am.evmkeeper, address, ctx)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %w", err)
	}
	nonceBigInt := new(big.Int).SetBytes(nonce.Bytes())
	// Increment nonce by 1 so that the signature can be used once over the message
	nonceBigInt.Add(nonceBigInt, big.NewInt(1))
	// Set the new nonce value
	am.evmkeeper.SetState(ctx, address, common.HexToHash(string(key)), nonceBigInt.Bytes())

	return nil
}

func buildDigest(owner, spender common.Address, amount, nonce, deadline *big.Int) ([]byte, error) {
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
	// - The EIP712 Domain Separator (name, version, chainId, this precompile address, salt). See: https://eips.ethereum.org/EIPS/eip-712#definition-of-domainseparator
	encodedData := append([]byte("\x19\x01"), DomainSeparator...)
	// - The Permit struct that includes: permit_typehash, owner, spender, value, nonce, deadline. See: https://eips.ethereum.org/EIPS/eip-2612#specification
	encodedData = append(encodedData, crypto.Keccak256(encodedPermitParams)...)
	// - The hash of the encoded data
	return crypto.Keccak256(encodedData), nil
}

// This functions implements the EIP712 domain separator for the permit function
// and produces the same result as the Solidity code seen e.g. in the OpenZeppelin
// lib https://github.com/OpenZeppelin/openzeppelin-contracts/blob/master/contracts/utils/cryptography/EIP712.sol#L89
// that is used by tBTC token https://github.com/keep-network/tbtc-v2/blob/main/solidity/contracts/token/TBTC.sol#L8
func buildDomainSeparator(chainID *big.Int) ([]byte, error) {
	// Hash the domain type
	domainType := "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"
	name := "BTC"
	version := "1"
	verifyingContract := common.HexToAddress("0x7b7c000000000000000000000000000000000000")

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

func getNonce(evmkeeper evmkeeper.Keeper, address common.Address, ctx sdk.Context) (common.Hash, []byte, error) {
	key := evmtypes.PrecompileBTCNonceKey(address)
	nonce := evmkeeper.GetState(ctx, address, common.HexToHash(string(key)))
	if len(nonce) == 0 {
		return common.Hash{}, nil, fmt.Errorf("failed to get nonce for address %s", address.Hex())
	}
	return nonce, key, nil
}

// TODO: Add Nonce read only method.
// TODO: Add DOMAIN_SEPARATOR read only method.
// TODO: Add PERMIT_TYPEHASH read only method.
