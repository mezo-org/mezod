<!--
order: 2
-->

# Begin-Block

Om each ABCI begin block call, the historical info will get stored and pruned
according to the `HistoricalEntries()` parameter returned by the `Keeper`.

## Historical Info Tracking

If the `HistoricalEntries` parameter is 0, then the `BeginBlock` performs a no-op.

Otherwise, the latest historical info is stored under the key
`HistoricalInfo|height`, while any entries older than
`height - HistoricalEntries` is deleted. In most cases, this results in a
single entry being pruned per block. However, if the parameter
`HistoricalEntries` has changed to a lower value there will be multiple entries
in the store that must be pruned.
