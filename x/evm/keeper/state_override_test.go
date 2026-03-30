package keeper

import (
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/holiman/uint256"
	"github.com/mezo-org/mezod/x/evm/statedb"
	"github.com/stretchr/testify/require"
)

var (
	testBlockHash = common.BigToHash(big.NewInt(9999))
	testTxConfig  = statedb.NewEmptyTxConfig(testBlockHash)
	testAddr1     = common.BigToAddress(big.NewInt(201))
	testAddr2     = common.BigToAddress(big.NewInt(202))
)

func newTestDB() (*statedb.StateDB, *statedb.MockKeeper) {
	keeper := statedb.NewMockKeeper()
	db := statedb.New(sdk.Context{}, keeper, testTxConfig)
	return db, keeper
}

func ptrUint64(v uint64) *hexutil.Uint64 {
	hv := hexutil.Uint64(v)
	return &hv
}

func ptrBig(v *big.Int) *hexutil.Big {
	hv := hexutil.Big(*v)
	return &hv
}

func ptrBytes(v []byte) *hexutil.Bytes {
	hv := hexutil.Bytes(v)
	return &hv
}

func TestApplyStateOverrides_Nonce(t *testing.T) {
	db, _ := newTestDB()

	overrides := stateOverride{
		testAddr1: overrideAccount{
			Nonce: ptrUint64(42),
		},
	}

	err := applyStateOverrides(db, overrides)
	require.NoError(t, err)
	require.Equal(t, uint64(42), db.GetNonce(testAddr1))
}

func TestApplyStateOverrides_Code(t *testing.T) {
	db, _ := newTestDB()
	code := []byte("contract code")

	overrides := stateOverride{
		testAddr1: overrideAccount{
			Code: ptrBytes(code),
		},
	}

	err := applyStateOverrides(db, overrides)
	require.NoError(t, err)
	require.Equal(t, code, db.GetCode(testAddr1))
}

func TestApplyStateOverrides_Balance(t *testing.T) {
	db, _ := newTestDB()

	overrides := stateOverride{
		testAddr1: overrideAccount{
			Balance: ptrBig(big.NewInt(1000)),
		},
	}

	err := applyStateOverrides(db, overrides)
	require.NoError(t, err)
	require.Equal(t, uint256.NewInt(1000), db.GetBalance(testAddr1))
}

func TestApplyStateOverrides_State(t *testing.T) {
	db, _ := newTestDB()
	newKey := common.BigToHash(big.NewInt(2))
	newVal := common.BigToHash(big.NewInt(20))

	overrides := stateOverride{
		testAddr1: overrideAccount{
			State: map[common.Hash]common.Hash{
				newKey: newVal,
			},
		},
	}

	err := applyStateOverrides(db, overrides)
	require.NoError(t, err)

	require.Equal(t, newVal, db.GetState(testAddr1, newKey))
}

func TestApplyStateOverrides_StateDiff(t *testing.T) {
	db, keeper := newTestDB()

	key1 := common.BigToHash(big.NewInt(1))
	val1 := common.BigToHash(big.NewInt(10))
	key2 := common.BigToHash(big.NewInt(2))
	val2 := common.BigToHash(big.NewInt(20))

	// Pre-populate storage
	db.SetState(testAddr1, key1, val1)
	db.SetState(testAddr1, key2, val2)
	require.NoError(t, db.Commit())

	// Apply stateDiff to change only key1
	db2 := statedb.New(sdk.Context{}, keeper, testTxConfig)
	newVal1 := common.BigToHash(big.NewInt(99))

	overrides := stateOverride{
		testAddr1: overrideAccount{
			StateDiff: map[common.Hash]common.Hash{
				key1: newVal1,
			},
		},
	}

	err := applyStateOverrides(db2, overrides)
	require.NoError(t, err)

	// key1 changed
	require.Equal(t, newVal1, db2.GetState(testAddr1, key1))
	// key2 untouched
	require.Equal(t, val2, db2.GetState(testAddr1, key2))
}

func TestApplyStateOverrides_StateAndStateDiffError(t *testing.T) {
	db, _ := newTestDB()

	overrides := stateOverride{
		testAddr1: overrideAccount{
			State: map[common.Hash]common.Hash{
				common.BigToHash(big.NewInt(1)): common.BigToHash(big.NewInt(10)),
			},
			StateDiff: map[common.Hash]common.Hash{
				common.BigToHash(big.NewInt(2)): common.BigToHash(big.NewInt(20)),
			},
		},
	}

	err := applyStateOverrides(db, overrides)
	require.Error(t, err)
	require.Contains(t, err.Error(), "has both state and stateDiff overrides")
}

func TestApplyStateOverrides_Combined(t *testing.T) {
	db, _ := newTestDB()
	code := []byte("new code")

	overrides := stateOverride{
		testAddr1: overrideAccount{
			Nonce:   ptrUint64(10),
			Code:    ptrBytes(code),
			Balance: ptrBig(big.NewInt(5000)),
			StateDiff: map[common.Hash]common.Hash{
				common.BigToHash(big.NewInt(1)): common.BigToHash(big.NewInt(100)),
			},
		},
	}

	err := applyStateOverrides(db, overrides)
	require.NoError(t, err)

	require.Equal(t, uint64(10), db.GetNonce(testAddr1))
	require.Equal(t, code, db.GetCode(testAddr1))
	require.Equal(t, uint256.NewInt(5000), db.GetBalance(testAddr1))
	require.Equal(t, common.BigToHash(big.NewInt(100)), db.GetState(testAddr1, common.BigToHash(big.NewInt(1))))
}

func TestApplyStateOverrides_MultipleAccounts(t *testing.T) {
	db, _ := newTestDB()

	overrides := stateOverride{
		testAddr1: overrideAccount{
			Nonce:   ptrUint64(1),
			Balance: ptrBig(big.NewInt(100)),
		},
		testAddr2: overrideAccount{
			Nonce:   ptrUint64(2),
			Balance: ptrBig(big.NewInt(200)),
		},
	}

	err := applyStateOverrides(db, overrides)
	require.NoError(t, err)

	require.Equal(t, uint64(1), db.GetNonce(testAddr1))
	require.Equal(t, uint256.NewInt(100), db.GetBalance(testAddr1))
	require.Equal(t, uint64(2), db.GetNonce(testAddr2))
	require.Equal(t, uint256.NewInt(200), db.GetBalance(testAddr2))
}

func TestApplyStateOverrides_NilFieldsSkipped(t *testing.T) {
	db, _ := newTestDB()

	// Set initial state
	db.SetNonce(testAddr1, 5)

	overrides := stateOverride{
		testAddr1: overrideAccount{
			// All fields nil. Nothing should change
		},
	}

	err := applyStateOverrides(db, overrides)
	require.NoError(t, err)

	// Nonce should be unchanged
	require.Equal(t, uint64(5), db.GetNonce(testAddr1))
}
