package keeper

import (
	"crypto/ecdsa"
	"math"
	"math/big"
	"testing"

	"cosmossdk.io/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/tracing"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	ethparams "github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"

	"github.com/mezo-org/mezod/x/evm/statedb"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

var testChainID = big.NewInt(31611)

// noPrecompiles is the default isPrecompile predicate for tests that don't
// exercise EIP-7702's precompile-target rejection. Returning false for every
// address preserves the pre-rejection behavior of validateSetCodeAuthorization.
func noPrecompiles(common.Address) bool { return false }

// testPrecompileSet is the address set behind testPrecompiles. Includes a
// representative stock-fork precompile per fork (Frontier/Istanbul/Cancun)
// plus two of Mezo's custom precompiles. Custom-precompile addresses are
// sourced from x/evm/types (not the precompile/ packages themselves) to
// avoid an import cycle: precompile/btctoken imports x/evm/keeper.
var testPrecompileSet = map[common.Address]struct{}{
	common.HexToAddress("0x01"):                                 {},
	common.HexToAddress("0x09"):                                 {},
	common.HexToAddress("0x0a"):                                 {},
	common.HexToAddress(evmtypes.BTCTokenPrecompileAddress):     {},
	common.HexToAddress(evmtypes.AssetsBridgePrecompileAddress): {},
}

func testPrecompiles(a common.Address) bool {
	_, ok := testPrecompileSet[a]
	return ok
}

func newAuthorityKey(t *testing.T) (*ecdsa.PrivateKey, common.Address) {
	t.Helper()
	priv, err := crypto.GenerateKey()
	require.NoError(t, err)
	return priv, crypto.PubkeyToAddress(priv.PublicKey)
}

// signAuth builds a SetCodeAuthorization signed by `priv` over (chainID,
// target, nonce). chainID==nil signs a 0-chain (cross-chain) authorization.
func signAuth(
	t *testing.T,
	chainID *big.Int,
	target common.Address,
	nonce uint64,
	priv *ecdsa.PrivateKey,
) ethtypes.SetCodeAuthorization {
	t.Helper()
	chainU256 := new(uint256.Int)
	if chainID != nil {
		chainU256 = uint256.MustFromBig(chainID)
	}
	signed, err := ethtypes.SignSetCode(priv, ethtypes.SetCodeAuthorization{
		ChainID: *chainU256,
		Address: target,
		Nonce:   nonce,
	})
	require.NoError(t, err)
	return signed
}

func TestApplySetCodeAuthorization_Validation(t *testing.T) {
	target := common.HexToAddress("0xdEAD000000000000000000000000000000000001")

	type tc struct {
		name string
		// build returns the auth + the authority address. The authority's
		// pre-existing nonce/code is set up in `setup`.
		build func(t *testing.T) (ethtypes.SetCodeAuthorization, common.Address)
		setup func(vmdb *statedb.StateDB, authority common.Address)
		// expectErr is the exact sentinel the validation branch must
		// return. Asserted via errors.Is so the wrapped invalid-signature
		// case (fmt.Errorf("%w: %v", ...)) still matches.
		expectErr error
		// expectWarmed reports whether the authority must be in the access
		// list after the (failing) call. Geth's ordering warms the authority
		// after signature recovery but before the destination-code and
		// nonce-mismatch checks, so failures past signature recovery leave
		// the authority warmed.
		expectWarmed bool
		// expectCodeUnchanged, when non-nil, asserts the authority retains
		// this exact code after the failing call (vs. nil/empty for fresh
		// authorities).
		expectCodeUnchanged []byte
	}

	cases := []tc{
		{
			name: "wrong chain id",
			build: func(t *testing.T) (ethtypes.SetCodeAuthorization, common.Address) {
				priv, authority := newAuthorityKey(t)
				wrong := new(big.Int).Add(testChainID, big.NewInt(1))
				return signAuth(t, wrong, target, 0, priv), authority
			},
			expectErr:    errSetCodeAuthorizationWrongChainID,
			expectWarmed: false,
		},
		{
			name: "nonce overflow",
			build: func(t *testing.T) (ethtypes.SetCodeAuthorization, common.Address) {
				priv, authority := newAuthorityKey(t)
				return signAuth(t, testChainID, target, math.MaxUint64, priv), authority
			},
			expectErr:    errSetCodeAuthorizationNonceOverflow,
			expectWarmed: false,
		},
		{
			name: "nonce mismatch",
			build: func(t *testing.T) (ethtypes.SetCodeAuthorization, common.Address) {
				priv, authority := newAuthorityKey(t)
				// authority has nonce 0 in state; sign with nonce 5.
				return signAuth(t, testChainID, target, 5, priv), authority
			},
			expectErr:    errSetCodeAuthorizationNonceMismatch,
			expectWarmed: true,
		},
		{
			// EIP-7702's destination-has-code check is on the AUTHORITY's
			// existing code (not the target's): an authority that already
			// holds non-delegation bytecode cannot be re-delegated.
			name: "authority already has non-delegation code",
			build: func(t *testing.T) (ethtypes.SetCodeAuthorization, common.Address) {
				priv, authority := newAuthorityKey(t)
				return signAuth(t, testChainID, target, 0, priv), authority
			},
			setup: func(vmdb *statedb.StateDB, authority common.Address) {
				vmdb.SetCode(authority, []byte{0x60, 0x60, 0x60, 0x40}, tracing.CodeChangeUnspecified)
				require.NoError(t, vmdb.Commit())
			},
			expectErr:           errSetCodeAuthorizationDestinationHasCode,
			expectWarmed:        true,
			expectCodeUnchanged: []byte{0x60, 0x60, 0x60, 0x40},
		},
		{
			name: "invalid signature",
			build: func(t *testing.T) (ethtypes.SetCodeAuthorization, common.Address) {
				priv, authority := newAuthorityKey(t)
				signed := signAuth(t, testChainID, target, 0, priv)
				// Mutate R to a value that fails ValidateSignatureValues
				// (R = 0 is canonical-form-invalid).
				signed.R = *uint256.NewInt(0)
				return signed, authority
			},
			expectErr: errSetCodeAuthorizationInvalidSignature,
			// Signature recovery fails before AddAddressToAccessList runs.
			expectWarmed: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			vmdb, _ := newTestDB()
			auth, authority := c.build(t)
			if c.setup != nil {
				c.setup(vmdb, authority)
			}

			err := applySetCodeAuthorization(vmdb, testChainID, &auth, noPrecompiles)
			require.ErrorIs(t, err, c.expectErr,
				"auth must be rejected with the matching sentinel")

			// No nonce bump on the authority.
			require.Equal(t, uint64(0), vmdb.GetNonce(authority))

			if c.expectCodeUnchanged != nil {
				require.Equal(t, c.expectCodeUnchanged, vmdb.GetCode(authority))
			} else {
				require.Empty(t, vmdb.GetCode(authority))
			}

			if c.expectWarmed {
				require.True(t, vmdb.AddressInAccessList(authority))
			} else {
				require.False(t, vmdb.AddressInAccessList(authority))
			}
		})
	}
}

