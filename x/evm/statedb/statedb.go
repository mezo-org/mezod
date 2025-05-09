// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE
package statedb

import (
	"fmt"
	"sort"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/tracing"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie/utils"
	"github.com/holiman/uint256"
	"github.com/mezo-org/mezod/x/evm/types"
)

// revision is the identifier of a version of state.
// it consists of an auto-increment id and a journal index.
// it's safer to use than using journal index alone.
type revision struct {
	id           int
	journalIndex int
}

// Number of address->curve point associations to keep.
const pointCacheSize = 4096

var _ vm.StateDB = &StateDB{}

// StateDB structs within the ethereum protocol are used to store anything
// within the merkle trie. StateDBs take care of caching and storing
// nested states. It's the general query interface to retrieve:
// * Contracts
// * Accounts
type StateDB struct {
	keeper     Keeper
	ctx        sdk.Context
	cachedCtx  sdk.Context
	flushCache func()

	// Journal of state modifications. This is the backbone of
	// Snapshot and RevertToSnapshot.
	journal        *journal
	validRevisions []revision
	nextRevisionID int

	logger *tracing.Hooks

	stateObjects map[common.Address]*stateObject

	txConfig TxConfig

	// The refund counter, also used by state transitioning.
	refund uint64

	// Per-transaction logs
	logs []*ethtypes.Log

	// Per-transaction access list
	accessList *accessList

	// Transient storage
	transientStorage transientStorage

	// State witness if cross validation is needed
	witness *stateless.Witness

	storageRootStrategy types.StorageRootStrategy

	// This counter is use to keep track of how many time
	// a precompile have been called during this execution. This
	// is implemented as a counter limit against abusing the use
	// of precompile as part of a single smart contract call. We
	// are added because every time a precompile contract is call,
	// a new copy of the state of the context is made, so we want
	// to prevent it eating up too much memory at once and any possible
	// attack related to this.
	// For backward compatibility a maxPrecompilesCallsPerExecution set to 0
	// means there's no limits.
	ongoingPrecompilesCallsCounter uint

	maxPrecompilesCallsPerExecution uint
}

// New creates a new state from a given trie.
func New(ctx sdk.Context, keeper Keeper, txConfig TxConfig) *StateDB {
	return &StateDB{
		keeper:                          keeper,
		ctx:                             ctx,
		stateObjects:                    make(map[common.Address]*stateObject),
		journal:                         newJournal(),
		accessList:                      newAccessList(),
		txConfig:                        txConfig,
		storageRootStrategy:             keeper.GetStorageRootStrategy(ctx),
		maxPrecompilesCallsPerExecution: keeper.GetMaxPrecompilesCallsPerExecution(ctx),
	}
}

// Keeper returns the underlying `Keeper`
func (s *StateDB) Keeper() Keeper {
	return s.keeper
}

// GetContext returns the transaction Context.
func (s *StateDB) GetContext() sdk.Context {
	return s.ctx
}

// AddLog adds a log, called by evm.
func (s *StateDB) AddLog(log *ethtypes.Log) {
	s.journal.append(addLogChange{})

	log.TxHash = s.txConfig.TxHash
	log.BlockHash = s.txConfig.BlockHash
	log.TxIndex = s.txConfig.TxIndex
	log.Index = s.txConfig.LogIndex + uint(len(s.logs))
	s.logs = append(s.logs, log)
}

// Logs returns the logs of current transaction.
func (s *StateDB) Logs() []*ethtypes.Log {
	return s.logs
}

// AddRefund adds gas to the refund counter
func (s *StateDB) AddRefund(gas uint64) {
	s.journal.append(refundChange{prev: s.refund})
	s.refund += gas
}

// SubRefund removes gas from the refund counter.
// This method will panic if the refund counter goes below zero
func (s *StateDB) SubRefund(gas uint64) {
	s.journal.append(refundChange{prev: s.refund})
	if gas > s.refund {
		panic(fmt.Sprintf("Refund counter below zero (gas: %d > refund: %d)", gas, s.refund))
	}
	s.refund -= gas
}

