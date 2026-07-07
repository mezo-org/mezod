package types_test

import (
	"bytes"
	"errors"
	"math/big"
	"runtime/debug"
	"testing"

	sdkmath "cosmossdk.io/math"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
	"github.com/ethereum/go-ethereum/common"

	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// validateSentinels enumerates every typed error Validate() is allowed
// to surface. The fuzz target asserts that every Validate() failure on
// a successfully-unmarshaled payload wraps one of these via errors.Is.
// Update this set whenever Validate() in
// x/evm/types/set_code_tx.go grows a new branch — forgetting to update
// will cause the fuzz target to flag legitimate errors as leaks.
var validateSentinels = []error{
	evmtypes.ErrInvalidGasCap,
	evmtypes.ErrInvalidAmount,
	evmtypes.ErrInvalidGasFee,
	evmtypes.ErrSetCodeMissingTo,
	evmtypes.ErrSetCodeEmptyAuthList,
	evmtypes.ErrInvalidSigner,
	errortypes.ErrInvalidChainID,
	errortypes.ErrInvalidAddress,
}

// mustMarshalSetCodeTx is a panic-on-failure marshal helper for the
// fuzz seed corpus. Seeds are hand-crafted; a marshal failure is a
// test-author bug, not a runtime condition.
func mustMarshalSetCodeTx(tx *evmtypes.SetCodeTx) []byte {
	b, err := proto.Marshal(tx)
	if err != nil {
		panic(err)
	}
	return b
}

// FuzzSetCodeTxUnmarshal mutates the wire payload of an EIP-7702
// SetCodeTx proto and asserts the unmarshal/validate/round-trip
// surface never panics and that every Validate() error wraps a typed
// sentinel from validateSentinels.
//
// No test-suite setup is needed: SetCodeTx is a self-contained proto
// message with no app/keeper dependency, so each iteration runs in
// isolation.
func FuzzSetCodeTxUnmarshal(f *testing.F) {
	to := common.HexToAddress("0x1111111111111111111111111111111111111111").Hex()
	target := common.HexToAddress("0x2222222222222222222222222222222222222222").Hex()
	chainID := sdkmath.NewInt(31611)
	one := sdkmath.NewInt(1)
	amount := sdkmath.NewInt(0)

	validEmptyAuth := &evmtypes.SetCodeTx{
		ChainID:   &chainID,
		Nonce:     1,
		GasTipCap: &one,
		GasFeeCap: &one,
		GasLimit:  21000,
		To:        to,
		Amount:    &amount,
		Data:      []byte{},
		AuthList:  nil,
	}

	authChainID := sdkmath.NewInt(31611)
	oneAuth := evmtypes.SetCodeAuthorization{
		ChainID: &authChainID,
		Address: target,
		Nonce:   0,
		V:       []byte{1},
		R:       big.NewInt(7).Bytes(),
		S:       big.NewInt(11).Bytes(),
	}
	validOneAuth := &evmtypes.SetCodeTx{
		ChainID:   &chainID,
		Nonce:     1,
		GasTipCap: &one,
		GasFeeCap: &one,
		GasLimit:  21000,
		To:        to,
		Amount:    &amount,
		Data:      []byte{},
		AuthList:  evmtypes.AuthorizationList{oneAuth},
	}

	tenAuths := make(evmtypes.AuthorizationList, 10)
	for i := range tenAuths {
		tenAuths[i] = oneAuth
	}
	validTenAuths := &evmtypes.SetCodeTx{
		ChainID:   &chainID,
		Nonce:     1,
		GasTipCap: &one,
		GasFeeCap: &one,
		GasLimit:  21000,
		To:        to,
		Amount:    &amount,
		Data:      []byte{},
		AuthList:  tenAuths,
	}

	// Each of the seeds below crosses exactly one Validate() boundary,
	// so the fuzzer enters Validate's later branches without having to
	// rediscover the precondition wire shape.

	invalidTo := *validOneAuth
	invalidTo.To = "0xinvalid"

	badChainID := sdkmath.NewInt(99999)
	wrongChainID := *validOneAuth
	wrongChainID.ChainID = &badChainID

	negativeAmount := sdkmath.NewInt(-1)
	negAmount := *validOneAuth
	negAmount.Amount = &negativeAmount

	// GasFeeCap = -1 fails the negative-cap branch, surfacing
	// ErrInvalidGasCap rather than ErrInvalidGasFee. The Fee()
	// overflow branch that produces ErrInvalidGasFee requires a
	// ~int256-sized cap and is left to fuzzer discovery.
	negativeFee := sdkmath.NewInt(-1)
	negFee := *validOneAuth
	negFee.GasFeeCap = &negativeFee

	emptyAuthList := *validOneAuth
	emptyAuthList.AuthList = nil

	// Auth V = 33 bytes with high bit set exceeds int256 bounds.
	oversizedV := make([]byte, 33)
	oversizedV[0] = 0xff
	oversizedAuth := oneAuth
	oversizedAuth.V = oversizedV
	authOversizedV := *validOneAuth
	authOversizedV.AuthList = evmtypes.AuthorizationList{oversizedAuth}

	// Second Mezo-allowed chain id (31612).
	altChainID := sdkmath.NewInt(31612)
	validOneAuthAltChain := *validOneAuth
	validOneAuthAltChain.ChainID = &altChainID

	// EIP-7702 any-chain branch: auth.ChainID = 0 is legal at
	// Validate time; per-tuple eligibility is enforced in the keeper.
	zeroAuthChainID := sdkmath.NewInt(0)
	zeroChainAuth := oneAuth
	zeroChainAuth.ChainID = &zeroAuthChainID
	validOneAuthZeroAuthChain := *validOneAuth
	validOneAuthZeroAuthChain.AuthList = evmtypes.AuthorizationList{zeroChainAuth}

	for _, seed := range [][]byte{
		nil,
		{},
		mustMarshalSetCodeTx(validEmptyAuth),
		mustMarshalSetCodeTx(validOneAuth),
		mustMarshalSetCodeTx(validTenAuths),
		mustMarshalSetCodeTx(&invalidTo),
		mustMarshalSetCodeTx(&wrongChainID),
		mustMarshalSetCodeTx(&negAmount),
		mustMarshalSetCodeTx(&negFee),
		mustMarshalSetCodeTx(&emptyAuthList),
		mustMarshalSetCodeTx(&authOversizedV),
		mustMarshalSetCodeTx(&validOneAuthAltChain),
		mustMarshalSetCodeTx(&validOneAuthZeroAuthChain),
		{0xff, 0x00, 0xff, 0x00, 0x55},
		{0x08, 0x01, 0x10, 0x02, 0x18, 0x03},
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic for input len=%d: %v\n%s",
					len(data), r, debug.Stack())
			}
		}()

		var tx evmtypes.SetCodeTx
		if err := proto.Unmarshal(data, &tx); err != nil {
			// Unmarshal rejection is a legal terminal state; the fuzz
			// invariant only constrains successful-unmarshal paths.
			return
		}

		err := tx.Validate()
		if err != nil {
			matched := false
			for _, sentinel := range validateSentinels {
				if errors.Is(err, sentinel) {
					matched = true
					break
				}
			}
			if !matched {
				t.Fatalf("Validate() returned untyped error: %v "+
					"(input len=%d)", err, len(data))
			}

			// Cheap accessors whose contract permits an unvalidated tx
			// must also be panic-safe (ante/keeper may touch them for
			// error formatting before Validate completes). Fee/Cost
			// route through fee() which Muls a nil *big.Int when
			// GasFeeCap is unset, mirroring DynamicFeeTx — those, like
			// AsEthereumData, are validated-path-only by construction
			// and are deliberately omitted here.
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("panic in invalid-branch "+
							"accessors for input len=%d: %v\n%s",
							len(data), r, debug.Stack())
					}
				}()
				_ = tx.GetGasTipCap()
				_ = tx.GetGasFeeCap()
				_ = tx.GetChainID()
				_ = tx.GetValue()
				_ = tx.GetTo()
				_ = tx.GetAccessList()
				_ = tx.GetAuthorizationList()
				_ = tx.Copy()
			}()
			return
		}

		// On a validated tx, downstream accessors must not panic. These
		// are the surface called from the keeper / ante on every tx; a
		// panic here would surface as a node-level crash.
		_ = tx.AsEthereumData()
		_ = tx.Copy()
		_ = tx.Fee()
		_ = tx.Cost()
		_ = tx.GetGasTipCap()

		// Round-trip pin: re-marshaling a validated tx must be
		// idempotent at the byte level. Comparing against the fuzzer's
		// input is unsound — proto inputs are not canonical — so we
		// pin Marshal->Unmarshal->Marshal byte-equality instead.
		out, err := proto.Marshal(&tx)
		if err != nil {
			t.Fatalf("re-marshal of validated tx failed: %v", err)
		}
		var round evmtypes.SetCodeTx
		if err := proto.Unmarshal(out, &round); err != nil {
			t.Fatalf("round-trip unmarshal failed: %v", err)
		}
		out2, err := proto.Marshal(&round)
		if err != nil {
			t.Fatalf("round-trip re-marshal failed: %v", err)
		}
		if !bytes.Equal(out, out2) {
			t.Fatalf("non-idempotent re-marshal: len(out)=%d len(out2)=%d",
				len(out), len(out2))
		}
		if err := round.Validate(); err != nil {
			t.Fatalf("round-trip Validate() failed: %v", err)
		}
	})
}
