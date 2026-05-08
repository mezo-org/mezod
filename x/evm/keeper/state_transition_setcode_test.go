package keeper_test

import (
	"crypto/ecdsa"
	"encoding/json"
	"math"
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/tracing"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	ethlogger "github.com/ethereum/go-ethereum/eth/tracers/logger"
	ethparams "github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"

	"github.com/mezo-org/mezod/crypto/ethsecp256k1"
	"github.com/mezo-org/mezod/server/config"
	utiltx "github.com/mezo-org/mezod/testutil/tx"
	mezotypes "github.com/mezo-org/mezod/types"
	"github.com/mezo-org/mezod/x/evm/keeper"
	"github.com/mezo-org/mezod/x/evm/statedb"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// signedAuth builds a SetCodeAuthorization signed by `priv` over (chainID,
// target, nonce). chainID==nil signs a 0-chain (cross-chain) authorization.
func (suite *KeeperTestSuite) signedAuth(
	chainID *big.Int,
	target common.Address,
	nonce uint64,
	priv *ecdsa.PrivateKey,
) ethtypes.SetCodeAuthorization {
	chainU256 := new(uint256.Int)
	if chainID != nil {
		chainU256 = uint256.MustFromBig(chainID)
	}
	auth := ethtypes.SetCodeAuthorization{
		ChainID: *chainU256,
		Address: target,
		Nonce:   nonce,
	}
	signed, err := ethtypes.SignSetCode(priv, auth)
	suite.Require().NoError(err)
	return signed
}

// makeAuthorityKey returns a fresh ECDSA key + its derived address. The
// authority account is *not* yet present in state until something funds it
// or sets its code/nonce.
func (suite *KeeperTestSuite) makeAuthorityKey() (*ecdsa.PrivateKey, common.Address) {
	priv, err := crypto.GenerateKey()
	suite.Require().NoError(err)
	return priv, crypto.PubkeyToAddress(priv.PublicKey)
}

// ensureAuthorityExists creates an empty account at `addr` and commits it,
// so subsequent reads via a fresh StateDB see it as existing.
func (suite *KeeperTestSuite) ensureAuthorityExists(addr common.Address) {
	acc := &mezotypes.EthAccount{
		BaseAccount: authtypes.NewBaseAccount(
			addr.Bytes(),
			nil,
			suite.app.AccountKeeper.NextAccountNumber(suite.ctx),
			0,
		),
		CodeHash: common.BytesToHash(crypto.Keccak256(nil)).String(),
	}
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
}

func (suite *KeeperTestSuite) TestApplySetCodeAuthorization_Validation() {
	chainID := suite.app.EvmKeeper.ChainID()
	target := common.HexToAddress("0xdEAD000000000000000000000000000000000001")

	type tc struct {
		name string
		// build returns the auth + the authority address. The authority's
		// pre-existing nonce/code is set up in `setup`.
		build func() (ethtypes.SetCodeAuthorization, common.Address)
		setup func(common.Address)
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
			build: func() (ethtypes.SetCodeAuthorization, common.Address) {
				priv, authority := suite.makeAuthorityKey()
				wrong := new(big.Int).Add(chainID, big.NewInt(1))
				return suite.signedAuth(wrong, target, 0, priv), authority
			},
			expectErr:    keeper.ErrSetCodeAuthorizationWrongChainID,
			expectWarmed: false,
		},
		{
			name: "nonce overflow",
			build: func() (ethtypes.SetCodeAuthorization, common.Address) {
				priv, authority := suite.makeAuthorityKey()
				return suite.signedAuth(chainID, target, math.MaxUint64, priv), authority
			},
			expectErr:    keeper.ErrSetCodeAuthorizationNonceOverflow,
			expectWarmed: false,
		},
		{
			name: "nonce mismatch",
			build: func() (ethtypes.SetCodeAuthorization, common.Address) {
				priv, authority := suite.makeAuthorityKey()
				// authority has nonce 0 in state; sign with nonce 5.
				return suite.signedAuth(chainID, target, 5, priv), authority
			},
			expectErr:    keeper.ErrSetCodeAuthorizationNonceMismatch,
			expectWarmed: true,
		},
		{
			// EIP-7702's destination-has-code check is on the AUTHORITY's
			// existing code (not the target's): an authority that already
			// holds non-delegation bytecode cannot be re-delegated.
			name: "authority already has non-delegation code",
			build: func() (ethtypes.SetCodeAuthorization, common.Address) {
				priv, authority := suite.makeAuthorityKey()
				return suite.signedAuth(chainID, target, 0, priv), authority
			},
			setup: func(authority common.Address) {
				vmdb := suite.StateDB()
				vmdb.SetCode(authority, []byte{0x60, 0x60, 0x60, 0x40}, tracing.CodeChangeUnspecified)
				suite.Require().NoError(vmdb.Commit())
			},
			expectErr:           keeper.ErrSetCodeAuthorizationDestinationHasCode,
			expectWarmed:        true,
			expectCodeUnchanged: []byte{0x60, 0x60, 0x60, 0x40},
		},
		{
			name: "invalid signature",
			build: func() (ethtypes.SetCodeAuthorization, common.Address) {
				priv, authority := suite.makeAuthorityKey()
				signed := suite.signedAuth(chainID, target, 0, priv)
				// Mutate R to a value that fails ValidateSignatureValues
				// (R = 0 is canonical-form-invalid).
				signed.R = *uint256.NewInt(0)
				return signed, authority
			},
			expectErr: keeper.ErrSetCodeAuthorizationInvalidSignature,
			// Signature recovery fails before AddAddressToAccessList runs.
			expectWarmed: false,
		},
	}

	for _, c := range cases {
		suite.Run(c.name, func() {
			suite.SetupTest()
			auth, authority := c.build()
			if c.setup != nil {
				c.setup(authority)
			}

			vmdb := suite.StateDB()
			err := keeper.ApplySetCodeAuthorization(vmdb, chainID, &auth)
			suite.Require().ErrorIs(err, c.expectErr,
				"auth must be rejected with the matching sentinel")

			// No nonce bump on the authority.
			suite.Require().Equal(uint64(0), vmdb.GetNonce(authority))

			if c.expectCodeUnchanged != nil {
				suite.Require().Equal(c.expectCodeUnchanged, vmdb.GetCode(authority))
			} else {
				suite.Require().Empty(vmdb.GetCode(authority))
			}

			if c.expectWarmed {
				suite.Require().True(vmdb.AddressInAccessList(authority))
			} else {
				suite.Require().False(vmdb.AddressInAccessList(authority))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestApplySetCodeAuthorization_InstallDelegation() {
	suite.SetupTest()
	chainID := suite.app.EvmKeeper.ChainID()
	target := common.HexToAddress("0xCAfE000000000000000000000000000000000001")

	priv, authority := suite.makeAuthorityKey()
	auth := suite.signedAuth(chainID, target, 0, priv)

	vmdb := suite.StateDB()
	err := keeper.ApplySetCodeAuthorization(vmdb, chainID, &auth)
	suite.Require().NoError(err)

	// Access list lives on the StateDB instance, not the backing store —
	// assert before commit/recreation.
	suite.Require().True(vmdb.AddressInAccessList(authority))

	suite.Require().NoError(vmdb.Commit())

	// Re-read to confirm persistence of nonce + delegation code.
	post := suite.StateDB()
	suite.Require().Equal(uint64(1), post.GetNonce(authority))
	suite.Require().Equal(ethtypes.AddressToDelegation(target), post.GetCode(authority))
}

func (suite *KeeperTestSuite) TestApplySetCodeAuthorization_ClearDelegation() {
	suite.SetupTest()
	chainID := suite.app.EvmKeeper.ChainID()
	priv, authority := suite.makeAuthorityKey()

	// Pre-install some delegation code AND a storage slot to confirm clear
	// only affects code, not storage.
	target := common.HexToAddress("0x0000000000000000000000000000000000000099")
	slot := common.HexToHash("0x01")
	val := common.HexToHash("0xbeef")
	vmdb := suite.StateDB()
	vmdb.SetCode(authority, ethtypes.AddressToDelegation(target), tracing.CodeChangeUnspecified)
	vmdb.SetState(authority, slot, val)
	suite.Require().NoError(vmdb.Commit())

	clearAuth := suite.signedAuth(chainID, common.Address{}, 0, priv)
	vmdb = suite.StateDB()
	err := keeper.ApplySetCodeAuthorization(vmdb, chainID, &clearAuth)
	suite.Require().NoError(err)
	suite.Require().NoError(vmdb.Commit())

	post := suite.StateDB()
	suite.Require().Equal(uint64(1), post.GetNonce(authority))
	suite.Require().Empty(post.GetCode(authority), "code should be cleared")
	suite.Require().Equal(val, post.GetState(authority, slot), "storage must persist after clear")
}

func (suite *KeeperTestSuite) TestApplySetCodeAuthorization_RotateDelegation() {
	suite.SetupTest()
	chainID := suite.app.EvmKeeper.ChainID()
	targetA := common.HexToAddress("0xAAaA000000000000000000000000000000000001")
	targetB := common.HexToAddress("0xBbBB000000000000000000000000000000000002")

	priv, authority := suite.makeAuthorityKey()

	// Install A.
	authA := suite.signedAuth(chainID, targetA, 0, priv)
	vmdb := suite.StateDB()
	suite.Require().NoError(keeper.ApplySetCodeAuthorization(vmdb, chainID, &authA))
	suite.Require().NoError(vmdb.Commit())

	// Install B (nonce now 1).
	authB := suite.signedAuth(chainID, targetB, 1, priv)
	vmdb = suite.StateDB()
	suite.Require().NoError(keeper.ApplySetCodeAuthorization(vmdb, chainID, &authB))
	suite.Require().NoError(vmdb.Commit())

	post := suite.StateDB()
	suite.Require().Equal(uint64(2), post.GetNonce(authority))
	suite.Require().Equal(ethtypes.AddressToDelegation(targetB), post.GetCode(authority))
}

// TestApplySetCodeAuthorization_TargetIsContract pins the spec semantics
// of EIP-7702's destination-has-code check: it gates on the AUTHORITY's
// existing code, not the target's. A target address that holds arbitrary
// non-delegation bytecode is a perfectly valid delegation target — the
// auth must succeed and install the delegation marker on the authority.
func (suite *KeeperTestSuite) TestApplySetCodeAuthorization_TargetIsContract() {
	suite.SetupTest()
	chainID := suite.app.EvmKeeper.ChainID()
	target := common.HexToAddress("0xc0DE000000000000000000000000000000000001")

	// Deploy non-delegation bytecode at target. The check the authority
	// goes through is on the authority's code, not target's, so this must
	// not block the delegation install.
	vmdb := suite.StateDB()
	vmdb.SetCode(target, []byte{0x60, 0x60, 0x60, 0x40}, tracing.CodeChangeUnspecified)
	suite.Require().NoError(vmdb.Commit())

	priv, authority := suite.makeAuthorityKey()
	auth := suite.signedAuth(chainID, target, 0, priv)

	vmdb = suite.StateDB()
	err := keeper.ApplySetCodeAuthorization(vmdb, chainID, &auth)
	suite.Require().NoError(err, "auth must succeed even when target holds non-delegation code")
	suite.Require().NoError(vmdb.Commit())

	post := suite.StateDB()
	suite.Require().Equal(uint64(1), post.GetNonce(authority))
	suite.Require().Equal(ethtypes.AddressToDelegation(target), post.GetCode(authority))
}

// TestApplySetCodeAuthorization_AuthorityWasDelegated pins EIP-7702's
// rotate semantics: the destination-has-code check parses the authority's
// existing code as a delegation marker — so an authority that ALREADY holds
// 0xef0100||addr is re-delegatable. The new tuple rotates the marker to the
// new target and bumps the authority's nonce.
func (suite *KeeperTestSuite) TestApplySetCodeAuthorization_AuthorityWasDelegated() {
	suite.SetupTest()
	chainID := suite.app.EvmKeeper.ChainID()

	priv, authority := suite.makeAuthorityKey()
	oldTarget := common.HexToAddress("0x0DD0000000000000000000000000000000000001")
	newTarget := common.HexToAddress("0xAce0000000000000000000000000000000000002")

	// Pre-install a delegation marker on the authority. ParseDelegation
	// must accept this so the destination-code check passes.
	vmdb := suite.StateDB()
	vmdb.SetCode(authority, ethtypes.AddressToDelegation(oldTarget), tracing.CodeChangeUnspecified)
	suite.Require().NoError(vmdb.Commit())

	auth := suite.signedAuth(chainID, newTarget, 0, priv)
	vmdb = suite.StateDB()
	err := keeper.ApplySetCodeAuthorization(vmdb, chainID, &auth)
	suite.Require().NoError(err, "auth must succeed when authority's existing code is a delegation marker")
	suite.Require().NoError(vmdb.Commit())

	post := suite.StateDB()
	suite.Require().Equal(uint64(1), post.GetNonce(authority))
	suite.Require().Equal(
		ethtypes.AddressToDelegation(newTarget),
		post.GetCode(authority),
		"delegation must rotate to the new target",
	)
}

// TestApplySetCodeAuthorization_ClearOnFreshAuthority pins clear-path
// semantics on an authority that has no pre-existing delegation: clear
// (auth.Address == 0x0) must succeed, bump the nonce to 1, and leave code
// empty. Sibling to the existing rotate-then-clear test, which exercises
// clear on a delegated authority.
func (suite *KeeperTestSuite) TestApplySetCodeAuthorization_ClearOnFreshAuthority() {
	suite.SetupTest()
	chainID := suite.app.EvmKeeper.ChainID()

	priv, authority := suite.makeAuthorityKey()
	clearAuth := suite.signedAuth(chainID, common.Address{}, 0, priv)

	vmdb := suite.StateDB()
	err := keeper.ApplySetCodeAuthorization(vmdb, chainID, &clearAuth)
	suite.Require().NoError(err, "clear must succeed on a fresh authority")
	suite.Require().NoError(vmdb.Commit())

	post := suite.StateDB()
	suite.Require().Equal(uint64(1), post.GetNonce(authority))
	suite.Require().Empty(post.GetCode(authority), "code must remain empty after clear on fresh authority")
}

// TestApplySetCodeAuthorization_CrossChainNilChainID pins EIP-7702's
// universal-authorization branch: a tuple signed with chainID == 0 is
// accepted on any chain. Drives signedAuth with chainID == nil (the helper
// builds a 0-chain auth) and asserts the delegation installs against the
// running mezod chain id.
func (suite *KeeperTestSuite) TestApplySetCodeAuthorization_CrossChainNilChainID() {
	suite.SetupTest()
	chainID := suite.app.EvmKeeper.ChainID()
	target := common.HexToAddress("0xCAfE000000000000000000000000000000000ccc")

	priv, authority := suite.makeAuthorityKey()
	auth := suite.signedAuth(nil, target, 0, priv)

	vmdb := suite.StateDB()
	err := keeper.ApplySetCodeAuthorization(vmdb, chainID, &auth)
	suite.Require().NoError(err, "cross-chain (chainID==0) auth must be accepted on this chain")
	suite.Require().NoError(vmdb.Commit())

	post := suite.StateDB()
	suite.Require().Equal(uint64(1), post.GetNonce(authority))
	suite.Require().Equal(ethtypes.AddressToDelegation(target), post.GetCode(authority))
}

func (suite *KeeperTestSuite) TestApplySetCodeAuthorization_RefundExistingAuthority() {
	suite.SetupTest()
	chainID := suite.app.EvmKeeper.ChainID()
	target := common.HexToAddress("0xCAfE000000000000000000000000000000000099")

	priv, authority := suite.makeAuthorityKey()
	suite.ensureAuthorityExists(authority)

	auth := suite.signedAuth(chainID, target, 0, priv)
	vmdb := suite.StateDB()

	before := vmdb.GetRefund()
	suite.Require().NoError(keeper.ApplySetCodeAuthorization(vmdb, chainID, &auth))
	after := vmdb.GetRefund()

	expected := ethparams.CallNewAccountGas - ethparams.TxAuthTupleGas
	suite.Require().Equal(expected, after-before, "refund must increment by 12500 when authority pre-exists")
}

func (suite *KeeperTestSuite) TestApplySetCodeAuthorization_NoRefundFreshAuthority() {
	suite.SetupTest()
	chainID := suite.app.EvmKeeper.ChainID()
	target := common.HexToAddress("0xCAfE0000000000000000000000000000000000Aa")

	priv, _ := suite.makeAuthorityKey()
	auth := suite.signedAuth(chainID, target, 0, priv)
	vmdb := suite.StateDB()
	before := vmdb.GetRefund()
	suite.Require().NoError(keeper.ApplySetCodeAuthorization(vmdb, chainID, &auth))
	after := vmdb.GetRefund()

	suite.Require().Equal(uint64(0), after-before, "no refund on fresh authority")
}

// TestApplyMessageWithConfig_AuthListInstallsDelegation drives a SetCodeTx
// through applyMessageWithConfig with a single auth tuple and asserts that
// the delegation is installed on the authority (code + nonce bump). This
// pins the auth-loop side-effects only; access-list warming is covered
// separately in TestApplyMessageWithConfig_PostLoopWarmsResolvedTarget.
func (suite *KeeperTestSuite) TestApplyMessageWithConfig_AuthListInstallsDelegation() {
	suite.SetupTest()
	chainID := suite.app.EvmKeeper.ChainID()
	priv, authority := suite.makeAuthorityKey()

	target := common.HexToAddress("0xDeAd0000000000000000000000000000000000aA")
	auth := suite.signedAuth(chainID, target, 0, priv)

	keeperParams := suite.app.EvmKeeper.GetParams(suite.ctx)
	ethCfg := keeperParams.ChainConfig.EthereumConfig(suite.app.EvmKeeper.ChainID())
	signer := ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	msg, err := newNativeMessage(
		suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address),
		suite.ctx.BlockHeight(),
		suite.address,
		ethCfg,
		suite.signer,
		signer,
		ethtypes.SetCodeTxType,
		authority,
		nil,
		nil,
		[]ethtypes.SetCodeAuthorization{auth},
		big.NewInt(suite.ctx.BlockTime().Unix()).Uint64(),
	)
	suite.Require().NoError(err)

	res, _, err := suite.app.EvmKeeper.ApplyMessage(suite.ctx, msg, nil, true)
	suite.Require().NoError(err)
	suite.Require().False(res.Failed(), "vm error: %s", res.VmError)

	// After commit, authority must hold the delegation marker.
	post := suite.StateDB()
	suite.Require().Equal(
		ethtypes.AddressToDelegation(target),
		post.GetCode(authority),
		"delegation must be installed on authority",
	)
	suite.Require().Equal(uint64(1), post.GetNonce(authority))
}

// TestApplyMessageWithConfig_PostLoopWarmsResolvedTarget pins the post-loop
// access-list warming branch in applyMessageWithConfig: after the auth loop
// runs, if msg.To resolves to a 0xef0100||addr delegation marker, the
// resolved target must be added to the access list. The test runs through
// the test-only ApplyMessageWithStateDB seam so the run StateDB instance
// (which holds the access list) is observable after the call returns —
// access lists do not survive Commit() and are not reachable via the
// public ApplyMessage entry point.
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
//
// Removing the post-loop warming branch (state_transition.go) makes the
// positive sub-cases fail at AddressInAccessList(target); making it
// unconditionally warm an extra address makes the negative sub-case fail.
func (suite *KeeperTestSuite) TestApplyMessageWithConfig_PostLoopWarmsResolvedTarget() {
	type tc struct {
		name string
		// build returns msg.To (the delegated EOA), the resolved target
		// (or zero when there is no delegation), and auth tuples to thread
		// into the SetCodeTx. setup runs against a committed StateDB
		// before the message is applied.
		build func() (msgTo common.Address, target common.Address, authList []ethtypes.SetCodeAuthorization)
		setup func(msgTo, target common.Address)
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
			build: func() (common.Address, common.Address, []ethtypes.SetCodeAuthorization) {
				priv, authority := suite.makeAuthorityKey()
				target := common.HexToAddress("0xDeAd0000000000000000000000000000000000aA")
				chainID := suite.app.EvmKeeper.ChainID()
				return authority, target, []ethtypes.SetCodeAuthorization{
					suite.signedAuth(chainID, target, 0, priv),
				}
			},
			expectTargetWarmed: true,
		},
		{
			name: "pre-existing delegation",
			build: func() (common.Address, common.Address, []ethtypes.SetCodeAuthorization) {
				_, delegated := suite.makeAuthorityKey()
				target := common.HexToAddress("0xCAfE000000000000000000000000000000000fff")
				return delegated, target, nil
			},
			setup: func(delegated, target common.Address) {
				vmdb := suite.StateDB()
				vmdb.SetCode(delegated, ethtypes.AddressToDelegation(target), tracing.CodeChangeUnspecified)
				suite.Require().NoError(vmdb.Commit())
			},
			expectTargetWarmed: true,
		},
		{
			// Negative pin: msg.To is a fresh EOA with no delegation; the
			// auth tuple points at a DIFFERENT authority so the loop has
			// real work to do but the post-loop branch must not warm an
			// unrelated sentinel address. Using a SetCodeTx (rather than
			// switching tx types) keeps the wire-shape consistent across
			// all sub-cases — the only varying axis is msg.To's resolved
			// code, which is what the post-loop branch keys on.
			name: "no delegation on msg.To",
			build: func() (common.Address, common.Address, []ethtypes.SetCodeAuthorization) {
				_, plainTo := suite.makeAuthorityKey()
				priv, _ := suite.makeAuthorityKey()
				otherTarget := common.HexToAddress("0xCAfE000000000000000000000000000000000123")
				chainID := suite.app.EvmKeeper.ChainID()
				return plainTo, otherTarget, []ethtypes.SetCodeAuthorization{
					suite.signedAuth(chainID, otherTarget, 0, priv),
				}
			},
			expectTargetWarmed: false,
		},
	}

	for _, c := range cases {
		suite.Run(c.name, func() {
			suite.SetupTest()
			msgTo, target, authList := c.build()
			if c.setup != nil {
				c.setup(msgTo, target)
			}

			keeperParams := suite.app.EvmKeeper.GetParams(suite.ctx)
			ethCfg := keeperParams.ChainConfig.EthereumConfig(suite.app.EvmKeeper.ChainID())
			signer := ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

			msg, err := newNativeMessage(
				suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address),
				suite.ctx.BlockHeight(),
				suite.address,
				ethCfg,
				suite.signer,
				signer,
				ethtypes.SetCodeTxType,
				msgTo,
				nil,
				nil,
				authList,
				big.NewInt(suite.ctx.BlockTime().Unix()).Uint64(),
			)
			suite.Require().NoError(err)

			cfg, err := suite.app.EvmKeeper.EVMConfig(
				suite.ctx,
				sdk.ConsAddress(suite.ctx.BlockHeader().ProposerAddress),
				suite.app.EvmKeeper.ChainID(),
			)
			suite.Require().NoError(err)
			txConfig := statedb.NewEmptyTxConfig(common.BytesToHash(suite.ctx.HeaderHash()))

			vmdb := statedb.New(suite.ctx, suite.app.EvmKeeper, txConfig)
			res, _, err := suite.app.EvmKeeper.ApplyMessageWithStateDB(
				suite.ctx,
				keeper.WrapMessage(msg),
				nil,
				false,
				cfg,
				txConfig,
				vmdb,
			)
			suite.Require().NoError(err)
			suite.Require().False(res.Failed(), "vm error: %s", res.VmError)

			// Sanity: msg.To is in the access list — Prepare always adds dst.
			suite.Require().True(
				vmdb.AddressInAccessList(msgTo),
				"msg.To must be in the access list (Prepare contract)",
			)

			if c.expectTargetWarmed {
				suite.Require().True(
					vmdb.AddressInAccessList(target),
					"resolved delegation target must be warmed by the post-loop branch",
				)
			} else {
				// Negative pin: post-loop branch must not blanket-warm
				// unrelated addresses when msg.To has no delegation.
				suite.Require().False(
					vmdb.AddressInAccessList(notWarmedSentinel),
					"sentinel address must NOT be in the access list — post-loop branch must be a no-op when msg.To has no delegation marker",
				)
			}
		})
	}
}

// TestApplyMessageWithConfig_InvalidTupleSilentlySkipped drives a SetCodeTx
// through ApplyMessage with two tuples: one signed against the wrong chainID
// (must be silently skipped) and one signed correctly (must apply). Asserts
// the spec's "invalid tuples are skipped, valid ones still apply" behavior
// at the keeper boundary. The wrong-chainID tuple also exercises the
// debug-log branch in applyMessageWithConfig that surfaces the rejection
// sentinel for operator diagnosis; the assertions below pin that production
// rejection path is reached without inspecting log output.
//
// Routes through ApplyMessageWithStateDB so the post-call StateDB instance
// (which holds the access list) is observable: validateSetCodeAuthorization
// rejects the wrong-chainID tuple BEFORE AddAddressToAccessList(authority)
// runs, so authA must NOT be warmed; the accepted authB tuple advances past
// signature recovery so it MUST be warmed. Without the access-list pin, a
// regression that warmed every authority unconditionally would still pass
// the nonce/code assertions.
func (suite *KeeperTestSuite) TestApplyMessageWithConfig_InvalidTupleSilentlySkipped() {
	suite.SetupTest()
	chainID := suite.app.EvmKeeper.ChainID()

	privA, authA := suite.makeAuthorityKey()
	privB, authB := suite.makeAuthorityKey()

	targetA := common.HexToAddress("0xAaAa000000000000000000000000000000000aaa")
	targetB := common.HexToAddress("0xBbBb000000000000000000000000000000000bbb")

	wrongChain := new(big.Int).Add(chainID, big.NewInt(1))
	tupleA := suite.signedAuth(wrongChain, targetA, 0, privA)
	tupleB := suite.signedAuth(chainID, targetB, 0, privB)

	keeperParams := suite.app.EvmKeeper.GetParams(suite.ctx)
	ethCfg := keeperParams.ChainConfig.EthereumConfig(suite.app.EvmKeeper.ChainID())
	signer := ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	// msg.To is irrelevant to this test; pick authB so the post-loop warming
	// branch has a concrete address to dereference.
	msg, err := newNativeMessage(
		suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address),
		suite.ctx.BlockHeight(),
		suite.address,
		ethCfg,
		suite.signer,
		signer,
		ethtypes.SetCodeTxType,
		authB,
		nil,
		nil,
		[]ethtypes.SetCodeAuthorization{tupleA, tupleB},
		big.NewInt(suite.ctx.BlockTime().Unix()).Uint64(),
	)
	suite.Require().NoError(err)

	cfg, err := suite.app.EvmKeeper.EVMConfig(
		suite.ctx,
		sdk.ConsAddress(suite.ctx.BlockHeader().ProposerAddress),
		suite.app.EvmKeeper.ChainID(),
	)
	suite.Require().NoError(err)
	txConfig := statedb.NewEmptyTxConfig(common.BytesToHash(suite.ctx.HeaderHash()))

	vmdb := statedb.New(suite.ctx, suite.app.EvmKeeper, txConfig)
	res, _, err := suite.app.EvmKeeper.ApplyMessageWithStateDB(
		suite.ctx,
		keeper.WrapMessage(msg),
		nil,
		true,
		cfg,
		txConfig,
		vmdb,
	)
	suite.Require().NoError(err)
	suite.Require().False(res.Failed(), "vm error: %s", res.VmError)

	// Access-list pin (observable only via the run StateDB instance):
	// authA was rejected at the wrong-chainID gate, which fires BEFORE
	// AddAddressToAccessList(authority); authB was accepted past sig
	// recovery and must be warmed. The map iteration order is irrelevant
	// since each authority is checked individually.
	suite.Require().False(
		vmdb.AddressInAccessList(authA),
		"rejected authority must NOT be in the access list — wrong-chainID rejection fires before AddAddressToAccessList",
	)
	suite.Require().True(
		vmdb.AddressInAccessList(authB),
		"accepted authority must be in the access list — sig recovery passes, warming runs",
	)

	post := suite.StateDB()

	// Tuple A: silently skipped — no nonce bump, no code installed.
	suite.Require().Equal(uint64(0), post.GetNonce(authA))
	suite.Require().Empty(post.GetCode(authA))

	// Tuple B: applied — nonce bumped, delegation installed.
	suite.Require().Equal(uint64(1), post.GetNonce(authB))
	suite.Require().Equal(ethtypes.AddressToDelegation(targetB), post.GetCode(authB))
}

// freshEthSecp256k1Account generates a fresh ethsecp256k1 key and registers
// the corresponding EthAccount on the suite's account keeper. Returns the
// keyring signer usable with MsgEthereumTx.Sign, the ECDSA private key for
// SignSetCode, and the derived address. The account is intentionally
// unfunded; the keeper's apply-message path does not deduct gas / fees
// from the sender's balance — that lives in the external ante / msg-server
// layer — so simulate / trace / direct ApplyMessage tests succeed with a
// zero-balance EOA.
func (suite *KeeperTestSuite) freshEthSecp256k1Account() (
	*ethsecp256k1.PrivKey, *ecdsa.PrivateKey, common.Address,
) {
	priv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	ecdsaPriv, err := priv.ToECDSA()
	suite.Require().NoError(err)
	addr := crypto.PubkeyToAddress(ecdsaPriv.PublicKey)

	acc := &mezotypes.EthAccount{
		BaseAccount: authtypes.NewBaseAccount(
			addr.Bytes(),
			nil,
			suite.app.AccountKeeper.NextAccountNumber(suite.ctx),
			0,
		),
		CodeHash: common.BytesToHash(crypto.Keccak256(nil)).String(),
	}
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
	return priv, ecdsaPriv, addr
}

// signedSelfSponsoredSetCodeTx builds a SetCodeTx where the signer is also
// the authority. Caller passes the authority's pre-bump state nonce; the
// auth tuple is signed against `stateNonce + 1` to match the geth-canonical
// ante-bump-then-execute order. Tx fields (gas, value, etc.) are cloned
// per call so the package-level templateSetCodeTx pointer is never mutated.
func (suite *KeeperTestSuite) signedSelfSponsoredSetCodeTx(
	priv *ethsecp256k1.PrivKey,
	ecdsaPriv *ecdsa.PrivateKey,
	authority common.Address,
	stateNonce uint64,
	target common.Address,
) *evmtypes.MsgEthereumTx {
	chainID := suite.app.EvmKeeper.ChainID()
	auth := suite.signedAuth(chainID, target, stateNonce+1, ecdsaPriv)

	txData := &ethtypes.SetCodeTx{
		ChainID:   uint256.MustFromBig(chainID),
		Nonce:     stateNonce,
		GasTipCap: uint256.NewInt(2),
		GasFeeCap: uint256.NewInt(10),
		Gas:       100_000,
		To:        authority,
		Value:     uint256.NewInt(0),
		Data:      []byte{},
		AuthList:  []ethtypes.SetCodeAuthorization{auth},
	}
	ethTx := ethtypes.NewTx(txData)

	msg := &evmtypes.MsgEthereumTx{}
	suite.Require().NoError(msg.FromEthereumTx(ethTx))
	msg.From = authority.Hex()

	signer := ethtypes.LatestSignerForChainID(chainID)
	suite.Require().NoError(msg.Sign(signer, utiltx.NewSigner(priv)))
	return msg
}

// TestTraceBlock_SelfSponsoredSetCodeAuth drives a self-sponsored
// EIP-7702 SetCodeTx through TraceBlock, which routes each tx through
// the singular traceTx with commit=true. The per-entry-point nonce
// bump in traceTx (mirroring EthIncrementSenderSequenceDecorator) is
// what lets the auth's `auth.Nonce == stateNonce + 1` validate against
// the post-bump value. Without that bump the auth loop would silently
// reject the tuple. TraceTx target-tx routing uses commit=false so the
// auth side-effects are observable only via TraceBlock or the
// predecessor-loop sibling test.
func (suite *KeeperTestSuite) TestTraceBlock_SelfSponsoredSetCodeAuth() {
	suite.SetupTest()

	priv, ecdsaPriv, authority := suite.freshEthSecp256k1Account()

	stateNonce := suite.app.EvmKeeper.GetNonce(suite.ctx, authority)
	suite.Require().Equal(uint64(0), stateNonce)

	target := common.HexToAddress("0xCafe000000000000000000000000000000007702")
	msg := suite.signedSelfSponsoredSetCodeTx(priv, ecdsaPriv, authority, stateNonce, target)

	traceReq := evmtypes.QueryTraceBlockRequest{
		Txs:         []*evmtypes.MsgEthereumTx{msg},
		TraceConfig: nil,
		ChainId:     suite.app.EvmKeeper.ChainID().Int64(),
		BlockNumber: suite.ctx.BlockHeight(),
	}
	res, err := suite.queryClient.TraceBlock(suite.ctx, &traceReq)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	// Inspect the trace payload itself: a regression that returned a
	// successful response but skipped the trace step (e.g. nil tracer
	// passed to applyMessageWithConfig) would produce empty Data even
	// though the state mutations below pass. TraceBlock returns
	// `[]TxTraceResult{Result: <ExecutionResult>}` JSON; parse the wrapper
	// and re-unmarshal Result into ExecutionResult, asserting positive
	// Gas to confirm the default struct logger actually ran.
	suite.Require().NotEmpty(res.Data, "trace payload must not be empty")
	var blockTraces []struct {
		Result json.RawMessage `json:"result"`
		Error  string          `json:"error,omitempty"`
	}
	suite.Require().NoError(json.Unmarshal(res.Data, &blockTraces),
		"TraceBlock payload must unmarshal to []TxTraceResult")
	suite.Require().Len(blockTraces, 1)
	suite.Require().Empty(blockTraces[0].Error, "per-tx trace must not error")
	var firstResult ethlogger.ExecutionResult
	suite.Require().NoError(json.Unmarshal(blockTraces[0].Result, &firstResult),
		"TxTraceResult.Result must unmarshal to ExecutionResult")
	suite.Require().Positive(firstResult.Gas,
		"per-tx ExecutionResult.Gas must be positive — the default struct logger ran")

	post := suite.StateDB()
	suite.Require().Equal(
		ethtypes.AddressToDelegation(target),
		post.GetCode(authority),
		"delegation must be installed on authority after self-sponsored trace",
	)
	suite.Require().Equal(stateNonce+2, post.GetNonce(authority))
}

// TestTraceTx_SelfSponsoredSetCodeAuth_Predecessor is the predecessor-loop
// counterpart to the target-tx test: the same self-sponsored SetCodeTx is
// replayed as a predecessor before the target trace, exercising the bump
// in TraceTx's predecessor loop.
func (suite *KeeperTestSuite) TestTraceTx_SelfSponsoredSetCodeAuth_Predecessor() {
	suite.SetupTest()

	priv, ecdsaPriv, authority := suite.freshEthSecp256k1Account()

	stateNonce := suite.app.EvmKeeper.GetNonce(suite.ctx, authority)
	target := common.HexToAddress("0xCafe000000000000000000000000000000017702")
	predecessor := suite.signedSelfSponsoredSetCodeTx(priv, ecdsaPriv, authority, stateNonce, target)

	// Build a trivial target tx so TraceTx has something to dispatch to
	// after the predecessor loop runs.
	contractAddr := suite.DeployTestContract(suite.T(), suite.address, sdkmath.NewIntWithDecimal(1000, 18).BigInt())
	suite.Commit()
	targetTx := suite.TransferERC20Token(
		suite.T(),
		contractAddr,
		suite.address,
		common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"),
		sdkmath.NewIntWithDecimal(1, 18).BigInt(),
	)
	suite.Commit()

	traceReq := evmtypes.QueryTraceTxRequest{
		Msg:          targetTx,
		Predecessors: []*evmtypes.MsgEthereumTx{predecessor},
		ChainId:      suite.app.EvmKeeper.ChainID().Int64(),
		BlockNumber:  suite.ctx.BlockHeight(),
	}
	res, err := suite.queryClient.TraceTx(suite.ctx, &traceReq)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	// Trace payload pin: a regression that returned a successful response
	// but skipped the trace step would slip past the state-mutation
	// assertions below. TraceTx returns a single ExecutionResult JSON
	// object (not an array — that is TraceBlock's shape).
	suite.Require().NotEmpty(res.Data, "trace payload must not be empty")
	var traceResult ethlogger.ExecutionResult
	suite.Require().NoError(json.Unmarshal(res.Data, &traceResult),
		"TraceTx payload must unmarshal to ExecutionResult")
	suite.Require().Positive(traceResult.Gas,
		"ExecutionResult.Gas must be positive — the default struct logger ran on the target tx")

	post := suite.StateDB()
	suite.Require().Equal(
		ethtypes.AddressToDelegation(target),
		post.GetCode(authority),
		"delegation must be installed on authority after predecessor replay",
	)
	suite.Require().Equal(stateNonce+2, post.GetNonce(authority))
}

// TestEthCall_SelfSponsoredSetCodeAuthDoesNotError pins the EthCall ingress
// path for a self-sponsored EIP-7702 auth payload: TransactionArgs.
// AuthorizationList must thread through ToMessage onto core.Message, and
// EthCall's per-entry-point pre-call sender-nonce bump must satisfy the
// auth loop's `auth.Nonce == sender.Nonce` check for self-sponsored
// tuples. A regression that drops args.AuthorizationList in ToMessage,
// or that removes the EthCall pre-bump, surfaces here as VmError.
//
// EthCall is read-only — its post-state is not observable from a fresh
// keeper StateDB — so the assertion is intentionally narrow: "EthCall
// does not error on a self-sponsored auth payload". The companion pin
// for actual delegation install on the committing path lives in
// TestApplyMessageWithConfig_AuthListInstallsDelegation. The companion
// pin for the per-tuple intrinsic-gas surcharge (which would be missed
// by an EthCall-only check) lives in TestEstimateGas_SetCodeAuthList_
// IntrinsicGas. A previous shape of this test bolted on a hand-built
// core.Message ApplyMessage block which masked the EthCall ingress (a
// regression dropping args.AuthorizationList in ToMessage would still
// pass through the second block) — that block has been removed.
func (suite *KeeperTestSuite) TestEthCall_SelfSponsoredSetCodeAuthDoesNotError() {
	suite.SetupTest()

	_, ecdsaPriv, authority := suite.freshEthSecp256k1Account()

	stateNonce := suite.app.EvmKeeper.GetNonce(suite.ctx, authority)
	chainID := suite.app.EvmKeeper.ChainID()
	target := common.HexToAddress("0xCafe000000000000000000000000000000027702")
	auth := suite.signedAuth(chainID, target, stateNonce+1, ecdsaPriv)

	to := authority
	gas := hexutil.Uint64(100_000)
	args := evmtypes.TransactionArgs{
		From:              &authority,
		To:                &to,
		Gas:               &gas,
		AuthorizationList: []ethtypes.SetCodeAuthorization{auth},
	}
	argsBytes, err := json.Marshal(args)
	suite.Require().NoError(err)

	res, err := suite.queryClient.EthCall(suite.ctx, &evmtypes.EthCallRequest{
		Args:            argsBytes,
		GasCap:          config.DefaultGasCap,
		ProposerAddress: suite.ctx.BlockHeader().ProposerAddress,
	})
	suite.Require().NoError(err)
	suite.Require().Empty(res.VmError,
		"EthCall must accept a self-sponsored auth payload — VmError here flags a regression in TransactionArgs.ToMessage's auth-list threading or EthCall's pre-bump")
}

// TestSimulateV1_SelfSponsoredAuthChain pins the live SimulateV1 pre-call
// sender-nonce bump (simulate_v1.go::processSimBlock) by driving the real
// public SimulateV1 gRPC entry point with a chained two-call block in
// validation=true mode. Validation=true is essential: without it the
// per-call nonce check in validateSimCall is skipped and the regression
// becomes invisible at the response boundary.
//
//   - Call 0 self-sponsors a SetCode auth (auth.Nonce = stateNonce+1).
//     With the live pre-bump, sdb.GetNonce(sender) reaches stateNonce+1
//     before the auth loop runs, the auth validates and bumps a second
//     time → final nonce stateNonce+2.
//   - Call 1 from the same sender pins args.Nonce = stateNonce+2. With
//     validation=true this is checked against sdb.GetNonce(sender).
//     If the call-0 pre-bump regresses (auth silently rejected → only one
//     bump, or none), sdb nonce after call 0 falls short of stateNonce+2
//     and call 1 fails with NonceTooHigh.
//
// Hand-rolled state-mutation ahead of ApplyMessage (the prior shape of
// this test) was a tautology — it asserted a value it had just written.
// This rewrite drives SimulateV1 end-to-end so a deletion of the live
// pre-bump in simulate_v1.go is caught at the response boundary.
func (suite *KeeperTestSuite) TestSimulateV1_SelfSponsoredAuthChain() {
	suite.SetupTest()

	_, ecdsaPriv, sender := suite.freshEthSecp256k1Account()
	stateNonce := suite.app.EvmKeeper.GetNonce(suite.ctx, sender)

	chainID := suite.app.EvmKeeper.ChainID()
	target := common.HexToAddress("0xCafe000000000000000000000000000000037702")
	auth := suite.signedAuth(chainID, target, stateNonce+1, ecdsaPriv)

	// stateOverrides fund the sender for validation=true balance check
	// and fix the starting nonce — simulate's stateOverrides are scoped
	// to the SimulateV1 sdb, so they do not mutate the keeper backing
	// store. Other validated tests in this package use the same shape.
	startNonceHex := hexutil.Uint64(stateNonce)
	state := map[common.Address]map[string]interface{}{
		sender: {"balance": validationFundedBalance, "nonce": &startNonceHex},
	}

	call0Nonce := hexutil.Uint64(stateNonce)
	call1Nonce := hexutil.Uint64(stateNonce + 2)
	gas := hexutil.Uint64(200_000)
	to := sender
	calls := []evmtypes.TransactionArgs{
		{
			From:              &sender,
			To:                &to,
			Nonce:             &call0Nonce,
			Gas:               &gas,
			MaxFeePerGas:      validationMaxFeePerGas,
			AuthorizationList: []ethtypes.SetCodeAuthorization{auth},
		},
		{
			From:         &sender,
			To:           &to,
			Nonce:        &call1Nonce,
			Gas:          &gas,
			MaxFeePerGas: validationMaxFeePerGas,
		},
	}

	resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx,
		suite.simulateV1Request(suite.validatedSimulateRequest(state, calls, true)))
	suite.Require().NoError(err)
	suite.Require().Nil(resp.Error,
		"validation=true chained calls must clear every gate; a non-nil resp.Error here flags a regression in the live SimulateV1 pre-call sender-nonce bump")

	results := suite.simulateV1BlockResults(resp)
	suite.Require().Len(results, 1)
	rawCalls, ok := results[0]["calls"].([]interface{})
	suite.Require().True(ok)
	suite.Require().Len(rawCalls, 2)
	suite.Require().Equal("0x1", rawCalls[0].(map[string]interface{})["status"],
		"call 0 must succeed with auth installed")
	suite.Require().Equal("0x1", rawCalls[1].(map[string]interface{})["status"],
		"call 1 must observe the chained nonce advance from call 0's pre-bump + auth-loop bump")
}