// Exist reports whether the given account address exists in the state.
// Notably this also returns true for suicided accounts.
func (s *StateDB) Exist(addr common.Address) bool {
	return s.getStateObject(addr) != nil
}

// Empty returns whether the state object is either non-existent
// or empty according to the EIP161 specification (balance = nonce = code = 0)
func (s *StateDB) Empty(addr common.Address) bool {
	so := s.getStateObject(addr)
	return so == nil || so.empty()
}

// GetBalance retrieves the balance from the given address or 0 if object not found
func (s *StateDB) GetBalance(addr common.Address) *uint256.Int {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Balance()
	}
	return uint256.NewInt(0)
}

// GetNonce returns the nonce of account, 0 if not exists.
func (s *StateDB) GetNonce(addr common.Address) uint64 {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Nonce()
	}

	return 0
}

// GetStorageRoot always returns an empty hash as the current implementation
// does not track storage roots.
func (s *StateDB) GetStorageRoot(addr common.Address) common.Hash {
	switch s.storageRootStrategy {
	case types.StorageRootStrategyDummyHash:
		return getStorageRootDummyHash(s, addr)
	case types.StorageRootStrategyEmptyHash:
		return getStorageRootEmptyHash(s, addr)
	default:
		panic("unknown storage root strategy")
	}
}

// getStorageRootDummyHash is the implementation of the types.StorageRootStrategyDummyHash strategy.
func getStorageRootDummyHash(s *StateDB, addr common.Address) common.Hash {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		// NOTE! The intention here is to return a state root hash to comply with
		// https://eips.ethereum.org/EIPS/eip-7610 Proper implementation is used
		// to revert contract creation if address already has the non-empty storage.
		// However, our current codebase does not support tracking the storage root
		// hash that is required by the currently used go-ethereum version. For now,
		// we just return a dummy hash to indicate that the storage exists behind this
		// address. This should be good enough for now.
		return common.HexToHash("6d657a6f")
	}
	return common.Hash{}
}

// getStorageRootEmptyHash is the implementation of the types.StorageRootStrategyEmptyHash strategy.
func getStorageRootEmptyHash(_ *StateDB, _ common.Address) common.Hash {
	// !!! WARNING !!!
	//
	// Mezo does not support tracking the storage roots and it is not possible
	// to return it from here. With the v1.14.8 go-ethereum dependency, this is
	// acceptable, because:
	// * Our Keeper's DeleteAccount() function for StateDB removes both the code
	//   and storage in one shot. This way,  deploying a new contract under
	//   a self-destructed address that holds a non-empty storage within the
	//   same transaction is not possible. Deploying a new contract will be
	//   possible as the nonce will be reset to 0 but the storage will be
	//   removed first.
	// * EIP-7610 was meant for certain specific cases in the old EVM state that
	//   are not possible to happen in Mezo. With EIP-161 the nonce is
	//   incremented and deploying a contract with a zero nonce and zero code
	//   length but with non-empty storage is not possible.
	//
	// Returning the empty hash here becomes potentially unsafe after we upgrade
	// the go-ethereum dependency. Also, without proper tracking of the storage
	// root incorporating future go-ethereum security upgrades may be
	// complicated.
	//
	// The work for tracking the storage root has been captured in
	// https://github.com/mezo-org/mezod/issues/369
	return common.Hash{}
}

// GetTransientState gets transient storage for a given account.
func (s *StateDB) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	return s.transientStorage.Get(addr, key)
}

func (s *StateDB) HasSelfDestructed(addr common.Address) bool {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.selfDestructed
	}
	return false
}

