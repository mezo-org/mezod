package evm_test

import (
	"errors"
	"math/big"
	"strings"

	sdkmath "cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethparams "github.com/ethereum/go-ethereum/params"
	utiltx "github.com/mezo-org/mezod/testutil/tx"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

func (suite *AnteTestSuite) TestAnteHandler() {
	var acc sdk.AccountI
	addr, privKey := utiltx.NewAddrKey()
	to := utiltx.GenerateAddress()
	const incorrectChainID string = "mezo_31600-1"

	setup := func() {
		suite.enableFeemarket = false
		suite.SetupTest() // reset

		acc = suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
		suite.Require().NoError(acc.SetSequence(1))
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

		err := suite.app.EvmKeeper.SetBalance(suite.ctx, addr, big.NewInt(10000000000))
		suite.Require().NoError(err)

		suite.app.FeeMarketKeeper.SetBaseFee(suite.ctx, big.NewInt(100))
	}

	ethContractCreationTxParams := &evmtypes.EvmTxArgs{
		ChainID:   suite.app.EvmKeeper.ChainID(),
		Nonce:     1,
		Amount:    big.NewInt(10),
		GasLimit:  100000,
		GasPrice:  big.NewInt(150),
		GasFeeCap: big.NewInt(200),
	}

	ethTxParams := &evmtypes.EvmTxArgs{
		ChainID:   suite.app.EvmKeeper.ChainID(),
		To:        &to,
		Nonce:     1,
		Amount:    big.NewInt(10),
		GasLimit:  100000,
		GasPrice:  big.NewInt(150),
		GasFeeCap: big.NewInt(200),
	}

	testCases := []struct {
		name      string
		txFn      func() sdk.Tx
		checkTx   bool
		reCheckTx bool
		expPass   bool
	}{
		{
			"success - DeliverTx (contract)",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTx(ethContractCreationTxParams)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			false, false, true,
		},
		{
			"success - CheckTx (contract)",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTx(ethContractCreationTxParams)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			true, false, true,
		},
		{
			"success - ReCheckTx (contract)",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTx(ethContractCreationTxParams)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			false, true, true,
		},
		{
			"success - DeliverTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			},
			false, false, true,
		},
		{
			"success - CheckTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			},
			true, false, true,
		},
		{
			"success - ReCheckTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			}, false, true, true,
		},
		{
			"success - CheckTx (cosmos tx not signed)",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			}, false, true, true,
		},
		{
			"fail - CheckTx (cosmos tx is not valid)",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)
				// bigger than MaxGasWanted
				txBuilder.SetGasLimit(uint64(1 << 63))
				return txBuilder.GetTx()
			}, true, false, false,
		},
		{
			"fail - CheckTx (memo too long)",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)
				txBuilder.SetMemo(strings.Repeat("*", 257))
				return txBuilder.GetTx()
			}, true, false, false,
		},
		{
			"fail - CheckTx (ExtensionOptionsEthereumTx not set)",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false, true)
				return txBuilder.GetTx()
			}, true, false, false,
		},
		// Based on EVMBackend.SendTransaction, for cosmos tx, forcing null for some fields except ExtensionOptions, Fee, MsgEthereumTx
		// should be part of consensus
		{
			"fail - DeliverTx (cosmos tx signed)",
			func() sdk.Tx {
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
				suite.Require().NoError(err)
				ethTxParams := &evmtypes.EvmTxArgs{
					ChainID:  suite.app.EvmKeeper.ChainID(),
					To:       &to,
					Nonce:    nonce,
					Amount:   big.NewInt(10),
					GasLimit: 100000,
					GasPrice: big.NewInt(1),
				}
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, true)
				return tx
			}, false, false, false,
		},
		{
			"fail - DeliverTx (cosmos tx with memo)",
			func() sdk.Tx {
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
				suite.Require().NoError(err)
				ethTxParams := &evmtypes.EvmTxArgs{
					ChainID:  suite.app.EvmKeeper.ChainID(),
					To:       &to,
					Nonce:    nonce,
					Amount:   big.NewInt(10),
					GasLimit: 100000,
					GasPrice: big.NewInt(1),
				}
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)
				txBuilder.SetMemo("memo for cosmos tx not allowed")
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - DeliverTx (cosmos tx with timeoutheight)",
			func() sdk.Tx {
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
				suite.Require().NoError(err)
				ethTxParams := &evmtypes.EvmTxArgs{
					ChainID:  suite.app.EvmKeeper.ChainID(),
					To:       &to,
					Nonce:    nonce,
					Amount:   big.NewInt(10),
					GasLimit: 100000,
					GasPrice: big.NewInt(1),
				}
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)
				txBuilder.SetTimeoutHeight(10)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - DeliverTx (invalid fee amount)",
			func() sdk.Tx {
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
				suite.Require().NoError(err)
				ethTxParams := &evmtypes.EvmTxArgs{
					ChainID:  suite.app.EvmKeeper.ChainID(),
					To:       &to,
					Nonce:    nonce,
					Amount:   big.NewInt(10),
					GasLimit: 100000,
					GasPrice: big.NewInt(1),
				}
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)

				txData, err := evmtypes.UnpackTxData(signedTx.Data)
				suite.Require().NoError(err)

				expFee := txData.Fee()
				invalidFee := new(big.Int).Add(expFee, big.NewInt(1))
				invalidFeeAmount := sdk.Coins{sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewIntFromBigInt(invalidFee))}
				txBuilder.SetFeeAmount(invalidFeeAmount)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - DeliverTx (invalid fee gaslimit)",
			func() sdk.Tx {
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
				suite.Require().NoError(err)
				ethTxParams := &evmtypes.EvmTxArgs{
					ChainID:  suite.app.EvmKeeper.ChainID(),
					To:       &to,
					Nonce:    nonce,
					Amount:   big.NewInt(10),
					GasLimit: 100000,
					GasPrice: big.NewInt(1),
				}
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)

				expGasLimit := signedTx.GetGas()
				invalidGasLimit := expGasLimit + 1
				txBuilder.SetGasLimit(invalidGasLimit)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - DeliverTx EIP712 signed Cosmos Tx with MsgSend",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder, err := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - DeliverTx EIP712 Multiple MsgSend",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MultipleMsgSend(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - DeliverTx EIP712 Multiple Different Msgs",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MultipleDifferentMsgs(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.RequireErrorForLegacyTypedData(err)
				return suite.TxForLegacyTypedData(txBuilder)
			}, false, false, false,
		},
		{
			"fail - DeliverTx EIP712 Same Msgs, Different Schemas",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712SameMsgDifferentSchemas(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.RequireErrorForLegacyTypedData(err)
				return suite.TxForLegacyTypedData(txBuilder)
			}, false, false, false,
		},
		{
			"fail - DeliverTx EIP712 Zero Value Array (Should Not Omit Field)",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712ZeroValueArray(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.RequireErrorForLegacyTypedData(err)
				return suite.TxForLegacyTypedData(txBuilder)
			}, false, false, false,
		},
		{
			"fail - DeliverTx EIP712 Zero Value Number (Should Not Omit Field)",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712ZeroValueNumber(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.RequireErrorForLegacyTypedData(err)
				return suite.TxForLegacyTypedData(txBuilder)
			}, false, false, false,
		},
		{
			"fail - DeliverTx EIP712 Multiple Signers",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MultipleSignerMsgs(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - DeliverTx EIP712 signed Cosmos Tx with wrong Chain ID",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder, err := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, incorrectChainID, gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - DeliverTx EIP712 signed Cosmos Tx with different gas fees",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder, err := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				txBuilder.SetGasLimit(uint64(300000))
				txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(30))))
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - DeliverTx EIP712 signed Cosmos Tx with empty signature",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder, err := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, incorrectChainID, gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - DeliverTx EIP712 signed Cosmos Tx with invalid sequence",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder, err := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
				suite.Require().NoError(err)
				sigsV2 := signing.SignatureV2{
					PubKey: privKey.PubKey(),
					Data: &signing.SingleSignatureData{
						SignMode: signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
					},
					Sequence: nonce - 1,
				}

				err = txBuilder.SetSignatures(sigsV2)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - DeliverTx EIP712 signed Cosmos Tx with invalid signMode",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder, err := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
				suite.Require().NoError(err)
				sigsV2 := signing.SignatureV2{
					PubKey: privKey.PubKey(),
					Data: &signing.SingleSignatureData{
						SignMode: signing.SignMode_SIGN_MODE_UNSPECIFIED,
					},
					Sequence: nonce,
				}
				err = txBuilder.SetSignatures(sigsV2)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - invalid from",
			func() sdk.Tx {
				msg := evmtypes.NewTx(ethContractCreationTxParams)
				msg.From = addr.Hex()
				tx := suite.CreateTestTx(msg, privKey, 1, false)
				msg = tx.GetMsgs()[0].(*evmtypes.MsgEthereumTx)
				msg.From = addr.Hex()
				return tx
			}, true, false, false,
		},
		{
			"fail - Single-signer EIP-712",
			func() sdk.Tx {
				msg := banktypes.NewMsgSend(
					sdk.AccAddress(privKey.PubKey().Address()),
					addr[:],
					sdk.NewCoins(
						sdk.NewCoin(
							"btc",
							sdkmath.NewInt(1),
						),
					),
				)

				txBuilder := suite.CreateTestSingleSignedTx(
					privKey,
					signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
					msg,
					suite.ctx.ChainID(),
					2000000,
					"EIP-712",
				)

				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - EIP-712 multi-key",
			func() sdk.Tx {
				numKeys := 5
				privKeys, pubKeys := suite.GenerateMultipleKeys(numKeys)
				pk := kmultisig.NewLegacyAminoPubKey(numKeys, pubKeys)

				msg := banktypes.NewMsgSend(
					sdk.AccAddress(pk.Address()),
					addr[:],
					sdk.NewCoins(
						sdk.NewCoin(
							"btc",
							sdkmath.NewInt(1),
						),
					),
				)

				txBuilder := suite.CreateTestSignedMultisigTx(
					privKeys,
					signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
					msg,
					suite.ctx.ChainID(),
					2000000,
					"EIP-712",
				)

				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - Mixed multi-key",
			func() sdk.Tx {
				numKeys := 5
				privKeys, pubKeys := suite.GenerateMultipleKeys(numKeys)
				pk := kmultisig.NewLegacyAminoPubKey(numKeys, pubKeys)

				msg := banktypes.NewMsgSend(
					sdk.AccAddress(pk.Address()),
					addr[:],
					sdk.NewCoins(
						sdk.NewCoin(
							"btc",
							sdkmath.NewInt(1),
						),
					),
				)

				txBuilder := suite.CreateTestSignedMultisigTx(
					privKeys,
					signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
					msg,
					suite.ctx.ChainID(),
					2000000,
					"mixed", // Combine EIP-712 and standard signatures
				)

				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"Fails - Multi-Key with incorrect Chain ID",
			func() sdk.Tx {
				numKeys := 5
				privKeys, pubKeys := suite.GenerateMultipleKeys(numKeys)
				pk := kmultisig.NewLegacyAminoPubKey(numKeys, pubKeys)

				msg := banktypes.NewMsgSend(
					sdk.AccAddress(pk.Address()),
					addr[:],
					sdk.NewCoins(
						sdk.NewCoin(
							"btc",
							sdkmath.NewInt(1),
						),
					),
				)

				txBuilder := suite.CreateTestSignedMultisigTx(
					privKeys,
					signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
					msg,
					incorrectChainID,
					2000000,
					"mixed",
				)

				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"Fails - Multi-Key with incorrect sign mode",
			func() sdk.Tx {
				numKeys := 5
				privKeys, pubKeys := suite.GenerateMultipleKeys(numKeys)
				pk := kmultisig.NewLegacyAminoPubKey(numKeys, pubKeys)

				msg := banktypes.NewMsgSend(
					sdk.AccAddress(pk.Address()),
					addr[:],
					sdk.NewCoins(
						sdk.NewCoin(
							"btc",
							sdkmath.NewInt(1),
						),
					),
				)

				txBuilder := suite.CreateTestSignedMultisigTx(
					privKeys,
					signing.SignMode_SIGN_MODE_DIRECT,
					msg,
					suite.ctx.ChainID(),
					2000000,
					"mixed",
				)

				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"Fails - Multi-Key with too little gas",
			func() sdk.Tx {
				numKeys := 5
				privKeys, pubKeys := suite.GenerateMultipleKeys(numKeys)
				pk := kmultisig.NewLegacyAminoPubKey(numKeys, pubKeys)

				msg := banktypes.NewMsgSend(
					sdk.AccAddress(pk.Address()),
					addr[:],
					sdk.NewCoins(
						sdk.NewCoin(
							"mezo",
							sdkmath.NewInt(1),
						),
					),
				)

				txBuilder := suite.CreateTestSignedMultisigTx(
					privKeys,
					signing.SignMode_SIGN_MODE_DIRECT,
					msg,
					suite.ctx.ChainID(),
					2000,
					"mixed", // Combine EIP-712 and standard signatures
				)

				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"Fails - Multi-Key with different payload than one signed",
			func() sdk.Tx {
				numKeys := 1
				privKeys, pubKeys := suite.GenerateMultipleKeys(numKeys)
				pk := kmultisig.NewLegacyAminoPubKey(numKeys, pubKeys)

				msg := banktypes.NewMsgSend(
					sdk.AccAddress(pk.Address()),
					addr[:],
					sdk.NewCoins(
						sdk.NewCoin(
							"btc",
							sdkmath.NewInt(1),
						),
					),
				)

				txBuilder := suite.CreateTestSignedMultisigTx(
					privKeys,
					signing.SignMode_SIGN_MODE_DIRECT,
					msg,
					suite.ctx.ChainID(),
					2000,
					"EIP-712",
				)

				msg.Amount[0].Amount = sdkmath.NewInt(5)
				err := txBuilder.SetMsgs(msg)
				suite.Require().NoError(err)

				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"Fails - Multi-Key with messages added after signing",
			func() sdk.Tx {
				numKeys := 1
				privKeys, pubKeys := suite.GenerateMultipleKeys(numKeys)
				pk := kmultisig.NewLegacyAminoPubKey(numKeys, pubKeys)

				msg := banktypes.NewMsgSend(
					sdk.AccAddress(pk.Address()),
					addr[:],
					sdk.NewCoins(
						sdk.NewCoin(
							"btc",
							sdkmath.NewInt(1),
						),
					),
				)

				txBuilder := suite.CreateTestSignedMultisigTx(
					privKeys,
					signing.SignMode_SIGN_MODE_DIRECT,
					msg,
					suite.ctx.ChainID(),
					2000,
					"EIP-712",
				)

				// Duplicate
				err := txBuilder.SetMsgs(msg, msg)
				suite.Require().NoError(err)

				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"Fails - Single-Signer EIP-712 with messages added after signing",
			func() sdk.Tx {
				msg := banktypes.NewMsgSend(
					sdk.AccAddress(privKey.PubKey().Address()),
					addr[:],
					sdk.NewCoins(
						sdk.NewCoin(
							"btc",
							sdkmath.NewInt(1),
						),
					),
				)

				txBuilder := suite.CreateTestSingleSignedTx(
					privKey,
					signing.SignMode_SIGN_MODE_DIRECT,
					msg,
					suite.ctx.ChainID(),
					2000,
					"EIP-712",
				)

				err := txBuilder.SetMsgs(msg, msg)
				suite.Require().NoError(err)

				return txBuilder.GetTx()
			}, false, false, false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			setup()

			suite.ctx = suite.ctx.WithIsCheckTx(tc.checkTx).WithIsReCheckTx(tc.reCheckTx)

			// expConsumed := params.TxGasContractCreation + params.TxGas
			_, err := suite.anteHandler(suite.ctx, tc.txFn(), false)

			// suite.Require().Equal(consumed, ctx.GasMeter().GasConsumed())

			if tc.expPass {
				suite.Require().NoError(err)
				// suite.Require().Equal(int(expConsumed), int(suite.ctx.GasMeter().GasConsumed()))
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *AnteTestSuite) TestAnteHandlerWithDynamicTxFee() {
	addr, privKey := utiltx.NewAddrKey()
	to := utiltx.GenerateAddress()

	ethContractCreationTxParams := &evmtypes.EvmTxArgs{
		ChainID:   suite.app.EvmKeeper.ChainID(),
		Nonce:     1,
		Amount:    big.NewInt(10),
		GasLimit:  100000,
		GasFeeCap: big.NewInt(ethparams.InitialBaseFee + 1),
		GasTipCap: big.NewInt(1),
		Accesses:  &types.AccessList{},
	}

	ethTxParams := &evmtypes.EvmTxArgs{
		ChainID:   suite.app.EvmKeeper.ChainID(),
		Nonce:     1,
		Amount:    big.NewInt(10),
		GasLimit:  100000,
		GasFeeCap: big.NewInt(ethparams.InitialBaseFee + 1),
		GasTipCap: big.NewInt(1),
		Accesses:  &types.AccessList{},
		To:        &to,
	}

	testCases := []struct {
		name           string
		txFn           func() sdk.Tx
		enableLondonHF bool
		checkTx        bool
		reCheckTx      bool
		expPass        bool
	}{
		{
			"success - DeliverTx (contract)",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTx(ethContractCreationTxParams)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			true,
			false, false, true,
		},
		{
			"success - CheckTx (contract)",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTx(ethContractCreationTxParams)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			true,
			true, false, true,
		},
		{
			"success - ReCheckTx (contract)",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTx(ethContractCreationTxParams)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			true,
			false, true, true,
		},
		{
			"success - DeliverTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			},
			true,
			false, false, true,
		},
		{
			"success - CheckTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			},
			true,
			true, false, true,
		},
		{
			"success - ReCheckTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			},
			true,
			false, true, true,
		},
		{
			"success - CheckTx (cosmos tx not signed)",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			},
			true,
			false, true, true,
		},
		{
			"fail - CheckTx (cosmos tx is not valid)",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)
				// bigger than MaxGasWanted
				txBuilder.SetGasLimit(uint64(1 << 63))
				return txBuilder.GetTx()
			},
			true,
			true, false, false,
		},
		{
			"fail - CheckTx (memo too long)",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)
				txBuilder.SetMemo(strings.Repeat("*", 257))
				return txBuilder.GetTx()
			},
			true,
			true, false, false,
		},
		{
			"fail - DynamicFeeTx without london hark fork",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTx(ethContractCreationTxParams)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			false,
			false, false, false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.enableFeemarket = true
			suite.enableLondonHF = tc.enableLondonHF
			suite.SetupTest() // reset

			acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
			suite.Require().NoError(acc.SetSequence(1))
			suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

			suite.ctx = suite.ctx.WithIsCheckTx(tc.checkTx).WithIsReCheckTx(tc.reCheckTx)
			err := suite.app.EvmKeeper.SetBalance(suite.ctx, addr, big.NewInt((ethparams.InitialBaseFee+10)*100000))
			suite.Require().NoError(err)

			_, err = suite.anteHandler(suite.ctx, tc.txFn(), false)
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
	suite.enableFeemarket = false
	suite.enableLondonHF = true
}

func (suite *AnteTestSuite) TestAnteHandlerWithParams() {
	addr, privKey := utiltx.NewAddrKey()
	to := utiltx.GenerateAddress()

	ethContractCreationTxParams := &evmtypes.EvmTxArgs{
		ChainID:   suite.app.EvmKeeper.ChainID(),
		Nonce:     1,
		Amount:    big.NewInt(10),
		GasLimit:  100000,
		GasFeeCap: big.NewInt(ethparams.InitialBaseFee + 1),
		GasTipCap: big.NewInt(1),
		Accesses:  &types.AccessList{},
	}

	ethTxParams := &evmtypes.EvmTxArgs{
		ChainID:   suite.app.EvmKeeper.ChainID(),
		Nonce:     1,
		Amount:    big.NewInt(10),
		GasLimit:  100000,
		GasFeeCap: big.NewInt(ethparams.InitialBaseFee + 1),
		GasTipCap: big.NewInt(1),
		Accesses:  &types.AccessList{},
		To:        &to,
	}

	testCases := []struct {
		name         string
		txFn         func() sdk.Tx
		enableCall   bool
		enableCreate bool
		expErr       error
	}{
		{
			"fail - Contract Creation Disabled",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTx(ethContractCreationTxParams)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			true, false,
			evmtypes.ErrCreateDisabled,
		},
		{
			"success - Contract Creation Enabled",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTx(ethContractCreationTxParams)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			true, true,
			nil,
		},
		{
			"fail - EVM Call Disabled",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			},
			false, true,
			evmtypes.ErrCallDisabled,
		},
		{
			"success - EVM Call Enabled",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			},
			true, true,
			nil,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.evmParamsOption = func(params *evmtypes.Params) {
				params.EnableCall = tc.enableCall
				params.EnableCreate = tc.enableCreate
			}
			suite.SetupTest() // reset

			acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
			suite.Require().NoError(acc.SetSequence(1))
			suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

			suite.ctx = suite.ctx.WithIsCheckTx(true)
			err := suite.app.EvmKeeper.SetBalance(suite.ctx, addr, big.NewInt((ethparams.InitialBaseFee+10)*100000))
			suite.Require().NoError(err)

			_, err = suite.anteHandler(suite.ctx, tc.txFn(), false)
			if tc.expErr == nil {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				suite.Require().True(errors.Is(err, tc.expErr))
			}
		})
	}
	suite.evmParamsOption = nil
}

func (suite *AnteTestSuite) TestAnteWithMulitpleSdkMsgs() {
	clientCtx := suite.clientCtx
	builder, ok := clientCtx.TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		suite.Fail("not a valid ExtensionOptionsTxBuilder")
	}

	option, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	if err != nil {
		suite.Fail("not a valid ExtensionOptionsTxBuilder")
	}

	builder.SetExtensionOptions(option)

	evmTxParams := &evmtypes.EvmTxArgs{
		ChainID:  big.NewInt(1),
		Nonce:    0,
		To:       &common.Address{},
		Amount:   big.NewInt(0),
		GasLimit: 210000000000000,
		GasPrice: big.NewInt(1000000000),
		Input:    []byte{},
	}

	msg := evmtypes.NewTx(evmTxParams)
	msgs := []sdk.Msg{
		msg, msg, // 2 transactions
	}

	_ = builder.SetMsgs(msgs...)
	tx := builder.GetTx().(sdk.Tx)
	_, err = suite.anteHandler(suite.ctx, tx, false)
	if suite.Error(err, "expected error") {
		suite.EqualError(err, "cannot submit more than one transaction at a time: feature not supported")
	}
}
