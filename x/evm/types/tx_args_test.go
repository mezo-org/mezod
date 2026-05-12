package types_test

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"

	"github.com/mezo-org/mezod/x/evm/types"
)

func (suite *TxDataTestSuite) TestTxArgsString() {
	testCases := []struct {
		name           string
		txArgs         types.TransactionArgs
		expectedString string
	}{
		{
			"empty tx args",
			types.TransactionArgs{},
			"TransactionArgs{From:<nil>, To:<nil>, Gas:<nil>, Nonce:<nil>, Data:<nil>, Input:<nil>, AccessList:<nil>, AuthorizationList:[]}",
		},
		{
			"tx args with fields",
			types.TransactionArgs{
				From:       &suite.addr,
				To:         &suite.addr,
				Gas:        &suite.hexUint64,
				Nonce:      &suite.hexUint64,
				Input:      &suite.hexInputBytes,
				Data:       &suite.hexDataBytes,
				AccessList: &ethtypes.AccessList{},
			},
			fmt.Sprintf("TransactionArgs{From:%v, To:%v, Gas:%v, Nonce:%v, Data:%v, Input:%v, AccessList:%v, AuthorizationList:%v}",
				&suite.addr,
				&suite.addr,
				&suite.hexUint64,
				&suite.hexUint64,
				&suite.hexDataBytes,
				&suite.hexInputBytes,
				&ethtypes.AccessList{},
				[]string{}),
		},
		{
			"tx args with auth list",
			types.TransactionArgs{
				AuthorizationList: []ethtypes.SetCodeAuthorization{
					{ChainID: *uint256.NewInt(31611), Address: suite.addr, Nonce: 7},
					{ChainID: *uint256.NewInt(31611), Address: suite.addr, Nonce: 8},
				},
			},
			fmt.Sprintf(
				"TransactionArgs{From:<nil>, To:<nil>, Gas:<nil>, Nonce:<nil>, Data:<nil>, Input:<nil>, AccessList:<nil>, AuthorizationList:[(chainId=31611 address=%s nonce=7) (chainId=31611 address=%s nonce=8)]}",
				suite.addr.Hex(), suite.addr.Hex(),
			),
		},
	}
	for _, tc := range testCases {
		outputString := tc.txArgs.String()
		suite.Require().Equal(outputString, tc.expectedString)
	}
}

func (suite *TxDataTestSuite) TestConvertTxArgsEthTx() {
	testCases := []struct {
		name   string
		txArgs types.TransactionArgs
	}{
		{
			"empty tx args",
			types.TransactionArgs{},
		},
		{
			"no nil args",
			types.TransactionArgs{
				From:                 &suite.addr,
				To:                   &suite.addr,
				Gas:                  &suite.hexUint64,
				GasPrice:             &suite.hexBigInt,
				MaxFeePerGas:         &suite.hexBigInt,
				MaxPriorityFeePerGas: &suite.hexBigInt,
				Value:                &suite.hexBigInt,
				Nonce:                &suite.hexUint64,
				Data:                 &suite.hexDataBytes,
				Input:                &suite.hexInputBytes,
				AccessList:           &ethtypes.AccessList{{Address: suite.addr, StorageKeys: []common.Hash{{0}}}},
				ChainID:              &suite.hexBigInt,
			},
		},
		{
			"max fee per gas nil, but access list not nil",
			types.TransactionArgs{
				From:                 &suite.addr,
				To:                   &suite.addr,
				Gas:                  &suite.hexUint64,
				GasPrice:             &suite.hexBigInt,
				MaxFeePerGas:         nil,
				MaxPriorityFeePerGas: &suite.hexBigInt,
				Value:                &suite.hexBigInt,
				Nonce:                &suite.hexUint64,
				Data:                 &suite.hexDataBytes,
				Input:                &suite.hexInputBytes,
				AccessList:           &ethtypes.AccessList{{Address: suite.addr, StorageKeys: []common.Hash{{0}}}},
				ChainID:              &suite.hexBigInt,
			},
		},
	}
	for _, tc := range testCases {
		res := tc.txArgs.ToTransaction()
		suite.Require().NotNil(res)
	}
}