// Prepare handles the preparatory steps for executing a state transition with.
// This method must be invoked before state transition.
//
// Berlin fork:
// - Add sender to access list (2929)
// - Add destination to access list (2929)
// - Add precompiles to access list (2929)
// - Add the contents of the optional tx access list (2930)
//
// Potential EIPs:
// - Reset access list (Berlin)
// - Add coinbase to access list (EIP-3651)
// - Reset transient storage (EIP-1153)
func (s *StateDB) Prepare(rules params.Rules, sender, coinbase common.Address, dst *common.Address, precompiles []common.Address, list ethtypes.AccessList) {
	if rules.IsEIP2929 && rules.IsEIP4762 {
		panic("eip2929 and eip4762 are both activated")
	}
	if rules.IsEIP2929 {
		// Clear out any leftover from previous executions
		al := newAccessList()
		s.accessList = al

		al.AddAddress(sender)
		if dst != nil {
			al.AddAddress(*dst)
			// If it's a create-tx, the destination will be added inside evm.create
		}
		for _, addr := range precompiles {
			al.AddAddress(addr)
		}
		for _, el := range list {
			al.AddAddress(el.Address)
			for _, key := range el.StorageKeys {
				al.AddSlot(el.Address, key)
			}
		}
		if rules.IsShanghai { // EIP-3651: warm coinbase
			al.AddAddress(coinbase)
		}
	}
	// Reset transient storage at the beginning of transaction execution
	s.transientStorage = newTransientStorage()
}

// SelfDestruct marks the given account as selfdestructed.
// This clears the account balance.
//
// The account's state object is still available until the state is committed,
// getStateObject will return a non-nil account after SelfDestruct.
func (s *StateDB) SelfDestruct(addr common.Address) {
	stateObject := s.getStateObject(addr)
	if stateObject == nil {
		return
	}
	var (
		prev = new(uint256.Int).Set(stateObject.Balance())
		n    = new(uint256.Int)
	)
	s.journal.append(selfDestructChange{
		account:     &addr,
		prev:        stateObject.selfDestructed,
		prevbalance: prev,
	})
	if s.logger != nil && s.logger.OnBalanceChange != nil && prev.Sign() > 0 {
		s.logger.OnBalanceChange(addr, prev.ToBig(), n.ToBig(), tracing.BalanceDecreaseSelfdestruct)
	}
	stateObject.markSelfdestructed()
	stateObject.account.Balance = n
}

func (s *StateDB) Selfdestruct6780(addr common.Address) {
	stateObject := s.getStateObject(addr)
	if stateObject == nil {
		return
	}
	if stateObject.newContract {
		s.SelfDestruct(addr)
	}
}

// SetTransientState sets transient storage for a given account. It
// adds the change to the journal so that it can be rolled back
// to its previous value if there is a revert.
func (s *StateDB) SetTransientState(addr common.Address, key, value common.Hash) {
	prev := s.GetTransientState(addr, key)
	if prev == value {
		return
	}
	s.journal.append(transientStorageChange{
		account:  &addr,
		key:      key,
		prevalue: prev,
	})
	s.setTransientState(addr, key, value)
}

// setTransientState is a lower level setter for transient storage. It
// is called during a revert to prevent modifications to the journal.
func (s *StateDB) setTransientState(addr common.Address, key, value common.Hash) {
	s.transientStorage.Set(addr, key, value)
}

// CreateContract is used whenever a contract is created. This may be preceded
// by CreateAccount, but that is not required if it already existed in the
// state due to funds sent beforehand.
// This operation sets the 'newContract'-flag, which is required in order to
// correctly handle EIP-6780 'delete-in-same-transaction' logic.
func (s *StateDB) CreateContract(addr common.Address) {
	obj := s.getStateObject(addr)
	if !obj.newContract {
		obj.newContract = true
		s.journal.append(createContractChange{account: addr})
	}
}

func (s *StateDB) PointCache() *utils.PointCache {
	return utils.NewPointCache(pointCacheSize)
}