// TestApplyMessageWithConfig_SelfSponsoredNoDoubleBump pins idempotence on
// the consensus path: when the sender has already been pre-bumped (as
// EthIncrementSenderSequenceDecorator does on CheckTx/DeliverTx), running
// applyMessageWithConfig over a self-sponsored SetCodeTx must produce
// final sender nonce of pre-bumped + 1 (from the auth loop) — never
// pre-bumped + 2. The CALL branch in applyMessageWithConfig must not
// re-bump the sender on top of the ante.
func (suite *KeeperTestSuite) TestApplyMessageWithConfig_SelfSponsoredNoDoubleBump() {
	suite.SetupTest()

	_, ecdsaPriv, authority := suite.freshEthSecp256k1Account()

	stateNonce := suite.app.EvmKeeper.GetNonce(suite.ctx, authority)

	// Mirror the ante's pre-bump.
	acct := suite.app.EvmKeeper.GetAccount(suite.ctx, authority)
	suite.Require().NotNil(acct)
	acct.Nonce = stateNonce + 1
	suite.Require().NoError(suite.app.EvmKeeper.SetAccount(suite.ctx, authority, *acct))

	chainID := suite.app.EvmKeeper.ChainID()
	target := common.HexToAddress("0xCafe000000000000000000000000000000047702")
	auth := suite.signedAuth(chainID, target, stateNonce+1, ecdsaPriv)

	to := authority
	coreMsg := core.Message{
		From:                  authority,
		To:                    &to,
		Nonce:                 stateNonce,
		Value:                 big.NewInt(0),
		GasLimit:              100_000,
		GasPrice:              big.NewInt(0),
		GasFeeCap:             big.NewInt(0),
		GasTipCap:             big.NewInt(0),
		Data:                  nil,
		AccessList:            ethtypes.AccessList{},
		SetCodeAuthorizations: []ethtypes.SetCodeAuthorization{auth},
		SkipNonceChecks:       true,
		SkipTransactionChecks: true,
	}

	res, _, err := suite.app.EvmKeeper.ApplyMessage(suite.ctx, coreMsg, nil, true)
	suite.Require().NoError(err)
	suite.Require().False(res.Failed(), "vm error: %s", res.VmError)

	post := suite.StateDB()
	suite.Require().Equal(
		stateNonce+2,
		post.GetNonce(authority),
		"final nonce must be pre-bump + auth-loop bump (no double-bump in the CALL branch)",
	)
	suite.Require().Equal(
		ethtypes.AddressToDelegation(target),
		post.GetCode(authority),
	)
}

