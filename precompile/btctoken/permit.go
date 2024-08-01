package btctoken

import (
	"bytes"
	"fmt"
	"math/big"
	"time"

	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v12/precompile"
)

// PermitMethodName is the name of the permit method that should match the name
// in the contract ABI.
const (
	PermitMethodName = "permit"
)

// Sets a `value` amount of tokens as the allowance of `spender` over the
// caller's tokens.
type permitMethod struct {
	bankKeeper  bankkeeper.Keeper
	authzkeeper authzkeeper.Keeper
}

func newPermitMethod(
	bankKeeper bankkeeper.Keeper,
	authzkeeper authzkeeper.Keeper,
) *permitMethod {
	return &permitMethod{
		bankKeeper:  bankKeeper,
		authzkeeper: authzkeeper,
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
	timestamp := time.Now().Unix() // Now

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

	deadline, ok := inputs[3].(int64)
	if !ok {
		return nil, fmt.Errorf("invalid deadline: %v", inputs[3])
	}
	// Check if deadline has passed
	if deadline < timestamp {
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
	v = v - 27

	r, ok := inputs[5].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid r component of the signature: %v", inputs[5])
	}

	s, ok := inputs[6].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid s component of the signature: %v", inputs[6])
	}
	// A boolean set to true checks the signature with `s` value against the lower
	// half of the secp256k1 curve's order and is considered valid.
	if !crypto.ValidateSignatureValues(v, r, s, true) {
		return nil, fmt.Errorf("invalid signature values")
	}

	// A message should consist of:
	// - The EIP712 Domain Separator (name, version, chainId, this precompile address, salt). See: https://eips.ethereum.org/EIPS/eip-712#definition-of-domainseparator
	// - The Permit struct that includes: permit_typehash, owner, spender, value, nonce?, deadline. See: https://eips.ethereum.org/EIPS/eip-2612#specification
	// - The hash of the message
	message := []byte("...") // TODO: constuct the real message
	digest := crypto.Keccak256Hash(message)

	// Concatenate r, s, and v to form the full signature
	signature := append(r.Bytes(), s.Bytes()...)
	signature = append(signature, v)

	// Recover the public key from the signature
	recoveredPubKey, err := crypto.SigToPub(digest.Bytes(), signature)
	if err != nil {
		return nil, fmt.Errorf("failed to recover public key from signature: %v", err)
	}

	// The recovered pub key is verified to ensure that the owner has signed the message.
	if !bytes.Equal(crypto.PubkeyToAddress(*recoveredPubKey).Bytes(), owner.Bytes()) {
		return nil, fmt.Errorf("signature verification failed")
	}

	approveMethod := newApproveMethod(am.bankKeeper, am.authzkeeper)
	authorization, expiration := am.authzkeeper.GetAuthorization(context.SdkCtx(), spender.Bytes(), owner.Bytes(), SendMsgURL)

	if authorization == nil {
		if amount.Sign() == 0 {
			// no authorization, amount 0 -> error
			err = fmt.Errorf("no existing approvals, cannot approve 0")
		} else {
			// no authorization, amount positive -> create a new authorization
			err = approveMethod.createAuthorization(context.SdkCtx(), spender, owner, amount)
		}
	} else {
		if amount.Sign() == 0 {
			// authorization exists, amount 0 -> delete authorization
			err = am.authzkeeper.DeleteGrant(context.SdkCtx(), spender.Bytes(), owner.Bytes(), SendMsgURL)
		} else {
			// authorization exists, amount positive -> update authorization
			err = approveMethod.updateAuthorization(context.SdkCtx(), spender, owner, amount, authorization, expiration)
		}
	}

	if err != nil {
		return nil, err
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