// Witness retrieves the current state witness being collected.
// As of now, witness is not initialized in the StateDB, but each reference
// perform a nil-check.
func (s *StateDB) Witness() *stateless.Witness {
	return s.witness
}

// GetCode returns the code of account, nil if not exists.
func (s *StateDB) GetCode(addr common.Address) []byte {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Code()
	}
	return nil
}

// GetCodeSize returns the code size of account.
func (s *StateDB) GetCodeSize(addr common.Address) int {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.CodeSize()
	}
	return 0
}

// GetCodeHash returns the code hash of account.
func (s *StateDB) GetCodeHash(addr common.Address) common.Hash {
	stateObject := s.getStateObject(addr)
	if stateObject == nil {
		return common.Hash{}
	}
	return common.BytesToHash(stateObject.CodeHash())
}

// GetState retrieves a value from the given account's storage trie.
func (s *StateDB) GetState(addr common.Address, hash common.Hash) common.Hash {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.GetState(hash)
	}
	return common.Hash{}
}

// GetCommittedState retrieves a value from the given account's committed storage trie.
func (s *StateDB) GetCommittedState(addr common.Address, hash common.Hash) common.Hash {
	stateObject := s.getStateObject(addr)
	if stateObject != nil {
		return stateObject.GetCommittedState(hash)
	}
	return common.Hash{}
}

// GetRefund returns the current value of the refund counter.
func (s *StateDB) GetRefund() uint64 {
	return s.refund
}

// AddPreimage records a SHA3 preimage seen by the VM.
// AddPreimage performs a no-op since the EnablePreimageRecording flag is disabled
// on the vm.Config during state transitions. No store trie preimages are written
// to the database.
func (s *StateDB) AddPreimage(_ common.Hash, _ []byte) {}

// getStateObject retrieves a state object given by the address, returning nil if
// the object is not found.
func (s *StateDB) getStateObject(addr common.Address) *stateObject {
	// Prefer live objects if any is available
	if obj := s.stateObjects[addr]; obj != nil {
		return obj
	}
	// If no live objects are available, load it from keeper
	account := s.keeper.GetAccount(s.ctx, addr)
	if account == nil {
		return nil
	}
	// Insert into the live set
	obj := newObject(s, addr, *account)
	s.setStateObject(obj)
	return obj
}

// getOrNewStateObject retrieves a state object or create a new state object if nil.
func (s *StateDB) getOrNewStateObject(addr common.Address) *stateObject {
	stateObject := s.getStateObject(addr)
	if stateObject == nil {
		stateObject, _ = s.createObject(addr)
	}
	return stateObject
}

// createObject creates a new state object. If there is an existing account with
// the given address, it is overwritten and returned as the second return value.
func (s *StateDB) createObject(addr common.Address) (newobj, prev *stateObject) {
	prev = s.getStateObject(addr)

	newobj = newObject(s, addr, Account{})
	if prev == nil {
		s.journal.append(createObjectChange{account: &addr})
	} else {
		s.journal.append(resetObjectChange{prev: prev})
	}
	s.setStateObject(newobj)
	if prev != nil {
		return newobj, prev
	}
	return newobj, nil
}

// CreateAccount explicitly creates a state object. If a state object with the address
// already exists the balance is carried over to the new account.
//
// CreateAccount is called during the EVM CREATE operation. The situation might arise that
// a contract does the following:
//
// 1. sends funds to sha(account ++ (nonce + 1))
// 2. tx_create(sha(account ++ nonce)) (note that this gets the address of 1)
//
// Carrying over the balance ensures that Ether doesn't disappear.
func (s *StateDB) CreateAccount(addr common.Address) {
	newObj, prev := s.createObject(addr)
	if prev != nil {
		newObj.setBalance(prev.account.Balance)
	}
}