// TestEstimateGas_SetCodeAuthList_IntrinsicGas pins that EstimateGas
// charges the per-tuple intrinsic-gas surcharge for EIP-7702 authorization
// lists. Geth's core.IntrinsicGas adds CallNewAccountGas (25_000) per
// tuple, so a SetCodeTx with N tuples must estimate >= TxGas +
// N*CallNewAccountGas. Without TransactionArgs threading the auth list
// into core.Message, the binary search would converge on a gas value
// that excludes the surcharge and wallets would produce under-funded
// txs.
//
// Two-axis pin:
//   - Per-N lower bound (sub-cases): rsp.Gas >= TxGas + N*CallNewAccountGas
//     catches surcharge-omission regressions even at N=1.
//   - Cross-N delta (final assertion): rspN2.Gas - rspN1.Gas must equal
//     exactly (N2-N1)*CallNewAccountGas. This pins the per-tuple cost
//     without depending on the absolute call overhead, so a regression
//     that under- or over-charged per tuple by a constant fails here even
//     if the lower bound still happens to clear.
func (suite *KeeperTestSuite) TestEstimateGas_SetCodeAuthList_IntrinsicGas() {
	chainID := suite.app.EvmKeeper.ChainID()
	target := common.HexToAddress("0xCafe000000000000000000000000000000050001")
	to := common.HexToAddress("0xCafe000000000000000000000000000000050002")

	cases := []struct {
		name string
		n    uint64
	}{
		{"one tuple", 1},
		{"three tuples", 3},
	}

	estimate := func(n uint64) uint64 {
		authList := make([]ethtypes.SetCodeAuthorization, 0, n)
		for i := uint64(0); i < n; i++ {
			priv, _ := suite.makeAuthorityKey()
			authList = append(authList, suite.signedAuth(chainID, target, 0, priv))
		}

		args := evmtypes.TransactionArgs{
			From:              &suite.address,
			To:                &to,
			AuthorizationList: authList,
		}
		argsBytes, err := json.Marshal(args)
		suite.Require().NoError(err)

		rsp, err := suite.queryClient.EstimateGas(suite.ctx, &evmtypes.EthCallRequest{
			Args:            argsBytes,
			GasCap:          config.DefaultGasCap,
			ProposerAddress: suite.ctx.BlockHeader().ProposerAddress,
		})
		suite.Require().NoError(err)
		return rsp.Gas
	}

	gasByN := make(map[uint64]uint64, len(cases))
	for _, tc := range cases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			gas := estimate(tc.n)
			gasByN[tc.n] = gas

			minGas := ethparams.TxGas + tc.n*ethparams.CallNewAccountGas
			suite.Require().GreaterOrEqual(
				gas, minGas,
				"EstimateGas must charge per-tuple CallNewAccountGas surcharge",
			)
		})
	}

	suite.Run("delta between N=1 and N=3 equals exactly 2*CallNewAccountGas", func() {
		// Per-tuple delta pin: independent of the absolute call overhead,
		// so it catches a regression that under/over-charged per tuple by
		// a fixed amount (which the per-N lower-bound check above could
		// miss). 2 == cases[1].n - cases[0].n.
		const dN = uint64(2)
		suite.Require().Equal(
			dN*ethparams.CallNewAccountGas,
			gasByN[3]-gasByN[1],
			"per-tuple surcharge must be exactly CallNewAccountGas",
		)
	})
}

