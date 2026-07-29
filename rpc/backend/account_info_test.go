package backend

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/cometbft/cometbft/libs/bytes"

	tmrpcclient "github.com/cometbft/cometbft/rpc/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	"github.com/mezo-org/mezod/rpc/backend/mocks"
	rpctypes "github.com/mezo-org/mezod/rpc/types"
	utiltx "github.com/mezo-org/mezod/testutil/tx"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

func (suite *BackendTestSuite) TestGetCode() {
	blockNr := rpctypes.NewBlockNumber(big.NewInt(1))
	contractCode := []byte("0xef616c92f3cfc9e92dc270d6acff9cea213cecc7020a76ee4395af09bdceb4837a1ebdb5735e11e7d3adb6104e0c3ac55180b4ddf5e54d022cc5e8837f6a4f971b")

	testCases := []struct {
		name          string
		addr          common.Address
		blockNrOrHash rpctypes.BlockNumberOrHash
		registerMock  func(common.Address)
		expPass       bool
		expCode       hexutil.Bytes
	}{
		{
			"fail - BlockHash and BlockNumber are both nil ",
			utiltx.GenerateAddress(),
			rpctypes.BlockNumberOrHash{},
			func(_ common.Address) {},
			false,
			nil,
		},
		{
			"fail - query client errors on getting Code",
			utiltx.GenerateAddress(),
			rpctypes.BlockNumberOrHash{BlockNumber: &blockNr},
			func(addr common.Address) {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterCodeError(queryClient, addr)
			},
			false,
			nil,
		},
		{
			"pass",
			utiltx.GenerateAddress(),
			rpctypes.BlockNumberOrHash{BlockNumber: &blockNr},
			func(addr common.Address) {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterCode(queryClient, addr, contractCode)
			},
			true,
			contractCode,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			tc.registerMock(tc.addr)

			code, err := suite.backend.GetCode(tc.addr, tc.blockNrOrHash)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expCode, code)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestGetProof() {
	blockNrInvalid := rpctypes.NewBlockNumber(big.NewInt(1))
	blockNr := rpctypes.NewBlockNumber(big.NewInt(4))
	address1 := utiltx.GenerateAddress()

	testCases := []struct {
		name          string
		addr          common.Address
		storageKeys   []string
		blockNrOrHash rpctypes.BlockNumberOrHash
		registerMock  func(rpctypes.BlockNumber, common.Address)
		expPass       bool
		expAccRes     *rpctypes.AccountResult
	}{
		{
			"fail - BlockNumeber = 1 (invalidBlockNumber)",
			address1,
			[]string{},
			rpctypes.BlockNumberOrHash{BlockNumber: &blockNrInvalid},
			func(bn rpctypes.BlockNumber, addr common.Address) {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlock(client, bn.Int64(), nil)
				suite.Require().NoError(err)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterAccount(queryClient, addr, blockNrInvalid.Int64())
			},
			false,
			&rpctypes.AccountResult{},
		},
		{
			"fail - Block doesn't exist",
			address1,
			[]string{},
			rpctypes.BlockNumberOrHash{BlockNumber: &blockNrInvalid},
			func(bn rpctypes.BlockNumber, _ common.Address) {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockError(client, bn.Int64())
			},
			false,
			&rpctypes.AccountResult{},
		},
		{
			"pass",
			address1,
			[]string{"0x0"},
			rpctypes.BlockNumberOrHash{BlockNumber: &blockNr},
			func(bn rpctypes.BlockNumber, addr common.Address) {
				suite.backend.ctx = rpctypes.ContextWithHeight(bn.Int64())

				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlock(client, bn.Int64(), nil)
				suite.Require().NoError(err)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterAccount(queryClient, addr, bn.Int64())

				// Use the IAVL height if a valid tendermint height is passed in.
				iavlHeight := bn.Int64()
				RegisterABCIQueryWithOptions(
					client,
					bn.Int64(),
					"store/evm/key",
					evmtypes.StateKey(address1, common.HexToHash("0x0").Bytes()),
					tmrpcclient.ABCIQueryOptions{Height: iavlHeight, Prove: true},
				)
				RegisterABCIQueryWithOptions(
					client,
					bn.Int64(),
					"store/acc/key",
					bytes.HexBytes(
						append(
							authtypes.AddressStoreKeyPrefix,
							sdk.AccAddress(address1.Bytes())...,
						),
					),
					tmrpcclient.ABCIQueryOptions{Height: iavlHeight, Prove: true},
				)
			},
			true,
			&rpctypes.AccountResult{
				Address:      address1,
				AccountProof: []string{""},
				Balance:      (*hexutil.Big)(big.NewInt(0)),
				CodeHash:     common.HexToHash(""),
				Nonce:        0x0,
				StorageHash:  common.Hash{},
				StorageProof: []rpctypes.StorageResult{
					{
						Key:   "0x0",
						Value: (*hexutil.Big)(big.NewInt(2)),
						Proof: []string{""},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			tc.registerMock(*tc.blockNrOrHash.BlockNumber, tc.addr)

			accRes, err := suite.backend.GetProof(tc.addr, tc.storageKeys, tc.blockNrOrHash)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expAccRes, accRes)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// registerGetProofMocks wires up everything GetProof queries for a successful
// call: the block lookup, the EVM account, one proven store query per given
// storage key, and the account proof. Keys must be given deduplicated, so that
// the number of registered store queries states how many are expected.
func (suite *BackendTestSuite) registerGetProofMocks(
	bn rpctypes.BlockNumber,
	addr common.Address,
	storageKeys []string,
) {
	suite.backend.ctx = rpctypes.ContextWithHeight(bn.Int64())

	client := suite.backend.clientCtx.Client.(*mocks.Client)
	_, err := RegisterBlock(client, bn.Int64(), nil)
	suite.Require().NoError(err)

	queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
	RegisterAccount(queryClient, addr, bn.Int64())

	opts := tmrpcclient.ABCIQueryOptions{Height: bn.Int64(), Prove: true}

	for _, key := range storageKeys {
		RegisterABCIQueryWithOptions(
			client,
			bn.Int64(),
			"store/evm/key",
			evmtypes.StateKey(addr, common.HexToHash(key).Bytes()),
			opts,
		)
	}

	RegisterABCIQueryWithOptions(
		client,
		bn.Int64(),
		"store/acc/key",
		bytes.HexBytes(
			append(
				authtypes.AddressStoreKeyPrefix,
				sdk.AccAddress(addr.Bytes())...,
			),
		),
		opts,
	)
}

func (suite *BackendTestSuite) TestGetProofStorageKeysCap() {
	blockNr := rpctypes.NewBlockNumber(big.NewInt(4))
	address := utiltx.GenerateAddress()

	testCases := []struct {
		name        string
		keysCap     int32
		storageKeys []string
		expPass     bool
	}{
		{
			"pass - fewer storage keys than the cap",
			3,
			[]string{"0x1", "0x2"},
			true,
		},
		{
			"pass - as many storage keys as the cap",
			2,
			[]string{"0x1", "0x2"},
			true,
		},
		{
			"pass - a cap of zero accepts any number of storage keys",
			0,
			[]string{"0x1", "0x2", "0x3"},
			true,
		},
		{
			"fail - one storage key more than the cap",
			2,
			[]string{"0x1", "0x2", "0x3"},
			false,
		},
		{
			// The cap counts the keys as given, so that it can be applied before
			// any store is touched, even though repeats cost a single query.
			"fail - repeated storage keys still count towards the cap",
			2,
			[]string{"0x1", "0x1", "0x1"},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			suite.backend.cfg.JSONRPC.GetProofStorageKeysCap = tc.keysCap

			// Nothing is registered for the failing case on purpose: going over
			// the cap must be reported before any store is queried.
			if tc.expPass {
				suite.registerGetProofMocks(blockNr, address, tc.storageKeys)
			}

			accRes, err := suite.backend.GetProof(
				address,
				tc.storageKeys,
				rpctypes.BlockNumberOrHash{BlockNumber: &blockNr},
			)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Len(accRes.StorageProof, len(tc.storageKeys))
			} else {
				suite.Require().EqualError(err, "max number of storage keys exceeded (max allowed 2)")
				suite.Require().Nil(accRes)
			}
		})
	}
}

func (suite *BackendTestSuite) TestGetProofInvalidStorageKeys() {
	blockNr := rpctypes.NewBlockNumber(big.NewInt(4))
	address := utiltx.GenerateAddress()

	testCases := []struct {
		name        string
		storageKeys []string
		expErr      string
	}{
		{
			"fail - storage key is not hex",
			[]string{"0xzz"},
			"invalid storage key at index 0: hex string invalid",
		},
		{
			"fail - storage key is longer than 32 bytes",
			[]string{"0x1", "0x" + strings.Repeat("ff", 33)},
			"invalid storage key at index 1: hex string too long, want at most 32 bytes",
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()

			// No mocks are registered: a malformed key must be reported before
			// any store is queried, so any query here fails the test.
			accRes, err := suite.backend.GetProof(
				address,
				tc.storageKeys,
				rpctypes.BlockNumberOrHash{BlockNumber: &blockNr},
			)

			suite.Require().EqualError(err, tc.expErr)
			suite.Require().Nil(accRes)
		})
	}
}

func (suite *BackendTestSuite) TestGetProofRepeatedStorageKeys() {
	blockNr := rpctypes.NewBlockNumber(big.NewInt(4))
	address := utiltx.GenerateAddress()

	// Three spellings of the same storage key, so one store query is registered.
	suite.registerGetProofMocks(blockNr, address, []string{"0x1"})

	accRes, err := suite.backend.GetProof(
		address,
		[]string{
			"0x1",
			"0x01",
			"0x0000000000000000000000000000000000000000000000000000000000000001",
		},
		rpctypes.BlockNumberOrHash{BlockNumber: &blockNr},
	)
	suite.Require().NoError(err)

	// One proven query for the repeated storage key plus one for the account.
	suite.backend.clientCtx.Client.(*mocks.Client).
		AssertNumberOfCalls(suite.T(), "ABCIQueryWithOptions", 2)

	// Every position is answered, and each key is echoed as it was asked for.
	suite.Require().Len(accRes.StorageProof, 3)
	suite.Require().Equal("0x1", accRes.StorageProof[0].Key)
	suite.Require().Equal("0x01", accRes.StorageProof[1].Key)
	suite.Require().Equal(
		"0x0000000000000000000000000000000000000000000000000000000000000001",
		accRes.StorageProof[2].Key,
	)

	for _, storageProof := range accRes.StorageProof {
		suite.Require().Equal(accRes.StorageProof[0].Value, storageProof.Value)
		suite.Require().Equal(accRes.StorageProof[0].Proof, storageProof.Proof)
	}
}

func TestDecodeStorageKey(t *testing.T) {
	testCases := []struct {
		name   string
		key    string
		expKey common.Hash
		expErr string
	}{
		{
			"a full 32-byte key is taken as it is",
			"0x00000000000000000000000000000000000000000000000000000000000000ff",
			common.HexToHash("0xff"),
			"",
		},
		{
			"a key shorter than 32 bytes is left-padded",
			"0x0102",
			common.HexToHash("0x0102"),
			"",
		},
		{
			"an odd number of hex digits is left-padded",
			"0x1",
			common.HexToHash("0x01"),
			"",
		},
		{
			"the 0X prefix is accepted",
			"0X0102",
			common.HexToHash("0x0102"),
			"",
		},
		{
			"a key without a prefix is accepted",
			"0102",
			common.HexToHash("0x0102"),
			"",
		},
		{
			"upper-case hex digits are accepted",
			"0xABcd",
			common.HexToHash("0xabcd"),
			"",
		},
		{
			"an empty key denotes the zero key",
			"",
			common.Hash{},
			"",
		},
		{
			"a bare prefix denotes the zero key",
			"0x",
			common.Hash{},
			"",
		},
		{
			"a key longer than 32 bytes is rejected",
			"0x" + strings.Repeat("ff", 33),
			common.Hash{},
			"hex string too long, want at most 32 bytes",
		},
		{
			"a key that is not hex is rejected",
			"0xzz",
			common.Hash{},
			"hex string invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key, err := decodeStorageKey(tc.key)

			if tc.expErr != "" {
				require.EqualError(t, err, tc.expErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.expKey, key)
		})
	}
}

func (suite *BackendTestSuite) TestGetStorageAt() {
	blockNr := rpctypes.NewBlockNumber(big.NewInt(1))

	testCases := []struct {
		name          string
		addr          common.Address
		key           string
		blockNrOrHash rpctypes.BlockNumberOrHash
		registerMock  func(common.Address, string, string)
		expPass       bool
		expStorage    hexutil.Bytes
	}{
		{
			"fail - BlockHash and BlockNumber are both nil",
			utiltx.GenerateAddress(),
			"0x0",
			rpctypes.BlockNumberOrHash{},
			func(_ common.Address, _ string, _ string) {},
			false,
			nil,
		},
		{
			"fail - query client errors on getting Storage",
			utiltx.GenerateAddress(),
			"0x0",
			rpctypes.BlockNumberOrHash{BlockNumber: &blockNr},
			func(addr common.Address, key string, _ string) {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterStorageAtError(queryClient, addr, key)
			},
			false,
			nil,
		},
		{
			"pass",
			utiltx.GenerateAddress(),
			"0x0",
			rpctypes.BlockNumberOrHash{BlockNumber: &blockNr},
			func(addr common.Address, key string, storage string) {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterStorageAt(queryClient, addr, key, storage)
			},
			true,
			hexutil.Bytes{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			tc.registerMock(tc.addr, tc.key, tc.expStorage.String())

			storage, err := suite.backend.GetStorageAt(tc.addr, tc.key, tc.blockNrOrHash)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expStorage, storage)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestGetBalance() {
	blockNr := rpctypes.NewBlockNumber(big.NewInt(1))

	testCases := []struct {
		name          string
		addr          common.Address
		blockNrOrHash rpctypes.BlockNumberOrHash
		registerMock  func(rpctypes.BlockNumber, common.Address)
		expPass       bool
		expBalance    *hexutil.Big
	}{
		{
			"fail - BlockHash and BlockNumber are both nil",
			utiltx.GenerateAddress(),
			rpctypes.BlockNumberOrHash{},
			func(_ rpctypes.BlockNumber, _ common.Address) {
			},
			false,
			nil,
		},
		{
			"fail - tendermint client failed to get block",
			utiltx.GenerateAddress(),
			rpctypes.BlockNumberOrHash{BlockNumber: &blockNr},
			func(bn rpctypes.BlockNumber, _ common.Address) {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockError(client, bn.Int64())
			},
			false,
			nil,
		},
		{
			"fail - query client failed to get balance",
			utiltx.GenerateAddress(),
			rpctypes.BlockNumberOrHash{BlockNumber: &blockNr},
			func(bn rpctypes.BlockNumber, addr common.Address) {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlock(client, bn.Int64(), nil)
				suite.Require().NoError(err)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBalanceError(queryClient, addr, bn.Int64())
			},
			false,
			nil,
		},
		{
			"fail - invalid balance",
			utiltx.GenerateAddress(),
			rpctypes.BlockNumberOrHash{BlockNumber: &blockNr},
			func(bn rpctypes.BlockNumber, addr common.Address) {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlock(client, bn.Int64(), nil)
				suite.Require().NoError(err)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBalanceInvalid(queryClient, addr, bn.Int64())
			},
			false,
			nil,
		},
		{
			"fail - pruned node state",
			utiltx.GenerateAddress(),
			rpctypes.BlockNumberOrHash{BlockNumber: &blockNr},
			func(bn rpctypes.BlockNumber, addr common.Address) {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlock(client, bn.Int64(), nil)
				suite.Require().NoError(err)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBalanceNegative(queryClient, addr, bn.Int64())
			},
			false,
			nil,
		},
		{
			"pass",
			utiltx.GenerateAddress(),
			rpctypes.BlockNumberOrHash{BlockNumber: &blockNr},
			func(bn rpctypes.BlockNumber, addr common.Address) {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlock(client, bn.Int64(), nil)
				suite.Require().NoError(err)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBalance(queryClient, addr, bn.Int64())
			},
			true,
			(*hexutil.Big)(big.NewInt(1)),
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()

			// avoid nil pointer reference
			if tc.blockNrOrHash.BlockNumber != nil {
				tc.registerMock(*tc.blockNrOrHash.BlockNumber, tc.addr)
			}

			balance, err := suite.backend.GetBalance(tc.addr, tc.blockNrOrHash)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expBalance, balance)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestGetTransactionCount() {
	testCases := []struct {
		name         string
		accExists    bool
		blockNum     rpctypes.BlockNumber
		registerMock func(common.Address, rpctypes.BlockNumber)
		expPass      bool
		expTxCount   hexutil.Uint64
	}{
		{
			"pass - account doesn't exist",
			false,
			rpctypes.NewBlockNumber(big.NewInt(1)),
			func(_ common.Address, _ rpctypes.BlockNumber) {
				var header metadata.MD
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterParams(queryClient, &header, 1)
			},
			true,
			hexutil.Uint64(0),
		},
		{
			"fail - block height is in the future",
			false,
			rpctypes.NewBlockNumber(big.NewInt(10000)),
			func(_ common.Address, _ rpctypes.BlockNumber) {
				var header metadata.MD
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterParams(queryClient, &header, 1)
			},
			false,
			hexutil.Uint64(0),
		},
		// TODO: Error mocking the GetAccount call - problem with Any type
		// {
		//	"pass - returns the number of transactions at the given address up to the given block number",
		//	true,
		//	rpctypes.NewBlockNumber(big.NewInt(1)),
		//	func(addr common.Address, bn rpctypes.BlockNumber) {
		//		client := suite.backend.clientCtx.Client.(*mocks.Client)
		//		account, err := suite.backend.clientCtx.AccountRetriever.GetAccount(suite.backend.clientCtx, suite.acc)
		//		suite.Require().NoError(err)
		//		request := &authtypes.QueryAccountRequest{Address: sdk.AccAddress(suite.acc.Bytes()).String()}
		//		requestMarshal, _ := request.Marshal()
		//		RegisterABCIQueryAccount(
		//			client,
		//			requestMarshal,
		//			tmrpcclient.ABCIQueryOptions{Height: int64(1), Prove: false},
		//			account,
		//		)
		//	},
		//	true,
		//	hexutil.Uint64(0),
		// },
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()

			addr := utiltx.GenerateAddress()
			if tc.accExists {
				addr = common.BytesToAddress(suite.acc.Bytes())
			}

			tc.registerMock(addr, tc.blockNum)

			txCount, err := suite.backend.GetTransactionCount(addr, tc.blockNum)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expTxCount, *txCount)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
