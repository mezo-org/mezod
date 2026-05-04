package eth

import (
	"errors"
	"math/big"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"

	"cosmossdk.io/log"
	tmrpctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/require"

	rpctypes "github.com/mezo-org/mezod/rpc/types"
	mezotypes "github.com/mezo-org/mezod/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// fakeBackend implements backend.EVMBackend for the kill-switch test.
// SimulateV1 increments simulateCalls so the test can prove the kill
// switch short-circuited without reaching the backend. Every other
// method is a no-op stub — they don't run on this code path.
type fakeBackend struct {
	disabled       bool
	simulateCalls  int
	simulateReturn []*evmtypes.SimBlockResult
}

func (f *fakeBackend) SimulateDisabled() bool { return f.disabled }
func (f *fakeBackend) SimulateV1(_ evmtypes.SimOpts, _ *rpctypes.BlockNumberOrHash) ([]*evmtypes.SimBlockResult, error) {
	f.simulateCalls++
	return f.simulateReturn, nil
}

// --- unused stubs --------------------------------------------------------

func (f *fakeBackend) Accounts() ([]common.Address, error) { return nil, nil }
func (f *fakeBackend) Syncing() (interface{}, error)       { return nil, nil }
func (f *fakeBackend) SetEtherbase(common.Address) bool    { return false }
func (f *fakeBackend) SetGasPrice(hexutil.Big) bool        { return false }
func (f *fakeBackend) ImportRawKey(string, string) (common.Address, error) {
	return common.Address{}, nil
}
func (f *fakeBackend) ListAccounts() ([]common.Address, error) { return nil, nil }
func (f *fakeBackend) NewMnemonic(string, keyring.Language, string, string, keyring.SignatureAlgo) (*keyring.Record, error) {
	return nil, nil
}
func (f *fakeBackend) UnprotectedAllowed() bool     { return false }
func (f *fakeBackend) RPCGasCap() uint64            { return 0 }
func (f *fakeBackend) RPCEVMTimeout() time.Duration { return 0 }
func (f *fakeBackend) RPCTxFeeCap() float64         { return 0 }
func (f *fakeBackend) RPCMinGasPrice() int64        { return 0 }
func (f *fakeBackend) Sign(common.Address, hexutil.Bytes) (hexutil.Bytes, error) {
	return nil, nil
}

func (f *fakeBackend) SendTransaction(evmtypes.TransactionArgs) (common.Hash, error) {
	return common.Hash{}, nil
}

func (f *fakeBackend) SignTypedData(common.Address, apitypes.TypedData) (hexutil.Bytes, error) {
	return nil, nil
}
func (f *fakeBackend) BlockNumber() (hexutil.Uint64, error) { return 0, nil }
func (f *fakeBackend) GetBlockByNumber(rpctypes.BlockNumber, bool) (map[string]interface{}, error) {
	return nil, nil
}

func (f *fakeBackend) GetBlockByHash(common.Hash, bool) (map[string]interface{}, error) {
	return nil, nil
}
func (f *fakeBackend) GetBlockTransactionCountByHash(common.Hash) *hexutil.Uint { return nil }
func (f *fakeBackend) GetBlockTransactionCountByNumber(rpctypes.BlockNumber) *hexutil.Uint {
	return nil
}

func (f *fakeBackend) TendermintBlockByNumber(rpctypes.BlockNumber) (*tmrpctypes.ResultBlock, error) {
	return nil, nil
}

func (f *fakeBackend) TendermintBlockResultByNumber(*int64) (*tmrpctypes.ResultBlockResults, error) {
	return nil, nil
}

func (f *fakeBackend) TendermintBlockByHash(common.Hash) (*tmrpctypes.ResultBlock, error) {
	return nil, nil
}

func (f *fakeBackend) BlockNumberFromTendermint(rpctypes.BlockNumberOrHash) (rpctypes.BlockNumber, error) {
	return 0, nil
}

func (f *fakeBackend) BlockNumberFromTendermintByHash(common.Hash) (*big.Int, error) {
	return nil, nil
}

func (f *fakeBackend) EthMsgsFromTendermintBlock(*tmrpctypes.ResultBlock, *tmrpctypes.ResultBlockResults) []*evmtypes.MsgEthereumTx {
	return nil
}

func (f *fakeBackend) BlockBloom(*tmrpctypes.ResultBlockResults) (ethtypes.Bloom, error) {
	return ethtypes.Bloom{}, nil
}

func (f *fakeBackend) HeaderByNumber(rpctypes.BlockNumber) (*ethtypes.Header, error) {
	return nil, nil
}
func (f *fakeBackend) HeaderByHash(common.Hash) (*ethtypes.Header, error) { return nil, nil }
func (f *fakeBackend) RPCBlockFromTendermintBlock(*tmrpctypes.ResultBlock, *tmrpctypes.ResultBlockResults, bool) (map[string]interface{}, error) {
	return nil, nil
}

func (f *fakeBackend) EthBlockByNumber(rpctypes.BlockNumber) (*ethtypes.Block, error) {
	return nil, nil
}

func (f *fakeBackend) EthBlockFromTendermintBlock(*tmrpctypes.ResultBlock, *tmrpctypes.ResultBlockResults) (*ethtypes.Block, error) {
	return nil, nil
}

func (f *fakeBackend) GetCode(common.Address, rpctypes.BlockNumberOrHash) (hexutil.Bytes, error) {
	return nil, nil
}

func (f *fakeBackend) GetBalance(common.Address, rpctypes.BlockNumberOrHash) (*hexutil.Big, error) {
	return nil, nil
}

func (f *fakeBackend) GetStorageAt(common.Address, string, rpctypes.BlockNumberOrHash) (hexutil.Bytes, error) {
	return nil, nil
}

func (f *fakeBackend) GetProof(common.Address, []string, rpctypes.BlockNumberOrHash) (*rpctypes.AccountResult, error) {
	return nil, nil
}

func (f *fakeBackend) GetTransactionCount(common.Address, rpctypes.BlockNumber) (*hexutil.Uint64, error) {
	return nil, nil
}
func (f *fakeBackend) ChainID() (*hexutil.Big, error)   { return nil, nil }
func (f *fakeBackend) ChainConfig() *params.ChainConfig { return nil }
func (f *fakeBackend) GlobalMinGasPrice() (sdkmath.LegacyDec, error) {
	return sdkmath.LegacyDec{}, nil
}
func (f *fakeBackend) BaseFee(*tmrpctypes.ResultBlockResults) (*big.Int, error) { return nil, nil }
func (f *fakeBackend) CurrentHeader() *ethtypes.Header                          { return nil }
func (f *fakeBackend) PendingTransactions() ([]*sdk.Tx, error)                  { return nil, nil }
func (f *fakeBackend) GetCoinbase() (sdk.AccAddress, error)                     { return nil, nil }
func (f *fakeBackend) FeeHistory(math.HexOrDecimal64, rpc.BlockNumber, []float64) (*rpctypes.FeeHistoryResult, error) {
	return nil, nil
}
func (f *fakeBackend) SuggestGasTipCap(*big.Int) (*big.Int, error) { return nil, nil }
func (f *fakeBackend) GetTransactionByHash(common.Hash) (*rpctypes.RPCTransaction, error) {
	return nil, nil
}
func (f *fakeBackend) GetTxByEthHash(common.Hash) (*mezotypes.TxResult, error) { return nil, nil }
func (f *fakeBackend) GetTransactionByBlockAndIndex(*tmrpctypes.ResultBlock, hexutil.Uint) (*rpctypes.RPCTransaction, error) {
	return nil, nil
}

func (f *fakeBackend) GetTransactionReceipt(common.Hash) (map[string]interface{}, error) {
	return nil, nil
}

func (f *fakeBackend) GetTransactionByBlockHashAndIndex(common.Hash, hexutil.Uint) (*rpctypes.RPCTransaction, error) {
	return nil, nil
}

func (f *fakeBackend) GetTransactionByBlockNumberAndIndex(rpctypes.BlockNumber, hexutil.Uint) (*rpctypes.RPCTransaction, error) {
	return nil, nil
}

func (f *fakeBackend) GetPseudoTransactionResult(*tmrpctypes.ResultBlock) *mezotypes.TxResult {
	return nil
}

func (f *fakeBackend) Resend(evmtypes.TransactionArgs, *hexutil.Big, *hexutil.Uint64) (common.Hash, error) {
	return common.Hash{}, nil
}

func (f *fakeBackend) SendRawTransaction(hexutil.Bytes) (common.Hash, error) {
	return common.Hash{}, nil
}

func (f *fakeBackend) SetTxDefaults(args evmtypes.TransactionArgs) (evmtypes.TransactionArgs, error) {
	return args, nil
}

func (f *fakeBackend) EstimateCost(evmtypes.TransactionArgs, *rpctypes.BlockNumber) (*rpctypes.EstimateCostResult, error) {
	return nil, nil
}

func (f *fakeBackend) EstimateGas(evmtypes.TransactionArgs, *rpctypes.BlockNumber, *evmtypes.StateOverride) (hexutil.Uint64, error) {
	return 0, nil
}

func (f *fakeBackend) DoCall(evmtypes.TransactionArgs, rpctypes.BlockNumber, *evmtypes.StateOverride) (*evmtypes.MsgEthereumTxResponse, error) {
	return nil, nil
}
func (f *fakeBackend) GasPrice() (*hexutil.Big, error)                   { return nil, nil }
func (f *fakeBackend) GetLogs(common.Hash) ([][]*ethtypes.Log, error)    { return nil, nil }
func (f *fakeBackend) GetLogsByHeight(*int64) ([][]*ethtypes.Log, error) { return nil, nil }
func (f *fakeBackend) BloomStatus() (uint64, uint64)                     { return 0, 0 }
func (f *fakeBackend) TraceTransaction(common.Hash, *evmtypes.TraceConfig) (interface{}, error) {
	return nil, nil
}

func (f *fakeBackend) TraceBlock(rpctypes.BlockNumber, *evmtypes.TraceConfig, *tmrpctypes.ResultBlock) ([]*evmtypes.TxTraceResult, error) {
	return nil, nil
}

// Kill switch returns -32601 without reaching the backend.
func TestSimulateV1_KillSwitch(t *testing.T) {
	be := &fakeBackend{disabled: true}
	api := NewPublicAPI(log.NewNopLogger(), be)

	got, err := api.SimulateV1(evmtypes.SimOpts{}, nil)
	require.Nil(t, got)
	require.Error(t, err)
	require.Equal(t, 0, be.simulateCalls, "kill switch must short-circuit before backend")

	var simErr *evmtypes.SimError
	require.True(t, errors.As(err, &simErr))
	require.Equal(t, evmtypes.SimErrCodeMethodNotFound, simErr.ErrorCode())
	require.Contains(t, simErr.Error(), "eth_simulateV1")
}

// When the kill switch is off the call must reach the backend.
func TestSimulateV1_KillSwitchOff(t *testing.T) {
	be := &fakeBackend{disabled: false}
	api := NewPublicAPI(log.NewNopLogger(), be)

	_, err := api.SimulateV1(evmtypes.SimOpts{}, nil)
	require.NoError(t, err)
	require.Equal(t, 1, be.simulateCalls, "kill switch off path must reach the backend")
}