// TestApplyMessageWithConfig_AuthSurvivesTopLevelRevert pins EIP-7702's
// requirement that auth-tuple side-effects (nonce bump + delegation install +
// refund) survive a revert of the top-level call. applySetCodeAuthorization
// writes through the StateDB journal before evm.Call takes its call-frame
// snapshot, so a revert in the call rolls back call-frame writes only — auth
// writes must persist through Commit().
//
// The target of msg.To is a 1-byte contract whose only opcode is 0xfd
// (REVERT). The call therefore reverts unconditionally; if a regression were
// to write any of {AddRefund, SetNonce, SetCode} through a non-journaled
// path, OR were to inadvertently roll the auth writes back as part of the
// revert, this test would fail at the post-commit nonce / code assertions.
func (suite *KeeperTestSuite) TestApplyMessageWithConfig_AuthSurvivesTopLevelRevert() {
	suite.SetupTest()
	chainID := suite.app.EvmKeeper.ChainID()

	priv, authority := suite.makeAuthorityKey()
	target := common.HexToAddress("0xCAfE000000000000000000000000000000007fd0")
	auth := suite.signedAuth(chainID, target, 0, priv)

	// Pre-deploy a 1-byte REVERT contract at revertAddr. The 0xfd opcode
	// reverts top-level execution unconditionally with no consumed input.
	revertAddr := common.HexToAddress("0x000000000000000000000000000000000000fd00")
	vmdb := suite.StateDB()
	vmdb.SetCode(revertAddr, []byte{0xfd}, tracing.CodeChangeUnspecified)
	suite.Require().NoError(vmdb.Commit())

	// Build a core.Message directly so the package-level templateSetCodeTx
	// pointer is not mutated. SkipNonceChecks/SkipTransactionChecks bypass
	// envelope-level guards that ApplyMessageWithConfig itself does not run
	// — the auth loop and call dispatch are what matter for this test.
	to := revertAddr
	msg := core.Message{
		From:                  suite.address,
		To:                    &to,
		Nonce:                 suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address),
		Value:                 big.NewInt(0),
		GasLimit:              200_000,
		GasPrice:              big.NewInt(0),
		GasFeeCap:             big.NewInt(0),
		GasTipCap:             big.NewInt(0),
		Data:                  nil,
		AccessList:            ethtypes.AccessList{},
		SetCodeAuthorizations: []ethtypes.SetCodeAuthorization{auth},
		SkipNonceChecks:       true,
		SkipTransactionChecks: true,
	}

	res, _, err := suite.app.EvmKeeper.ApplyMessage(suite.ctx, msg, nil, true)
	suite.Require().NoError(err, "ApplyMessage itself must not error on revert; vmErr lives on the response")
	suite.Require().True(res.Failed(), "top-level call must have reverted")
	suite.Require().NotEmpty(res.VmError, "VmError must surface the revert")

	// Fresh post-commit view: auth side-effects must have persisted despite
	// the top-level revert.
	post := suite.StateDB()
	suite.Require().Equal(
		uint64(1),
		post.GetNonce(authority),
		"authority nonce bump must survive the top-level revert",
	)
	suite.Require().Equal(
		ethtypes.AddressToDelegation(target),
		post.GetCode(authority),
		"delegation marker must be installed on authority despite the top-level revert",
	)
	suite.Require().Equal(
		uint64(1),
		suite.app.EvmKeeper.GetNonce(suite.ctx, authority),
		"keeper-level nonce read must agree with StateDB view",
	)
}

