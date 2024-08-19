package keeper_test

import (
	"encoding/json"
	"math"
	"math/big"
	"time"

	simutils "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/mezo-org/mezod/utils"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	poatypes "github.com/mezo-org/mezod/x/poa/types"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/mezo-org/mezod/app"
	"github.com/mezo-org/mezod/crypto/ethsecp256k1"
	"github.com/mezo-org/mezod/encoding"
	"github.com/mezo-org/mezod/testutil"
	utiltx "github.com/mezo-org/mezod/testutil/tx"
	mezotypes "github.com/mezo-org/mezod/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/mezo-org/mezod/x/feemarket/types"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
)

func (suite *KeeperTestSuite) SetupApp(checkTx bool, chainID string) {
	t := suite.T()
	// account key
	accPriv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.address = common.BytesToAddress(accPriv.PubKey().Address().Bytes())
	suite.signer = utiltx.NewSigner(accPriv)

	// consensus key (must use pure secp256k1 curve due to Tendermint requirements)
	priv := secp256k1.GenPrivKey()
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	header := testutil.NewHeader(
		1, time.Now().UTC(), chainID, suite.consAddress, nil, nil,
	)

	suite.ctx = suite.app.BaseApp.NewContextLegacy(checkTx, header)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.FeeMarketKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	nextAccNumber := suite.app.AccountKeeper.NextAccountNumber(suite.ctx)

	acc := &mezotypes.EthAccount{
		BaseAccount: authtypes.NewBaseAccount(
			suite.address.Bytes(),
			nil,
			nextAccNumber,
			0,
		),
		CodeHash: common.BytesToHash(crypto.Keccak256(nil)).String(),
	}

	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := poatypes.NewValidator(valAddr, priv.PubKey(), poatypes.Description{})
	suite.Require().NoError(err)

	err = suite.app.PoaKeeper.SubmitApplication(
		suite.ctx,
		sdk.AccAddress(validator.GetOperator()),
		validator,
	)
	require.NoError(t, err)

	owner := suite.app.PoaKeeper.GetOwner(suite.ctx)

	err = suite.app.PoaKeeper.ApproveApplication(suite.ctx, owner, validator.GetOperator())
	require.NoError(t, err)

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	suite.ethSigner = ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())
	suite.appCodec = encodingConfig.Codec
	suite.denom = evmtypes.DefaultEVMDenom
}

// Commit commits and starts a new block with an updated context.
func (suite *KeeperTestSuite) Commit() {
	suite.CommitAfter(time.Second * 0)
}

// Commit commits a block at a given time.
func (suite *KeeperTestSuite) CommitAfter(t time.Duration) {
	var err error
	suite.ctx, err = testutil.Commit(suite.ctx, suite.app, t, nil)
	suite.Require().NoError(err)
	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.FeeMarketKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

// setupTestWithContext sets up a test chain with an example Cosmos send msg,
// given a local (validator config) and a global (feemarket param) minGasPrice
func setupTestWithContext(valMinGasPrice string, minGasPrice sdkmath.LegacyDec, baseFee sdkmath.Int) (*ethsecp256k1.PrivKey, banktypes.MsgSend) {
	chainID := utils.TestnetChainID + "-1"
	privKey, msg := setupTest(valMinGasPrice+s.denom, chainID)
	params := types.DefaultParams()
	params.MinGasPrice = minGasPrice
	// Disable base fee recalculation in the x/feemarket BeginBlock hook to
	// not overwrite the value set in this test.
	params.EnableHeight = math.MaxInt64
	err := s.app.FeeMarketKeeper.SetParams(s.ctx, params)
	s.Require().NoError(err)
	s.app.FeeMarketKeeper.SetBaseFee(s.ctx, baseFee.BigInt())
	s.Commit()

	return privKey, msg
}

func setupTest(localMinGasPrices, chainID string) (*ethsecp256k1.PrivKey, banktypes.MsgSend) {
	setupChain(localMinGasPrices, chainID)

	address, privKey := utiltx.NewAccAddressAndKey()
	amount, ok := sdkmath.NewIntFromString("10000000000000000000")
	s.Require().True(ok)
	initBalance := sdk.Coins{sdk.Coin{
		Denom:  s.denom,
		Amount: amount,
	}}
	err := testutil.FundAccount(s.ctx, s.app.BankKeeper, address, initBalance)
	s.Require().NoError(err)

	msg := banktypes.MsgSend{
		FromAddress: address.String(),
		ToAddress:   address.String(),
		Amount: sdk.Coins{sdk.Coin{
			Denom:  s.denom,
			Amount: sdkmath.NewInt(10000),
		}},
	}
	s.Commit()
	return privKey, msg
}

func setupChain(localMinGasPricesStr, chainID string) {
	// Initialize the app, so we can use SetMinGasPrices to set the
	// validator-specific min-gas-prices setting
	db := dbm.NewMemDB()
	newapp := app.NewMezo(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		app.DefaultNodeHome,
		5,
		encoding.MakeConfig(app.ModuleBasics),
		simutils.NewAppOptionsWithFlagHome(app.DefaultNodeHome),
		baseapp.SetChainID(chainID),
		baseapp.SetMinGasPrices(localMinGasPricesStr),
	)

	genesisState := app.NewTestGenesisState(newapp.AppCodec())
	genesisState[types.ModuleName] = newapp.AppCodec().MustMarshalJSON(types.DefaultGenesisState())

	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	s.Require().NoError(err)

	// Initialize the chain
	_, err = newapp.InitChain(
		&abci.RequestInitChain{
			ChainId:         chainID,
			Validators:      []abci.ValidatorUpdate{},
			AppStateBytes:   stateBytes,
			ConsensusParams: app.DefaultConsensusParams,
		},
	)
	s.Require().NoError(err)

	s.app = newapp
	s.SetupApp(false, chainID)
}

func getNonce(addressBytes []byte) uint64 {
	return s.app.EvmKeeper.GetNonce(
		s.ctx,
		common.BytesToAddress(addressBytes),
	)
}

func buildEthTx(
	priv *ethsecp256k1.PrivKey,
	to *common.Address,
	gasPrice *big.Int,
	gasFeeCap *big.Int,
	gasTipCap *big.Int,
	accesses *ethtypes.AccessList,
) *evmtypes.MsgEthereumTx {
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := getNonce(from.Bytes())
	data := make([]byte, 0)
	gasLimit := uint64(100000)
	ethTxParams := &evmtypes.EvmTxArgs{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        to,
		GasLimit:  gasLimit,
		GasPrice:  gasPrice,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Input:     data,
		Accesses:  accesses,
	}
	msgEthereumTx := evmtypes.NewTx(ethTxParams)
	msgEthereumTx.From = from.String()
	return msgEthereumTx
}
