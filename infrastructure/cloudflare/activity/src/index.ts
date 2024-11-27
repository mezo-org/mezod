import { WorkerEntrypoint } from "cloudflare:workers"
import { BlockScoutAPI, ContractItem } from "#/blockscout"

type Env = {
  BLOCKSCOUT_API_URL: string
  FAUCET_ADDRESS: string
  DB: D1Database
  UPDATE_BATCH_SIZE: string
}

async function updateActivity(env: Env) {
  // TODO: The update approach taken here is quite naive and may bite us
  //       when there is a lot of data. Potential issues are too many
  //       concurrent DB connections, slow processing time, or memory issues.
  //       Consider a more sophisticated approach for the next iteration.

  const activity = await fetchActivity(env)

  const stmts = activity.map((item) => {
    return env.DB
      .prepare(`INSERT OR REPLACE INTO activity (address, tx_count, deployed_contracts, deployed_contracts_tx_count, claimed_btc) VALUES (?1, ?2, ?3, ?4, ?5)`)
      .bind(item.address, item.txCount, item.deployedContracts, item.deployedContractsTxCount, item.claimedBTC)
  })

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

async function getActivity(env: Env): Promise<ActivityItem[]> {
  const { results } = await env.DB.prepare("SELECT * FROM activity").all<ActivityItem>()
  return results
}

async function fetchActivity(env: Env): Promise<ActivityItem[]> {
  const bsAPI = new BlockScoutAPI(env.BLOCKSCOUT_API_URL)

  // TODO: We may want to process addresses in batches if number of addresses is large.
  const addresses = await bsAPI.addresses()
  const contracts = await bsAPI.contracts()
  const txsFromFaucet = await bsAPI.txsFromAddress(env.FAUCET_ADDRESS)

  const addressToContracts = contracts.reduce(
    (acc, contract) => {
      if (!acc[contract.deployer]) {
        acc[contract.deployer] = []
      }
      acc[contract.deployer].push(contract)
      return acc
    },
    {} as Record<string, ContractItem[]>
  )

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

  return addresses
    .filter((item) => item.address !== env.FAUCET_ADDRESS) // Filter out the faucet address.
    .map((item) => {
      const contracts = addressToContracts[item.address] || []

      const deployedContractsTxCount = contracts.reduce(
        (acc, contract) => acc + contract.txCount,
        0
      )

      return {
        address: item.address,
        txCount: item.txCount,
        deployedContracts: contracts.length,
        deployedContractsTxCount: deployedContractsTxCount,
        claimedBTC: addressToClaimedBTC[item.address]?.toString() || "0"
      }
    })
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
  async scheduled(_: unknown, env: Env) {
    await updateActivity(env)
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
