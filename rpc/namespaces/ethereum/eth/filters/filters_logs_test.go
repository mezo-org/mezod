package filters

import (
	"context"
	"math/big"
	"testing"

	"cosmossdk.io/log"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	gethfilters "github.com/ethereum/go-ethereum/eth/filters"
	"github.com/stretchr/testify/require"

	rpctypes "github.com/mezo-org/mezod/rpc/types"
	mezodtypes "github.com/mezo-org/mezod/types"
)

// nilResultBlockByHashBackend simulates TendermintBlockByHash when the node returns
// no block for the hash without an RPC error (see rpc/backend/blocks.go).
type nilResultBlockByHashBackend struct{}

func (nilResultBlockByHashBackend) GetBlockByNumber(rpctypes.BlockNumber, bool) (map[string]interface{}, error) {
	return nil, nil
}

func (nilResultBlockByHashBackend) HeaderByNumber(rpctypes.BlockNumber) (*ethtypes.Header, error) {
	return nil, nil
}

func (nilResultBlockByHashBackend) HeaderByHash(common.Hash) (*ethtypes.Header, error) {
	return nil, nil
}

func (nilResultBlockByHashBackend) GetPseudoTransactionResult(*coretypes.ResultBlock) *mezodtypes.TxResult {
	return nil
}

func (nilResultBlockByHashBackend) TendermintBlockByHash(common.Hash) (*coretypes.ResultBlock, error) {
	return nil, nil
}

func (nilResultBlockByHashBackend) TendermintBlockByNumber(rpctypes.BlockNumber) (*coretypes.ResultBlock, error) {
	return nil, nil
}

func (nilResultBlockByHashBackend) TendermintBlockResultByNumber(*int64) (*coretypes.ResultBlockResults, error) {
	return nil, nil
}

func (nilResultBlockByHashBackend) GetLogs(common.Hash) ([][]*ethtypes.Log, error) {
	return nil, nil
}

func (nilResultBlockByHashBackend) GetLogsByHeight(*int64) ([][]*ethtypes.Log, error) {
	return nil, nil
}

func (nilResultBlockByHashBackend) BlockBloom(*coretypes.ResultBlockResults) (ethtypes.Bloom, error) {
	return ethtypes.Bloom{}, nil
}

func (nilResultBlockByHashBackend) BloomStatus() (uint64, uint64) {
	return 0, 0
}

func (nilResultBlockByHashBackend) RPCFilterCap() int32   { return 0 }
func (nilResultBlockByHashBackend) RPCLogsCap() int32     { return 10000 }
func (nilResultBlockByHashBackend) RPCBlockRangeCap() int32 { return 10000 }
func (nilResultBlockByHashBackend) RPCLogsFilterAddrCap() int32 {
	return 0
}

func TestFilterLogs_blockHashMissingDoesNotPanic(t *testing.T) {
	t.Parallel()

	hash := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000abc")
	criteria := gethfilters.FilterCriteria{
		BlockHash: &hash,
		FromBlock: big.NewInt(-1),
		ToBlock:   big.NewInt(-1),
	}

	f := NewBlockFilter(log.NewNopLogger(), nilResultBlockByHashBackend{}, criteria)
	logs, err := f.Logs(context.Background(), 10000, 10000)
	require.NoError(t, err)
	require.Nil(t, logs)
}
