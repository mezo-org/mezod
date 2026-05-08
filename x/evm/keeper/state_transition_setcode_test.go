package keeper_test

import (
	"crypto/ecdsa"
	"encoding/json"
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
	suite.T().Helper()
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
	suite.T().Helper()
	priv, err := crypto.GenerateKey()
	suite.Require().NoError(err)
	return priv, crypto.PubkeyToAddress(priv.PublicKey)
}

// TestApplyMessageWithConfig_AuthListInstallsDelegation drives a SetCodeTx
// through applyMessageWithConfig with a single auth tuple and asserts that
// the delegation is installed on the authority (code + nonce bump). The
// per-tuple validation, install/clear/rotate, and access-list warming
// branches are pinned at the helper level in
// set_code_authorization_test.go.
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

// TestApplyMessageWithConfig_InvalidTupleSilentlySkipped drives a SetCodeTx
// through ApplyMessageWithConfig with two tuples: one signed against the
// wrong chainID (must be silently skipped) and one signed correctly (must
// apply). Pins the spec's "invalid tuples are skipped, valid ones still
// apply" behavior at the keeper boundary using only post-commit observable
// state.
//
// The access-list ordering pin (rejected authority NOT warmed; accepted
// authority warmed) is covered at the loop level by
// TestApplySetCodeAuthorizations_InvalidTupleSilentlySkipped in
// set_code_authorization_test.go — that pin needs the run StateDB
// instance, which is not reachable through the public entry point.
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
	ethCfg := keeperParams.ChainConfig.EthereumConfig(chainID)
	signer := ethtypes.LatestSignerForChainID(chainID)

	// msg.To is irrelevant to this test; any address works.
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
		chainID,
	)
	suite.Require().NoError(err)
	txConfig := statedb.NewEmptyTxConfig(common.BytesToHash(suite.ctx.HeaderHash()))

	res, _, err := suite.app.EvmKeeper.ApplyMessageWithConfig(
		suite.ctx,
		keeper.WrapMessage(msg),
		nil,
		true,
		cfg,
		txConfig,
	)
	suite.Require().NoError(err)
	suite.Require().False(res.Failed(), "vm error: %s", res.VmError)

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