// TestApplyMessageWithConfig_RejectsAuthListWithNilTo pins the keeper-level
// guard that mirrors geth's preCheck: a core.Message carrying a non-empty
// SetCodeAuthorizations list with To == nil must be rejected up-front, before
// the create branch can charge per-tuple intrinsic gas without applying any
// tuple. SetCodeTx.Validate and the Prague-gating ante cover the consensus
// path; this guard is the chokepoint for simulate / RPC ingress that builds a
// core.Message directly.
func (suite *KeeperTestSuite) TestApplyMessageWithConfig_RejectsAuthListWithNilTo() {
	suite.SetupTest()
	chainID := suite.app.EvmKeeper.ChainID()

	priv, authority := suite.makeAuthorityKey()
	target := common.HexToAddress("0xCafe000000000000000000000000000000007cc1")
	auth := suite.signedAuth(chainID, target, 0, priv)

	msg := core.Message{
		From:                  suite.address,
		To:                    nil,
		Nonce:                 suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address),
		Value:                 big.NewInt(0),
		GasLimit:              200_000,
		GasPrice:              big.NewInt(0),
		GasFeeCap:             big.NewInt(0),
		GasTipCap:             big.NewInt(0),
		Data:                  nil,
		AccessList:            ethtypes.AccessList{},
		SetCodeAuthorizations: []ethtypes.SetCodeAuthorization{auth},
		SkipNonceChecks:       true,
		SkipTransactionChecks: true,
	}

	res, _, err := suite.app.EvmKeeper.ApplyMessage(suite.ctx, msg, nil, true)
	suite.Require().Error(err, "auth list + nil To must be rejected")
	suite.Require().ErrorIs(err, core.ErrSetCodeTxCreate,
		"rejection must surface geth's EIP-7702 sentinel via errors.Is")
	suite.Require().Nil(res, "no response on early reject")

	// No state side-effects: authority's nonce must be untouched and no
	// delegation marker installed.
	post := suite.StateDB()
	suite.Require().Equal(uint64(0), post.GetNonce(authority),
		"authority nonce must not advance when the message is rejected")
	suite.Require().Empty(post.GetCode(authority),
		"no delegation marker must be installed when the message is rejected")
}