func TestApplySetCodeAuthorization_InstallDelegation(t *testing.T) {
	target := common.HexToAddress("0xCAfE000000000000000000000000000000000001")
	priv, authority := newAuthorityKey(t)
	auth := signAuth(t, testChainID, target, 0, priv)

	vmdb, _ := newTestDB()
	require.NoError(t, applySetCodeAuthorization(vmdb, testChainID, &auth, noPrecompiles))

	// Access list lives on the StateDB instance, not the backing store —
	// assert before commit/recreation.
	require.True(t, vmdb.AddressInAccessList(authority))

	require.NoError(t, vmdb.Commit())

	require.Equal(t, uint64(1), vmdb.GetNonce(authority))
	require.Equal(t, ethtypes.AddressToDelegation(target), vmdb.GetCode(authority))
}

func TestApplySetCodeAuthorization_ClearDelegation(t *testing.T) {
	priv, authority := newAuthorityKey(t)

	// Pre-install some delegation code AND a storage slot to confirm clear
	// only affects code, not storage.
	target := common.HexToAddress("0x0000000000000000000000000000000000000099")
	slot := common.HexToHash("0x01")
	val := common.HexToHash("0xbeef")
	vmdb, _ := newTestDB()
	vmdb.SetCode(authority, ethtypes.AddressToDelegation(target), tracing.CodeChangeUnspecified)
	vmdb.SetState(authority, slot, val)
	require.NoError(t, vmdb.Commit())

	clearAuth := signAuth(t, testChainID, common.Address{}, 0, priv)
	require.NoError(t, applySetCodeAuthorization(vmdb, testChainID, &clearAuth, noPrecompiles))
	require.NoError(t, vmdb.Commit())

	require.Equal(t, uint64(1), vmdb.GetNonce(authority))
	require.Empty(t, vmdb.GetCode(authority), "code should be cleared")
	require.Equal(t, val, vmdb.GetState(authority, slot), "storage must persist after clear")
}

