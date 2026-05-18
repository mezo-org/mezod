package statedb_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"

	"github.com/mezo-org/mezod/x/evm/statedb"
)

func (suite *StateDBTestSuite) TestTracingHooks() {
	type balanceChange struct {
		prev   string
		new    string
		reason tracing.BalanceChangeReason
	}
	type nonceChange struct {
		prev   uint64
		new    uint64
		reason tracing.NonceChangeReason
	}
	type codeChange struct {
		prev   []byte
		new    []byte
		reason tracing.CodeChangeReason
	}
	type storageChange struct {
		prev common.Hash
		new  common.Hash
	}

	db := statedb.New(suite.ctx, statedb.NewMockKeeper(), emptyTxConfig)
	key := common.BigToHash(big.NewInt(1))
	value := common.BigToHash(big.NewInt(2))

	var balanceChanges []balanceChange
	var nonceChanges []nonceChange
	var codeChanges []codeChange
	var storageChanges []storageChange

	removeTracingHooks := db.AddTracingHooks(&tracing.Hooks{
		OnBalanceChange: func(_ common.Address, prev, newBalance *big.Int, reason tracing.BalanceChangeReason) {
			balanceChanges = append(balanceChanges, balanceChange{prev: prev.String(), new: newBalance.String(), reason: reason})
		},
		OnNonceChangeV2: func(_ common.Address, prev, newNonce uint64, reason tracing.NonceChangeReason) {
			nonceChanges = append(nonceChanges, nonceChange{prev: prev, new: newNonce, reason: reason})
		},
		OnCodeChangeV2: func(_ common.Address, _ common.Hash, prevCode []byte, _ common.Hash, code []byte, reason tracing.CodeChangeReason) {
			codeChanges = append(codeChanges, codeChange{prev: prevCode, new: code, reason: reason})
		},
		OnStorageChange: func(_ common.Address, _ common.Hash, prev, newValue common.Hash) {
			storageChanges = append(storageChanges, storageChange{prev: prev, new: newValue})
		},
	})
	defer removeTracingHooks()

	db.AddBalance(address, uint256.NewInt(10), tracing.BalanceChangeTransfer)
	db.SubBalance(address, uint256.NewInt(3), tracing.BalanceDecreaseGasBuy)
	db.SetNonce(address, 7, tracing.NonceChangeAuthorization)
	db.SetCode(address, []byte("code"), tracing.CodeChangeAuthorization)
	db.SetState(address, key, value)
	db.SelfDestruct(address)

	suite.Require().Equal(
		[]balanceChange{
			{prev: "0", new: "10", reason: tracing.BalanceChangeTransfer},
			{prev: "10", new: "7", reason: tracing.BalanceDecreaseGasBuy},
			{prev: "7", new: "0", reason: tracing.BalanceDecreaseSelfdestruct},
		},
		balanceChanges,
	)
	suite.Require().Equal(
		[]nonceChange{{prev: 0, new: 7, reason: tracing.NonceChangeAuthorization}},
		nonceChanges,
	)
	suite.Require().Equal(
		[]codeChange{
			{prev: nil, new: []byte("code"), reason: tracing.CodeChangeAuthorization},
			{prev: []byte("code"), new: nil, reason: tracing.CodeChangeSelfDestruct},
		},
		codeChanges,
	)
	suite.Require().Equal(
		[]storageChange{{prev: common.Hash{}, new: value}},
		storageChanges,
	)
}

func (suite *StateDBTestSuite) TestAddTracingHooks() {
	db := statedb.New(suite.ctx, statedb.NewMockKeeper(), emptyTxConfig)
	var firstLogs int
	var secondLogs int

	firstHooks := &tracing.Hooks{OnLog: func(_ *ethtypes.Log) { firstLogs++ }}
	secondHooks := &tracing.Hooks{OnLog: func(_ *ethtypes.Log) { secondLogs++ }}

	removeFirst := db.AddTracingHooks(firstHooks)
	removeFirstDuplicate := db.AddTracingHooks(firstHooks)
	removeSecond := db.AddTracingHooks(secondHooks)

	db.AddLog(&ethtypes.Log{})
	suite.Require().Equal(1, firstLogs)
	suite.Require().Equal(1, secondLogs)

	removeFirstDuplicate()
	db.AddLog(&ethtypes.Log{})
	suite.Require().Equal(2, firstLogs)
	suite.Require().Equal(2, secondLogs)

	removeFirst()
	db.AddLog(&ethtypes.Log{})
	suite.Require().Equal(2, firstLogs)
	suite.Require().Equal(3, secondLogs)

	removeSecond()
	db.AddLog(&ethtypes.Log{})
	suite.Require().Equal(2, firstLogs)
	suite.Require().Equal(3, secondLogs)
}