// ForEachStorage iterate the contract storage, the iteration order is not defined.
func (s *StateDB) ForEachStorage(addr common.Address, cb func(key, value common.Hash) bool) error {
	so := s.getStateObject(addr)
	if so == nil {
		return nil
	}
	s.keeper.ForEachStorage(s.ctx, addr, func(key, value common.Hash) bool {
		if value, dirty := so.dirtyStorage[key]; dirty {
			return cb(key, value)
		}
		if len(value) > 0 {
			return cb(key, value)
		}
		return true
	})
	return nil
}

func (s *StateDB) setStateObject(object *stateObject) {
	s.stateObjects[object.Address()] = object
}

/*
 * SETTERS
 */

// AddBalance adds amount to the account associated with addr.
func (s *StateDB) AddBalance(addr common.Address, amount *uint256.Int, _ tracing.BalanceChangeReason) {
	stateObject := s.getOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.AddBalance(amount)
	}
}

// SubBalance subtracts amount from the account associated with addr.
func (s *StateDB) SubBalance(addr common.Address, amount *uint256.Int, _ tracing.BalanceChangeReason) {
	stateObject := s.getOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SubBalance(amount)
	}
}

// RegisterCachedContextCheckpoint ... Register a cached context checkpoint
// in the journal entries.
func (s *StateDB) RegisterCachedCtxCheckpoint(addr common.Address, cachedCtxCheckpoint *CachedCtxCheckpoint) error {
	// add the cache ctx checkpoint to the journal,
	// this cannot realistically fail
	s.getOrNewStateObject(addr).RegisterCachedCtxCheckpoint(cachedCtxCheckpoint)

	// here we increment the state DB counter of precompile calls.
	// we are doing it here because we are registering a cached context checkpoint.
	// a cached checkpoint context is registered every time a precompile is executed.
	s.ongoingPrecompilesCallsCounter++

	// for backward compatibility when maxPrecompilesCallsPerExecution == 0
	// we do not have any check, the first assertion bypass the check.
	if s.maxPrecompilesCallsPerExecution > 0 && s.ongoingPrecompilesCallsCounter > s.maxPrecompilesCallsPerExecution {
		return fmt.Errorf("transaction have exceeded the maximum number of precompile calls per execution, max allowed: %v, attempted: %v", s.maxPrecompilesCallsPerExecution, s.ongoingPrecompilesCallsCounter)
	}

	return nil
}

// SetNonce sets the nonce of account.
func (s *StateDB) SetNonce(addr common.Address, nonce uint64) {
	stateObject := s.getOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetNonce(nonce)
	}
}

// SetCode sets the code of account.
func (s *StateDB) SetCode(addr common.Address, code []byte) {
	stateObject := s.getOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetCode(crypto.Keccak256Hash(code), code)
	}
}

// SetState sets the contract state.
func (s *StateDB) SetState(addr common.Address, key, value common.Hash) {
	stateObject := s.getOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetState(key, value)
	}
}

// AddAddressToAccessList adds the given address to the access list
func (s *StateDB) AddAddressToAccessList(addr common.Address) {
	if s.accessList.AddAddress(addr) {
		s.journal.append(accessListAddAccountChange{&addr})
	}
}

// AddSlotToAccessList adds the given (address, slot)-tuple to the access list
func (s *StateDB) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	addrMod, slotMod := s.accessList.AddSlot(addr, slot)
	if addrMod {
		// In practice, this should not happen, since there is no way to enter the
		// scope of 'address' without having the 'address' become already added
		// to the access list (via call-variant, create, etc).
		// Better safe than sorry, though
		s.journal.append(accessListAddAccountChange{&addr})
	}
	if slotMod {
		s.journal.append(accessListAddSlotChange{
			address: &addr,
			slot:    &slot,
		})
	}
}

// AddressInAccessList returns true if the given address is in the access list.
func (s *StateDB) AddressInAccessList(addr common.Address) bool {
	return s.accessList.ContainsAddress(addr)
}