func TestApplySetCodeAuthorization_RotateDelegation(t *testing.T) {
	targetA := common.HexToAddress("0xAAaA000000000000000000000000000000000001")
	targetB := common.HexToAddress("0xBbBB000000000000000000000000000000000002")

	priv, authority := newAuthorityKey(t)
	vmdb, _ := newTestDB()

	// Install A.
	authA := signAuth(t, testChainID, targetA, 0, priv)
	require.NoError(t, applySetCodeAuthorization(vmdb, testChainID, &authA, noPrecompiles))

	// Install B (nonce now 1). The journal carries A's writes, so no
	// intermediate Commit is needed.
	authB := signAuth(t, testChainID, targetB, 1, priv)
	require.NoError(t, applySetCodeAuthorization(vmdb, testChainID, &authB, noPrecompiles))
	require.NoError(t, vmdb.Commit())

	require.Equal(t, uint64(2), vmdb.GetNonce(authority))
	require.Equal(t, ethtypes.AddressToDelegation(targetB), vmdb.GetCode(authority))
}

// TestApplySetCodeAuthorization_TargetIsContract pins the spec semantics
// of EIP-7702's destination-has-code check: it gates on the AUTHORITY's
// existing code, not the target's. A target address that holds arbitrary
// non-delegation bytecode is a perfectly valid delegation target — the
// auth must succeed and install the delegation marker on the authority.
func TestApplySetCodeAuthorization_TargetIsContract(t *testing.T) {
	target := common.HexToAddress("0xc0DE000000000000000000000000000000000001")

	// Deploy non-delegation bytecode at target. The check the authority
	// goes through is on the authority's code, not target's, so this must
	// not block the delegation install.
	vmdb, _ := newTestDB()
	vmdb.SetCode(target, []byte{0x60, 0x60, 0x60, 0x40}, tracing.CodeChangeUnspecified)
	require.NoError(t, vmdb.Commit())

	priv, authority := newAuthorityKey(t)
	auth := signAuth(t, testChainID, target, 0, priv)

	require.NoError(t, applySetCodeAuthorization(vmdb, testChainID, &auth, noPrecompiles),
		"auth must succeed even when target holds non-delegation code")
	require.NoError(t, vmdb.Commit())

	require.Equal(t, uint64(1), vmdb.GetNonce(authority))
	require.Equal(t, ethtypes.AddressToDelegation(target), vmdb.GetCode(authority))
}

// TestApplySetCodeAuthorization_AuthorityWasDelegated pins EIP-7702's
// rotate semantics: the destination-has-code check parses the authority's
// existing code as a delegation marker — so an authority that ALREADY holds
// 0xef0100||addr is re-delegatable. The new tuple rotates the marker to the
// new target and bumps the authority's nonce.
func TestApplySetCodeAuthorization_AuthorityWasDelegated(t *testing.T) {
	priv, authority := newAuthorityKey(t)
	oldTarget := common.HexToAddress("0x0DD0000000000000000000000000000000000001")
	newTarget := common.HexToAddress("0xAce0000000000000000000000000000000000002")

	// Pre-install a delegation marker on the authority. ParseDelegation
	// must accept this so the destination-code check passes.
	vmdb, _ := newTestDB()
	vmdb.SetCode(authority, ethtypes.AddressToDelegation(oldTarget), tracing.CodeChangeUnspecified)
	require.NoError(t, vmdb.Commit())

	auth := signAuth(t, testChainID, newTarget, 0, priv)
	require.NoError(t, applySetCodeAuthorization(vmdb, testChainID, &auth, noPrecompiles),
		"auth must succeed when authority's existing code is a delegation marker")
	require.NoError(t, vmdb.Commit())

	require.Equal(t, uint64(1), vmdb.GetNonce(authority))
	require.Equal(t,
		ethtypes.AddressToDelegation(newTarget),
		vmdb.GetCode(authority),
		"delegation must rotate to the new target",
	)
}

