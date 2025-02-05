package keeper_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

func (suite *KeeperTestSuite) TestExecuteContractCall() {
	// Deploy a test ERC20 contract with 1000 tokens of supply.
	contract := suite.DeployTestContract(suite.T(), suite.address, big.NewInt(1000))

	// Prepare a transfer of 800 tokens to a recipient.
	transferRecipient := common.HexToAddress("0x2B66aeB8C31619FE9d06A64772Df147878F69054")
	transferData, err := evmtypes.ERC20Contract.ABI.Pack("transfer", transferRecipient, big.NewInt(800))
	suite.Require().NoError(err)

	// Execute the transfer call.
	_, err = suite.app.EvmKeeper.ExecuteContractCall(suite.ctx, &testCall{
		from: suite.address,
		to:   &contract,
		data: transferData,
	})
	suite.Require().NoError(err)

	balanceOfCallFn := func(address common.Address) *testCall {
		data, err := evmtypes.ERC20Contract.ABI.Pack("balanceOf", address)
		suite.Require().NoError(err)

		return &testCall{
			from: suite.address,
			to:   &contract,
			data: data,
		}
	}

	// Check the balance of the sender.
	response, err := suite.app.EvmKeeper.ExecuteContractCall(suite.ctx, balanceOfCallFn(suite.address))
	suite.Require().NoError(err)
	suite.Require().Equal(common.LeftPadBytes(big.NewInt(200).Bytes(), 32), response.Ret)

	// Check the balance of the recipient.
	response, err = suite.app.EvmKeeper.ExecuteContractCall(suite.ctx, balanceOfCallFn(transferRecipient))
	suite.Require().NoError(err)
	suite.Require().Equal(common.LeftPadBytes(big.NewInt(800).Bytes(), 32), response.Ret)
}

type testCall struct {
	from common.Address
	to   *common.Address
	data []byte
}

func (tc *testCall) From() common.Address {
	return tc.from
}

func (tc *testCall) To() *common.Address {
	return tc.to
}

func (tc *testCall) Data() []byte {
	return tc.data
}