// SlotInAccessList returns true if the given (address, slot)-tuple is in the access list.
func (s *StateDB) SlotInAccessList(addr common.Address, slot common.Hash) (addressPresent bool, slotPresent bool) {
	return s.accessList.Contains(addr, slot)
}

// Snapshot returns an identifier for the current revision of the state.
func (s *StateDB) Snapshot() int {
	id := s.nextRevisionID
	s.nextRevisionID++
	s.validRevisions = append(s.validRevisions, revision{id, s.journal.length()})
	return id
}

// RevertToSnapshot reverts all state changes made since the given revision.
func (s *StateDB) RevertToSnapshot(revid int) {
	// Find the snapshot in the stack of valid snapshots.
	idx := sort.Search(len(s.validRevisions), func(i int) bool {
		return s.validRevisions[i].id >= revid
	})
	if idx == len(s.validRevisions) || s.validRevisions[idx].id != revid {
		panic(fmt.Errorf("revision id %v cannot be reverted", revid))
	}
	snapshot := s.validRevisions[idx].journalIndex

	// Replay the journal to undo changes and remove invalidated snapshots
	s.journal.Revert(s, snapshot)
	s.validRevisions = s.validRevisions[:idx]
}

// Commit writes the dirty states to keeper
// the StateDB object should be discarded after committed.
func (s *StateDB) commit(ctx sdk.Context) error {
	for _, addr := range s.journal.sortedDirties() {
		obj := s.stateObjects[addr]
		if obj.selfDestructed {
			if err := s.keeper.DeleteAccount(ctx, obj.Address(), obj.Balance().ToBig()); err != nil {
				return errorsmod.Wrap(err, "failed to delete account")
			}
		} else {
			if obj.code != nil && obj.dirtyCode {
				s.keeper.SetCode(ctx, obj.CodeHash(), obj.code)
			}
			if err := s.keeper.SetAccount(ctx, obj.Address(), obj.account); err != nil {
				return errorsmod.Wrap(err, "failed to set account")
			}
			for _, key := range obj.dirtyStorage.SortedKeys() {
				s.keeper.SetState(ctx, obj.Address(), key, obj.dirtyStorage[key].Bytes())
			}
		}
	}
	return nil
}

func (s *StateDB) Commit() error {
	// if this is set, this means a cache context
	// existed as well, let's flush it first
	if s.flushCache != nil {
		s.flushCache()
	}

	return s.commit(s.ctx)
}

func (s *StateDB) CommitCacheContext() error {
	return s.commit(s.cachedCtx)
}

type CachedCtxCheckpoint struct {
	ms     storetypes.CacheMultiStore
	events sdk.Events
}

func (ccc *CachedCtxCheckpoint) Revert(stateDB *StateDB) {
	// first we load back the state in the context
	stateDB.cachedCtx = stateDB.cachedCtx.WithMultiStore(ccc.ms)
	// then replace the flush cache function
	// we write our own flushCache function here which will be used at the
	// time of rollback.
	stateDB.flushCache = func() {
		// we capture  the events and the actual context
		stateDB.ctx.EventManager().EmitEvents(ccc.events)

		// and we capture the copy of the cache multistore
		// at time of creation
		ccc.ms.Write()
	}
}

func (s *StateDB) CacheContext() (sdk.Context, *CachedCtxCheckpoint) {
	// here we create a cache context on the very first
	// call to this function
	if s.flushCache == nil {
		s.cachedCtx, s.flushCache = s.ctx.CacheContext()
	}

	ccp := CachedCtxCheckpoint{
		// we do a copy of the state here so we can just hot swap it later on?
		ms: s.cachedCtx.MultiStore().(storetypes.CacheMultiStore).Clone(),
		// we copy the events from the cache context, just to restore them
		// the same way later.
		events: s.cachedCtx.EventManager().Events(),
	}

	return s.cachedCtx, &ccp
}