// TestApplySetCodeAuthorization_ClearOnFreshAuthority pins clear-path
// semantics on an authority that has no pre-existing delegation: clear
// (auth.Address == 0x0) must succeed, bump the nonce to 1, and leave code
// empty. Sibling to the existing rotate-then-clear test, which exercises
// clear on a delegated authority.
func TestApplySetCodeAuthorization_ClearOnFreshAuthority(t *testing.T) {
	priv, authority := newAuthorityKey(t)
	clearAuth := signAuth(t, testChainID, common.Address{}, 0, priv)

	vmdb, _ := newTestDB()
	require.NoError(t, applySetCodeAuthorization(vmdb, testChainID, &clearAuth, noPrecompiles),
		"clear must succeed on a fresh authority")
	require.NoError(t, vmdb.Commit())

	require.Equal(t, uint64(1), vmdb.GetNonce(authority))
	require.Empty(t, vmdb.GetCode(authority),
		"code must remain empty after clear on fresh authority")
}

// TestApplySetCodeAuthorization_CrossChainNilChainID pins EIP-7702's
// universal-authorization branch: a tuple signed with chainID == 0 is
// accepted on any chain. Drives signAuth with chainID == nil (the helper
// builds a 0-chain auth) and asserts the delegation installs against the
// running chain id.
func TestApplySetCodeAuthorization_CrossChainNilChainID(t *testing.T) {
	target := common.HexToAddress("0xCAfE000000000000000000000000000000000ccc")

	priv, authority := newAuthorityKey(t)
	auth := signAuth(t, nil, target, 0, priv)

	vmdb, _ := newTestDB()
	require.NoError(t, applySetCodeAuthorization(vmdb, testChainID, &auth, noPrecompiles),
		"cross-chain (chainID==0) auth must be accepted on this chain")
	require.NoError(t, vmdb.Commit())

	require.Equal(t, uint64(1), vmdb.GetNonce(authority))
	require.Equal(t, ethtypes.AddressToDelegation(target), vmdb.GetCode(authority))
}

// TestApplySetCodeAuthorization_Refund pins EIP-7702's refund rule: the
// per-tuple new-account intrinsic-gas overcharge (CallNewAccountGas -
// TxAuthTupleGas) is refunded iff the authority already exists in state.
func TestApplySetCodeAuthorization_Refund(t *testing.T) {
	target := common.HexToAddress("0xCAfE000000000000000000000000000000000099")

	cases := []struct {
		name           string
		authorityExist bool
		expectedDelta  uint64
	}{
		{
			name:           "existing authority refunds new-account overcharge",
			authorityExist: true,
			expectedDelta:  ethparams.CallNewAccountGas - ethparams.TxAuthTupleGas,
		},
		{
			name:           "fresh authority has no refund",
			authorityExist: false,
			expectedDelta:  0,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			priv, authority := newAuthorityKey(t)
			vmdb, _ := newTestDB()
			if c.authorityExist {
				// Materialize the authority so Exist(authority) is true at
				// apply time. AddBalance is the cheapest state-creating write.
				vmdb.AddBalance(authority, uint256.NewInt(1), tracing.BalanceChangeUnspecified)
				require.NoError(t, vmdb.Commit())
			}

			auth := signAuth(t, testChainID, target, 0, priv)
			before := vmdb.GetRefund()
			require.NoError(t, applySetCodeAuthorization(vmdb, testChainID, &auth, noPrecompiles))
			after := vmdb.GetRefund()

			require.Equal(t, c.expectedDelta, after-before)
		})
	}
}

