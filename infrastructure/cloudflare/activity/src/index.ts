import { WorkerEntrypoint } from "cloudflare:workers"
import { AddressItem, BlockScoutAPI } from "#/blockscout"

type Env = {
  BLOCKSCOUT_API_URL: string
  FAUCET_ADDRESS: string
  DB: D1Database
  UPDATE_BATCH_SIZE: string
}

async function updateAddresses(env: Env) {
  const addresses = await fetchAddresses(env)

  // This statement inserts new addresses with the value of `updated_at` equal
  // to unix timestamp 0. This will put them in the front of the queue for activity updates.
  const stmts = addresses.map((item) => {
    return env.DB
      .prepare(`INSERT OR IGNORE INTO activity (address, tx_count, claimed_btc, deployer, updated_at) VALUES (?1, ?2, ?3, ?4, ?5)`)
      .bind(item.address, 0, "0", item.deployer ? item.deployer : null, 0)
  })

  await batchExecuteStmts(env, stmts)
}

async function updateActivity(env: Env) {
  const activity = await fetchActivity(env)

  const stmts = activity.map((item) => {
    return env.DB
      .prepare(`UPDATE activity SET tx_count = ?1, claimed_btc = ?2, updated_at = datetime('now') WHERE address = ?3`)
      .bind(item.txCount, item.claimedBTC, item.address)
  })

  await batchExecuteStmts(env, stmts)
}

async function batchExecuteStmts(env: Env, stmts: D1PreparedStatement[]): Promise<void> {
  const batchSize = parseInt(env.UPDATE_BATCH_SIZE);
  const batches = [];

  for (let i = 0; i < stmts.length; i += batchSize) {
    const batch = stmts.slice(i, i + batchSize);
    batches.push(batch);
  }

  await Promise.all(
    batches.map(async (batch) => {
      await env.DB.batch(batch)
    })
  )
}

async function getAddresses(env: Env, limit: number): Promise<string[]> {
  // Return oldest addresses first.
  const { results } = await env.DB.prepare(`SELECT address FROM activity ORDER BY updated_at ASC LIMIT ${limit}`).all<{ address: string }>()
  return results.map((item) => item.address)
}

async function getActivity(env: Env): Promise<ActivityItem[]> {
  const { results } = await env.DB.prepare(
    `
    WITH contract_info AS (
        SELECT deployer, count(*) as deployed_contracts, sum(tx_count) as deployed_contracts_tx_count
        FROM activity
        WHERE deployer IS NOT NULL
        GROUP BY deployer
    )
    SELECT 
        a.address, 
        a.tx_count, 
        COALESCE(ci.deployed_contracts, 0) as deployed_contracts, 
        COALESCE(ci.deployed_contracts_tx_count, 0) as deployed_contracts_tx_count,
        a.claimed_btc
    FROM activity a
    LEFT JOIN contract_info ci ON a.address = ci.deployer
    WHERE a.deployer IS NULL;
    `
  ).all<ActivityItem>()

  return results
}

async function fetchAddresses(env: Env): Promise<AddressItem[]> {
  const bsAPI = new BlockScoutAPI(env.BLOCKSCOUT_API_URL)
  const EOAs = await bsAPI.externallyOwnedAccounts()
  const contracts = await bsAPI.contracts()
  return [...EOAs, ...contracts].filter((item) => item.address !== env.FAUCET_ADDRESS)
}

async function fetchActivity(env: Env): Promise<Pick<ActivityItem, "address" | "txCount" | "claimedBTC">[]> {
  const bsAPI = new BlockScoutAPI(env.BLOCKSCOUT_API_URL)
  const txsFromFaucet = await bsAPI.txs(env.FAUCET_ADDRESS, "from")

  const addressToClaimedBTC = txsFromFaucet.reduce(
    (acc, tx) => {
      if (!acc[tx.to]) {
        acc[tx.to] = BigInt(0)
      }
      acc[tx.to] += BigInt(tx.value)
      return acc
    },
    {} as Record<string, bigint>
  )

  const batchSize = parseInt(env.UPDATE_BATCH_SIZE);
  const addressBatch = await getAddresses(env, batchSize)

  const addressToTxCount = new Map<string, number>
  for (const address of addressBatch) {
    try {
      const txCount = await bsAPI.txCount(address)
      addressToTxCount.set(address, txCount)
    } catch (error) {
      // Log error and continue to next address.
      console.error(`error fetching tx count for: ${address}: ${error}`)
    }
  }

  const activity = addressBatch
    .filter((address) => addressToTxCount.has(address))
    .map((address) => {
      return {
        address: address,
        txCount: addressToTxCount.get(address)!,
        claimedBTC: addressToClaimedBTC[address]?.toString() || "0"
      }
    })

  console.log(`fetched activity for ${activity.length}/${addressBatch.length} addresses in the batch`)

  return activity
}

export default {
  async fetch(request: Request, env: Env, _: ExecutionContext) {
    try {
      const activity = await getActivity(env)
      return Response.json({ success: true, activity: activity })
    } catch (error) {
      return Response.json({ success: false, errorMsg: `${error}` })
    }
  },
  async scheduled(_: { cron: unknown; scheduledTime: string }, env: Env) {
    console.log(`updating addresses index`)
    await updateAddresses(env)
    console.log(`addresses index updated`)

    console.log(`updating activity data`)
    await updateActivity(env)
    console.log(`activity data updated`)
  },
}

export class InternalEntrypoint extends WorkerEntrypoint {
  async get(): Promise<InternalGetResponse> {
    try {
      const activity = await getActivity(this.env as Env)
      return { success: true, activity: activity }
    } catch (error) {
      return { success: false, errorMsg: `${error}` }
    }
  }
}

export type InternalGetResponse = {
  success: boolean
  activity?: ActivityItem[]
  errorMsg?: string
}

export type ActivityItem = {
  address: string
  txCount: number,
  deployedContracts: number,
  deployedContractsTxCount: number
  claimedBTC: string // 1e18 precision
}
