package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

func (suite *KeeperTestSuite) TestEndBlock() {
	em := suite.ctx.EventManager()
	suite.Require().Equal(0, len(em.Events()))

	err := suite.app.EvmKeeper.EndBlock(suite.ctx)
	suite.Require().NoError(err)

	// should emit 1 EventTypeBlockBloom event on EndBlock
	suite.Require().Equal(1, len(em.Events()))
	suite.Require().Equal(evmtypes.EventTypeBlockBloom, em.Events()[0].Type)
}

func (suite *KeeperTestSuite) TestEndBlockWithZeroBalance() {
	feeCollectorAddr := authtypes.NewModuleAddress(authtypes.FeeCollectorName)
	balance := suite.app.BankKeeper.GetBalance(suite.ctx, feeCollectorAddr, suite.app.EvmKeeper.GetParams(suite.ctx).EvmDenom)

	// Set balance to zero for FeeCollector
	zeroAddress := common.Address{}.Bytes()
	err := suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, authtypes.FeeCollectorName, zeroAddress, sdk.NewCoins(balance))
	suite.Require().NoError(err)

	err = suite.app.EvmKeeper.EndBlock(suite.ctx)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestEndBlockWithZeroChainFeeSplitterAddress() {
	// Configure chain fee splitter address to zero address
	params := suite.app.EvmKeeper.GetParams(suite.ctx)
	params.ChainFeeSplitterAddress = common.Address{}.Hex() // zero address
	suite.app.EvmKeeper.SetParams(suite.ctx, params)

	coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(42))

	// Set balance to non-zero
	err := addBalanceToFeeCollector(suite, coinAmount)
	suite.Require().NoError(err)

	err = suite.app.EvmKeeper.EndBlock(suite.ctx)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestEndBlockSuccessfulTransfer() {
	// Set a valid chain fee splitter address
	targetAddress := common.HexToAddress("0x1234567890AbcdEF1234567890aBcdef12345678")
	suite.app.EvmKeeper.SetParams(suite.ctx, evmtypes.Params{
		EvmDenom:                suite.app.EvmKeeper.GetParams(suite.ctx).EvmDenom,
		ChainFeeSplitterAddress: targetAddress.Hex(),
	})

	coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(42))

	// Set balance to non-zero
	err := addBalanceToFeeCollector(suite, coinAmount)
	suite.Require().NoError(err)

	err = suite.app.EvmKeeper.EndBlock(suite.ctx)
	suite.Require().NoError(err)

	// Check fee collector balance
	feeCollectorBalance := suite.app.BankKeeper.GetBalance(suite.ctx, suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), suite.app.EvmKeeper.GetParams(suite.ctx).EvmDenom)
	suite.Require().Equal(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(0)), feeCollectorBalance)

	// Check target address balance
	sdkValidAddress := sdk.AccAddress(targetAddress.Bytes())
	chainFeeSplitterBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sdkValidAddress, suite.app.EvmKeeper.GetParams(suite.ctx).EvmDenom)
	suite.Require().Equal(coinAmount, chainFeeSplitterBalance)
}

func addBalanceToFeeCollector(suite *KeeperTestSuite, coinAmount sdk.Coin) error {
	amount := sdk.NewCoins(coinAmount)
	err := suite.app.BankKeeper.MintCoins(suite.ctx, evmtypes.ModuleName, amount)
	suite.Require().NoError(err)
	// Send coins to FeeCollector
	suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, evmtypes.ModuleName, authtypes.FeeCollectorName, amount)

	return err
}