// TestApplySetCodeAuthorizations_PostLoopWarmsResolvedTarget pins the
// post-loop access-list warming branch in applySetCodeAuthorizations:
// after the auth loop runs, if msg.To resolves to a 0xef0100||addr
// delegation marker, the resolved target must be added to the access list.
//
// Coverage:
//   - "installed in this tx": the auth tuple installs a delegation on
//     msg.To during the loop; the post-loop branch must then warm the
//     resolved target.
//   - "pre-existing delegation": msg.To already holds a delegation marker
//     in committed state (no auth tuples in the tx); the post-loop branch
//     must still warm the resolved target.
//   - "no delegation on msg.To": negative pin — msg.To is a fresh EOA with
//     no delegation marker; the post-loop branch must NOT add an unrelated
//     sentinel address to the access list. Without this case a regression
//     that warmed every authority/target unconditionally would still pass
//     the two positive cases above.
func TestApplySetCodeAuthorizations_PostLoopWarmsResolvedTarget(t *testing.T) {
	type tc struct {
		name string
		// build returns msg.To (the delegated EOA), the resolved target
		// (or zero when there is no delegation), and auth tuples to thread
		// into the call.
		build func(t *testing.T) (msgTo common.Address, target common.Address, authList []ethtypes.SetCodeAuthorization)
		// setup runs against the StateDB before the call.
		setup func(vmdb *statedb.StateDB, msgTo, target common.Address)
		// expectTargetWarmed reports whether `target` must be in the
		// access list after the call. False for the negative pin where
		// msg.To has no delegation; we then assert a sentinel address
		// is NOT warmed instead.
		expectTargetWarmed bool
	}

	// notWarmedSentinel is an arbitrary deterministic address never
	// touched by any code path under test. Used by the negative case to
	// pin "the post-loop branch does not blanket-warm unrelated addrs".
	notWarmedSentinel := common.HexToAddress("0xDEadDeAdDeAdDeAdDeAdDeAdDeAdDeAd00000C0C")

	cases := []tc{
		{
			name: "installed in this tx",
			build: func(t *testing.T) (common.Address, common.Address, []ethtypes.SetCodeAuthorization) {
				priv, authority := newAuthorityKey(t)
				target := common.HexToAddress("0xDeAd0000000000000000000000000000000000aA")
				return authority, target, []ethtypes.SetCodeAuthorization{
					signAuth(t, testChainID, target, 0, priv),
				}
			},
			expectTargetWarmed: true,
		},
		{
			name: "pre-existing delegation",
			build: func(_ *testing.T) (common.Address, common.Address, []ethtypes.SetCodeAuthorization) {
				delegated := common.HexToAddress("0xDEAd000000000000000000000000000000000DE1")
				target := common.HexToAddress("0xCAfE000000000000000000000000000000000fff")
				return delegated, target, nil
			},
			setup: func(vmdb *statedb.StateDB, delegated, target common.Address) {
				vmdb.SetCode(delegated, ethtypes.AddressToDelegation(target), tracing.CodeChangeUnspecified)
				require.NoError(t, vmdb.Commit())
			},
			expectTargetWarmed: true,
		},
		{
			// Negative pin: msg.To is a fresh EOA with no delegation; the
			// auth tuple points at a DIFFERENT authority so the loop has
			// real work to do but the post-loop branch must not warm an
			// unrelated sentinel address.
			name: "no delegation on msg.To",
			build: func(t *testing.T) (common.Address, common.Address, []ethtypes.SetCodeAuthorization) {
				plainTo := common.HexToAddress("0xDEAd000000000000000000000000000000000DE2")
				priv, _ := newAuthorityKey(t)
				otherTarget := common.HexToAddress("0xCAfE000000000000000000000000000000000123")
				return plainTo, otherTarget, []ethtypes.SetCodeAuthorization{
					signAuth(t, testChainID, otherTarget, 0, priv),
				}
			},
			expectTargetWarmed: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			vmdb, _ := newTestDB()
			msgTo, target, authList := c.build(t)
			if c.setup != nil {
				c.setup(vmdb, msgTo, target)
			}

			msg := core.Message{
				To:                    &msgTo,
				SetCodeAuthorizations: authList,
			}
			applySetCodeAuthorizations(log.NewNopLogger(), vmdb, testChainID, msg, noPrecompiles)

			if c.expectTargetWarmed {
				require.True(t, vmdb.AddressInAccessList(target),
					"resolved delegation target must be warmed by the post-loop branch")
			} else {
				require.False(t, vmdb.AddressInAccessList(notWarmedSentinel),
					"sentinel address must NOT be in the access list — post-loop branch must be a no-op when msg.To has no delegation marker")
			}
		})
	}
}

