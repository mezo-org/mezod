package keeper

import (
	"math/big"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	rpctypes "github.com/mezo-org/mezod/rpc/types"
	"github.com/mezo-org/mezod/x/evm/statedb"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/stretchr/testify/require"
)

var (
	testBlockHash = common.BigToHash(big.NewInt(9999))
	testTxConfig  = statedb.NewEmptyTxConfig(testBlockHash)
	testAddr1     = common.BigToAddress(big.NewInt(201))
	testAddr2     = common.BigToAddress(big.NewInt(202))
	// testRules enables stdlib precompiles 0x01..0x09 (Istanbul+). Sufficient
	// for every MovePrecompileTo test that names sha256 (0x02) or ripemd (0x03).
	testRules = params.Rules{IsIstanbul: true, IsBerlin: true}
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

	moves, err := applyStateOverrides(db, overrides, testRules)
	require.NoError(t, err)
	require.Nil(t, moves)
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

	moves, err := applyStateOverrides(db, overrides, testRules)
	require.NoError(t, err)
	require.Nil(t, moves)
	require.Equal(t, code, db.GetCode(testAddr1))
}

func TestApplyStateOverrides_Balance(t *testing.T) {
	db, _ := newTestDB()

	overrides := stateOverride{
		testAddr1: overrideAccount{
			Balance: ptrBig(big.NewInt(1000)),
		},
	}

	moves, err := applyStateOverrides(db, overrides, testRules)
	require.NoError(t, err)
	require.Nil(t, moves)
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

	moves, err := applyStateOverrides(db, overrides, testRules)
	require.NoError(t, err)
	require.Nil(t, moves)

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

	moves, err := applyStateOverrides(db2, overrides, testRules)
	require.NoError(t, err)
	require.Nil(t, moves)

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

	_, err := applyStateOverrides(db, overrides, testRules)
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

	moves, err := applyStateOverrides(db, overrides, testRules)
	require.NoError(t, err)
	require.Nil(t, moves)

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

	moves, err := applyStateOverrides(db, overrides, testRules)
	require.NoError(t, err)
	require.Nil(t, moves)

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

	moves, err := applyStateOverrides(db, overrides, testRules)
	require.NoError(t, err)
	require.Nil(t, moves)

	// Nonce should be unchanged
	require.Equal(t, uint64(5), db.GetNonce(testAddr1))
}

func ptrAddr(a common.Address) *common.Address { return &a }

func TestApplyStateOverrides_MovePrecompileTo_CollectsMove(t *testing.T) {
	db, _ := newTestDB()

	sha256Addr := common.BytesToAddress([]byte{0x02})
	dest := common.BigToAddress(big.NewInt(0x1234))

	overrides := stateOverride{
		sha256Addr: overrideAccount{
			MovePrecompileTo: ptrAddr(dest),
		},
	}

	moves, err := applyStateOverrides(db, overrides, testRules)
	require.NoError(t, err)

	require.Len(t, moves, 1)
	require.Equal(t, dest, moves[sha256Addr])
}

func TestApplyStateOverrides_MovePrecompileTo_SourceNotPrecompile(t *testing.T) {
	db, _ := newTestDB()

	nonPrecompile := common.BigToAddress(big.NewInt(500))
	dest := common.BigToAddress(big.NewInt(600))

	overrides := stateOverride{
		nonPrecompile: overrideAccount{
			MovePrecompileTo: ptrAddr(dest),
		},
	}

	_, err := applyStateOverrides(db, overrides, testRules)
	require.Error(t, err)

	rpcErr, ok := err.(*rpctypes.RPCError)
	require.True(t, ok, "expected *RPCError, got %T", err)
	require.Equal(t, rpctypes.SimErrCodeInvalidParams, rpcErr.Code)
	require.Contains(t, rpcErr.Message, "is not a precompile")
}

func TestApplyStateOverrides_MovePrecompileTo_SelfReference(t *testing.T) {
	db, _ := newTestDB()

	sha256Addr := common.BytesToAddress([]byte{0x02})

	overrides := stateOverride{
		sha256Addr: overrideAccount{
			MovePrecompileTo: ptrAddr(sha256Addr),
		},
	}

	_, err := applyStateOverrides(db, overrides, testRules)
	require.Error(t, err)

	rpcErr, ok := err.(*rpctypes.RPCError)
	require.True(t, ok, "expected *RPCError, got %T", err)
	require.Equal(t, rpctypes.SimErrCodeMovePrecompileSelfReference, rpcErr.Code)
	require.Contains(t, rpcErr.Message, "referenced itself")
}