func (suite *TxDataTestSuite) TestToTransactionSetCode() {
	auth := ethtypes.SetCodeAuthorization{
		ChainID: *uint256.NewInt(31611),
		Address: suite.addr,
		Nonce:   42,
		V:       1,
		R:       *uint256.NewInt(7),
		S:       *uint256.NewInt(11),
	}

	// wantAccesses lets callers distinguish a populated AccessList from the
	// empty fallback used when args.AccessList is nil.
	assertSetCodeFields := func(setCodeTx *types.SetCodeTx, wantAccesses types.AccessList) {
		suite.Require().Equal(suite.addr.Hex(), setCodeTx.To)
		suite.Require().NotNil(setCodeTx.ChainID)
		suite.Require().Equal(suite.hexBigInt.ToInt(), setCodeTx.ChainID.BigInt())
		suite.Require().Equal(uint64(suite.hexUint64), setCodeTx.Nonce)
		suite.Require().Equal(uint64(suite.hexUint64), setCodeTx.GasLimit)
		suite.Require().NotNil(setCodeTx.GasFeeCap)
		suite.Require().Equal(suite.hexBigInt.ToInt(), setCodeTx.GasFeeCap.BigInt())
		suite.Require().NotNil(setCodeTx.GasTipCap)
		suite.Require().Equal(suite.hexBigInt.ToInt(), setCodeTx.GasTipCap.BigInt())
		suite.Require().NotNil(setCodeTx.Amount)
		suite.Require().Equal(suite.hexBigInt.ToInt(), setCodeTx.Amount.BigInt())
		// Input is preferred over Data when both are set (see GetData).
		suite.Require().Equal([]byte(suite.hexInputBytes), setCodeTx.Data)
		suite.Require().Equal(wantAccesses, setCodeTx.Accesses)
		suite.Require().Len(setCodeTx.AuthList, 1)
		suite.Require().Equal(auth, setCodeTx.AuthList[0].ToEthAuthorization())
	}

	suite.Run("with auth list, builds SetCodeTx", func() {
		accessList := ethtypes.AccessList{{Address: suite.addr, StorageKeys: []common.Hash{{0}}}}
		args := types.TransactionArgs{
			From:                 &suite.addr,
			To:                   &suite.addr,
			Gas:                  &suite.hexUint64,
			MaxFeePerGas:         &suite.hexBigInt,
			MaxPriorityFeePerGas: &suite.hexBigInt,
			Value:                &suite.hexBigInt,
			Nonce:                &suite.hexUint64,
			Data:                 &suite.hexDataBytes,
			Input:                &suite.hexInputBytes,
			AccessList:           &accessList,
			AuthorizationList:    []ethtypes.SetCodeAuthorization{auth},
			ChainID:              &suite.hexBigInt,
		}

		msg := args.ToTransaction()
		suite.Require().NotNil(msg)

		txData, err := types.UnpackTxData(msg.Data)
		suite.Require().NoError(err)

		setCodeTx, ok := txData.(*types.SetCodeTx)
		suite.Require().True(ok, "expected *SetCodeTx, got %T", txData)
		assertSetCodeFields(setCodeTx, types.NewAccessList(&accessList))

		// Round-trip through AsTransaction() must succeed and yield a
		// non-empty hash with the SetCode type.
		ethTx := msg.AsTransaction()
		suite.Require().Equal(uint8(ethtypes.SetCodeTxType), ethTx.Type())
		suite.Require().NotEmpty(ethTx.Hash().Hex())
	})

	suite.Run("nil access list, builds SetCodeTx with empty Accesses", func() {
		args := types.TransactionArgs{
			From:                 &suite.addr,
			To:                   &suite.addr,
			Gas:                  &suite.hexUint64,
			MaxFeePerGas:         &suite.hexBigInt,
			MaxPriorityFeePerGas: &suite.hexBigInt,
			Value:                &suite.hexBigInt,
			Nonce:                &suite.hexUint64,
			Data:                 &suite.hexDataBytes,
			Input:                &suite.hexInputBytes,
			AccessList:           nil,
			AuthorizationList:    []ethtypes.SetCodeAuthorization{auth},
			ChainID:              &suite.hexBigInt,
		}

		msg := args.ToTransaction()
		suite.Require().NotNil(msg)

		txData, err := types.UnpackTxData(msg.Data)
		suite.Require().NoError(err)

		setCodeTx, ok := txData.(*types.SetCodeTx)
		suite.Require().True(ok, "expected *SetCodeTx, got %T", txData)
		// When args.AccessList is nil the branch must fall back to an
		// empty (non-nil) AccessList{}, not propagate nil.
		assertSetCodeFields(setCodeTx, types.AccessList{})
		suite.Require().NotNil(setCodeTx.Accesses)

		// Round-trip through AsTransaction() must not panic and must
		// produce a non-empty hash.
		ethTx := msg.AsTransaction()
		suite.Require().Equal(uint8(ethtypes.SetCodeTxType), ethTx.Type())
		suite.Require().NotEmpty(ethTx.Hash().Hex())
	})

	suite.Run("nil auth list, builds DynamicFeeTx", func() {
		args := types.TransactionArgs{
			From:                 &suite.addr,
			To:                   &suite.addr,
			Gas:                  &suite.hexUint64,
			MaxFeePerGas:         &suite.hexBigInt,
			MaxPriorityFeePerGas: &suite.hexBigInt,
			Value:                &suite.hexBigInt,
			Nonce:                &suite.hexUint64,
			Data:                 &suite.hexDataBytes,
			Input:                &suite.hexInputBytes,
			AccessList:           &ethtypes.AccessList{},
			AuthorizationList:    nil,
			ChainID:              &suite.hexBigInt,
		}

		msg := args.ToTransaction()
		suite.Require().NotNil(msg)

		txData, err := types.UnpackTxData(msg.Data)
		suite.Require().NoError(err)

		_, ok := txData.(*types.DynamicFeeTx)
		suite.Require().True(ok, "expected *DynamicFeeTx, got %T", txData)
	})

	suite.Run("empty non-nil auth list, builds SetCodeTx", func() {
		args := types.TransactionArgs{
			From:                 &suite.addr,
			To:                   &suite.addr,
			Gas:                  &suite.hexUint64,
			MaxFeePerGas:         &suite.hexBigInt,
			MaxPriorityFeePerGas: &suite.hexBigInt,
			Value:                &suite.hexBigInt,
			Nonce:                &suite.hexUint64,
			Data:                 &suite.hexDataBytes,
			Input:                &suite.hexInputBytes,
			AccessList:           &ethtypes.AccessList{},
			AuthorizationList:    []ethtypes.SetCodeAuthorization{},
			ChainID:              &suite.hexBigInt,
		}

		msg := args.ToTransaction()
		suite.Require().NotNil(msg)

		txData, err := types.UnpackTxData(msg.Data)
		suite.Require().NoError(err)

		setCodeTx, ok := txData.(*types.SetCodeTx)
		suite.Require().True(ok, "expected *SetCodeTx, got %T", txData)
		suite.Require().Empty(setCodeTx.AuthList)
	})

	suite.Run("gasPrice + auth list, falls through to LegacyTx", func() {
		auth := ethtypes.SetCodeAuthorization{
			ChainID: *uint256.NewInt(31611),
			Address: suite.addr,
			Nonce:   1,
		}
		args := types.TransactionArgs{
			From:              &suite.addr,
			To:                &suite.addr,
			Gas:               &suite.hexUint64,
			GasPrice:          &suite.hexBigInt,
			Value:             &suite.hexBigInt,
			Nonce:             &suite.hexUint64,
			Data:              &suite.hexDataBytes,
			Input:             &suite.hexInputBytes,
			AuthorizationList: []ethtypes.SetCodeAuthorization{auth},
			ChainID:           &suite.hexBigInt,
		}

		msg := args.ToTransaction()
		suite.Require().NotNil(msg)

		txData, err := types.UnpackTxData(msg.Data)
		suite.Require().NoError(err)

		_, ok := txData.(*types.LegacyTx)
		suite.Require().True(ok, "expected *LegacyTx, got %T", txData)
		suite.Require().NotEmpty(msg.AsTransaction().Hash().Hex())
	})
}

