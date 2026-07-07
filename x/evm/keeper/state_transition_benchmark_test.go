package keeper_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/stretchr/testify/require"
)

var templateAccessListTx = &ethtypes.AccessListTx{
	GasPrice: big.NewInt(1),
	Gas:      21000,
	To:       &common.Address{},
	Value:    big.NewInt(0),
	Data:     []byte{},
}

var templateLegacyTx = &ethtypes.LegacyTx{
	GasPrice: big.NewInt(1),
	Gas:      21000,
	To:       &common.Address{},
	Value:    big.NewInt(0),
	Data:     []byte{},
}

var templateDynamicFeeTx = &ethtypes.DynamicFeeTx{
	GasFeeCap: big.NewInt(10),
	GasTipCap: big.NewInt(2),
	Gas:       21000,
	To:        &common.Address{},
	Value:     big.NewInt(0),
	Data:      []byte{},
}

var templateSetCodeTx = &ethtypes.SetCodeTx{
	GasFeeCap: uint256.NewInt(10),
	GasTipCap: uint256.NewInt(2),
	Gas:       100000,
	To:        common.Address{},
	Value:     uint256.NewInt(0),
	Data:      []byte{},
}

func newSignedEthTx(
	txData ethtypes.TxData,
	nonce uint64,
	addr sdk.Address,
	krSigner keyring.Signer,
	ethSigner ethtypes.Signer,
) (*ethtypes.Transaction, error) {
	var ethTx *ethtypes.Transaction
	switch txData := txData.(type) {
	case *ethtypes.AccessListTx:
		txData.Nonce = nonce
		ethTx = ethtypes.NewTx(txData)
	case *ethtypes.LegacyTx:
		txData.Nonce = nonce
		ethTx = ethtypes.NewTx(txData)
	case *ethtypes.DynamicFeeTx:
		txData.Nonce = nonce
		ethTx = ethtypes.NewTx(txData)
	case *ethtypes.SetCodeTx:
		txData.Nonce = nonce
		ethTx = ethtypes.NewTx(txData)
	default:
		return nil, errors.New("unknown transaction type")
	}

	sig, _, err := krSigner.SignByAddress(addr, ethTx.Hash().Bytes(), signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
	if err != nil {
		return nil, err
	}

	ethTx, err = ethTx.WithSignature(ethSigner, sig)
	if err != nil {
		return nil, err
	}

	return ethTx, nil
}

func newEthMsgTx(
	nonce uint64,
	address common.Address,
	krSigner keyring.Signer,
	ethSigner ethtypes.Signer,
	txType byte,
	to common.Address,
	data []byte,
	accessList ethtypes.AccessList,
	authList []ethtypes.SetCodeAuthorization,
) (*evmtypes.MsgEthereumTx, *big.Int, error) {
	var (
		ethTx   *ethtypes.Transaction
		baseFee *big.Int
	)
	switch txType {
	case ethtypes.LegacyTxType:
		templateLegacyTx.Nonce = nonce
		if data != nil {
			templateLegacyTx.Data = data
		}
		ethTx = ethtypes.NewTx(templateLegacyTx)
	case ethtypes.AccessListTxType:
		templateAccessListTx.Nonce = nonce
		if data != nil {
			templateAccessListTx.Data = data
		} else {
			templateAccessListTx.Data = []byte{}
		}

		templateAccessListTx.AccessList = accessList
		ethTx = ethtypes.NewTx(templateAccessListTx)
	case ethtypes.DynamicFeeTxType:
		templateDynamicFeeTx.Nonce = nonce

		if data != nil {
			templateAccessListTx.Data = data
		} else {
			templateAccessListTx.Data = []byte{}
		}
		templateAccessListTx.AccessList = accessList
		ethTx = ethtypes.NewTx(templateDynamicFeeTx)
		baseFee = big.NewInt(3)
	case ethtypes.SetCodeTxType:
		// Clone the package-level template so per-test mutations don't
		// leak across benchmarks/tests via the shared pointer.
		txData := *templateSetCodeTx
		txData.Nonce = nonce
		txData.To = to
		if data != nil {
			txData.Data = data
		} else {
			txData.Data = []byte{}
		}
		txData.AccessList = accessList
		txData.AuthList = authList
		// SetCodeTx requires a non-zero ChainID; signers populate it via
		// LatestSignerForChainID, but ethtypes.NewTx still needs a value
		// here so the unsigned tx isn't malformed.
		txData.ChainID = uint256.MustFromBig(ethSigner.ChainID())
		ethTx = ethtypes.NewTx(&txData)
		baseFee = big.NewInt(3)
	default:
		return nil, baseFee, errors.New("unsupport tx type")
	}

	msg := &evmtypes.MsgEthereumTx{}
	err := msg.FromEthereumTx(ethTx)
	if err != nil {
		return nil, nil, err
	}

	msg.From = address.Hex()

	return msg, baseFee, msg.Sign(ethSigner, krSigner)
}

func newNativeMessage(
	nonce uint64,
	blockHeight int64,
	address common.Address,
	cfg *params.ChainConfig,
	krSigner keyring.Signer,
	ethSigner ethtypes.Signer,
	txType byte,
	to common.Address,
	data []byte,
	accessList ethtypes.AccessList,
	authList []ethtypes.SetCodeAuthorization,
	blockTime uint64,
) (core.Message, error) {
	msgSigner := ethtypes.MakeSigner(cfg, big.NewInt(blockHeight), blockTime)

	msg, baseFee, err := newEthMsgTx(nonce, address, krSigner, ethSigner, txType, to, data, accessList, authList)
	if err != nil {
		return core.Message{}, err
	}

	m, err := msg.AsMessage(msgSigner, baseFee)
	if err != nil {
		return core.Message{}, err
	}

	return m, nil
}

func BenchmarkApplyTransaction(b *testing.B) {
	suite := KeeperTestSuite{enableLondonHF: true}
	suite.SetupTestWithT(b)

	ethSigner := ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tx, err := newSignedEthTx(templateAccessListTx,
			suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address),
			sdk.AccAddress(suite.address.Bytes()),
			suite.signer,
			ethSigner,
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, err := suite.app.EvmKeeper.ApplyTransaction(suite.ctx, tx)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}

func BenchmarkApplyTransactionWithLegacyTx(b *testing.B) {
	suite := KeeperTestSuite{enableLondonHF: true}
	suite.SetupTestWithT(b)

	ethSigner := ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tx, err := newSignedEthTx(templateLegacyTx,
			suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address),
			sdk.AccAddress(suite.address.Bytes()),
			suite.signer,
			ethSigner,
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, err := suite.app.EvmKeeper.ApplyTransaction(suite.ctx, tx)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}

func BenchmarkApplyTransactionWithDynamicFeeTx(b *testing.B) {
	suite := KeeperTestSuite{enableFeemarket: true, enableLondonHF: true}
	suite.SetupTestWithT(b)

	ethSigner := ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tx, err := newSignedEthTx(templateDynamicFeeTx,
			suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address),
			sdk.AccAddress(suite.address.Bytes()),
			suite.signer,
			ethSigner,
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, err := suite.app.EvmKeeper.ApplyTransaction(suite.ctx, tx)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}

//nolint:all
func BenchmarkApplyMessage(b *testing.B) {
	suite := KeeperTestSuite{enableLondonHF: true}
	suite.SetupTestWithT(b)

	params := suite.app.EvmKeeper.GetParams(suite.ctx)
	ethCfg := params.ChainConfig.EthereumConfig(suite.app.EvmKeeper.ChainID())
	signer := ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		m, err := newNativeMessage(
			suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address),
			suite.ctx.BlockHeight(),
			suite.address,
			ethCfg,
			suite.signer,
			signer,
			ethtypes.AccessListTxType,
			common.Address{},
			nil,
			nil,
			nil,
			big.NewInt(suite.ctx.BlockTime().Unix()).Uint64(),
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, _, err := suite.app.EvmKeeper.ApplyMessage(suite.ctx, m, nil, true)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}

//nolint:all
func BenchmarkApplyMessageWithLegacyTx(b *testing.B) {
	suite := KeeperTestSuite{enableLondonHF: true}
	suite.SetupTestWithT(b)

	params := suite.app.EvmKeeper.GetParams(suite.ctx)
	ethCfg := params.ChainConfig.EthereumConfig(suite.app.EvmKeeper.ChainID())
	signer := ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		m, err := newNativeMessage(
			suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address),
			suite.ctx.BlockHeight(),
			suite.address,
			ethCfg,
			suite.signer,
			signer,
			ethtypes.LegacyTxType,
			common.Address{},
			nil,
			nil,
			nil,
			big.NewInt(suite.ctx.BlockTime().Unix()).Uint64(),
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, _, err := suite.app.EvmKeeper.ApplyMessage(suite.ctx, m, nil, true)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}

func BenchmarkApplyMessageWithDynamicFeeTx(b *testing.B) {
	suite := KeeperTestSuite{enableFeemarket: true, enableLondonHF: true}
	suite.SetupTestWithT(b)

	params := suite.app.EvmKeeper.GetParams(suite.ctx)
	ethCfg := params.ChainConfig.EthereumConfig(suite.app.EvmKeeper.ChainID())
	signer := ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		m, err := newNativeMessage(
			suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address),
			suite.ctx.BlockHeight(),
			suite.address,
			ethCfg,
			suite.signer,
			signer,
			ethtypes.DynamicFeeTxType,
			common.Address{},
			nil,
			nil,
			nil,
			big.NewInt(suite.ctx.BlockTime().Unix()).Uint64(),
		)
		require.NoError(b, err)

		b.StartTimer()
		resp, _, err := suite.app.EvmKeeper.ApplyMessage(suite.ctx, m, nil, true)
		b.StopTimer()

		require.NoError(b, err)
		require.False(b, resp.Failed())
	}
}