// TestApplyMessageWithConfig_DuplicateAuthorityLastWins pins the spec
// language for repeated authorities in the same tuple list: geth processes
// tuples in order, with last-write semantics on the authority's code.
// Two tuples for the same authority key signing different targets at
// consecutive nonces (n, n+1) must both apply — the final delegation
// marker points to the SECOND target and the authority's nonce ends at
// the pre-state nonce + 2.
func (suite *KeeperTestSuite) TestApplyMessageWithConfig_DuplicateAuthorityLastWins() {
	suite.SetupTest()
	chainID := suite.app.EvmKeeper.ChainID()

	priv, authority := suite.makeAuthorityKey()
	targetA := common.HexToAddress("0xAaAa000000000000000000000000000000000777")
	targetB := common.HexToAddress("0xBbBb000000000000000000000000000000000888")

	// Same authority signs n and n+1 against two different targets.
	tupleA := suite.signedAuth(chainID, targetA, 0, priv)
	tupleB := suite.signedAuth(chainID, targetB, 1, priv)

	keeperParams := suite.app.EvmKeeper.GetParams(suite.ctx)
	ethCfg := keeperParams.ChainConfig.EthereumConfig(chainID)
	signer := ethtypes.LatestSignerForChainID(chainID)

	msg, err := newNativeMessage(
		suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address),
		suite.ctx.BlockHeight(),
		suite.address,
		ethCfg,
		suite.signer,
		signer,
		ethtypes.SetCodeTxType,
		authority,
		nil,
		nil,
		[]ethtypes.SetCodeAuthorization{tupleA, tupleB},
		big.NewInt(suite.ctx.BlockTime().Unix()).Uint64(),
	)
	suite.Require().NoError(err)

	res, _, err := suite.app.EvmKeeper.ApplyMessage(suite.ctx, msg, nil, true)
	suite.Require().NoError(err)
	suite.Require().False(res.Failed(), "vm error: %s", res.VmError)

	post := suite.StateDB()
	suite.Require().Equal(
		uint64(2),
		post.GetNonce(authority),
		"both tuples must apply — authority nonce must reach pre-state + 2",
	)
	suite.Require().Equal(
		ethtypes.AddressToDelegation(targetB),
		post.GetCode(authority),
		"last-wins: final delegation must point to the second target",
	)
}

