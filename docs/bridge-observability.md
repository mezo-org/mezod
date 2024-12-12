# Bridge observability

## Overview

Information on bridged assets is stored in special type of transactions called
pseudo-transactions. These pseudo-transactions contain information on
`AssetsLocked` events. Pseudo-transactions can be viewed using the same
commands as ordinary transactions.

## Pseudo-transactions

There is at most one pseudo-transaction per block. If it is present in a block,
it is always located at index `0` and appears as the block's first transaction.
JSON-RPC commands return information on pseudo-transactions in such a way that
it appears the transactions called the `AssetsBridge` precompile's `bridge`
method. This allows displaying the `AssetsLocked` events in block explorers.
A pseudo-transaction contains at most 10 events.

## Enabling bridge observability

Pseudo-transactions are only stored by Mezo nodes if a custom indexer is enabled.

It can be done by setting the following option in `config/app.toml`: `enable-indexer = true`

If the custom indexer is not enabled, the pseudo-transactions will not be stored
and therefore will not be retrieved using JSON-RPC commands. They will also not
be seen by tools such as block explorers.

## Limitations

When the custom indexer is not enabled, pseudo-transactions will not be stored.
JSON-RPC commands will not be able to retrieve pseudo-transactions and will act
as if there were no pseudo-transactions present in the Mezo network (blocks
returned from commands will not contain pseudo-transactions, regular ETH
transactions will be indexed starting from `0`).
