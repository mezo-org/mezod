<!--
order: 1
-->

# State

## Params

Params is a module-wide configuration structure that stores system parameters
and defines the overall functioning of the PoA module.

### Keys

- `Params`: `0x10 | types.Params`

`Params` is the primary entry storing the module's parameters.

### Types

```go
type Params struct {
    MaxValidators uint32 // Maximum number of validators
}
```

## Owner

The owner is the entity that has the right to change the validator set
and the module's parameters. The owner can transfer the ownership to another
account, in a 2-step process. The new owner must accept the ownership transfer
before the transfer is completed.

### Keys

- `Owner`: `0x20 | sdk.AccAddress`
- `CandidateOwner`: `0x21 | sdk.AccAddress`

`Owner` is the primary entry storing the current module's owner.

`CandidateOwner` is the primary entry storing the candidate owner during
module's ownership transfer.

### Types

The owner's state does not use any custom types. It works with the
Cosmos SDK `sdk.AccAddress` type.

## Application

The application pool tracks all the current applications. An operator can only
be one validator, therefore the application can be accessed by the operator
address.

### Keys

- `Application`: `0x30 | sdk.ValAddress -> type.Application`
- `ApplicationByConsAddr`: `0x31 | sdk.ConsAddress -> sdk.ValAddress`

`Application` is the primary index - it ensures that each operator can have
only one application

`ApplicationByConsAddr` is an additional index to ensure there is no two
applications with the same consensus public key.

### Types

An application is stored in an `Application` structure. The `Validator` field
represents the potential new validator.

```go
type Application struct {
    // Validator is the candidate that is subject of the application.
    Validator Validator
}
```

## Validator

Validators objects should be primarily stored and accessed by the
`Operator` key, an `sdk.ValAddress` for the operator of the validator.
`ValidatorByConsAddr` is maintained per validator object in order to fulfill
the required lookups for validator-set updates.

### Keys

- `Validator`: `0x40 | sdk.ValAddress -> types.Validator`
- `ValidatorByConsAddr`: `0x41 | sdk.ConsAddress -> sdk.ValAddress`
- `ValidatorState`: `0x42 | sdk.ValAddress -> types.ValidatorState`
- `ValidatorsByPrivilege`: `0x43 | string -> []sdk.ConsAddress`

`Validator` is the primary index - it ensures that each operator can have only one
associated validator, where the public key of that validator can change in the
future.

`ValidatorByConsAddr` is an additional index that enables lookups for future
uses (like automatic kick for misbehaving).

`ValidatorState` holds the state of a validator. The validator can have 3
states: joining, active or leaving. This state allows the End Blocker to know
how to update the Tendermint consensus validator state.

`ValidatorsByPrivilege` holds the information about validators' privileges.
The key is a string that represents the privilege name. The value is a set
of consensus addresses corresponding to validators that have the privilege.
Consensus addresses are ordered lexicographically, in ascending order.

### Types

Each validator's state is stored in a `Validator` struct:

```go
type Validator struct {
    // Bech32 encoded address of the validator's operator. 
    // Unmarshals to `sdk.ValAddress`. 
    OperatorBech32 string 
    // Bech32 encoded consensus public key of the validator.
    // Unmarshals to `crypto.PubKey`. 
    ConsPubKeyBech32 string 
    // Human-readable information about the validator.
    Description Description
}

type Description struct {
    // Name of the validator.
    Moniker string
	// Optional identity signature (ex. UPort or Keybase).	
    Identity string
	// Optional website link.
    Website string
	// Optional email for security contact
    SecurityContact string
    // Optional details.	
    Details string 
}

const (
    // ValidatorStateUnknown is the default state of a validator.
    ValidatorStateUnknown ValidatorState = iota
    // ValidatorStateJoining means that the validator is not yet present in the
    // Tendermint consensus validator set and will join it at the end of the block.
    ValidatorStateJoining
    // ValidatorStateActive means that the validator is present in the
    // Tendermint consensus validator set.
    ValidatorStateActive
    // ValidatorStateLeaving means that the validator will leave the Tendermint
    // consensus validator set at the end of the block.
    ValidatorStateLeaving
)
```

## Historical info

`HistoricalInfo` objects are stored and pruned at each block such that the PoA
keeper persists the `n` most recent historical info defined by the
`HistoricalEntries()` parameter returned by the `Keeper`.

At the beginning of each block, the PoA `Keeper` will persist the current
`Header` and the active validators of the current block in a `HistoricalInfo`
object. The validators are sorted on their operator address to ensure that they
are in a deterministic order. The oldest entries will be pruned to ensure that
there only exist the parameter-defined number of historical entries.

### Keys

- `HistoricalInfo`: `0x40 | int64 -> types.HistoricalInfo`

`HistoricalInfo` is the primary index storing the historical info by block height.

### Types

```go
type HistoricalInfo struct {
    // The block's header.
    Header types.Header
    // The active validator set at the block.
    Valset []Validator
}
```
