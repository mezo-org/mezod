package keeper_test

import (
	"math/big"
	"slices"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// runCreate builds, signs, and runs a contract-creation EVM tx with the given
// init code, returning the response (rsp.VmError holds any EVM error).
func (suite *KeeperTestSuite) runCreate(input []byte) *evmtypes.MsgEthereumTxResponse {
	chainID := suite.app.EvmKeeper.ChainID()
	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	var ethTxParams *evmtypes.EvmTxArgs
	if suite.enableFeemarket {
		ethTxParams = &evmtypes.EvmTxArgs{
			ChainID:   chainID,
			Nonce:     nonce,
			GasLimit:  500_000,
			GasFeeCap: suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx),
			GasTipCap: big.NewInt(1),
			Input:     input,
			Accesses:  &ethtypes.AccessList{},
		}
	} else {
		ethTxParams = &evmtypes.EvmTxArgs{
			ChainID:  chainID,
			Nonce:    nonce,
			GasLimit: 500_000,
			Input:    input,
		}
	}

	tx := evmtypes.NewTx(ethTxParams)
	tx.From = suite.address.Hex()
	err := tx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	suite.Require().NoError(err)

	rsp, err := suite.app.EvmKeeper.EthereumTx(suite.ctx, tx)
	suite.Require().NoError(err)
	return rsp
}

// TestSelfdestructDisabledByEIP90000 checks that EIP 90000 turns SELFDESTRUCT
// into an invalid opcode. It is on by default (genesis params); removing it
// from the params lets the opcode run again.
func (suite *KeeperTestSuite) TestSelfdestructDisabledByEIP90000() {
	// PUSH1 0x00; SELFDESTRUCT -- self-destruct to the zero address.
	initCode := []byte{0x60, 0x00, 0xff}

	// Default params carry EIP 90000, so SELFDESTRUCT is an invalid opcode.
	suite.Require().Contains(
		suite.app.EvmKeeper.GetParams(suite.ctx).ExtraEIPs,
		evmtypes.SelfdestructDisableEIP,
	)
	rsp := suite.runCreate(initCode)
	suite.Require().Contains(rsp.VmError, "invalid opcode")

	// Remove EIP 90000 -> the opcode runs and the tx succeeds.
	params := suite.app.EvmKeeper.GetParams(suite.ctx)
	params.ExtraEIPs = slices.DeleteFunc(params.ExtraEIPs, func(eip int64) bool {
		return eip == evmtypes.SelfdestructDisableEIP
	})
	suite.Require().NoError(suite.app.EvmKeeper.SetParams(suite.ctx, params))

	rsp = suite.runCreate(initCode)
	suite.Require().Empty(rsp.VmError)
}