// TestApplySetCodeAuthorizations_InvalidTupleSilentlySkipped pins the
// spec's "invalid tuples are skipped, valid ones still apply" behavior at
// the loop level: a tuple signed against the wrong chainID must be
// silently rejected (no error, no state mutation, no access-list warming
// — validateSetCodeAuthorization rejects BEFORE AddAddressToAccessList
// runs), while a correctly-signed tuple in the same loop must apply
// normally (delegation installed, nonce bumped, authority warmed).
func TestApplySetCodeAuthorizations_InvalidTupleSilentlySkipped(t *testing.T) {
	privA, authA := newAuthorityKey(t)
	privB, authB := newAuthorityKey(t)

	targetA := common.HexToAddress("0xAaAa000000000000000000000000000000000aaa")
	targetB := common.HexToAddress("0xBbBb000000000000000000000000000000000bbb")

	wrongChain := new(big.Int).Add(testChainID, big.NewInt(1))
	tupleA := signAuth(t, wrongChain, targetA, 0, privA)
	tupleB := signAuth(t, testChainID, targetB, 0, privB)

	vmdb, _ := newTestDB()
	// msg.To is irrelevant to this test; any address works.
	to := authB
	msg := core.Message{
		To:                    &to,
		SetCodeAuthorizations: []ethtypes.SetCodeAuthorization{tupleA, tupleB},
	}
	applySetCodeAuthorizations(log.NewNopLogger(), vmdb, testChainID, msg, noPrecompiles)

	// Rejected tuple: no state mutation, no warming.
	require.Empty(t, vmdb.GetCode(authA),
		"rejected tuple must leave authority code untouched")
	require.Equal(t, uint64(0), vmdb.GetNonce(authA),
		"rejected tuple must leave authority nonce untouched")
	require.False(t, vmdb.AddressInAccessList(authA),
		"wrong-chainID rejection happens before AddAddressToAccessList")

	// Accepted tuple: delegation installed, nonce bumped, authority warmed.
	require.Equal(t, ethtypes.AddressToDelegation(targetB), vmdb.GetCode(authB),
		"accepted tuple must install delegation marker")
	require.Equal(t, uint64(1), vmdb.GetNonce(authB),
		"accepted tuple must bump authority nonce")
	require.True(t, vmdb.AddressInAccessList(authB),
		"accepted tuple must warm authority")
}

// TestApplySetCodeAuthorization_RejectsPrecompile pins the Mezo-specific
// rule that an EIP-7702 authorization whose target is any precompile —
// stock Ethereum or a Mezo custom precompile — must be rejected.
//
// For each rejection case we assert:
//   - the matching sentinel error is returned;
//   - the authority's nonce did NOT advance and code did NOT change;
//   - the authority is NOT in the access list — the precompile-target
//     check sits before AddAddressToAccessList, so warming never runs.
//
// A clear-delegation auth (auth.Address == 0x0) is included as a negative
// control: zero-address is not a precompile, so the path must succeed.
func TestApplySetCodeAuthorization_RejectsPrecompile(t *testing.T) {
	type tc struct {
		name      string
		target    common.Address
		expectErr error
	}

	cases := []tc{
		{
			name:      "stock_precompile_ecrecover",
			target:    common.HexToAddress("0x01"),
			expectErr: errSetCodeAuthorizationTargetIsPrecompile,
		},
		{
			name:      "stock_precompile_blake2f",
			target:    common.HexToAddress("0x09"),
			expectErr: errSetCodeAuthorizationTargetIsPrecompile,
		},
		{
			name:      "stock_precompile_kzg",
			target:    common.HexToAddress("0x0a"),
			expectErr: errSetCodeAuthorizationTargetIsPrecompile,
		},
		{
			name:      "custom_precompile_btctoken",
			target:    common.HexToAddress(evmtypes.BTCTokenPrecompileAddress),
			expectErr: errSetCodeAuthorizationTargetIsPrecompile,
		},
		{
			name:      "custom_precompile_assetsbridge",
			target:    common.HexToAddress(evmtypes.AssetsBridgePrecompileAddress),
			expectErr: errSetCodeAuthorizationTargetIsPrecompile,
		},
		{
			// Negative control: zero address is not a precompile in any
			// fork or in Mezo's custom set. The clear-delegation path
			// must remain unaffected by the new rejection rule.
			name:      "clear_delegation_unaffected",
			target:    common.Address{},
			expectErr: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			priv, authority := newAuthorityKey(t)
			auth := signAuth(t, testChainID, c.target, 0, priv)

			vmdb, _ := newTestDB()
			err := applySetCodeAuthorization(vmdb, testChainID, &auth, testPrecompiles)

			if c.expectErr == nil {
				require.NoError(t, err)
				require.Equal(t, uint64(1), vmdb.GetNonce(authority),
					"successful auth must bump authority nonce")
				return
			}

			require.ErrorIs(t, err, c.expectErr)
			require.Equal(t, uint64(0), vmdb.GetNonce(authority),
				"rejected auth must leave authority nonce untouched")
			require.Empty(t, vmdb.GetCode(authority),
				"rejected auth must leave authority code untouched")
			require.False(t, vmdb.AddressInAccessList(authority),
				"precompile-target rejection happens before AddAddressToAccessList — authority must not be warmed")
		})
	}
}