func (suite *TxDataTestSuite) TestToMessageEVM() {
	testCases := []struct {
		name         string
		txArgs       types.TransactionArgs
		globalGasCap uint64
		baseFee      *big.Int
		expError     bool
	}{
		{
			"empty tx args",
			types.TransactionArgs{},
			uint64(0),
			nil,
			false,
		},
		{
			"specify gasPrice and (maxFeePerGas or maxPriorityFeePerGas)",
			types.TransactionArgs{
				From:                 &suite.addr,
				To:                   &suite.addr,
				Gas:                  &suite.hexUint64,
				GasPrice:             &suite.hexBigInt,
				MaxFeePerGas:         &suite.hexBigInt,
				MaxPriorityFeePerGas: &suite.hexBigInt,
				Value:                &suite.hexBigInt,
				Nonce:                &suite.hexUint64,
				Data:                 &suite.hexDataBytes,
				Input:                &suite.hexInputBytes,
				AccessList:           &ethtypes.AccessList{{Address: suite.addr, StorageKeys: []common.Hash{{0}}}},
				ChainID:              &suite.hexBigInt,
			},
			uint64(0),
			nil,
			true,
		},
		{
			"non-1559 execution, zero gas cap",
			types.TransactionArgs{
				From:                 &suite.addr,
				To:                   &suite.addr,
				Gas:                  &suite.hexUint64,
				GasPrice:             &suite.hexBigInt,
				MaxFeePerGas:         nil,
				MaxPriorityFeePerGas: nil,
				Value:                &suite.hexBigInt,
				Nonce:                &suite.hexUint64,
				Data:                 &suite.hexDataBytes,
				Input:                &suite.hexInputBytes,
				AccessList:           &ethtypes.AccessList{{Address: suite.addr, StorageKeys: []common.Hash{{0}}}},
				ChainID:              &suite.hexBigInt,
			},
			uint64(0),
			nil,
			false,
		},
		{
			"non-1559 execution, nonzero gas cap",
			types.TransactionArgs{
				From:                 &suite.addr,
				To:                   &suite.addr,
				Gas:                  &suite.hexUint64,
				GasPrice:             &suite.hexBigInt,
				MaxFeePerGas:         nil,
				MaxPriorityFeePerGas: nil,
				Value:                &suite.hexBigInt,
				Nonce:                &suite.hexUint64,
				Data:                 &suite.hexDataBytes,
				Input:                &suite.hexInputBytes,
				AccessList:           &ethtypes.AccessList{{Address: suite.addr, StorageKeys: []common.Hash{{0}}}},
				ChainID:              &suite.hexBigInt,
			},
			uint64(1),
			nil,
			false,
		},
		{
			"1559-type execution, nil gas price",
			types.TransactionArgs{
				From:                 &suite.addr,
				To:                   &suite.addr,
				Gas:                  &suite.hexUint64,
				GasPrice:             nil,
				MaxFeePerGas:         &suite.hexBigInt,
				MaxPriorityFeePerGas: &suite.hexBigInt,
				Value:                &suite.hexBigInt,
				Nonce:                &suite.hexUint64,
				Data:                 &suite.hexDataBytes,
				Input:                &suite.hexInputBytes,
				AccessList:           &ethtypes.AccessList{{Address: suite.addr, StorageKeys: []common.Hash{{0}}}},
				ChainID:              &suite.hexBigInt,
			},
			uint64(1),
			suite.bigInt,
			false,
		},
		{
			"1559-type execution, non-nil gas price",
			types.TransactionArgs{
				From:                 &suite.addr,
				To:                   &suite.addr,
				Gas:                  &suite.hexUint64,
				GasPrice:             &suite.hexBigInt,
				MaxFeePerGas:         nil,
				MaxPriorityFeePerGas: nil,
				Value:                &suite.hexBigInt,
				Nonce:                &suite.hexUint64,
				Data:                 &suite.hexDataBytes,
				Input:                &suite.hexInputBytes,
				AccessList:           &ethtypes.AccessList{{Address: suite.addr, StorageKeys: []common.Hash{{0}}}},
				ChainID:              &suite.hexBigInt,
			},
			uint64(1),
			suite.bigInt,
			false,
		},
	}
	for _, tc := range testCases {
		res, err := tc.txArgs.ToMessage(tc.globalGasCap, tc.baseFee)

		if tc.expError {
			suite.Require().NotNil(err)
		} else {
			suite.Require().Nil(err)
			suite.Require().NotNil(res)
		}
	}
}