func TestApplyStateOverrides_MovePrecompileTo_DuplicateDestination(t *testing.T) {
	db, _ := newTestDB()

	sha256Addr := common.BytesToAddress([]byte{0x02})
	ripemdAddr := common.BytesToAddress([]byte{0x03})
	dest := common.BigToAddress(big.NewInt(0x1234))

	overrides := stateOverride{
		sha256Addr: overrideAccount{
			MovePrecompileTo: ptrAddr(dest),
		},
		ripemdAddr: overrideAccount{
			MovePrecompileTo: ptrAddr(dest),
		},
	}

	_, err := applyStateOverrides(db, overrides, testRules)
	require.Error(t, err)

	rpcErr, ok := err.(*rpctypes.RPCError)
	require.True(t, ok, "expected *RPCError, got %T", err)
	require.Equal(t, rpctypes.SimErrCodeMovePrecompileDuplicateDest, rpcErr.Code)
	require.Contains(t, rpcErr.Message, dest.Hex())
}

func TestApplyStateOverrides_MovePrecompileTo_MezoCustomBlocked(t *testing.T) {
	dest := common.BigToAddress(big.NewInt(0xabc))

	for _, pv := range evmtypes.DefaultPrecompilesVersions {
		t.Run(pv.PrecompileAddress, func(t *testing.T) {
			db, _ := newTestDB()

			customAddr := common.HexToAddress(pv.PrecompileAddress)

			overrides := stateOverride{
				customAddr: overrideAccount{
					MovePrecompileTo: ptrAddr(dest),
				},
			}

			_, err := applyStateOverrides(db, overrides, testRules)
			require.Error(t, err)

			rpcErr, ok := err.(*rpctypes.RPCError)
			require.True(t, ok, "expected *RPCError, got %T", err)
			require.Equal(t, rpctypes.SimErrCodeInvalidParams, rpcErr.Code)
			require.Contains(t, rpcErr.Message, "cannot move mezo custom precompile")
		})
	}
}

func TestApplyStateOverrides_MovePrecompileTo_DestAlreadyOverridden(t *testing.T) {
	db, _ := newTestDB()

	sha256Addr := common.BytesToAddress([]byte{0x02})
	dest := common.BigToAddress(big.NewInt(0x1234))

	overrides := stateOverride{
		sha256Addr: overrideAccount{MovePrecompileTo: ptrAddr(dest)},
		dest:       overrideAccount{Code: ptrBytes([]byte("conflicting"))},
	}

	_, err := applyStateOverrides(db, overrides, testRules)
	require.Error(t, err)

	rpcErr, ok := err.(*rpctypes.RPCError)
	require.True(t, ok, "expected *RPCError, got %T", err)
	require.Equal(t, rpctypes.SimErrCodeInvalidParams, rpcErr.Code)
	require.Contains(t, rpcErr.Message, "is already overridden")
}

func TestApplyStateOverrides_MovePrecompileTo_ChainedMoveRejected(t *testing.T) {
	db, _ := newTestDB()

	sha256Addr := common.BytesToAddress([]byte{0x02})
	ripemdAddr := common.BytesToAddress([]byte{0x03})
	thirdAddr := common.BigToAddress(big.NewInt(0xabcd))

	overrides := stateOverride{
		sha256Addr: overrideAccount{MovePrecompileTo: ptrAddr(ripemdAddr)},
		ripemdAddr: overrideAccount{MovePrecompileTo: ptrAddr(thirdAddr)},
	}

	_, err := applyStateOverrides(db, overrides, testRules)
	require.Error(t, err)

	rpcErr, ok := err.(*rpctypes.RPCError)
	require.True(t, ok, "expected *RPCError, got %T", err)
	require.Equal(t, rpctypes.SimErrCodeInvalidParams, rpcErr.Code)
	// Either guard may fire depending on iteration order; both are
	// acceptable so long as one of them always does.
	require.True(t,
		strings.Contains(rpcErr.Message, "already overridden by a precompile") ||
			strings.Contains(rpcErr.Message, "is already overridden"),
		"unexpected rejection message: %s", rpcErr.Message,
	)
}

func TestApplyStateOverrides_MovePrecompileTo_WithSourceCodeOverwrite(t *testing.T) {
	db, _ := newTestDB()

	sha256Addr := common.BytesToAddress([]byte{0x02})
	dest := common.BigToAddress(big.NewInt(0x1234))
	newCode := []byte("overwriting source code")

	overrides := stateOverride{
		sha256Addr: overrideAccount{
			MovePrecompileTo: ptrAddr(dest),
			Code:             ptrBytes(newCode),
			Nonce:            ptrUint64(7),
		},
	}

	moves, err := applyStateOverrides(db, overrides, testRules)
	require.NoError(t, err)

	require.Len(t, moves, 1)
	require.Equal(t, dest, moves[sha256Addr])

	// Regular account overrides on the source address still land on the stateDB.
	require.Equal(t, newCode, db.GetCode(sha256Addr))
	require.Equal(t, uint64(7), db.GetNonce(sha256Addr))
}
