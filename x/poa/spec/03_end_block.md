<!--
order: 3
-->

# End-Block

Each abci end block call, the operations to update the validator set
changes are specified to execute.

## Validator Set Changes

The staking validator set is updated during this process by state transitions
that run at the end of every block. As a part of this process any updated
validators are also returned back to Tendermint for inclusion in the Tendermint
validator set which is responsible for validating Tendermint messages at the
consensus layer.