// TestApplySetCodeAuthorizations_PrecompileTupleSilentlySkipped pins the
// loop-level behavior of the precompile-target rejection: when a multi-tuple
// message contains one tuple targeting a precompile and one targeting a
// regular contract, the bad tuple is silently skipped while the good tuple
// applies normally. Also asserts the post-loop msg.To warming branch still
// runs.
func TestApplySetCodeAuthorizations_PrecompileTupleSilentlySkipped(t *testing.T) {
	privBad, authBad := newAuthorityKey(t)
	privGood, authGood := newAuthorityKey(t)

	precompileTarget := common.HexToAddress(evmtypes.BTCTokenPrecompileAddress)
	goodTarget := common.HexToAddress("0xCAfE000000000000000000000000000000000777")

	tupleBad := signAuth(t, testChainID, precompileTarget, 0, privBad)
	tupleGood := signAuth(t, testChainID, goodTarget, 0, privGood)

	vmdb, _ := newTestDB()
	// msg.To holds a pre-existing delegation marker so the post-loop
	// resolved-target warming branch has something to warm.
	msgTo := common.HexToAddress("0xDEAd000000000000000000000000000000000A0A")
	msgToTarget := common.HexToAddress("0xCAfE000000000000000000000000000000000B0B")
	vmdb.SetCode(msgTo, ethtypes.AddressToDelegation(msgToTarget), tracing.CodeChangeUnspecified)
	require.NoError(t, vmdb.Commit())

	msg := core.Message{
		To:                    &msgTo,
		SetCodeAuthorizations: []ethtypes.SetCodeAuthorization{tupleBad, tupleGood},
	}
	applySetCodeAuthorizations(log.NewNopLogger(), vmdb, testChainID, msg, testPrecompiles)

	// Bad tuple: silently skipped — no state mutation, no warming.
	require.Empty(t, vmdb.GetCode(authBad),
		"precompile-target tuple must leave authority code untouched")
	require.Equal(t, uint64(0), vmdb.GetNonce(authBad),
		"precompile-target tuple must leave authority nonce untouched")
	require.False(t, vmdb.AddressInAccessList(authBad),
		"precompile-target rejection happens before AddAddressToAccessList")

	// Good tuple: delegation installed, nonce bumped, authority warmed.
	require.Equal(t, ethtypes.AddressToDelegation(goodTarget), vmdb.GetCode(authGood),
		"good tuple must install delegation marker")
	require.Equal(t, uint64(1), vmdb.GetNonce(authGood),
		"good tuple must bump authority nonce")
	require.True(t, vmdb.AddressInAccessList(authGood),
		"good tuple must warm authority")

	// Post-loop branch: msg.To's resolved delegation target must be warmed.
	require.True(t, vmdb.AddressInAccessList(msgToTarget),
		"post-loop warming of msg.To's resolved target must still run")
}