// TestApplyMessageWithConfig_EmptyAuthList pins keeper behavior for a
// SetCode-shaped core.Message carrying a length-0 (non-nil) auth list:
// the auth loop iterates zero times, no nonce bumps occur, and msg.To
// dispatches as a normal call. SetCodeTx envelope validation runs in
// the ante / preCheck layer; this test pins the keeper-internal slice
// for simulate / RPC ingress that may construct the message directly.
func (suite *KeeperTestSuite) TestApplyMessageWithConfig_EmptyAuthList() {
	suite.SetupTest()

	to := common.HexToAddress("0xCAfE000000000000000000000000000000000eee")
	preNonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	msg := core.Message{
		From:                  suite.address,
		To:                    &to,
		Nonce:                 preNonce,
		Value:                 big.NewInt(0),
		GasLimit:              100_000,
		GasPrice:              big.NewInt(0),
		GasFeeCap:             big.NewInt(0),
		GasTipCap:             big.NewInt(0),
		Data:                  nil,
		AccessList:            ethtypes.AccessList{},
		SetCodeAuthorizations: []ethtypes.SetCodeAuthorization{},
		SkipNonceChecks:       true,
		SkipTransactionChecks: true,
	}

	res, _, err := suite.app.EvmKeeper.ApplyMessage(suite.ctx, msg, nil, true)
	suite.Require().NoError(err, "empty auth list must be a no-op for the auth loop, not an error")
	suite.Require().False(res.Failed(), "vm error: %s", res.VmError)

	// No nonce bumps anywhere: msg.To is an EOA so the call is a no-op,
	// and the empty auth list contributes no per-tuple bumps.
	post := suite.StateDB()
	suite.Require().Empty(post.GetCode(to), "msg.To must remain code-less after no-op call")
	suite.Require().Equal(uint64(0), post.GetNonce(to))
}