func (suite *TxDataTestSuite) TestGetFrom() {
	testCases := []struct {
		name       string
		txArgs     types.TransactionArgs
		expAddress common.Address
	}{
		{
			"empty from field",
			types.TransactionArgs{},
			common.Address{},
		},
		{
			"non-empty from field",
			types.TransactionArgs{
				From: &suite.addr,
			},
			suite.addr,
		},
	}
	for _, tc := range testCases {
		retrievedAddress := tc.txArgs.GetFrom()
		suite.Require().Equal(retrievedAddress, tc.expAddress)
	}
}

func (suite *TxDataTestSuite) TestAuthorizationListJSONRoundTrip() {
	testCases := []struct {
		name     string
		input    string
		expected []ethtypes.SetCodeAuthorization
		isNil    bool
	}{
		{
			"empty list must survive marshal/unmarshal",
			`{"authorizationList":[]}`,
			[]ethtypes.SetCodeAuthorization{},
			false,
		},
		{
			"absent field stays nil",
			`{}`,
			nil,
			true,
		},
	}
	for _, tc := range testCases {
		var args types.TransactionArgs
		err := json.Unmarshal([]byte(tc.input), &args)
		suite.Require().NoError(err, tc.name)
		if tc.isNil {
			suite.Require().Nil(args.AuthorizationList, tc.name)
		} else {
			suite.Require().NotNil(args.AuthorizationList, tc.name)
			suite.Require().Equal(tc.expected, args.AuthorizationList, tc.name)
		}

		bz, err := json.Marshal(&args)
		suite.Require().NoError(err, tc.name)

		var round types.TransactionArgs
		suite.Require().NoError(json.Unmarshal(bz, &round), tc.name)
		if tc.isNil {
			suite.Require().Nil(round.AuthorizationList, tc.name)
		} else {
			suite.Require().NotNil(round.AuthorizationList, tc.name)
			suite.Require().Equal(tc.expected, round.AuthorizationList, tc.name)
		}
	}
}

func (suite *TxDataTestSuite) TestGetData() {
	testCases := []struct {
		name           string
		txArgs         types.TransactionArgs
		expectedOutput []byte
	}{
		{
			"empty input and data fields",
			types.TransactionArgs{
				Data:  nil,
				Input: nil,
			},
			nil,
		},
		{
			"empty input field, non-empty data field",
			types.TransactionArgs{
				Data:  &suite.hexDataBytes,
				Input: nil,
			},
			[]byte("data"),
		},
		{
			"non-empty input and data fields",
			types.TransactionArgs{
				Data:  &suite.hexDataBytes,
				Input: &suite.hexInputBytes,
			},
			[]byte("input"),
		},
	}
	for _, tc := range testCases {
		retrievedData := tc.txArgs.GetData()
		suite.Require().Equal(retrievedData, tc.expectedOutput)
	}
}