// BenchmarkApplyMessageWithSetCodeTx exercises the EIP-7702 auth-loop hot
// path through ApplyMessage at varying tuple counts. Each tuple is signed
// by a fresh ECDSA key so geth's signer cache cannot artificially share
// per-tuple ECDSA recovery cost across tuples within one message — the
// benchmark reflects the worst case where every tuple forces a real
// recovery. The core.Message is constructed inline (mirroring
// state_transition_setcode_test.go's ApplyMessageWithConfig_RejectsAuthListWithNilTo
// pattern) to avoid expanding the template-helper signature with a gas
// override; the gas budget is sized as TxGas + N*CallNewAccountGas + 50_000
// so the per-tuple intrinsic charge fits at all tested N.
func BenchmarkApplyMessageWithSetCodeTx(b *testing.B) {
	cases := []struct {
		name string
		n    int
	}{
		{"N=0", 0},
		{"N=1", 1},
		{"N=16", 16},
		{"N=64", 64},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			suite := KeeperTestSuite{enableLondonHF: true}
			suite.SetupTestWithT(b)

			chainID := suite.app.EvmKeeper.ChainID()
			gas := params.TxGas + uint64(c.n)*params.CallNewAccountGas + 50_000 //nolint:gosec

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				b.StopTimer()

				// Build N tuples, each signed by a fresh ECDSA key so the
				// per-tuple ecrecover cost isn't amortized via geth's signer
				// cache (signer keys ECDSA recovery results by signature).
				authList := make([]ethtypes.SetCodeAuthorization, c.n)
				for j := 0; j < c.n; j++ {
					priv, err := crypto.GenerateKey()
					require.NoError(b, err)
					target := common.BigToAddress(big.NewInt(int64(j + 1)))
					auth, err := ethtypes.SignSetCode(priv, ethtypes.SetCodeAuthorization{
						ChainID: *uint256.MustFromBig(chainID),
						Address: target,
						Nonce:   0,
					})
					require.NoError(b, err)
					authList[j] = auth
				}

				// Pick a concrete msg.To so the post-loop warming branch has
				// an address to dereference; the value is irrelevant to cost.
				to := common.BigToAddress(big.NewInt(0xabcd))
				m := core.Message{
					From:                  suite.address,
					To:                    &to,
					Nonce:                 suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address),
					Value:                 big.NewInt(0),
					GasLimit:              gas,
					GasPrice:              big.NewInt(0),
					GasFeeCap:             big.NewInt(0),
					GasTipCap:             big.NewInt(0),
					Data:                  nil,
					AccessList:            ethtypes.AccessList{},
					SetCodeAuthorizations: authList,
					SkipNonceChecks:       true,
					SkipTransactionChecks: true,
				}

				b.StartTimer()
				resp, _, err := suite.app.EvmKeeper.ApplyMessage(suite.ctx, m, nil, true)
				b.StopTimer()

				require.NoError(b, err)
				require.False(b, resp.Failed(), "vm error: %s", resp.VmError)
			}
		})
	}
}
