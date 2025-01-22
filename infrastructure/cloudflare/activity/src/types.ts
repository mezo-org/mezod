export type Env = {
  BLOCKSCOUT_API_URL: string
  FAUCET_ADDRESS: string
  DB: D1Database
  UPDATE_BATCH_SIZE: string
}

export enum FetchProgressKey {
  ADDRESS = "address",
  TX_COUNT = "tx_count",
  ADDRESS_TXS = "address_txs",
  CONTRACT_TXS = "contract_txs",
}
