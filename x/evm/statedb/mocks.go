package statedb

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/mezo-org/mezod/x/evm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

var (
	_          Keeper = &MockKeeper{}
	errAddress        = common.BigToAddress(big.NewInt(100))
)

type MockAccount struct {
	Account Account
	States  Storage
}

type MockKeeper struct {
	Accounts map[common.Address]MockAccount
	Codes    map[common.Hash][]byte
}

func NewMockKeeper() *MockKeeper {
	return &MockKeeper{
		Accounts: make(map[common.Address]MockAccount),
		Codes:    make(map[common.Hash][]byte),
	}
}

func (k MockKeeper) GetAccount(_ sdk.Context, addr common.Address) *Account {
	acct, ok := k.Accounts[addr]
	if !ok {
		return nil
	}
	return &acct.Account
}

func (k MockKeeper) GetState(_ sdk.Context, addr common.Address, key common.Hash) common.Hash {
	return k.Accounts[addr].States[key]
}

func (k MockKeeper) GetCode(_ sdk.Context, codeHash common.Hash) []byte {
	return k.Codes[codeHash]
}

func (k MockKeeper) ForEachStorage(_ sdk.Context, addr common.Address, cb func(key, value common.Hash) bool) {
	if acct, ok := k.Accounts[addr]; ok {
		for k, v := range acct.States {
			if !cb(k, v) {
				return
			}
		}
	}
}

func (k MockKeeper) SetAccount(_ sdk.Context, addr common.Address, account Account) error {
	if addr == errAddress {
		return errors.New("mock db error")
	}
	acct, exists := k.Accounts[addr]
	if exists {
		// update
		acct.Account = account
		k.Accounts[addr] = acct
	} else {
		k.Accounts[addr] = MockAccount{Account: account, States: make(Storage)}
	}
	return nil
}

func (k MockKeeper) SetState(_ sdk.Context, addr common.Address, key common.Hash, value []byte) {
	if acct, ok := k.Accounts[addr]; ok {
		if len(value) == 0 {
			delete(acct.States, key)
		} else {
			acct.States[key] = common.BytesToHash(value)
		}
	}
}

func (k MockKeeper) SetCode(_ sdk.Context, codeHash []byte, code []byte) {
	k.Codes[common.BytesToHash(codeHash)] = code
}

func (k MockKeeper) DeleteAccount(_ sdk.Context, addr common.Address) error {
	if addr == errAddress {
		return errors.New("mock db error")
	}
	old := k.Accounts[addr]
	delete(k.Accounts, addr)
	if !bytes.Equal(old.Account.CodeHash, emptyCodeHash) {
		delete(k.Codes, common.BytesToHash(old.Account.CodeHash))
	}
	return nil
}

func (k MockKeeper) Clone() *MockKeeper {
	accounts := make(map[common.Address]MockAccount, len(k.Accounts))
	for k, v := range k.Accounts {
		accounts[k] = v
	}
	codes := make(map[common.Hash][]byte, len(k.Codes))
	for k, v := range k.Codes {
		codes[k] = v
	}
	return &MockKeeper{accounts, codes}
}

func (k MockKeeper) GetStorageRootStrategy(_ sdk.Context) types.StorageRootStrategy {
	return types.StorageRootStrategyEmptyHash
}

func (k MockKeeper) GetMaxPrecompilesCallsPerExecution(_ sdk.Context) uint {
	return 10
}
